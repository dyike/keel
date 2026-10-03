// Search, copy buttons, demo frames and the mobile menu. No dependencies.
(() => {
  const root = document.body.dataset.root || "./";
  const dark = matchMedia("(prefers-color-scheme: dark)").matches;

  // Demo frames load the gallery in the reader's color scheme.
  for (const f of document.querySelectorAll(".demo iframe[data-src]")) {
    f.src = f.dataset.src + (dark ? "&theme=dark" : "");
  }

  // Copy buttons on code blocks.
  for (const b of document.querySelectorAll(".code .copy")) {
    b.addEventListener("click", async () => {
      const code = b.parentElement.querySelector("pre").innerText;
      try {
        await navigator.clipboard.writeText(code);
        b.textContent = "已复制";
      } catch {
        b.textContent = "复制失败";
      }
      setTimeout(() => (b.textContent = "复制"), 1500);
    });
  }

  // Mobile menu.
  document.querySelector(".menu")?.addEventListener("click", () => document.body.classList.toggle("nav-open"));
  document.querySelector('.side [aria-current]')?.scrollIntoView({ block: "center" });

  // Highlight the outline entry of the section in view.
  const toc = [...document.querySelectorAll(".toc a")];
  if (toc.length) {
    const targets = toc.map((a) => document.getElementById(decodeURIComponent(a.hash.slice(1)))).filter(Boolean);
    const onScroll = () => {
      let current = targets[0];
      for (const t of targets) if (t.getBoundingClientRect().top < 120) current = t;
      for (const a of toc) a.classList.toggle("active", current && decodeURIComponent(a.hash.slice(1)) === current.id);
    };
    addEventListener("scroll", onScroll, { passive: true });
    onScroll();
  }

  // Search over titles, headings and text; titles weigh most.
  const input = document.querySelector(".search input");
  const list = document.querySelector(".search .results");
  let index = null;
  let active = -1;
  const load = async () => {
    if (!index) index = await (await fetch(root + "search.json")).json();
    return index;
  };
  const escape = (s) => s.replace(/[&<>"]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" })[c]);
  const snippet = (text, q) => {
    const i = text.toLowerCase().indexOf(q);
    if (i < 0) return "";
    return (i > 20 ? "…" : "") + text.slice(Math.max(0, i - 20), i + 60);
  };
  const search = async () => {
    const q = input.value.trim().toLowerCase();
    if (!q) {
      list.hidden = true;
      return;
    }
    const pages = await load();
    const hits = [];
    for (const p of pages) {
      let score = 0;
      const title = p.t.toLowerCase();
      if (title === q) score += 100;
      else if (title.startsWith(q)) score += 60;
      else if (title.includes(q)) score += 40;
      const h = (p.h || []).find(([text]) => text.toLowerCase().includes(q));
      if (h) score += 15;
      if (p.x.toLowerCase().includes(q)) score += 5;
      if (score) hits.push({ p, score, h });
    }
    hits.sort((a, b) => b.score - a.score);
    active = -1;
    list.innerHTML = hits.slice(0, 12).map(({ p, h }) =>
      `<li><a href="${root + p.u + (h ? "#" + encodeURIComponent(h[1]) : "")}">${escape(p.t)}<span class="group">${escape(p.g || "")}</span><small>${escape(h ? h[0] : snippet(p.x, q))}</small></a></li>`
    ).join("") || '<li><a>没有找到</a></li>';
    list.hidden = false;
  };
  input?.addEventListener("input", search);
  input?.addEventListener("focus", () => input.value && search());
  input?.addEventListener("keydown", (e) => {
    const items = [...list.querySelectorAll("a[href]")];
    if (e.key === "ArrowDown" || e.key === "ArrowUp") {
      e.preventDefault();
      active = Math.max(0, Math.min(items.length - 1, active + (e.key === "ArrowDown" ? 1 : -1)));
      items.forEach((a, i) => a.classList.toggle("active", i === active));
    } else if (e.key === "Enter" && items.length) {
      location.href = items[Math.max(active, 0)].href;
    } else if (e.key === "Escape") {
      list.hidden = true;
      input.blur();
    }
  });
  document.addEventListener("click", (e) => {
    if (!e.target.closest(".search")) list.hidden = true;
  });
  document.addEventListener("keydown", (e) => {
    if (e.key === "/" && document.activeElement !== input && !/INPUT|TEXTAREA/.test(document.activeElement.tagName)) {
      e.preventDefault();
      input.focus();
    }
  });
})();
