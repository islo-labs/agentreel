package vt

import (
	"strconv"
	"strings"
)

// Color represents a terminal color.
type Color struct {
	R, G, B  uint8
	Default  bool // use theme default fg/bg
}

var DefaultFG = Color{Default: true}
var DefaultBG = Color{Default: true}

// Cell is a single character cell on the screen.
type Cell struct {
	Char rune
	FG   Color
	BG   Color
	Bold bool
}

// Screen is a minimal VT100 terminal emulator.
type Screen struct {
	Width, Height    int
	Cells            [][]Cell
	CursorX, CursorY int

	// Current text attributes
	fg   Color
	bg   Color
	bold bool

	// Parser state
	state    int
	escBuf   []byte
	oscBuf   []byte
}

const (
	stateGround = iota
	stateEscape
	stateCSI
	stateOSC
)

// ANSI basic colors (normal)
var ansiColors = [8]Color{
	{0x21, 0x22, 0x2c, false}, // black
	{0xff, 0x55, 0x55, false}, // red
	{0x50, 0xfa, 0x7b, false}, // green
	{0xf1, 0xfa, 0x8c, false}, // yellow
	{0xbd, 0x93, 0xf9, false}, // blue
	{0xff, 0x79, 0xc6, false}, // magenta
	{0x8b, 0xe9, 0xfd, false}, // cyan
	{0xf8, 0xf8, 0xf2, false}, // white
}

// AnsiColors returns the standard 8 ANSI colors.
func AnsiColors() [8]Color { return ansiColors }

// AnsiBrightColors returns the 8 bright ANSI colors.
func AnsiBrightColors() [8]Color { return ansiBrightColors }

// Bright variants
var ansiBrightColors = [8]Color{
	{0x62, 0x72, 0xa4, false}, // bright black
	{0xff, 0x6e, 0x6e, false}, // bright red
	{0x69, 0xff, 0x94, false}, // bright green
	{0xff, 0xff, 0xa5, false}, // bright yellow
	{0xd6, 0xac, 0xff, false}, // bright blue
	{0xff, 0x92, 0xdf, false}, // bright magenta
	{0xa4, 0xfb, 0xff, false}, // bright cyan
	{0xff, 0xff, 0xff, false}, // bright white
}

func NewScreen(width, height int) *Screen {
	s := &Screen{
		Width:  width,
		Height: height,
		fg:     DefaultFG,
		bg:     DefaultBG,
	}
	s.Cells = make([][]Cell, height)
	for y := range s.Cells {
		s.Cells[y] = make([]Cell, width)
		for x := range s.Cells[y] {
			s.Cells[y][x] = Cell{Char: ' ', FG: DefaultFG, BG: DefaultBG}
		}
	}
	return s
}

// Clone returns a deep copy of the screen state.
func (s *Screen) Clone() *Screen {
	ns := &Screen{
		Width: s.Width, Height: s.Height,
		CursorX: s.CursorX, CursorY: s.CursorY,
		fg: s.fg, bg: s.bg, bold: s.bold,
		state: stateGround,
	}
	ns.Cells = make([][]Cell, s.Height)
	for y := range ns.Cells {
		ns.Cells[y] = make([]Cell, s.Width)
		copy(ns.Cells[y], s.Cells[y])
	}
	return ns
}

// Equal returns true if the visible content matches.
func (s *Screen) Equal(other *Screen) bool {
	if s.Width != other.Width || s.Height != other.Height {
		return false
	}
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			a, b := s.Cells[y][x], other.Cells[y][x]
			if a.Char != b.Char || a.FG != b.FG || a.BG != b.BG || a.Bold != b.Bold {
				return false
			}
		}
	}
	return true
}

// Write processes terminal output data with UTF-8 support.
func (s *Screen) Write(data []byte) {
	i := 0
	for i < len(data) {
		b := data[i]
		// If in an escape sequence or it's a control/ASCII byte, process as byte
		if s.state != stateGround || b < 0x80 {
			s.processByte(b)
			i++
			continue
		}
		// UTF-8 multi-byte sequence in ground state
		r, size := decodeUTF8(data[i:])
		if size > 0 && r >= 0x20 {
			s.putChar(r)
		}
		i += size
	}
}

func decodeUTF8(data []byte) (rune, int) {
	if len(data) == 0 {
		return 0, 0
	}
	b := data[0]
	switch {
	case b < 0xC0:
		return rune(b), 1
	case b < 0xE0 && len(data) >= 2:
		return rune(b&0x1F)<<6 | rune(data[1]&0x3F), 2
	case b < 0xF0 && len(data) >= 3:
		return rune(b&0x0F)<<12 | rune(data[1]&0x3F)<<6 | rune(data[2]&0x3F), 3
	case b < 0xF8 && len(data) >= 4:
		return rune(b&0x07)<<18 | rune(data[1]&0x3F)<<12 | rune(data[2]&0x3F)<<6 | rune(data[3]&0x3F), 4
	default:
		return 0xFFFD, 1 // replacement character
	}
}

func (s *Screen) processByte(b byte) {
	switch s.state {
	case stateGround:
		switch {
		case b == 0x1b: // ESC
			s.state = stateEscape
			s.escBuf = s.escBuf[:0]
		case b == '\r':
			s.CursorX = 0
		case b == '\n':
			s.linefeed()
		case b == '\b':
			if s.CursorX > 0 {
				s.CursorX--
			}
		case b == '\t':
			s.CursorX = (s.CursorX + 8) &^ 7
			if s.CursorX >= s.Width {
				s.CursorX = s.Width - 1
			}
		case b == 0x07: // BEL — ignore
		case b >= 0x20:
			s.putChar(rune(b))
		}

	case stateEscape:
		switch b {
		case '[':
			s.state = stateCSI
			s.escBuf = s.escBuf[:0]
		case ']':
			s.state = stateOSC
			s.oscBuf = s.oscBuf[:0]
		case '(':
			// Character set designation — consume next byte
			s.state = stateGround
		default:
			s.state = stateGround
		}

	case stateCSI:
		if b >= 0x20 && b <= 0x3f {
			// Parameter or intermediate byte
			s.escBuf = append(s.escBuf, b)
		} else if b >= 0x40 && b <= 0x7e {
			// Final byte — execute
			s.executeCSI(b)
			s.state = stateGround
		} else {
			s.state = stateGround
		}

	case stateOSC:
		if b == 0x07 || b == 0x1b { // BEL or ESC terminates OSC
			s.state = stateGround
		}
		// Consume OSC content silently
	}
}

func (s *Screen) putChar(ch rune) {
	if s.CursorX >= s.Width {
		s.CursorX = 0
		s.linefeed()
	}
	if s.CursorY >= 0 && s.CursorY < s.Height && s.CursorX >= 0 && s.CursorX < s.Width {
		s.Cells[s.CursorY][s.CursorX] = Cell{
			Char: ch,
			FG:   s.fg,
			BG:   s.bg,
			Bold: s.bold,
		}
	}
	s.CursorX++
}

func (s *Screen) linefeed() {
	s.CursorY++
	if s.CursorY >= s.Height {
		s.scrollUp()
		s.CursorY = s.Height - 1
	}
}

func (s *Screen) scrollUp() {
	copy(s.Cells[0:], s.Cells[1:])
	s.Cells[s.Height-1] = make([]Cell, s.Width)
	for x := range s.Cells[s.Height-1] {
		s.Cells[s.Height-1][x] = Cell{Char: ' ', FG: DefaultFG, BG: DefaultBG}
	}
}

func (s *Screen) executeCSI(final byte) {
	params := string(s.escBuf)

	// Strip leading '?' for private mode sequences
	private := false
	if len(params) > 0 && params[0] == '?' {
		private = true
		params = params[1:]
	}

	parts := splitParams(params)

	switch final {
	case 'm': // SGR — Select Graphic Rendition
		if !private {
			s.executeSGR(parts)
		}
	case 'H', 'f': // Cursor position
		row, col := 1, 1
		if len(parts) >= 1 && parts[0] > 0 {
			row = parts[0]
		}
		if len(parts) >= 2 && parts[1] > 0 {
			col = parts[1]
		}
		s.CursorY = clamp(row-1, 0, s.Height-1)
		s.CursorX = clamp(col-1, 0, s.Width-1)
	case 'A': // Cursor up
		n := paramDefault(parts, 0, 1)
		s.CursorY = clamp(s.CursorY-n, 0, s.Height-1)
	case 'B': // Cursor down
		n := paramDefault(parts, 0, 1)
		s.CursorY = clamp(s.CursorY+n, 0, s.Height-1)
	case 'C': // Cursor forward
		n := paramDefault(parts, 0, 1)
		s.CursorX = clamp(s.CursorX+n, 0, s.Width-1)
	case 'D': // Cursor back
		n := paramDefault(parts, 0, 1)
		s.CursorX = clamp(s.CursorX-n, 0, s.Width-1)
	case 'G': // Cursor horizontal absolute
		n := paramDefault(parts, 0, 1)
		s.CursorX = clamp(n-1, 0, s.Width-1)
	case 'J': // Erase in display
		n := paramDefault(parts, 0, 0)
		s.eraseDisplay(n)
	case 'K': // Erase in line
		n := paramDefault(parts, 0, 0)
		s.eraseLine(n)
	case 'L': // Insert lines
		n := paramDefault(parts, 0, 1)
		s.insertLines(n)
	case 'M': // Delete lines
		n := paramDefault(parts, 0, 1)
		s.deleteLines(n)
	case 'P': // Delete characters
		n := paramDefault(parts, 0, 1)
		s.deleteChars(n)
	case 'd': // Cursor vertical absolute
		n := paramDefault(parts, 0, 1)
		s.CursorY = clamp(n-1, 0, s.Height-1)
	case 'r': // Set scrolling region — ignore for now
	case 'h', 'l': // Set/reset mode — ignore
	case 'n': // Device status report — ignore
	case 's': // Save cursor — ignore
	case 'u': // Restore cursor — ignore
	}
}

func (s *Screen) executeSGR(parts []int) {
	if len(parts) == 0 {
		parts = []int{0}
	}
	for i := 0; i < len(parts); i++ {
		p := parts[i]
		switch {
		case p == 0: // Reset
			s.fg = DefaultFG
			s.bg = DefaultBG
			s.bold = false
		case p == 1:
			s.bold = true
		case p == 2 || p == 22:
			s.bold = false
		case p == 7: // Reverse
			s.fg, s.bg = s.bg, s.fg
		case p == 27: // Reverse off
			// Not perfectly trackable, reset to defaults
		case p >= 30 && p <= 37:
			s.fg = ansiColors[p-30]
		case p == 38: // Extended foreground
			if i+1 < len(parts) && parts[i+1] == 5 && i+2 < len(parts) {
				s.fg = color256(parts[i+2])
				i += 2
			} else if i+1 < len(parts) && parts[i+1] == 2 && i+4 < len(parts) {
				s.fg = Color{uint8(parts[i+2]), uint8(parts[i+3]), uint8(parts[i+4]), false}
				i += 4
			}
		case p == 39:
			s.fg = DefaultFG
		case p >= 40 && p <= 47:
			s.bg = ansiColors[p-40]
		case p == 48: // Extended background
			if i+1 < len(parts) && parts[i+1] == 5 && i+2 < len(parts) {
				s.bg = color256(parts[i+2])
				i += 2
			} else if i+1 < len(parts) && parts[i+1] == 2 && i+4 < len(parts) {
				s.bg = Color{uint8(parts[i+2]), uint8(parts[i+3]), uint8(parts[i+4]), false}
				i += 4
			}
		case p == 49:
			s.bg = DefaultBG
		case p >= 90 && p <= 97:
			s.fg = ansiBrightColors[p-90]
		case p >= 100 && p <= 107:
			s.bg = ansiBrightColors[p-100]
		}
	}
}

func (s *Screen) eraseDisplay(mode int) {
	switch mode {
	case 0: // Cursor to end
		s.eraseLine(0)
		for y := s.CursorY + 1; y < s.Height; y++ {
			s.clearRow(y)
		}
	case 1: // Start to cursor
		for y := 0; y < s.CursorY; y++ {
			s.clearRow(y)
		}
		for x := 0; x <= s.CursorX && x < s.Width; x++ {
			s.Cells[s.CursorY][x] = Cell{Char: ' ', FG: DefaultFG, BG: DefaultBG}
		}
	case 2, 3: // Entire screen
		for y := 0; y < s.Height; y++ {
			s.clearRow(y)
		}
	}
}

func (s *Screen) eraseLine(mode int) {
	y := s.CursorY
	if y < 0 || y >= s.Height {
		return
	}
	switch mode {
	case 0: // Cursor to end
		for x := s.CursorX; x < s.Width; x++ {
			s.Cells[y][x] = Cell{Char: ' ', FG: DefaultFG, BG: DefaultBG}
		}
	case 1: // Start to cursor
		for x := 0; x <= s.CursorX && x < s.Width; x++ {
			s.Cells[y][x] = Cell{Char: ' ', FG: DefaultFG, BG: DefaultBG}
		}
	case 2: // Entire line
		s.clearRow(y)
	}
}

func (s *Screen) clearRow(y int) {
	for x := 0; x < s.Width; x++ {
		s.Cells[y][x] = Cell{Char: ' ', FG: DefaultFG, BG: DefaultBG}
	}
}

func (s *Screen) insertLines(n int) {
	for i := 0; i < n && s.CursorY+i < s.Height; i++ {
		copy(s.Cells[s.CursorY+1:], s.Cells[s.CursorY:s.Height-1])
		s.clearRow(s.CursorY)
	}
}

func (s *Screen) deleteLines(n int) {
	for i := 0; i < n && s.CursorY < s.Height; i++ {
		copy(s.Cells[s.CursorY:], s.Cells[s.CursorY+1:])
		s.clearRow(s.Height - 1)
	}
}

func (s *Screen) deleteChars(n int) {
	y := s.CursorY
	if y < 0 || y >= s.Height {
		return
	}
	for i := 0; i < n; i++ {
		if s.CursorX < s.Width-1 {
			copy(s.Cells[y][s.CursorX:], s.Cells[y][s.CursorX+1:])
		}
		s.Cells[y][s.Width-1] = Cell{Char: ' ', FG: DefaultFG, BG: DefaultBG}
	}
}

// Resize updates the screen dimensions.
func (s *Screen) Resize(width, height int) {
	ns := NewScreen(width, height)
	for y := 0; y < min(s.Height, height); y++ {
		for x := 0; x < min(s.Width, width); x++ {
			ns.Cells[y][x] = s.Cells[y][x]
		}
	}
	s.Width = width
	s.Height = height
	s.Cells = ns.Cells
	s.CursorX = clamp(s.CursorX, 0, width-1)
	s.CursorY = clamp(s.CursorY, 0, height-1)
}

func splitParams(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ";")
	result := make([]int, len(parts))
	for i, p := range parts {
		n, _ := strconv.Atoi(p)
		result[i] = n
	}
	return result
}

func paramDefault(parts []int, idx, def int) int {
	if idx < len(parts) && parts[idx] > 0 {
		return parts[idx]
	}
	return def
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func color256(n int) Color {
	if n < 8 {
		return ansiColors[n]
	}
	if n < 16 {
		return ansiBrightColors[n-8]
	}
	if n < 232 {
		// 216-color cube
		n -= 16
		b := n % 6
		g := (n / 6) % 6
		r := n / 36
		return Color{uint8(r * 51), uint8(g * 51), uint8(b * 51), false}
	}
	// Grayscale
	v := uint8((n-232)*10 + 8)
	return Color{v, v, v, false}
}
