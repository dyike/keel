package gogio

import (
	"encoding/json"
	"os/exec"
	"testing"
)

func TestWebStartupReportsExitAndFailure(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node required for startup glue tests")
	}
	source, _ := json.Marshal(jsStartGo)
	script := `
const vm=require('node:vm'),assert=require('node:assert/strict');
(async()=>{
 for(const mode of ['compile','run','exit']) {
  const events=[];
  class Go {
   constructor(){this.argv=[];this.env={};this.importObject={};this.exit=()=>{}}
   async run(){if(mode==='run')throw Error('panic');this.exit(2)}
  }
  const ctx={Go,go:{argv:[],env:{},importObject:{}},fetch:async()=>({}),
   CustomEvent:class{constructor(type,options){this.type=type;this.detail=options.detail}},
   dispatchEvent:e=>events.push(e),console:{error(){}},
   WebAssembly:{instantiateStreaming:async()=>{if(mode==='compile')throw Error('bad wasm');return {instance:{}}}}
  };
  ctx.window=ctx;
  vm.runInNewContext(SOURCE,ctx);
  await new Promise(resolve=>setImmediate(resolve));
  assert.equal(events.length,1);
  assert.equal(events[0].type,mode==='exit'?'keel-exit':'keel-error');
  if(mode==='exit')assert.equal(events[0].detail.code,2);
 }
})().catch(e=>{console.error(e);process.exitCode=1});`
	cmd := exec.Command(node, "-e", "const SOURCE="+string(source)+";\n"+script)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, output)
	}
}
