package main

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("terminal", "data", func() core.Widget {
		term := kit.Terminal().Name(demoText("Terminal session", "终端会话")).Height(380)
		prompt := "\x1b[32mkeel\x1b[0m:\x1b[34m~\x1b[0m$ "
		line := ""
		welcomed := false
		running := false
		generation := 0
		write := func(s string) { term.WriteString(s) }
		help := func() { write("help · colors · progress · echo <text> · clear\r\n") }
		var tick func(int, int)
		tick = func(token, done int) {
			time.AfterFunc(60*time.Millisecond, func() {
				core.UpdateIf(func() bool {
					if generation != token || !running {
						return false
					}
					next := done + 5
					write(fmt.Sprintf("\r\x1b[36m%s\x1b[2m%s\x1b[0m %d%%", strings.Repeat("█", next/5), strings.Repeat("░", 20-next/5), next))
					if next == 100 {
						running = false
						write("\r\n" + prompt)
					} else {
						tick(token, next)
					}
					return true
				})
			})
		}
		term.OnResize(func(cols, rows int) {
			if !welcomed {
				welcomed = true
				write("Keel terminal · pure Go\r\n")
				help()
				write(prompt)
			}
		}).OnData(func(data string) {
			if strings.ContainsRune(data, 3) {
				generation++
				running = false
				line = ""
				write("^C\r\n" + prompt)
				return
			}
			if running {
				return
			}
			if strings.HasPrefix(data, "\x1b") {
				return
			}
			for _, r := range data {
				switch r {
				case '\r':
					write("\r\n")
					command := strings.TrimSpace(line)
					line = ""
					name, rest, _ := strings.Cut(command, " ")
					switch name {
					case "help":
						help()
					case "clear":
						write("\x1b[3J\x1b[2J\x1b[H")
					case "echo":
						write(rest + "\r\n")
					case "colors":
						for i := 0; i < 16; i++ {
							write(fmt.Sprintf("\x1b[48;5;%dm  \x1b[0m", i))
						}
						write("\r\n")
						for i := 0; i < 32; i++ {
							write(fmt.Sprintf("\x1b[48;2;%d;96;%dm \x1b[0m", 255-i*8, i*8))
						}
						write("\r\n")
					case "progress":
						running = true
						generation++
						tick(generation, 0)
					case "":
					default:
						write("\x1b[31mUnknown command: " + name + "\x1b[0m\r\n")
					}
					if !running {
						write(prompt)
					}
				case 0x7f:
					if line != "" {
						_, n := utf8.DecodeLastRuneInString(line)
						line = line[:len(line)-n]
						write("\b \b")
					}
				default:
					if r >= ' ' {
						line += string(r)
						write(string(r))
					}
				}
			}
		})
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(16).Gap(12).ScrollY().Child(el.Text(demoText("Type help. Ctrl+C cancels progress; Ctrl+Shift+C copies a selection.", "输入 help。Ctrl+C 取消进度；Ctrl+Shift+C 复制选区。")), term.Render(cx))
		}))
	})
}
