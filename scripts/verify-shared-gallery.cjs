// Experimental shared-runtime acceptance. Requires externally installed Playwright.
const {chromium}=require('playwright');
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
(async()=>{
 const base=process.argv[2];if(!base)throw Error('usage: node scripts/verify-shared-gallery.cjs http://127.0.0.1:8767/ [report-dir]');
 const dir=process.argv[3]||'work/gallery-shared';fs.mkdirSync(dir,{recursive:true});
 const browser=await chromium.launch({channel:process.env.KEEL_BROWSER_CHANNEL || "chrome",args:['--enable-unsafe-swiftshader']});
 try {
  const page=await browser.newPage({viewport:{width:390,height:844},isMobile:true,hasTouch:true,permissions:['clipboard-read','clipboard-write']});
  await page.addInitScript(()=>{window.galleryReady=false;addEventListener('keel-ready',()=>window.galleryReady=true)});
  const errors=[],requests=[];page.on('pageerror',e=>errors.push(String(e)));page.on('response',r=>requests.push(r.url()));
  const frames=()=>page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));
  await page.goto(new URL('?lang=en#component=input',base).href);
  await page.waitForFunction(()=>window.galleryReady);await frames();
  const config=await page.evaluate(()=>window.keelDemo || null);
  const build=config ? {sections:Object.fromEntries(Object.entries(config.shared.sections).map(([name,file])=>[name,config.assets[file].url]))} : await (await page.request.get(new URL('shared-build.json',base).href)).json();
  const dialogChild=build.sections?.dialog || build.dialog;
  const isDialog=url=>new URL(url).pathname.endsWith('/'+dialogChild);
  assert.ok(!requests.some(isDialog),'child downloaded before demand');
  const initialMemoryBytes=await page.evaluate(()=>{window.initialMemory=sharedPrimary.exports.mem;return initialMemory.buffer.byteLength});
  // Actual touch focus, text entry, later selection/copy: no Go state backdoor.
  await page.touchscreen.tap(120,150);await frames();
  await page.waitForFunction(()=>document.activeElement.tagName==='TEXTAREA');
  await page.keyboard.insertText('Retained across modules');await frames();
  // Motion also uses DialogView: shared code cannot belong only to Dialog.
  await page.evaluate(()=>{location.hash='component=motion'});await frames();
  assert.ok(!requests.some(x=>isDialog(x)),'shared DialogView forced a demo download');
  await page.evaluate(()=>{location.hash='component=input'});await frames();

  await page.evaluate(()=>{location.hash='component=dialog'});
  await page.waitForFunction(()=>keelSectionStatus('dialog')==='');await frames();
  assert.equal(requests.filter(x=>isDialog(x)).length,1);
  await page.touchscreen.tap(75,225);await frames();
  await page.screenshot({path:path.join(dir,'dialog-open.png')});
  await page.keyboard.press('Escape');await frames();
  await page.evaluate(()=>{location.hash='component=input'});await frames();
  await page.touchscreen.tap(120,150);await frames();
  await page.keyboard.press('Control+a');await page.keyboard.press('Control+c');
  await page.waitForFunction(async()=> (await navigator.clipboard.readText())==='Retained across modules');
  const copied=await page.evaluate(()=>navigator.clipboard.readText());
  assert.ok(await page.evaluate(()=>initialMemory===sharedPrimary.exports.mem));
  await page.evaluate(()=>loadDialog());assert.equal(requests.filter(x=>isDialog(x)).length,1);
  await page.screenshot({path:path.join(dir,'retained.png')});
  // When built with multiple partitions, each must remain independently lazy.
  const extraPartitions = ['nav_stack','emoji_picker','terminal'].filter(name => build[name]);
  for (const name of extraPartitions) {
   const requested = () => requests.filter(url => new URL(url).pathname.endsWith('/'+build[name])).length;
   assert.equal(requested(),0, name+' loaded before navigation');
   await page.evaluate(name=>{location.hash='component='+name},name);
   await page.waitForFunction(name=>keelSectionStatus(name)==='',name);await frames();
   assert.equal(requested(),1,name+' did not load exactly once');
   assert.ok(await page.evaluate(()=>initialMemory===sharedPrimary.exports.mem));
   await page.screenshot({path:path.join(dir,name+'.png')});
  }
  if (extraPartitions.length) {
   await page.evaluate(()=>{location.hash='component=input'});await frames();
   await page.touchscreen.tap(120,150);await frames();
   await page.keyboard.press('Control+a');await page.keyboard.press('Control+c');
   await page.waitForFunction(async()=>await navigator.clipboard.readText()==='Retained across modules');
  }

  let attempts=0;
  await page.route('**/'+dialogChild, route => ++attempts===1 ? route.abort() : route.continue());
  await page.goto(new URL('?lang=en&scenario=retry#component=dialog',base).href);
  await page.waitForFunction(()=>window.galleryReady && keelSectionStatus('dialog')==='error');await frames();
  await page.screenshot({path:path.join(dir,'load-error.png')});
  // The retry button is visible in the native-rendered 390dp error page.
  await page.touchscreen.tap(195,191);
  await page.waitForFunction(()=>keelSectionStatus('dialog')==='');await frames();
  assert.equal(attempts,2,'retry did not issue exactly one new child request');
  await page.screenshot({path:path.join(dir,'retried.png')});
  await page.unroute('**/'+dialogChild);
  let heldRoute, releaseRequest;
  const requested = new Promise(resolve => { releaseRequest = resolve; });
  await page.route('**/'+dialogChild, route => { heldRoute=route; releaseRequest(); });
  await page.goto(new URL('?lang=en&scenario=history#component=input',base).href);
  await page.waitForFunction(()=>window.galleryReady);await frames();
  await page.touchscreen.tap(120,150);await frames();
  await page.keyboard.insertText('Kept during pending navigation');await frames();
  await page.evaluate(()=>{location.hash='component=dialog'});
  await Promise.race([requested, new Promise((_, reject) => setTimeout(() => reject(Error("child request did not start")), 30000).unref())]);await frames();
  await page.goBack();await frames();
  assert.ok(page.url().endsWith('#component=input'));
  await heldRoute.continue();
  await page.waitForFunction(()=>keelSectionStatus('dialog')==='');await frames();
  assert.ok(page.url().endsWith('#component=input'),'late completion replaced navigation');
  await page.touchscreen.tap(120,150);await frames();
  await page.keyboard.press('Control+a');await page.keyboard.press('Control+c');
  await page.waitForFunction(async()=> (await navigator.clipboard.readText())==='Kept during pending navigation');
  await page.goForward();await frames();
  assert.ok(page.url().endsWith('#component=dialog'));
  await page.touchscreen.tap(75,225);await frames();
  await page.screenshot({path:path.join(dir,'forward-after-load.png')});
  assert.deepEqual(errors,[]);
  fs.writeFileSync(path.join(dir,'report.json'),JSON.stringify({browser:browser.version(),extraPartitions,initialMemoryBytes,requests,errors,copied,sharedMemory:true,lazyDownload:true,navigationTriggeredLoad:true,retryAfterNetworkFailure:attempts===2,backDuringLoad:true,forwardAfterLoad:true},null,2));
 }finally{await browser.close();}
})().catch(e=>{console.error(e);process.exitCode=1});
