package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// copyDemo publishes gogio's build of the component gallery as demo/, with
// a page of our own that shows download progress: the WebAssembly binary is
// tens of megabytes, about ten once compressed in transit.
func (s *site) copyDemo(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	size := int64(0)
	for _, e := range entries {
		if e.IsDir() || e.Name() == "index.html" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return err
		}
		if e.Name() == "main.wasm" {
			size = int64(len(data))
		}
		if err := s.write("demo/"+e.Name(), data); err != nil {
			return err
		}
	}
	if size == 0 {
		return fmt.Errorf("%s has no main.wasm; build it with gogio -target js", dir)
	}
	return s.write("demo/index.html", fmt.Appendf(nil, demoPage, strconv.FormatInt(size, 10)))
}

// demoPage wraps gogio's wasm.js. It replaces fetch("main.wasm") with a
// streaming copy that reports progress against the uncompressed size (the
// server's Content-Length is the compressed one), and hides the overlay
// once the program has drawn its canvas.
const demoPage = `<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, user-scalable=no">
<title>Keel 组件库</title>
<style>
  html, body { margin: 0; padding: 0; height: 100%%; overflow: hidden; }
  body { background: #f5f6f8; font: 14px -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif; color: #1f2328; }
  #loading { position: fixed; inset: 0; display: grid; place-content: center; gap: 12px; text-align: center; background: inherit; z-index: 1; }
  #bar { width: 240px; height: 6px; border-radius: 3px; background: #e3e5e8; overflow: hidden; }
  #fill { width: 0; height: 100%%; background: #2563eb; transition: width .15s; }
  #note { color: #6b7280; font-size: 12px; }
  @media (prefers-color-scheme: dark) { body { background: #111827; color: #f3f4f6; } #bar { background: #374151; } #note { color: #9ca3af; } }
</style>
<script>
(() => {
  const total = %s;
  const plain = window.fetch.bind(window);
  window.fetch = async (url, init) => {
    if (String(url) !== "main.wasm") return plain(url, init);
    const resp = await plain(url, init);
    if (!resp.body) return resp;
    const fill = () => document.getElementById("fill");
    const note = () => document.getElementById("note");
    let got = 0;
    const reader = resp.body.getReader();
    const stream = new ReadableStream({
      async pull(c) {
        const { done, value } = await reader.read();
        if (done) {
          note() && (note().textContent = "正在启动…");
          c.close();
          return;
        }
        got += value.length;
        const pct = Math.min(100, Math.round(got / total * 100));
        fill() && (fill().style.width = pct + "%%");
        note() && (note().textContent = pct + "%% · " + (got / 1048576).toFixed(1) + " / " + (total / 1048576).toFixed(1) + " MB");
        c.enqueue(value);
      },
    });
    return new Response(stream, { headers: { "Content-Type": "application/wasm" } });
  };
  new MutationObserver((_, obs) => {
    if (document.querySelector("canvas")) {
      document.getElementById("loading")?.remove();
      obs.disconnect();
    }
  }).observe(document.documentElement, { childList: true, subtree: true });
})();
</script>
<script src="wasm.js"></script>
</head>
<body>
<div id="loading"><div>正在加载 Keel 组件库</div><div id="bar"><div id="fill"></div></div><div id="note">首次加载约 10 MB，之后由浏览器缓存</div></div>
</body>
</html>
`
