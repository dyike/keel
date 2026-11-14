// Experimental single-runtime loader. Child URLs must belong to this primary build.
(function (root) {
  function install(primary, sections, options = {}) {
    const instantiate = options.instantiate || (url => WebAssembly.instantiateStreaming(fetch(url), {primary: primary.exports}));
    const notify = options.notify || (name => root.dispatchEvent(new CustomEvent('keel-section-loaded', {detail: name})));
    const states = new Map();
    function ensure(name) {
      const url = sections[name];
      if (!url) return Promise.resolve();
      let state = states.get(url);
      if (state) return state.pending;
      state = {status: 'loading', pending: null};
      states.set(url, state);
      state.pending = Promise.resolve().then(() => instantiate(url)).then(result => {
        state.status = '';
        state.instance = result.instance;
        for (const [section, child] of Object.entries(sections)) if (child === url) notify(section);
        return result.instance;
      }, error => {
        state.status = 'error';
        for (const [section, child] of Object.entries(sections)) if (child === url) notify(section);
        throw error;
      });
      // Render-triggered loads report failures through visible UI, not an unhandled rejection.
      state.pending.catch(() => {});
      return state.pending;
    }
    function status(name) {
      if (!sections[name]) return '';
      ensure(name);
      return states.get(sections[name]).status;
    }
    function retry(name) {
      const url = sections[name], state = states.get(url);
      if (state?.status === 'error') states.delete(url);
      ensure(name);
    }
    root.keelSectionStatus = status;
    root.keelRetrySection = retry;
    return {ensure, status, retry};
  }
  if (typeof module !== 'undefined') module.exports = install;
  else root.installSharedGalleryLoader = install;
})(globalThis);
