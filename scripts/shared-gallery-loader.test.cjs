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
