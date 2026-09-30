// Chat shows streaming Markdown answers the way an AI assistant produces them:
// text arrives a few characters at a time, the view follows it to the bottom
// unless you scroll up, and the answer can be stopped. Answers are canned; a
// real app would read tokens from a model API in the same goroutine.
package main

import (
	"flag"
	"math/rand/v2"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/markdown"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

type message struct {
	user bool
	text string        // user messages
	doc  *markdown.Doc // assistant messages
}

type chat struct {
	msgs   []*message
	prompt string
	stop   *atomic.Bool // set while an answer streams
	delay  time.Duration
}

func (c *chat) Render(cx *el.Context) el.Element {
	status := "在线"
	if c.stop != nil {
		status = "正在回答…"
	}
	// Follow a streaming answer; sending a message jumps to the end even if the
	// user had scrolled up.
	msgs := el.Div().ID("messages").Grow().ScrollY().StickToBottom().ScrollToEndOn(len(c.msgs)).Px(20).Py(16).Gap(18)
	if len(c.msgs) == 0 {
		msgs.Child(el.Text("问点什么，比如：用 Go 写一个并发下载器").TextColor(theme.Muted))
	}
	for i, m := range c.msgs {
		msgs.Child(c.bubble(cx, i, m))
	}
	return el.Div().Child(
		el.Div().Row().Items(el.Center).Px(20).Py(12).Bg(theme.Surface).Child(
			el.Text("AI 助手").Bold().Grow(),
			el.Text(status).TextSize(13).TextColor(theme.Muted),
		),
		el.Div().H(el.Dp(1)).Bg(theme.Border),
		msgs,
		el.Div().H(el.Dp(1)).Bg(theme.Border),
		el.Div().Row().Gap(8).Items(el.Center).P(12).Bg(theme.Surface).Child(
			el.Input().ID("prompt").Name("消息").Placeholder("输入消息，回车发送").Bind(&c.prompt).
				OnSubmit(func(string) { c.send() }).Grow(),
			c.action(),
		),
	)
}

func (c *chat) bubble(cx *el.Context, i int, m *message) el.Element {
	if m.user {
		return el.Div().Row().Justify(el.End).Child(
			el.Div().MaxW(el.Frac(0.75)).Bg(theme.Primary).TextColor(theme.OnColor).Rounded(12).Px(14).Py(10).
				Child(el.Text(m.text)),
		)
	}
	answer := el.Div().Grow().Pt(4).Gap(8).Child(m.doc.Render(cx))
	if !m.doc.Streaming() {
		// Copy the whole answer as Markdown; text inside a paragraph or code
		// block can also be selected and copied with Cmd+C.
		src := m.doc.Source()
		answer.Child(el.Div().Row().Child(
			el.Div().ID("copy-answer").Px(8).Py(3).Rounded(4).TextSize(12).TextColor(theme.Muted).
				CursorPointer().Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) }).
				OnClick(func() { el.WriteClipboard(src) }).Child(el.Text("复制全文")),
		))
	}
	return el.Div().Row().Gap(10).Items(el.Start).Child(
		el.Div().Size(el.Dp(28)).Rounded(14).Bg(theme.Subtle).Center().NoShrink().
			Child(el.Text("AI").TextSize(11).Bold().TextColor(theme.Muted)),
		answer,
	)
}

func (c *chat) action() el.Element {
	label, bg, hover, fn := "发送", theme.Primary, theme.PrimaryHover, c.send
	if c.stop != nil {
		label, bg, hover, fn = "停止", theme.Danger, theme.DangerHover, func() { c.stop.Store(true) }
	}
	return el.Div().ID(label).Px(16).Py(8).Rounded(6).Bg(bg).TextColor(theme.OnColor).TextSize(14).
		CursorPointer().Hover(func(s *el.Style) { s.Bg(hover) }).OnClick(fn).Child(el.Text(label))
}

// send posts the prompt and streams an answer from a goroutine, the way a
// model API would, handing each token to the UI through core.Update.
func (c *chat) send() {
	q := strings.TrimSpace(c.prompt)
	if q == "" || c.stop != nil {
		return
	}
	c.prompt = ""
	doc := markdown.New("").OnLink(func(url string) { exec.Command("open", url).Start() })
	doc.SetStreaming(true)
	c.msgs = append(c.msgs, &message{user: true, text: q}, &message{doc: doc})
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

func answerFor(q string) string {
	if strings.Contains(q, "表") {
		return tableAnswer
	}
	return downloaderAnswer
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
	flag.Parse()
	c := &chat{delay: *delay}
	window.Open(window.Options{Title: "AI 助手", Width: 720, Height: 640, Content: el.Root(c)})
	window.Main()
}
