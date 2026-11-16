// Verify an already-open page across replacement of the published entry HTML.
const {chromium}=require('playwright');
const fs=require('node:fs'),http=require('node:http'),path=require('node:path'),assert=require('node:assert/strict');
const [dir,previousIndex,reportDir]=process.argv.slice(2);
if(!dir || !previousIndex || !reportDir)throw Error('usage: node scripts/verify-gallery-retention.cjs DEMO_DIR PREVIOUS_INDEX REPORT_DIR');
const html=fs.readFileSync(previousIndex),nextHTML=fs.readFileSync(path.join(dir,'index.html'));
fs.mkdirSync(reportDir,{recursive:true});
let published=false;const requests=[];
const server=http.createServer((req,res)=>{
 const name=new URL(req.url,'http://localhost').pathname.split('/').pop();
 requests.push({name,published});
 if(!name){res.setHeader('Content-Type','text/html');res.end(published?nextHTML:html);return;}
 // Favicon is site-owned, not a versioned runtime asset.
 const file=name==='favicon.svg'?path.join(dir,'..',name):path.join(dir,name);if(!fs.existsSync(file)){res.writeHead(404);res.end();return;}
 res.setHeader('Content-Type',name.endsWith('.wasm')?'application/wasm':name.endsWith('.js')?'text/javascript':'application/octet-stream');res.end(fs.readFileSync(file));
});
(async()=>{await new Promise(r=>server.listen(0,'127.0.0.1',r));let browser,page;const report={passed:false,requests};
try{
 browser=await chromium.launch({channel:process.env.KEEL_BROWSER_CHANNEL || 'chrome',args:['--enable-unsafe-swiftshader']});
 page=await browser.newPage({viewport:{width:390,height:844},hasTouch:true,isMobile:true,permissions:['clipboard-read','clipboard-write']});
 const errors=[];report.errors=errors;page.on('pageerror',e=>errors.push(String(e)));
 page.on('console',m=>{if(m.type()==='error')errors.push(m.text())});
 await page.goto(`http://127.0.0.1:${server.address().port}/?lang=en#component=input`);
 await page.waitForFunction(()=>window.keelSectionStatus && keelSectionStatus('input')==='');
 const frames=()=>page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));await frames();
 await page.touchscreen.tap(120,150);await page.waitForFunction(()=>document.activeElement.tagName==='TEXTAREA');
 await page.keyboard.insertText('Kept after deployment');await frames();
 const child=await page.evaluate(()=>{window.oldMemory=sharedPrimary.exports.mem;return keelDemo.assets[keelDemo.shared.sections.dialog].url});
 assert.ok(!requests.some(r=>r.name===child));
 // Atomically replace the entry document; the old page stays open.
 published=true;
 assert.equal(await (await page.request.get(page.url())).text(),nextHTML.toString(),'server did not switch to the new entry');
 await page.evaluate(()=>{location.hash='component=dialog'});
 await page.waitForFunction(()=>keelSectionStatus('dialog')==='');await frames();
 assert.ok(requests.some(r=>r.name===child && r.published));
 assert.equal(await page.evaluate(()=>oldMemory===sharedPrimary.exports.mem),true);
 await page.evaluate(()=>{location.hash='component=input'});await frames();
 await page.touchscreen.tap(120,150);await frames();await page.keyboard.press('Control+a');await page.keyboard.press('Control+c');
 await page.waitForFunction(async()=>await navigator.clipboard.readText()==='Kept after deployment');
 assert.deepEqual(errors,[]);
 Object.assign(report,{passed:true,browser:browser.version(),oldChildLoadedAfterPublication:child,inputRetained:true,sharedMemory:true});
}catch(error){report.failure=String(error);throw error;}finally{
 if(page)await page.screenshot({path:path.join(reportDir,'final-state.png')}).catch(()=>{});
 fs.writeFileSync(path.join(reportDir,'report.json'),JSON.stringify(report,null,2));
 await browser?.close();await new Promise(r=>server.close(r));
}
})().catch(e=>{console.error(e);process.exitCode=1});
