// Run with Playwright installed externally (NODE_PATH may point to its modules).
// This verifies a served build in Chromium, not physical-device IME behavior.
const {chromium} = require('playwright');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
(async () => {
 const base = process.argv[2];
 if (!base) throw Error('usage: node scripts/verify-gallery-web.cjs http://127.0.0.1:8765/demo/ [report-dir]');
 const dir = process.argv[3] || 'work/gallery-web';
 fs.mkdirSync(dir,{recursive:true});
 const browser = await chromium.launch({channel:process.env.KEEL_BROWSER_CHANNEL || "chrome",headless:true,args:['--enable-unsafe-swiftshader']});
 try {
  const page = await browser.newPage({viewport:{width:390,height:844},isMobile:true,hasTouch:true,permissions:['clipboard-read','clipboard-write']});
  const errors = [], resources = [];
  page.on('pageerror',e=>errors.push(String(e)));
  page.on('response',r=>{if (/\.(wasm|otf)(\?|$)/.test(r.url())) resources.push({url:r.url(),status:r.status()});});
  const ready = async()=>{
   await page.waitForSelector('#loading',{state:'detached',timeout:60000});
   await page.waitForSelector('canvas');
   // Wait two display frames for the submitted canvas operations to present.
   await page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));
  };
  await page.goto(new URL('?lang=en',base).href);await ready();
  await page.screenshot({path:path.join(dir,'index.png')});
  // The gallery's first 52dp row follows its 226dp header at this viewport.
  await page.touchscreen.tap(90,252);
  await page.waitForURL('**#component=button');
  await page.screenshot({path:path.join(dir,'button.png')});
  await page.goBack();await page.waitForLoadState('load');await ready();assert.equal(new URL(page.url()).hash,'');
  // Default full gallery must navigate across categories without restarting.
  await page.evaluate(()=>{window.retainedRuntimeToken='same-document'});
  const beforeNavigation=resources.filter(r=>r.url.includes('.wasm')).length;
  await page.evaluate(()=>{setTimeout(()=>{location.hash='component=input'},0)});
  await page.waitForURL('**#component=input');await ready();
  assert.equal(new URL(page.url()).hash,'#component=input');
  assert.equal(await page.evaluate(()=>window.retainedRuntimeToken),'same-document');
  assert.equal(resources.filter(r=>r.url.includes('.wasm')).length,beforeNavigation);
  await page.touchscreen.tap(120,150);await ready();
  await page.waitForFunction(()=>document.activeElement.tagName==='TEXTAREA');
  await page.keyboard.insertText('Retained across categories');await ready();
  await page.evaluate(()=>{location.hash='component=dialog'});await ready();
  await page.touchscreen.tap(75,225);await ready();
  await page.keyboard.press('Escape');await ready();
  await page.goBack();await ready();
  assert.equal(new URL(page.url()).hash,'#component=input');
  await page.touchscreen.tap(120,150);await ready();
  await page.keyboard.press('Control+a');await page.keyboard.press('Control+c');
  await page.waitForFunction(async()=>await navigator.clipboard.readText()==='Retained across categories');
  assert.equal(resources.filter(r=>r.url.includes('.wasm')).length,beforeNavigation);
  await page.screenshot({path:path.join(dir,'input.png')});
  // A stale category parameter must not override the component destination.
  await page.goto(new URL('?lang=en&bundle=general#component=input',base).href);await ready();
  assert.ok(resources.slice(-2).some(r=>r.url.includes('/groups/inputs/')));
  assert.deepEqual(errors,[]);
  assert.ok(resources.every(r=>r.status===200 || r.status===304));
  fs.writeFileSync(path.join(dir,'report.json'),JSON.stringify({browser:browser.version(),errors,resources,checks:['mobile startup','touch navigation','browser back','hash cross-category navigation','actual input retained across categories and browser back','no runtime redownload','stale bundle route']},null,2));
 } finally {await browser.close();}
})().catch(e=>{console.error(e);process.exitCode=1});
