const test = require('node:test');
const assert = require('node:assert/strict');
const install = require('./shared-gallery-loader.js');
test('single child, shared sections, failure retry and retained successful instance', async () => {
 let calls = 0; const notifications = [], instance = {};
 const loader = install({}, {dialog:'child.wasm', alias:'child.wasm'}, {
  notify: name => notifications.push(name),
  instantiate: async () => { if (++calls === 1) throw Error('offline'); return {instance}; }
 });
 assert.equal(loader.status('input'), '');
 assert.equal(calls, 0);
 assert.equal(loader.status('dialog'), 'loading');
 assert.strictEqual(loader.ensure('dialog'), loader.ensure('alias'));
 await assert.rejects(loader.ensure('dialog'), /offline/);
 assert.equal(loader.status('alias'), 'error');
 assert.equal(calls, 1);
 loader.retry('dialog');
 assert.strictEqual(await loader.ensure('alias'), instance);
 assert.equal(loader.status('dialog'), '');
 loader.retry('dialog');
 assert.strictEqual(await loader.ensure('dialog'), instance);
 assert.equal(calls, 2);
 assert.deepEqual(notifications, ['dialog', 'alias', 'dialog', 'alias']);
});
test('multi-shard section deduplicates shared dependency across sections and retries failed shard only', async () => {
 const fetched = [], notifications = [];
 let failLeaf = true;
 const loader = install({}, {
  code_editor: ['shared_01.wasm', 'code_editor.wasm'],
  terminal: ['shared_01.wasm', 'terminal.wasm'],
 }, {
  notify: name => notifications.push(name),
  instantiate: async url => {
   fetched.push(url);
   if (url === 'code_editor.wasm' && failLeaf) {
    failLeaf = false;
    throw Error('leaf failed');
   }
   return {instance: {url}};
  }
 });
 await assert.rejects(loader.ensure('code_editor'), /leaf failed/);
 assert.equal(loader.status('code_editor'), 'error');
 assert.deepEqual(fetched, ['shared_01.wasm', 'code_editor.wasm']);
 loader.retry('code_editor');
 assert.deepEqual(await loader.ensure('code_editor'), {url: 'code_editor.wasm'});
 assert.equal(loader.status('code_editor'), '');
 assert.deepEqual(await loader.ensure('terminal'), {url: 'terminal.wasm'});
 assert.deepEqual(fetched, ['shared_01.wasm', 'code_editor.wasm', 'code_editor.wasm', 'terminal.wasm']);
 assert.deepEqual(notifications, ['code_editor', 'code_editor', 'terminal']);
});
