package vmap

import "strings"

// brailleBuffer is a port of mapscii's BrailleBuffer: a pixel buffer where
// each terminal cell encodes a 2x4 dot grid as a braille character
// (U+2800..U+28FF), with per-cell foreground/background 256-color codes
// and a character overlay for text labels.
type brailleBuffer struct {
	width, height int // pixel dimensions (must be multiples of 2 and 4)
	cellsW        int
	pixel         []uint8
	chars         []rune
	fg            []uint8
	bg            []uint8
	globalBG      uint8 // 0 = unset, like mapscii's Buffer.alloc default
}

// brailleMap mirrors BrailleBuffer.brailleMap
var brailleDots = [4][2]uint8{
	{0x01, 0x08},
	{0x02, 0x10},
	{0x04, 0x20},
	{0x40, 0x80},
}

func newBrailleBuffer(width, height int) *brailleBuffer {
	size := width * height / 8
	return &brailleBuffer{
		width:  width,
		height: height,
		cellsW: width / 2,
		pixel:  make([]uint8, size),
		chars:  make([]rune, size),
		fg:     make([]uint8, size),
		bg:     make([]uint8, size),
	}
}

func (b *brailleBuffer) setGlobalBackground(color uint8) {
	b.globalBG = color
}

// project mirrors BrailleBuffer._project
func (b *brailleBuffer) project(x, y int) int {
	return (x >> 1) + (b.width>>1)*(y>>2)
}

func (b *brailleBuffer) inBounds(x, y int) bool {
	return x >= 0 && x < b.width && y >= 0 && y < b.height
}

func (b *brailleBuffer) setPixel(x, y int, color uint8) {
	if !b.inBounds(x, y) {
		return
	}
	idx := b.project(x, y)
	b.pixel[idx] |= brailleDots[y&3][x&1]
	b.fg[idx] = color
}

// setBackground sets the background color of the cell containing pixel (x, y).
func (b *brailleBuffer) setBackground(x, y int, color uint8) {
	if !b.inBounds(x, y) {
		return
	}
	b.bg[b.project(x, y)] = color
}

// setChar mirrors BrailleBuffer.setChar: places a text character in the
// cell containing pixel (x, y); it takes precedence over braille dots.
func (b *brailleBuffer) setChar(ch rune, x, y int, color uint8) {
	if !b.inBounds(x, y) {
		return
	}
	idx := b.project(x, y)
	b.chars[idx] = ch
	b.fg[idx] = color
}

// writeText mirrors BrailleBuffer.writeText: one character per cell,
// advancing 2 pixels per character.
func (b *brailleBuffer) writeText(text string, x, y int, color uint8, center bool) {
	runes := []rune(text)
	if center {
		x -= len(runes)/2 + 1
	}
	for i, ch := range runes {
		b.setChar(ch, x+i*2, y, color)
	}
}

// termColor mirrors BrailleBuffer._termColor, except that an explicit
// per-cell background wins over the global background (mapscii ORs the
// color codes together, which is only correct when one of them is zero —
// here rain overlays set real per-cell backgrounds).
func (b *brailleBuffer) termColor(fg, bg uint8) string {
	if bg == 0 {
		bg = b.globalBG
	}
	switch {
	case fg != 0 && bg != 0:
		return "\x1b[38;5;" + itoa(fg) + ";48;5;" + itoa(bg) + "m"
	case fg != 0:
		return "\x1b[49;38;5;" + itoa(fg) + "m"
	case bg != 0:
		return "\x1b[39;48;5;" + itoa(bg) + "m"
	default:
		return termReset
	}
}

const termReset = "\x1b[39;49m"

func itoa(v uint8) string {
	// small, allocation-free uint8 formatting
	var buf [3]byte
	i := len(buf)
	for {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
		if v == 0 {
			break
		}
	}
	return string(buf[i:])
}

// frame mirrors BrailleBuffer.frame: compose the cell grid into a string,
// emitting color codes only on change. Each row is reset and newline
// terminated so the output embeds cleanly in a larger TUI view.
func (b *brailleBuffer) frame() string {
	var out strings.Builder
	out.Grow(len(b.pixel) * 4)
	currentColor := ""

	rows := b.height / 4
	cols := b.width / 2
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			idx := y*cols + x
			color := b.termColor(b.fg[idx], b.bg[idx])
			if color != currentColor {
				currentColor = color
				out.WriteString(color)
			}
			if ch := b.chars[idx]; ch != 0 {
				out.WriteRune(ch)
			} else {
				out.WriteRune(0x2800 + rune(b.pixel[idx]))
			}
		}
		out.WriteString(termReset)
		currentColor = termReset
		out.WriteByte('\n')
	}
	return out.String()
}
