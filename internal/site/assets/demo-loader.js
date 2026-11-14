// Only the demo's known local assets are redirected to content-addressed URLs.
// No polling: progress follows stream reads; readiness comes from the Go app.
(() => {
  const config = window.keelDemo;
  const params = new URLSearchParams(location.search);
  const chinese = params.get("lang") === "zh-CN";
  const section = params.get("section") || new URLSearchParams(location.hash.slice(1)).get("component");
  // Each category is a standalone Go program, not a dynamically linked plugin.
  // Explicit "all" keeps the complete runtime and its cross-category state.
  if (params.get("bundle") !== "all" && section !== "all") {
    const name = params.get("bundle") || config.sections?.[section] || "general";
    const group = config.groups?.[name];
    if (group) {
      config.assets["main.wasm"] = config.assets[group.wasm];
      config.assets["wasm.js"] = config.assets[group.script];
    }
  }
  const label = chinese ? {
    title: "正在加载 Keel 组件库", starting: "正在启动…", retry: "重新加载",
    failed: "组件库加载失败", slow: "加载时间较长，可继续等待或重新加载。",
    note: "首次访问需要下载，后续访问可使用浏览器缓存。"
  } : {
    title: "Loading Keel components", starting: "Starting…", retry: "Reload",
    failed: "Unable to load the gallery", slow: "Loading is taking longer. You can wait or reload.",
    note: "Downloaded on first visit; reusable from the browser cache afterwards."
  };
  document.documentElement.lang = chinese ? "zh-CN" : "en";
  document.title = chinese ? "Keel 组件库" : "Keel component gallery";
  const overlay = document.getElementById("loading");
  const title = document.getElementById("loading-title");
  const note = document.getElementById("note");
  const retry = document.getElementById("retry");
  const progress = document.getElementById("progress");
  title.textContent = label.title;
  note.textContent = label.note;
  retry.textContent = label.retry;
  retry.addEventListener("click", () => location.reload());
  let finished = false;
  let failed = false;
  const fail = (error) => {
    if (finished || failed) return;
    failed = true;
    clearTimeout(slow);
    overlay.setAttribute("role", "alert");
    title.textContent = label.failed;
    note.textContent = error instanceof Error ? error.message : String(error || "");
    retry.hidden = false;
    progress.hidden = true;
  };
  const slow = setTimeout(() => {
    note.textContent = label.slow;
    retry.hidden = false;
  }, 60000);
  const ready = () => {
    if (failed) return;
    finished = true;
    clearTimeout(slow);
    overlay.remove();
    // Keep asset routing for any fonts/assets requested after the first frame.
  };
  window.addEventListener("keel-ready", ready, {once: true});
  window.addEventListener("keel-error", (event) => fail(event.detail));
  window.addEventListener("keel-exit", (event) => fail(new Error(chinese
    ? `应用在就绪前退出（${event.detail.code}）`
    : `Application exited before it was ready (${event.detail.code})`)));
  window.addEventListener("unhandledrejection", (event) => fail(event.reason));
  window.addEventListener("error", (event) => fail(event.error || event.message));
  const originalFetch = window.fetch;
  const fetchAsset = originalFetch.bind(window);
  // Overlap runtime/font transfer with loading the Go support script. Each
  // response is consumed once; the Go fetch calls still drive error/progress UI.
  const pending = new Map();
  for (const name of ["main.wasm", "font.otf"]) {
    const asset = config.assets[name];
    if (!asset) continue;
    const response = fetchAsset(asset.url);
    response.catch(() => {});
    pending.set(asset.url, response);
  }
  const redirects = new Map(Object.entries(config.assets).map(([name, asset]) => [new URL(name, location.href).href, asset]));
  window.fetch = async (input, init) => {
    const address = new URL(typeof input === "string" || input instanceof URL ? input : input.url, location.href).href;
    const asset = redirects.get(address);
    if (!asset) return fetchAsset(input, init);
    try {
      const prefetched = pending.get(asset.url);
      pending.delete(asset.url);
      const response = await (prefetched || fetchAsset(asset.url, init));
      if (!response.ok) throw new Error(`HTTP ${response.status}: ${asset.url}`);
      if (address !== new URL("main.wasm", location.href).href || !response.body) return response;
      const reader = response.body.getReader();
      let received = 0;
      const stream = new ReadableStream({
        async pull(controller) {
          try {
            const {done, value} = await reader.read();
            if (done) {
              note.textContent = label.starting;
              controller.close();
              return;
            }
            received += value.length;
            progress.value = Math.min(100, Math.round(received / asset.bytes * 100));
            note.textContent = `${progress.value}% · ${(received / 1048576).toFixed(1)} / ${(asset.bytes / 1048576).toFixed(1)} MiB`;
            controller.enqueue(value);
          } catch (error) { fail(error); controller.error(error); }
        },
        cancel(reason) { return reader.cancel(reason); }
      });
      return new Response(stream, {headers: {"Content-Type": "application/wasm"}});
    } catch (error) { fail(error); throw error; }
  };
  const script = document.createElement("script");
  script.src = config.assets["wasm.js"].url;
  script.onerror = () => fail(new Error(`Unable to load ${script.src}`));
  document.body.append(script);
})();
