// Reproducible local HTTP acceptance, not a production CDN or physical-phone test.
const {chromium} = require('playwright');
const fs = require('node:fs'), path = require('node:path'), http = require('node:http');
const os = require('node:os');
const zlib = require('node:zlib'), assert = require('node:assert/strict');
const root = path.resolve(process.argv[2] || '_site');
const reportDir = process.argv[3] || 'work/wasm-loading';
const samples = Number(process.env.KEEL_PERF_SAMPLES || 5);
assert.ok(samples >= 3, 'at least three samples are required');
const assets = new Map();
// Compress before timing, so compression CPU is excluded from measurements.
function walk(dir) {
 for (const e of fs.readdirSync(dir,{withFileTypes:true})) {
  const file=path.join(dir,e.name); if(e.isDirectory()) {walk(file);continue;}
  const bytes=fs.readFileSync(file), ext=path.extname(file);
  assets.set('/'+path.relative(root,file).split(path.sep).join('/'), {
   bytes:zlib.gzipSync(bytes,{level:9}), type:({'.wasm':'application/wasm','.js':'text/javascript','.html':'text/html','.css':'text/css','.otf':'font/otf','.json':'application/json'})[ext] || 'application/octet-stream',
   cache:/\.[a-f0-9]{16,}\./.test(e.name) ? 'public, max-age=31536000, immutable' : 'no-cache'
  });
 }
}
walk(root);
const server=http.createServer((req,res)=>{
 let name=new URL(req.url,'http://localhost').pathname; if(name.endsWith('/'))name+='index.html';
 const asset=assets.get(name); if(!asset){res.writeHead(404);res.end();return;}
 res.writeHead(200,{'Content-Type':asset.type,'Content-Encoding':'gzip','Content-Length':asset.bytes.length,'Cache-Control':asset.cache});res.end(asset.bytes);
});
const median = values => [...values].sort((a,b)=>a-b)[Math.floor(values.length/2)];
(async()=>{
 fs.mkdirSync(reportDir,{recursive:true});
 await new Promise(r=>server.listen(0,'127.0.0.1',r));
 let browser;
 const report={profile:{samples,viewport:'390x844',networkMbps:6,latencyMs:40,cpuSlowdown:2,encoding:'gzip-9',server:'local controlled HTTP; immutable hashed assets',cache:'fresh normal Chrome profile per cold sample; default disk-cache capacity'},runs:[],errors:[]};
 try {
  for(let i=0;i<samples;i++) for(const mode of (i%2 ? ['full','shared'] : ['shared','full'])) {
   // A normal fresh profile has a disk cache; incognito contexts may reject
   // large WASM entries from their small in-memory cache. Do not enlarge it.
   const profile=fs.mkdtempSync(path.join(os.tmpdir(),'keel-loading-'));
   const context=await chromium.launchPersistentContext(profile,{channel:process.env.KEEL_BROWSER_CHANNEL || 'chrome',args:['--enable-unsafe-swiftshader'],viewport:{width:390,height:844},isMobile:true,hasTouch:true});
   browser=context.browser();report.browser=browser.version();
   try {
    const page=await context.newPage(), session=await context.newCDPSession(page);
    page.on('pageerror',e=>report.errors.push(String(e)));
    page.on('console',m=>{if(m.type()==='error')report.errors.push(m.text())});
    await session.send('Network.enable');
    await session.send('Network.emulateNetworkConditions',{offline:false,latency:40,downloadThroughput:6*1000000/8,uploadThroughput:6*1000000/8});
    await session.send('Emulation.setCPUThrottlingRate',{rate:2});
    await page.addInitScript(()=>{
     window.readyAt=0; addEventListener('keel-ready',()=>window.readyAt=performance.now());
     addEventListener('keel-exit',e=>console.error('Go exited: '+JSON.stringify(e.detail)));
    });
    const settled=async(section)=>{
     await page.waitForFunction(section=>window.readyAt && (!window.keelSectionStatus || keelSectionStatus(section)===''),section,{timeout:90000});
     return page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve(performance.now())))));
    };
    const metrics=async()=>page.evaluate(()=>({
     readyMs:window.readyAt,
     totalTransfer:performance.getEntriesByType('resource').reduce((sum,r)=>sum+r.transferSize,0)+performance.getEntriesByType('navigation')[0].transferSize,
     resources:performance.getEntriesByType('resource').filter(r=>new URL(r.name).pathname.endsWith('.wasm')).map(r=>({file:new URL(r.name).pathname.split('/').pop(),transfer:r.transferSize,encoded:r.encodedBodySize,decoded:r.decodedBodySize,duration:r.duration})),
     wasmMemoryBytes:window.sharedPrimary?.exports.mem.buffer.byteLength || null
    }));
    const url=`http://127.0.0.1:${server.address().port}/demo/?lang=en${mode==='full'?'&bundle=all':''}#component=input`;
    await page.goto(url);
    const coldMs=await settled('input'), cold=await metrics();
    report.inProgress={mode,sample:i,coldMs,cold};
    assert.equal(cold.resources.length,mode==='full'?1:2,'cold load must fetch only the primary and requested category');
    await page.reload();
    const warmMs=await settled('input'), warm=await metrics();
    Object.assign(report.inProgress,{warmMs,warm});
    assert.ok(warm.resources.every(r=>r.transfer===0),'hashed WASM should load from browser cache: '+JSON.stringify(warm));
    const before=await page.evaluate(()=>performance.now());
    await page.evaluate(()=>{location.hash='component=dialog'});
    const switchMs=(await settled('dialog'))-before;
    const afterSwitch=await metrics();
    assert.equal(afterSwitch.resources.length,mode==='full'?1:3,'switch downloads only its own child');
    const beforeRepeat=await page.evaluate(()=>{location.hash='component=input';return performance.now()});
    await settled('input');
    await page.evaluate(()=>{location.hash='component=dialog'});
    const repeatMs=(await settled('dialog'))-beforeRepeat;
    assert.equal((await metrics()).resources.length,afterSwitch.resources.length,'revisit must not redownload');
    if(i===0)await page.screenshot({path:path.join(reportDir,mode+'.png')});
    report.runs.push({mode,sample:i,coldMs,warmMs,switchMs,repeatMs,cold,warm,afterSwitch});
    delete report.inProgress;
    console.log(`${mode} sample ${i+1}/${samples}: cold=${coldMs.toFixed(0)}ms warm=${warmMs.toFixed(0)}ms switch=${switchMs.toFixed(0)}ms`);
   } finally {await context.close();browser=null;fs.rmSync(profile,{recursive:true,force:true});}
  }
  report.summary=Object.fromEntries(['shared','full'].map(mode=>{
   const runs=report.runs.filter(r=>r.mode===mode);
   return [mode,Object.fromEntries(['coldMs','warmMs','switchMs','repeatMs','coldWasmTransfer','coldTotalTransfer'].map(key=>[key,median(runs.map(r=>key==='coldTotalTransfer'?r.cold.totalTransfer:key==='coldWasmTransfer'?r.cold.resources.reduce((n,a)=>n+a.transfer,0):r[key]))]))];
  }));
  const {shared,full}=report.summary;
  report.gates={noBrowserErrors:report.errors.length===0,firstWasmTransferReduction:shared.coldWasmTransfer<=full.coldWasmTransfer*0.95,firstTotalTransferReduction:shared.coldTotalTransfer<=full.coldTotalTransfer*0.95,coldStartup:shared.coldMs<=full.coldMs*1.15+100,warmStartup:shared.warmMs<=full.warmMs*1.15+100,firstCategorySwitch:shared.switchMs<3000};
  fs.writeFileSync(path.join(reportDir,'report.json'),JSON.stringify(report,null,2)+'\n');
  console.log(JSON.stringify({browser:report.browser,summary:report.summary,gates:report.gates},null,2));
  assert.ok(Object.values(report.gates).every(Boolean),'loading acceptance gate failed; see report.json');
 } finally {
  if(!report.gates)fs.writeFileSync(path.join(reportDir,'report.json'),JSON.stringify(report,null,2)+'\n');
  await browser?.close();await new Promise(r=>server.close(r));
 }
})().catch(e=>{console.error(e);process.exitCode=1});
