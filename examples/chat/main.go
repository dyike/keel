// Chat shows streaming Markdown answers the way an AI assistant produces them:
// text arrives a few characters at a time, the view follows it to the bottom
// unless you scroll up, and the answer can be stopped. Answers are canned; a
// real app would read tokens from a model API in the same goroutine.
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/markdown"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

type message struct {
	user   bool
	text   string        // user messages
	doc    *markdown.Doc // assistant messages
	copier *kit.CopyButtonView
}

// copy is the answer's copy button, kept so its 已复制 feedback survives frames.
func (m *message) copy(src string) *kit.CopyButtonView {
	if m.copier == nil {
		m.copier = kit.CopyButton(func() string { return src })
	}
	return m.copier
}

type chat struct {
	msgs     []*message
	prompt   string
	stop     *atomic.Bool // set while an answer streams
	delay    time.Duration
	preview  bool // a complete startup sample opens at the top
	scroller *kit.MessageScrollerView
}

func newChat(delay time.Duration) *chat {
	c := &chat{delay: delay}
	c.scroller = kit.MessageScroller(nil, 160, c.item)
	c.syncKeys()
	return c
}

func (c *chat) Render(cx *el.Context) el.Element {
	status := "在线"
	if c.stop != nil {
		status = "正在回答…"
	}
	c.scroller.SetFollow(!c.preview)
	return el.Div().Child(
		el.Div().Row().Items(el.Center).Px(20).Py(12).Bg(theme.Surface).Child(
			el.Text("AI 助手").Bold().Grow(),
			el.Text(status).TextSize(13).TextColor(theme.Muted),
		),
		el.Div().H(el.Dp(1)).Bg(theme.Border),
		c.scroller.Render(cx),
		el.Div().H(el.Dp(1)).Bg(theme.Border),
		el.Div().Row().Gap(8).Items(el.Center).P(12).Bg(theme.Surface).Child(
			el.Input().ID("prompt").Name("消息").Placeholder("输入消息，回车发送").Bind(&c.prompt).
				OnSubmit(func(string) { c.send() }).Grow(),
			c.action(cx),
		),
	)
}

// items lists the conversation for the scroller, or sample prompts before it starts.
func (c *chat) items(cx *el.Context) []el.Element {
	if len(c.msgs) == 0 {
		out := []el.Element{el.Text("选择一个 Markdown 样例，或输入消息查看预设的流式回答。").TextColor(theme.Muted)}
		for _, s := range demoSamples {
			out = append(out, el.Div().Row().Name(s.title).Gap(12).Items(el.Center).P(8).Rounded(6).
				CursorPointer().Hover(func(st *el.Style) { st.Bg(theme.SubtleHover) }).
				OnClick(func() { c.prompt = s.title; c.send() }).Child(
				el.Text(s.title).Bold().W(el.Dp(110)),
				el.Text(s.description).TextSize(13).TextColor(theme.Muted).Grow(),
			))
		}
		return append(out, el.Div().P(8).Rounded(6).CursorPointer().
			Hover(func(st *el.Style) { st.Bg(theme.SubtleHover) }).
			OnClick(func() { c.prompt = "全部样例"; c.send() }).Child(el.Text("查看全部样例")))
	}
	return nil
}
func (c *chat) item(cx *el.Context, i int) el.Element {
	if len(c.msgs) == 0 {
		return c.items(cx)[i]
	}
	m := c.msgs[i]
	if m.user {
		return kit.Message("我", el.ViewFunc(func(*el.Context) el.Element { return el.Text(m.text) })).User().Render(cx)
	}
	msg := kit.Message("AI", m.doc)
	if !m.doc.Streaming() {
		msg.Actions(m.copy(m.doc.Source()))
	}
	return msg.Render(cx)
}
func (c *chat) syncKeys() {
	keys := make([]string, len(c.msgs))
	if len(keys) == 0 {
		keys = make([]string, len(demoSamples)+2)
		for i := range keys {
			keys[i] = fmt.Sprintf("welcome-%d", i)
		}
	} else {
		for i, m := range c.msgs {
			keys[i] = fmt.Sprintf("message-%p", m)
		}
	}
	c.scroller.SetKeys(keys)
}

func (c *chat) action(cx *el.Context) el.Element {
	if c.stop != nil {
		return kit.Button("停止", func() { c.stop.Store(true) }).Variant(kit.ButtonDanger).Render(cx)
	}
	return kit.Button("发送", c.send).Render(cx)
}

// send posts the prompt and streams an answer from a goroutine, the way a
// model API would, handing each token to the UI through core.Update.
func (c *chat) send() {
	q := strings.TrimSpace(c.prompt)
	if q == "" || c.stop != nil {
		return
	}
	c.prompt = ""
	c.preview = false
	doc := newDocument("")
	doc.SetStreaming(true)
	c.msgs = append(c.msgs, &message{user: true, text: q}, &message{doc: doc})
	c.syncKeys()
	c.scroller.ScrollToEnd() // sending jumps to the end even if the user had scrolled up
	stop := new(atomic.Bool)
	c.stop = stop
	answer := answerFor(q)
	go func() {
		for rest := []rune(answer); len(rest) > 0 && !stop.Load(); {
			n := min(1+rand.IntN(4), len(rest)) // a token is a few characters
			tok := string(rest[:n])
			rest = rest[n:]
			core.Update(func() { doc.Append(tok) })
			time.Sleep(c.delay)
		}
		core.Update(func() {
			if stop.Load() {
				doc.Append("\n\n*（已停止）*")
			}
			doc.SetStreaming(false)
			c.stop = nil
		})
	}()
}

func newDocument(src string) *markdown.Doc {
	return markdown.New(src).OnLink(func(url string) { exec.Command("open", url).Start() })
}

const downloaderAnswer = "## 并发下载器\n\n" +
	"下面用 **goroutine** 加 **有缓冲的 channel** 限制并发数，每个下载任务独立处理错误。关键点：\n\n" +
	"1. 用 `sync.WaitGroup` 等待全部任务结束；\n" +
	"2. 用容量为 `n` 的 channel 当信号量，限制同时下载的数量；\n" +
	"3. 用 `context` 支持整体取消。\n\n" +
	"```go\nfunc Download(ctx context.Context, urls []string, n int) error {\n" +
	"\tsem := make(chan struct{}, n)\n\tvar wg sync.WaitGroup\n\terrs := make(chan error, len(urls))\n" +
	"\tfor _, u := range urls {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n" +
	"\t\t\tsem <- struct{}{}        // 占一个名额\n\t\t\tdefer func() { <-sem }()\n" +
	"\t\t\tif err := fetch(ctx, u); err != nil {\n\t\t\t\terrs <- fmt.Errorf(\"%s: %w\", u, err)\n\t\t\t}\n\t\t}()\n\t}\n" +
	"\twg.Wait()\n\tclose(errs)\n\treturn errors.Join(slices.Collect(chanValues(errs))...)\n}\n```\n\n" +
	"| 并发数 | 100 个文件耗时 | 说明 |\n|--:|--:|:--|\n| 1 | 42.0s | 串行 |\n| 8 | 5.6s | 推荐 |\n| 64 | 4.9s | 服务端可能限流 |\n\n" +
	"> 注意：并发数不是越大越好，超过带宽或对方的限流阈值后收益很小。\n\n" +
	"- [x] 限制并发\n- [x] 汇总错误\n- [ ] 断点续传（留作练习）\n\n" +
	"更多用法见 [Go 并发模式](https://go.dev/blog/pipelines)。下载器完成。"

const tableAnswer = "三种方案对比：\n\n" +
	"| 方案 | 延迟 | 成本 | 适合 |\n|:--|--:|--:|:--|\n" +
	"| 本地缓存 | 1ms | 低 | 读多写少 |\n| Redis | 3ms | 中 | 多实例共享 |\n| 数据库 | 15ms | 高 | 强一致 |\n\n" +
	"结论：先用**本地缓存**，多实例后再加 Redis。对比完成。"

func main() {
	delay := flag.Duration("delay", 15*time.Millisecond, "pause between streamed tokens")
	sample := flag.String("sample", "", "show a complete sample immediately: all, selection, math, code-scroll, click-selection, references, images, table, downloader")
	flag.Parse()
	c := newChat(*delay)
	if *sample != "" {
		src, ok := sourceFor(*sample)
		if !ok {
			log.Fatalf("unknown sample %q; run with -help for available names", *sample)
		}
		c.msgs = []*message{{doc: newDocument(src)}}
		c.preview = true
		c.syncKeys()
	}
	window.Open(window.Options{Title: "AI 助手", Width: 720, Height: 640, Content: el.Root(c)})
	window.Main()
}
