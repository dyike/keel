// Experimental single-runtime loader. Child URLs must belong to this primary build.
(function (root) {
  function install(primary, sections, options = {}) {
    const instantiate = options.instantiate || (url => WebAssembly.instantiateStreaming(fetch(url), {primary: primary.exports}));
    const notify = options.notify || (name => root.dispatchEvent(new CustomEvent('keel-section-loaded', {detail: name})));
    const states = new Map();
    const urlsOf = name => Array.isArray(sections[name]) ? sections[name] : sections[name] ? [sections[name]] : [];
    function settled(url) {
      for (const section of Object.keys(sections)) {
        const urls = urlsOf(section);
        if (urls.includes(url) && urls.every(u => states.get(u) && states.get(u).status !== 'loading')) {
          notify(section);
        }
      }
    }
    function loadUrl(url) {
      let state = states.get(url);
      if (state) return state.pending;
      state = {status: 'loading', pending: null};
      states.set(url, state);
      state.pending = Promise.resolve().then(() => instantiate(url)).then(result => {
        state.status = '';
        state.instance = result.instance;
        settled(url);
        return result.instance;
      }, error => {
        state.status = 'error';
        settled(url);
        throw error;
      });
      // Render-triggered loads report failures through visible UI, not an unhandled rejection.
      state.pending.catch(() => {});
      return state.pending;
    }
    function ensure(name) {
      const urls = urlsOf(name);
      if (!urls.length) return Promise.resolve();
      if (urls.length === 1) return loadUrl(urls[0]);
      const pending = Promise.all(urls.map(loadUrl)).then(instances => instances[instances.length - 1]);
      pending.catch(() => {});
      return pending;
    }
    function status(name) {
      const urls = urlsOf(name);
      if (!urls.length) return '';
      ensure(name);
      if (urls.some(u => states.get(u).status === 'loading')) return 'loading';
      if (urls.some(u => states.get(u).status === 'error')) return 'error';
      return '';
    }
    function retry(name) {
      for (const url of urlsOf(name)) {
        if (states.get(url)?.status === 'error') states.delete(url);
      }
      ensure(name);
    }
    root.keelSectionStatus = status;
    root.keelRetrySection = retry;
    return {ensure, status, retry};
  }
  if (typeof module !== 'undefined') module.exports = install;
  else root.installSharedGalleryLoader = install;
})(globalThis);
