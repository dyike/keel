const {test} = require('node:test');
const assert = require('node:assert/strict');
const vm = require('node:vm');
const fs = require('node:fs');
const source = fs.readFileSync('assets/demo-loader.js','utf8');
function boot(fetch, search='', hash='') {
 const handlers={}, timers=[], elements={},requests=[];
 const element=(id)=>elements[id]??=( {hidden:true,value:0,textContent:'',setAttribute(){},remove(){this.removed=true},addEventListener(name,fn){this[name]=fn}} );
 const location={search,hash,href:'https://example.test/sub/demo/'+search+hash,reload(){this.reloaded=true}};
 const assets={ 'main.wasm':{url:'main.hash.wasm',bytes:4}, 'wasm.js':{url:'wasm.hash.js',bytes:1}, 'font.otf':{url:'font.hash.otf',bytes:1}, 'general.wasm':{url:'general.hash.wasm',bytes:4}, 'general.js':{url:'general.hash.js',bytes:1} };
 const context={URL,URLSearchParams,Request,Response,ReadableStream,Error,location,
  document:{documentElement:{},getElementById:element,createElement:()=>({}),body:{append(script){elements.script=script}}},
  keelDemo:{assets,groups:{general:{wasm:'general.wasm',script:'general.js'}},sections:{button:'general'}},
  fetch:async(url,init)=>{requests.push(url);return fetch(url,init)},
  addEventListener:(name,fn)=>handlers[name]=fn,
  setTimeout:(fn)=>{timers.push(fn);return timers.length},clearTimeout:(id)=>{timers[id-1]=null}
 };
 context.window=context;vm.runInNewContext(source,context);
 return {context,elements,handlers,timers,requests,location};
}
test('category routing, streamed progress, ready and immutable asset URLs',async()=>{
 const b=boot(async()=>new Response(new Uint8Array([0,1,2,3])), '?section=button');
 assert.equal(b.elements.script.src,'general.hash.js');
 const response=await b.context.fetch('main.wasm');await response.arrayBuffer();
 assert.deepEqual(b.requests,['general.hash.wasm','font.hash.otf']);
 assert.equal(b.elements.progress.value,100);
 assert.equal(b.elements.note.textContent,'Starting…');
 assert.ok(!b.elements.loading.removed);
 b.handlers['keel-ready']();assert.ok(b.elements.loading.removed);assert.equal(b.timers[0],null);
 await b.context.fetch('font.otf');assert.equal(b.requests.filter(x=>x==='font.hash.otf').length,1);
});
test('explicit complete runtime and nonasset requests',async()=>{
 const b=boot(async()=>new Response('ok'),'?bundle=all');
 assert.equal(b.elements.script.src,'wasm.hash.js');
 await b.context.fetch('https://elsewhere.test/data');assert.equal(b.requests.at(-1),'https://elsewhere.test/data');
});
test('HTTP errors and reload preserve route',async()=>{
 const b=boot(async()=>new Response('',{status:404}),'?lang=zh-CN','#component=button');
 await assert.rejects(b.context.fetch('main.wasm'));
 assert.equal(b.elements['loading-title'].textContent,'组件库加载失败');
 assert.equal(b.elements.retry.hidden,false);assert.equal(b.elements.progress.hidden,true);
 b.handlers['keel-ready']();assert.ok(!b.elements.loading.removed);
 b.elements.retry.click();assert.ok(b.location.reloaded);
});
test('network, compilation, script and stream failures are visible',async()=>{
 const network=boot(async()=>{throw Error('offline')});await assert.rejects(network.context.fetch('main.wasm'));assert.match(network.elements.note.textContent,/offline/);
 const compile=boot(async()=>new Response(''));compile.handlers.unhandledrejection({reason:Error('compile failed')});assert.equal(compile.elements.retry.hidden,false);
 const script=boot(async()=>new Response(''));script.elements.script.onerror();assert.equal(script.elements.retry.hidden,false);
 const stream=boot(async()=>new Response(new ReadableStream({pull(){throw Error('disconnected')}})));
 const response=await stream.context.fetch('main.wasm');await assert.rejects(response.arrayBuffer());assert.match(stream.elements.note.textContent,/disconnected/);
});
test('slow download offers retry without polling or aborting',()=>{
 const b=boot(async()=>new Response(''));assert.equal(b.timers.length,1);b.timers[0]();assert.equal(b.elements.retry.hidden,false);assert.match(b.elements.note.textContent,/longer/);
});
test('runtime exit before readiness fails; after readiness it stays dismissed',()=>{
 const early=boot(async()=>new Response(''));early.handlers['keel-exit']({detail:{code:2}});assert.equal(early.elements.retry.hidden,false);
 const ready=boot(async()=>new Response(''));ready.handlers['keel-ready']();ready.handlers['keel-exit']({detail:{code:0}});assert.ok(ready.elements.loading.removed);
});
