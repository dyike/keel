// Candidate category partition acceptance; requires external Playwright.
const {chromium}=require('playwright');
const fs=require('node:fs'), path=require('node:path'), assert=require('node:assert/strict');
(async()=>{
 const base=process.argv[2], dir=process.argv[3]||'work/gallery-shared-categories';
 if(!base) throw Error('usage: node scripts/verify-shared-categories.cjs URL [report-dir]');
 fs.mkdirSync(dir,{recursive:true});
 const browser=await chromium.launch({channel:process.env.KEEL_BROWSER_CHANNEL || "chrome",args:['--enable-unsafe-swiftshader']});
 try {
  const page=await browser.newPage({viewport:{width:390,height:844},isMobile:true,hasTouch:true,permissions:['clipboard-read','clipboard-write']});
  const errors=[],requests=[],consoleErrors=[];
  page.on('console',m=>{if(m.type()==='error')consoleErrors.push(m.text())});
  page.on('pageerror',e=>errors.push(String(e)));
  page.on('response',r=>{if(new URL(r.url()).pathname.endsWith('.wasm'))requests.push({url:r.url(),status:r.status()});});
  await page.addInitScript(()=>{
   window.runtimeExits=[];window.galleryReady=false;
   addEventListener('keel-ready',()=>window.galleryReady=true);
   addEventListener('keel-exit',e=>runtimeExits.push(e.detail));
  });
  const frames=()=>page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));
  await page.goto(new URL('?lang=en#component=input',base).href);
  await page.waitForFunction(()=>window.galleryReady);await frames();
  await page.waitForFunction(()=>keelSectionStatus('input')==='');await frames();
  const siteConfig=await page.evaluate(()=>window.keelDemo || null);
  const build=siteConfig ? {
   primary:siteConfig.assets[siteConfig.shared.primary].url,
   fullWasm:siteConfig.assets[siteConfig.shared.fullWasm].url,
   sections:Object.fromEntries(Object.entries(siteConfig.shared.sections).map(([section,file])=>[section,siteConfig.assets[file].url])),
   dependencies:siteConfig.shared.dependencies ? Object.fromEntries(Object.entries(siteConfig.shared.dependencies).map(([k,files])=>[k,files.map(f=>siteConfig.assets[f].url)])) : null
  } : await (await page.request.get(new URL('shared-build.json',base).href)).json();
  assert.ok(build.sections && Object.keys(build.sections).length>100,'missing full category ownership manifest');
  await page.evaluate(()=>{window.initialMemory=sharedPrimary.exports.mem;window.runtimeToken='retained'});
  await page.touchscreen.tap(120,150);await frames();
  await page.waitForFunction(()=>document.activeElement.tagName==='TEXTAREA');
  await page.keyboard.insertText('Kept across every category');await frames();
  const visited=[];
  for(const section of Object.keys(build.sections).sort()) {
   await page.evaluate(section=>{location.hash='component='+section},section);
   await page.waitForFunction(section=>keelSectionStatus(section)==='',section);await frames();
   assert.equal(await page.evaluate(()=>initialMemory===sharedPrimary.exports.mem && runtimeToken==='retained'),true,section+' replaced runtime');
   assert.deepEqual(errors,[],section+' raised an error');
   assert.deepEqual(await page.evaluate(()=>runtimeExits),[],section+' exited Go');
   visited.push(section);
   await page.screenshot({path:path.join(dir,section+'.png')});
  }
  await page.evaluate(()=>{location.hash='component=input'});await frames();
  await page.touchscreen.tap(120,150);await frames();
  await page.keyboard.press('Control+a');await page.keyboard.press('Control+c');
  await page.waitForFunction(async()=>await navigator.clipboard.readText()==='Kept across every category');
  const children=[...new Set(Object.values(build.sections))];
  for(const child of children) assert.equal(requests.filter(r=>new URL(r.url).pathname.endsWith('/'+child)).length,1,child+' fetched more than once');
  assert.ok(requests.every(r=>r.status===200 || r.status===304));
  const coldStarts=[];
  for (const child of children) {
   const sections=Object.keys(build.sections).filter(section=>build.sections[section]===child).sort();
   const start=requests.length;
   await page.goto(new URL('?lang=en&cold='+encodeURIComponent(child)+'#component='+sections[0],base).href);
   await page.waitForFunction(()=>window.galleryReady);await frames();
   for(const section of sections) {
    await page.evaluate(section=>{location.hash='component='+section},section);
    await page.waitForFunction(section=>keelSectionStatus(section)==='',section);await frames();
    assert.deepEqual(errors,[],section+' failed in a cold category');
    assert.deepEqual(await page.evaluate(()=>runtimeExits),[],section+' exited a cold runtime');
   }
   const loaded=requests.slice(start).map(r=>new URL(r.url).pathname.split('/').at(-1));
   const expected=[build.primary,...new Set(sections.flatMap(s=>build.dependencies?.[s]||[child]))];
   assert.deepEqual(loaded.sort(),expected.sort(),child+' loaded unexpected dependencies');
   coldStarts.push({child,sections,loaded});
  }
  const embeds=[];
  if(siteConfig) {
   for(const section of ['input','inputs']) {
    const start=requests.length;
    await page.goto(new URL('?lang=en&section='+section,base).href);
    await page.waitForFunction(()=>window.galleryReady);await frames();
    assert.deepEqual(errors,[],section+' embed failed');
    assert.deepEqual(await page.evaluate(()=>runtimeExits),[],section+' embed exited');
    const loaded=requests.slice(start).map(r=>new URL(r.url).pathname.split('/').at(-1));
    const expected=[build.primary,...(build.dependencies?.[section]||[build.sections[section]||build.sections.input])];
    assert.deepEqual(loaded.sort(),expected.sort());
    embeds.push(section);
   }
  }
  let fullFallback=false;
  if(siteConfig) {
   const start=requests.length;
   await page.goto(new URL('?lang=en&bundle=all#component=input',base).href);
   await page.waitForFunction(()=>window.galleryReady);await frames();
   assert.equal(await page.evaluate(()=>typeof keelSectionStatus),'undefined');
   assert.deepEqual(requests.slice(start).map(r=>new URL(r.url).pathname.split('/').at(-1)),[build.fullWasm]);
   await page.touchscreen.tap(120,150);await frames();
   await page.keyboard.insertText('Full fallback works');await frames();
   await page.keyboard.press('Control+a');await page.keyboard.press('Control+c');
   await page.waitForFunction(async()=>await navigator.clipboard.readText()==='Full fallback works');
   fullFallback=true;
  }
  fs.writeFileSync(path.join(dir,'report.json'),JSON.stringify({browser:browser.version(),visited,children,requests,errors,consoleErrors,coldStarts,embeds,fullFallback,retainedInput:true,sharedMemory:true},null,2));
  assert.deepEqual(consoleErrors,[],'browser console errors remain');
 } finally {await browser.close();}
})().catch(e=>{console.error(e);process.exitCode=1});
