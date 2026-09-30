package main

import (
	"bytes"
	"image/png"
	"os"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
)

func TestAnswerForSampleRouting(t *testing.T) {
	for _, tc := range []struct{ prompt, sample string }{
		{"双击选择英文词", "click-selection"},
		{"三击选段", "click-selection"},
		{"展示数学公式", "math"}, // must not fall into the existing "表" route
		{"LaTeX 公式", "math"},
		{"看看代码块横向滚动", "code-scroll"},
		{"长代码", "code-scroll"},
		{"跨块引用式链接", "references"},
		{"脚注", "references"},
		{"展示图片", "images"},
		{"跨段落选择", "selection"},
		{"看看表格", "table"},
		{"用 Go 写一个并发下载器", "downloader"},
		{"全部样例", "all"},
		{"TODO 验证", "all"},
		{"验证示例", "all"},
	} {
		t.Run(tc.prompt, func(t *testing.T) {
			want, ok := sourceFor(tc.sample)
			if !ok || want == "" {
				t.Fatalf("missing sample %q", tc.sample)
			}
			if got := answerFor(tc.prompt); got != want {
				t.Fatalf("prompt %q did not route to %q", tc.prompt, tc.sample)
			}
		})
	}
	// Every catalogue button sends its title through the same prompt router.
	for _, s := range demoSamples {
		src, _ := sourceFor(s.name)
		if answerFor(s.title) != src {
			t.Errorf("catalogue title %q routed to another sample", s.title)
		}
	}
	if _, ok := sourceFor("unknown-sample"); ok {
		t.Fatal("unknown startup sample should be rejected")
	}
}

func TestAllSamplesIncludeEveryCase(t *testing.T) {
	all, _ := sourceFor("all")
	for _, s := range demoSamples {
		src, ok := sourceFor(s.name)
		if !ok || !strings.Contains(all, src) {
			t.Errorf("all is missing %q", s.name)
		}
	}
	// These syntax forms are the cases the manual acceptance checks depend on.
	for _, marker := range []string{
		"$a^2 + b^2 = c^2$", "$$\n", `\frac`, `\begin{pmatrix}`,
		"END_OF_LONG_LINE", "END_OF_SPACED_LINE", "END_OF_SECOND_BLOCK",
		"render**ing**", "[^note]", "\n\n[go-docs]:", "\n\n[^note]:",
		"![红绿蓝色块]", "![](examples/chat/assets/colors.png)", "assets/missing.png",
	} {
		if !strings.Contains(all, marker) {
			t.Errorf("all is missing acceptance case %q", marker)
		}
	}
}

func TestLocalImageFixture(t *testing.T) {
	f, err := os.Open("assets/colors.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 360 || img.Bounds().Dy() != 120 {
		t.Fatalf("image bounds %v, want 360x120", img.Bounds())
	}
	if _, err := os.Stat("assets/missing.png"); !os.IsNotExist(err) {
		t.Fatalf("missing-image fixture must remain absent: %v", err)
	}
}

func TestReferenceControlResolvesWithinItsChunk(t *testing.T) {
	src, _ := sourceFor("references")
	for _, chunk := range strings.Split(src, "\n\n") {
		if !strings.Contains(chunk, "[同块链接对照]") {
			continue
		}
		var out bytes.Buffer
		if err := goldmark.Convert([]byte(chunk), &out); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), `<a href="https://go.dev/doc/">同块链接对照</a>`) {
			t.Fatalf("the same-chunk control must resolve as a link: %s", out.String())
		}
		return
	}
	t.Fatal("missing same-chunk reference control")
}
