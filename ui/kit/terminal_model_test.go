package kit

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTerminalModelProgressCursorAndErase(t *testing.T) {
	m := newTerminalModel(20, 3, 10)
	m.write([]byte("\x1b[32mkeel\x1b[0m$ progress\r\n0%"))
	m.write([]byte("\r\x1b[2K100%\r\n"))
	if got := m.text(); got != "keel$ progress\n100%\n" {
		t.Fatal(got)
	}
	if m.screen.lines[0][0].style.fg != 3 || m.screen.lines[0][4].style.fg != 0 {
		t.Fatal("SGR leaked")
	}
	m.write([]byte("\x1b[2J\x1b[Hready"))
	if got := m.text(); got != "ready\n\n" {
		t.Fatal(got)
	}
	m.write([]byte("\x1b[2;3HX\x1b[1D!"))
	if terminalLineText(m.screen.lines[1]) != "  !" {
		t.Fatal(m.text())
	}
}
func TestTerminalModelColorsAndAttributes(t *testing.T) {
	m := newTerminalModel(20, 2, 0)
	m.write([]byte("\x1b[1;2;3;4;7;9;91;104mA\x1b[22;23;24;27;29;38;5;123;48;2;12;34;56mB\x1b[0mC"))
	a, b, c := m.screen.lines[0][0], m.screen.lines[0][1], m.screen.lines[0][2]
	if a.style.fg != 10 || a.style.bg != 13 || a.style.flags != terminalBold|terminalDim|terminalItalic|terminalUnderline|terminalInverse|terminalStrike {
		t.Fatal(a.style)
	}
	if b.style.fg != 124 || b.style.bg != 1<<24|12<<16|34<<8|56 || b.style.flags != 0 {
		t.Fatal(b.style)
	}
	if c.style != (terminalStyle{}) {
		t.Fatal(c.style)
	}
	m.write([]byte("\x1b[38;2;999;2;3mD"))
	if m.screen.lines[0][3].style.fg != 0 {
		t.Fatal("invalid color")
	}
}
func TestTerminalModelScrollbackBoundedAndRegion(t *testing.T) {
	m := newTerminalModel(8, 2, 2)
	m.write([]byte("a\r\nb\r\nc\r\nd\r\ne"))
	if m.text() != "b\nc\nd\ne" || len(m.history) != 2 {
		t.Fatal(m.text())
	}
	m.write([]byte("\x1b[3J"))
	if m.text() != "d\ne" || len(m.history) != 0 {
		t.Fatal(m.text())
	}
	m = newTerminalModel(4, 4, 4)
	m.write([]byte("A\x1b[2;1HB\x1b[3;1HC\x1b[4;1HD\x1b[2;3r\x1b[3;1H\n"))
	if m.text() != "A\nC\n\nD" || len(m.history) != 0 {
		t.Fatal(m.text())
	}
	m.write([]byte("\x1b[r\x1b[H\x1b[M"))
	if len(m.history) != 0 {
		t.Fatal("delete lines created scrollback")
	}
}
func TestTerminalModelChunkedUTF8AndGraphemes(t *testing.T) {
	for _, text := range []string{"中文", "e\u0301", "👨‍👩‍👧‍👦", "👍🏽", "🇨🇳", "1️⃣", "✈️"} {
		t.Run(text, func(t *testing.T) {
			m := newTerminalModel(30, 2, 0)
			for _, b := range []byte(text) {
				m.write([]byte{b})
			}
			if got := terminalLineText(m.screen.lines[0]); got != text {
				t.Fatal(got)
			}
			want := 2
			if text == "中文" {
				want = 4
			} else if text == "e\u0301" {
				want = 1
			}
			if m.screen.x != want {
				t.Fatal("grapheme cell width", text, m.screen.x, want)
			}
			if !utf8.ValidString(m.text()) || m.utfN != 0 {
				t.Fatal("invalid UTF8")
			}
			assertTerminalInvariant(t, m)
		})
	}
	m := newTerminalModel(10, 2, 0)
	m.write([]byte{0xe4})
	if m.version != 0 {
		t.Fatal("incomplete UTF8 painted")
	}
	m.write([]byte{0xb8, 0xad})
	if m.screen.lines[0][0].text != "中" || m.screen.lines[0][0].width != 2 {
		t.Fatal(m.text())
	}
	m.write([]byte{0xe4, 'A'})
	if !strings.Contains(m.text(), "�A") {
		t.Fatal("invalid sequence recovery", m.text())
	}
}
func TestTerminalModelWideOverwriteWrapAndResize(t *testing.T) {
	m := newTerminalModel(4, 2, 2)
	m.write([]byte("abc中"))
	if m.text() != "abc\n中" {
		t.Fatal(m.text())
	}
	m.write([]byte("\x1b[2;2HX"))
	if m.text() != "abc\n X" {
		t.Fatal("overwriting continuation", m.text())
	}
	m.write([]byte("\x1b[1;1H中語\x1b[1;2H\x1b[P"))
	assertTerminalInvariant(t, m)
	m.resize(1, 1)
	assertTerminalInvariant(t, m)
	m.resize(8, 4)
	assertTerminalInvariant(t, m)
	before := m.version
	m.resize(8, 4)
	if m.version != before {
		t.Fatal("unchanged resize redraw")
	}
	m = newTerminalModel(3, 2, 0)
	m.write([]byte("abcX"))
	if m.text() != "abc\nX" {
		t.Fatal("delayed wrap", m.text())
	}
	m = newTerminalModel(3, 2, 0)
	m.write([]byte("\x1b[?7labcd"))
	if m.text() != "abd\n" {
		t.Fatal("wrap disabled", m.text())
	}
}
func TestTerminalModelAlternateScreenAndModes(t *testing.T) {
	m := newTerminalModel(10, 3, 3)
	m.write([]byte("main\x1b[?1;25;2004h\x1b[?25l\x1b[?1049hALT"))
	if m.normal == nil || !m.applicationCursor || !m.bracketedPaste || m.cursorVisible {
		t.Fatal("modes")
	}
	if m.text() != "ALT\n\n" {
		t.Fatal(m.text())
	}
	m.write([]byte("\x1b[?1049h"))
	if m.text() != "ALT\n\n" {
		t.Fatal("repeated enter erased alternate")
	}
	m.resize(8, 2)
	m.write([]byte("\x1b[?1049l!"))
	if m.text() != "main!\n" {
		t.Fatal("restore", m.text())
	}
	assertTerminalInvariant(t, m)
}
func TestTerminalModelIgnoresBoundedControlStrings(t *testing.T) {
	m := newTerminalModel(20, 2, 0)
	m.write([]byte("a\x1b]52;c;clipboard payload\aB\x1bPignored\x1b\\C\x1b(0D"))
	if m.text() != "aBCD\n" {
		t.Fatal(m.text())
	}
	m.write([]byte("\x1b[" + strings.Repeat("1;", 10000) + "mE"))
	if terminalLineText(m.screen.lines[0]) != "aBCDE" || len(m.sequence) > 128 {
		t.Fatal("oversized CSI leaked")
	}
	m.write([]byte("\x1b]" + strings.Repeat("x", 10000)))
	if len(m.sequence) > 128 {
		t.Fatal("OSC retained")
	}
	m.write([]byte("\x1b\\F"))
	if !strings.HasSuffix(terminalLineText(m.screen.lines[0]), "EF") {
		t.Fatal("ST recovery", m.text())
	}
}
func TestTerminalModelFragmentedSequenceEquivalent(t *testing.T) {
	input := []byte("\x1b[38;2;12;34;56m中👩🏽‍💻\r\nnext\x1b[1A\r\x1b[Kdone\x1b]0;title\x1b\\")
	expected := newTerminalModel(16, 3, 8)
	expected.write(input)
	for split := 0; split <= len(input); split++ {
		m := newTerminalModel(16, 3, 8)
		m.write(input[:split])
		m.write(input[split:])
		if !reflect.DeepEqual(m, expected) {
			t.Fatal("split", split)
		}
	}
}
func assertTerminalInvariant(t *testing.T, m *terminalModel) {
	t.Helper()
	if len(m.screen.lines) != m.rows || m.screen.x < 0 || m.screen.x >= m.cols || m.screen.y < 0 || m.screen.y >= m.rows || len(m.history) > m.historyLimit || len(m.sequence) > 128 || m.utfN >= utf8.UTFMax {
		t.Fatal("bounds")
	}
	for _, line := range m.screen.lines {
		if len(line) != m.cols {
			t.Fatal("line width")
		}
		for x, c := range line {
			if !utf8.ValidString(c.text) || len(c.text) > 260 || c.width > 2 {
				t.Fatal("cell")
			}
			if c.width == 0 && (x == 0 || line[x-1].width != 2) {
				t.Fatal("orphan continuation")
			}
			if c.width == 2 && (x+1 >= len(line) || line[x+1].width != 0) {
				t.Fatal("orphan wide")
			}
		}
	}
}
func FuzzTerminalModel(f *testing.F) {
	for _, s := range []string{"hello", "\x1b[2J\x1b[H", "中👩‍💻\rprogress", "\x1b[?1049hALT\x1b[?1049l"} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 8192 {
			t.Skip()
		}
		m := newTerminalModel(13, 5, 12)
		for i, b := range data {
			m.write([]byte{b})
			if i%31 == 0 {
				m.resize(1+int(b)%20, 1+int(b)%8)
			}
			assertTerminalInvariant(t, m)
		}
	})
}
func BenchmarkTerminalModelOutput(b *testing.B) {
	payload := []byte("\x1b[32mkeel\x1b[0m: 中文 and 👩‍💻 output\r\n")
	m := newTerminalModel(100, 30, 1000)
	b.ReportAllocs()
	b.SetBytes(int64(len(payload)))
	for b.Loop() {
		m.write(payload)
	}
}

func TestTerminalModelRecyclingKeepsHistoryIndependent(t *testing.T) {
	m := newTerminalModel(8, 3, 4)
	for _, line := range []string{"00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19"} {
		m.write([]byte(line + "\r\n"))
	}
	if m.text() != "14\n15\n16\n17\n18\n19\n" {
		t.Fatal("recycled visible/history alias", m.text())
	}
	m.write([]byte("\x1b[Hchanged"))
	if terminalLineText(m.historyLine(3)) != "17" {
		t.Fatal("screen mutation changed history")
	}
	m.resize(3, 2)
	assertTerminalInvariant(t, m)
}
