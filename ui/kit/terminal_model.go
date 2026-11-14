package kit

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dyike/keel/typesetting/segmenter"
	"golang.org/x/text/width"
)

const (
	terminalBold uint16 = 1 << iota
	terminalDim
	terminalItalic
	terminalUnderline
	terminalInverse
	terminalStrike
)

// Colors: zero means theme default, 1..256 are palette indices plus one,
// and bit24 marks a literal 24-bit RGB value.
type terminalStyle struct {
	fg, bg uint32
	flags  uint16
}
type terminalCell struct {
	text  string
	style terminalStyle
	width uint8
}
type terminalBuffer struct {
	lines                             [][]terminalCell
	x, y, savedX, savedY, top, bottom int
	style, savedStyle                 terminalStyle
	pending                           bool
}
type terminalModel struct {
	discarded                                              uint64
	cols, rows, historyLimit, historyStart                 int
	history                                                [][]terminalCell
	screen                                                 terminalBuffer
	normal                                                 *terminalBuffer
	state                                                  byte
	sequence                                               []byte
	utf                                                    [utf8.UTFMax]byte
	utfN                                                   int
	wrap, cursorVisible, applicationCursor, bracketedPaste bool
	joinX, joinY                                           int
	join                                                   bool
	version                                                uint64
}

func newTerminalModel(cols, rows, history int) *terminalModel {
	m := &terminalModel{cols: min(max(cols, 1), 512), rows: min(max(rows, 1), 256), historyLimit: min(max(history, 0), 10000), wrap: true, cursorVisible: true}
	m.screen = m.newBuffer()
	return m
}
func (m *terminalModel) newBuffer() terminalBuffer {
	b := terminalBuffer{bottom: m.rows - 1}
	b.lines = make([][]terminalCell, m.rows)
	for i := range b.lines {
		b.lines[i] = m.blankLine(terminalStyle{})
	}
	return b
}
func (m *terminalModel) blankLine(style terminalStyle) []terminalCell {
	return m.resetLine(nil, style)
}
func (m *terminalModel) resetLine(line []terminalCell, style terminalStyle) []terminalCell {
	if cap(line) < m.cols {
		line = make([]terminalCell, m.cols)
	} else {
		line = line[:m.cols]
		clear(line[:cap(line)][m.cols:])
	}
	for i := range line {
		line[i] = terminalCell{style: style, width: 1}
	}
	return line
}
func (m *terminalModel) remember(line []terminalCell) []terminalCell {
	if m.historyLimit == 0 || m.normal != nil {
		return line
	}
	if len(m.history) < m.historyLimit {
		m.history = append(m.history, line)
		return nil
	} else {
		m.discarded++
		recycled := m.history[m.historyStart]
		m.history[m.historyStart] = line
		m.historyStart = (m.historyStart + 1) % m.historyLimit
		return recycled
	}
}
func (m *terminalModel) historyLine(i int) []terminalCell {
	return m.history[(m.historyStart+i)%len(m.history)]
}
func (m *terminalModel) scroll(up bool, count int, record bool) {
	b := &m.screen
	count = min(max(count, 1), b.bottom-b.top+1)
	for range count {
		if up {
			recycled := b.lines[b.top]
			if record && b.top == 0 && b.bottom == m.rows-1 {
				recycled = m.remember(b.lines[0])
			}
			copy(b.lines[b.top:b.bottom], b.lines[b.top+1:b.bottom+1])
			b.lines[b.bottom] = m.resetLine(recycled, b.style)
		} else {
			recycled := b.lines[b.bottom]
			copy(b.lines[b.top+1:b.bottom+1], b.lines[b.top:b.bottom])
			b.lines[b.top] = m.resetLine(recycled, b.style)
		}
	}
	m.version++
}
func (m *terminalModel) lineFeed() {
	b := &m.screen
	b.pending = false
	m.join = false
	if b.y == b.bottom {
		m.scroll(true, 1, true)
	} else if b.y < m.rows-1 {
		b.y++
		m.version++
	}
}
func (m *terminalModel) move(x, y int) {
	b := &m.screen
	x, y = min(max(x, 0), m.cols-1), min(max(y, 0), m.rows-1)
	if b.x != x || b.y != y || b.pending {
		m.version++
	}
	b.x, b.y, b.pending = x, y, false
	m.join = false
}
func terminalRuneWidth(r rune) int {
	if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || r == 0x200d || r == 0xfe0f || r == 0xfe0e {
		return 0
	}
	k := width.LookupRune(r).Kind()
	if k == width.EastAsianWide || k == width.EastAsianFullwidth || r >= 0x1f000 && r <= 0x1faff {
		return 2
	}
	return 1
}
func terminalJoins(text string, r rune) bool {
	if r < 128 {
		return false
	}
	var seg segmenter.Segmenter
	seg.Init(append([]rune(text), r))
	it := seg.GraphemeIterator()
	return it.Next() && !it.Next()
}
func (m *terminalModel) clearCell(y, x int) {
	line := m.screen.lines[y]
	a, z := x, x+1
	if line[x].width == 0 && x > 0 {
		a = x - 1
	}
	if line[a].width == 2 {
		z = min(a+2, m.cols)
	}
	for i := a; i < z; i++ {
		line[i] = terminalCell{style: m.screen.style, width: 1}
	}
}
func (m *terminalModel) put(r rune) {
	b := &m.screen
	if m.join && m.joinY == b.y && m.joinX < m.cols {
		cell := &b.lines[m.joinY][m.joinX]
		if terminalJoins(cell.text, r) {
			// Bound hostile runs of combining characters without splitting ordinary
			// Unicode graphemes or retaining an unbounded scratch buffer.
			if len(cell.text) >= 256 {
				return
			}
			cell.text += string(r)
			w := int(cell.width)
			if terminalRuneWidth(r) == 2 || r == 0xfe0f || r == 0x20e3 {
				w = 2
			}
			if w > int(cell.width) && m.joinX+1 < m.cols {
				m.clearCell(b.y, m.joinX+1)
				cell = &b.lines[m.joinY][m.joinX]
				cell.width = 2
				b.lines[b.y][m.joinX+1] = terminalCell{style: cell.style, width: 0}
				b.x = min(m.joinX+2, m.cols-1)
				b.pending = m.joinX+2 >= m.cols
			}
			m.version++
			return
		}
	}
	w := max(1, terminalRuneWidth(r))
	w = min(w, m.cols)
	if b.pending && m.wrap {
		b.x = 0
		m.lineFeed()
	}
	b.pending = false
	if w == 2 && b.x == m.cols-1 {
		if m.wrap {
			b.x = 0
			m.lineFeed()
		} else {
			w = 1
		}
	}
	m.clearCell(b.y, b.x)
	if w == 2 {
		m.clearCell(b.y, b.x+1)
	}
	b.lines[b.y][b.x] = terminalCell{text: string(r), style: b.style, width: uint8(w)}
	if w == 2 {
		b.lines[b.y][b.x+1] = terminalCell{style: b.style, width: 0}
	}
	m.joinX, m.joinY, m.join = b.x, b.y, true
	end := b.x + w
	b.x = min(end, m.cols-1)
	b.pending = end >= m.cols
	m.version++
}
func terminalLineText(line []terminalCell) string {
	var out strings.Builder
	for _, c := range line {
		if c.width == 0 {
			continue
		}
		if c.text == "" {
			out.WriteByte(' ')
		} else {
			out.WriteString(c.text)
		}
	}
	return strings.TrimRight(out.String(), " ")
}
func (m *terminalModel) text() string {
	var lines []string
	if m.normal == nil {
		for i := range m.history {
			lines = append(lines, terminalLineText(m.historyLine(i)))
		}
	}
	for _, line := range m.screen.lines {
		lines = append(lines, terminalLineText(line))
	}
	return strings.Join(lines, "\n")
}
func terminalNormalize(line []terminalCell) {
	for i := range line {
		if line[i].width == 0 && (i == 0 || line[i-1].width != 2) {
			line[i] = terminalCell{width: 1}
		}
		if line[i].width == 2 && (i+1 >= len(line) || line[i+1].width != 0) {
			line[i] = terminalCell{width: 1}
		}
	}
}
func (m *terminalModel) resize(cols, rows int) {
	cols, rows = min(max(cols, 1), 512), min(max(rows, 1), 256)
	if cols == m.cols && rows == m.rows {
		return
	}
	m.cols, m.rows = cols, rows
	resize := func(b *terminalBuffer) {
		for i, line := range b.lines {
			next := m.blankLine(b.style)
			copy(next, line)
			terminalNormalize(next)
			b.lines[i] = next
		}
		remove := max(0, b.y-rows+1)
		if remove > 0 {
			if b == &m.screen {
				for _, line := range b.lines[:remove] {
					m.remember(line)
				}
			}
			b.lines = append([][]terminalCell(nil), b.lines[remove:]...)
			b.y -= remove
		}
		if len(b.lines) > rows {
			b.lines = b.lines[:rows]
		}
		for len(b.lines) < rows {
			b.lines = append(b.lines, m.blankLine(b.style))
		}
		b.x, b.y = min(b.x, cols-1), min(b.y, rows-1)
		b.top, b.bottom, b.pending = 0, rows-1, false
	}
	resize(&m.screen)
	if m.normal != nil {
		resize(m.normal)
	}
	m.join = false
	m.version++
}
