package kit

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

func (m *terminalModel) write(data []byte) {
	for _, b := range data {
		m.byte(b)
	}
}
func (m *terminalModel) byte(b byte) {
	if m.state != 0 && (b == 0x18 || b == 0x1a) {
		m.state = 0
		m.sequence = m.sequence[:0]
		m.join = false
		return
	}
	switch m.state {
	case 'g':
		m.state = 0
		return
	case 'o', 'd': // OSC (BEL or ST), DCS/PM/APC (ST); payload never retained.
		if b == 0x18 || b == 0x1a {
			m.state = 0
			return
		}
		if b == 0x1b {
			if m.state == 'o' {
				m.state = 'O'
			} else {
				m.state = 'D'
			}
		} else if b == 7 && m.state == 'o' {
			m.state = 0
		}
		return
	case 'O', 'D':
		if b == '\\' {
			m.state = 0
		} else if b == 0x1b {
			return
		} else if m.state == 'O' {
			m.state = 'o'
			if b == 7 {
				m.state = 0
			}
		} else {
			m.state = 'd'
		}
		return
	case 'c', 'x':
		if b == 0x1b {
			m.state = 'e'
			m.sequence = m.sequence[:0]
			return
		}
		if b == 0x18 || b == 0x1a {
			m.state = 0
			m.sequence = m.sequence[:0]
			return
		}
		if b >= 0x40 && b <= 0x7e {
			if m.state == 'c' {
				m.csi(b, string(m.sequence))
			}
			m.state = 0
			m.sequence = m.sequence[:0]
			return
		}
		if b >= 0x20 && b <= 0x3f && m.state == 'c' {
			if len(m.sequence) < 128 {
				m.sequence = append(m.sequence, b)
			} else {
				m.state = 'x'
				m.sequence = m.sequence[:0]
			}
		}
		return
	case 'e':
		m.state = 0
		switch b {
		case '(', ')', '*', '+', '-', '.', '/':
			m.state = 'g'
		case '[':
			m.state = 'c'
			m.sequence = m.sequence[:0]
		case ']':
			m.state = 'o'
		case 'P', '^', '_':
			m.state = 'd'
		case '7':
			m.save()
		case '8':
			m.restore()
		case 'D':
			m.lineFeed()
		case 'E':
			m.move(0, m.screen.y)
			m.lineFeed()
		case 'M':
			if m.screen.y == m.screen.top {
				m.scroll(false, 1, false)
			} else {
				m.move(m.screen.x, m.screen.y-1)
			}
		case 'c':
			m.screen = m.newBuffer()
			m.normal = nil
			m.history = nil
			m.historyStart = 0
			m.wrap, m.cursorVisible, m.applicationCursor, m.bracketedPaste = true, true, false, false
			m.version++
		}
		return
	}
	if m.utfN > 0 && b&0xc0 != 0x80 {
		for range m.utfN {
			m.put(utf8.RuneError)
		}
		m.utfN = 0
	}
	if b >= 0x80 {
		m.utf[m.utfN] = b
		m.utfN++
		for m.utfN > 0 && utf8.FullRune(m.utf[:m.utfN]) {
			r, n := utf8.DecodeRune(m.utf[:m.utfN])
			m.put(r)
			copy(m.utf[:], m.utf[n:m.utfN])
			m.utfN -= n
		}
		return
	}
	switch b {
	case 0x1b:
		m.state = 'e'
		m.join = false
	case '\r':
		m.move(0, m.screen.y)
	case '\n', '\v', '\f':
		m.lineFeed()
	case '\b':
		m.move(m.screen.x-1, m.screen.y)
	case '\t':
		m.move(min((m.screen.x/8+1)*8, m.cols-1), m.screen.y)
	default:
		if b >= 0x20 && b != 0x7f {
			m.put(rune(b))
		}
	}
}
func (m *terminalModel) save() {
	b := &m.screen
	b.savedX, b.savedY, b.savedStyle = b.x, b.y, b.style
}
func (m *terminalModel) restore() {
	b := &m.screen
	m.move(b.savedX, b.savedY)
	b.style = b.savedStyle
}
func terminalParams(s string) []int {
	fields := strings.Split(s, ";")
	p := make([]int, len(fields))
	for i, f := range fields {
		if len(f) > 7 {
			p[i] = 1000000
			continue
		}
		n, err := strconv.Atoi(f)
		if err == nil && n >= 0 {
			p[i] = min(n, 1000000)
		}
	}
	return p
}
func (m *terminalModel) csi(final byte, raw string) {
	private := strings.HasPrefix(raw, "?")
	if private {
		raw = raw[1:]
	}
	// Ignore unsupported intermediate/colon forms as complete sequences.
	for _, r := range raw {
		if r != ';' && (r < '0' || r > '9') {
			return
		}
	}
	p := terminalParams(raw)
	arg := func(i, def int) int {
		if i >= len(p) || p[i] == 0 {
			return def
		}
		return p[i]
	}
	b := &m.screen
	if private {
		if final == 'h' || final == 'l' {
			on := final == 'h'
			for _, mode := range p {
				switch mode {
				case 1:
					m.applicationCursor = on
				case 7:
					m.wrap = on
				case 25:
					if m.cursorVisible != on {
						m.cursorVisible = on
						m.version++
					}
				case 2004:
					m.bracketedPaste = on
				case 1049:
					if on && m.normal == nil {
						saved := m.screen
						m.normal = &saved
						m.screen = m.newBuffer()
						m.join = false
						m.version++
					}
					if !on && m.normal != nil {
						m.screen = *m.normal
						m.normal = nil
						m.join = false
						m.version++
					}
				}
			}
		}
		return
	}
	switch final {
	case 'A':
		m.move(b.x, b.y-arg(0, 1))
	case 'B', 'e':
		m.move(b.x, b.y+arg(0, 1))
	case 'C', 'a':
		m.move(b.x+arg(0, 1), b.y)
	case 'D':
		m.move(b.x-arg(0, 1), b.y)
	case 'E':
		m.move(0, b.y+arg(0, 1))
	case 'F':
		m.move(0, b.y-arg(0, 1))
	case 'G', '`':
		m.move(arg(0, 1)-1, b.y)
	case 'd':
		m.move(b.x, arg(0, 1)-1)
	case 'H', 'f':
		m.move(arg(1, 1)-1, arg(0, 1)-1)
	case 's':
		m.save()
	case 'u':
		m.restore()
	case 'm':
		m.sgr(p)
	case 'K':
		m.eraseLine(p[0])
	case 'J':
		m.eraseDisplay(p[0])
	case 'S':
		m.scroll(true, arg(0, 1), true)
	case 'T':
		m.scroll(false, arg(0, 1), false)
	case 'X':
		for x := b.x; x < min(m.cols, b.x+arg(0, 1)); x++ {
			m.clearCell(b.y, x)
		}
		m.version++
		m.join = false
	case '@', 'P':
		n := min(arg(0, 1), m.cols-b.x)
		line := b.lines[b.y]
		if final == '@' {
			copy(line[b.x+n:], line[b.x:])
			for x := b.x; x < b.x+n; x++ {
				line[x] = terminalCell{style: b.style, width: 1}
			}
		} else {
			copy(line[b.x:], line[b.x+n:])
			for x := m.cols - n; x < m.cols; x++ {
				line[x] = terminalCell{style: b.style, width: 1}
			}
		}
		terminalNormalize(line)
		m.join = false
		m.version++
	case 'r':
		top, bottom := arg(0, 1)-1, arg(1, m.rows)-1
		if top >= 0 && bottom < m.rows && top < bottom {
			b.top, b.bottom = top, bottom
			m.move(0, 0)
		}
	case 'L', 'M':
		if b.y >= b.top && b.y <= b.bottom {
			top := b.top
			b.top = b.y
			m.scroll(final == 'M', arg(0, 1), false)
			b.top = top
			m.join = false
		}
	}
}
func (m *terminalModel) eraseLine(mode int) {
	b := &m.screen
	a, z := b.x, m.cols
	switch mode {
	case 1:
		a, z = 0, b.x+1
	case 2:
		a = 0
	case 0:
	default:
		return
	}
	for x := a; x < z; x++ {
		m.clearCell(b.y, x)
	}
	b.pending = false
	m.join = false
	m.version++
}
func (m *terminalModel) eraseDisplay(mode int) {
	b := &m.screen
	switch mode {
	case 0:
		m.eraseLine(0)
		for y := b.y + 1; y < m.rows; y++ {
			b.lines[y] = m.blankLine(b.style)
		}
	case 1:
		m.eraseLine(1)
		for y := 0; y < b.y; y++ {
			b.lines[y] = m.blankLine(b.style)
		}
	case 2:
		for y := range b.lines {
			b.lines[y] = m.blankLine(b.style)
		}
	case 3:
		m.history = nil
		m.historyStart = 0
	default:
		return
	}
	b.pending = false
	m.join = false
	m.version++
}
func (m *terminalModel) sgr(p []int) {
	s := &m.screen.style
	for i := 0; i < len(p); i++ {
		n := p[i]
		switch {
		case n == 0:
			*s = terminalStyle{}
		case n == 1:
			s.flags |= terminalBold
		case n == 2:
			s.flags |= terminalDim
		case n == 3:
			s.flags |= terminalItalic
		case n == 4:
			s.flags |= terminalUnderline
		case n == 7:
			s.flags |= terminalInverse
		case n == 9:
			s.flags |= terminalStrike
		case n == 22:
			s.flags &^= terminalBold | terminalDim
		case n == 23:
			s.flags &^= terminalItalic
		case n == 24:
			s.flags &^= terminalUnderline
		case n == 27:
			s.flags &^= terminalInverse
		case n == 29:
			s.flags &^= terminalStrike
		case n >= 30 && n <= 37:
			s.fg = uint32(n - 30 + 1)
		case n >= 40 && n <= 47:
			s.bg = uint32(n - 40 + 1)
		case n >= 90 && n <= 97:
			s.fg = uint32(n - 90 + 9)
		case n >= 100 && n <= 107:
			s.bg = uint32(n - 100 + 9)
		case n == 39:
			s.fg = 0
		case n == 49:
			s.bg = 0
		case n == 38 || n == 48:
			var c uint32
			valid := false
			if i+2 < len(p) && p[i+1] == 5 {
				if p[i+2] <= 255 {
					c = uint32(p[i+2] + 1)
					valid = true
				}
				i += 2
			} else if i+4 < len(p) && p[i+1] == 2 {
				if p[i+2] <= 255 && p[i+3] <= 255 && p[i+4] <= 255 {
					c = 1<<24 | uint32(p[i+2])<<16 | uint32(p[i+3])<<8 | uint32(p[i+4])
					valid = true
				}
				i += 4
			} else {
				return
			}
			if valid {
				if n == 38 {
					s.fg = c
				} else {
					s.bg = c
				}
			}
		}
	}
}
