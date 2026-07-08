package vmap

// labelBuffer is a port of mapscii's LabelBuffer: tracks label areas in
// cell coordinates to avoid overlaps. mapscii uses an RBush spatial index;
// a linear scan is equivalent at the handful of labels per frame.
type labelBuffer struct {
	areas []labelArea
	seen  map[string]bool
}

type labelArea struct {
	minX, minY, maxX, maxY float64
}

// writeIfPossible mirrors LabelBuffer.writeIfPossible: (x, y) are pixel
// coordinates. Space is checked without margin; the reserved area is
// inserted with margin — exactly like the original.
func (l *labelBuffer) writeIfPossible(text string, x, y int, margin int) bool {
	// project to cell coordinates (LabelBuffer.project)
	cx := float64(x) / 2
	cy := float64(y) / 4
	if cx < 0 {
		cx = float64(int(cx) - 1)
	} else {
		cx = float64(int(cx))
	}
	if cy < 0 {
		cy = float64(int(cy) - 1)
	} else {
		cy = float64(int(cy))
	}

	if l.collides(l.calculateArea(text, cx, cy, 0)) {
		return false
	}
	l.areas = append(l.areas, l.calculateArea(text, cx, cy, float64(margin)))
	return true
}

// calculateArea mirrors LabelBuffer._calculateArea
func (l *labelBuffer) calculateArea(text string, x, y, margin float64) labelArea {
	return labelArea{
		minX: x - margin,
		minY: y - margin/2,
		maxX: x + margin + float64(len([]rune(text))),
		maxY: y + margin/2,
	}
}

// placed reports whether a label with this text has already been drawn
// this frame; markPlaced records one.
func (l *labelBuffer) placed(text string) bool { return l.seen[text] }

func (l *labelBuffer) markPlaced(text string) {
	if l.seen == nil {
		l.seen = map[string]bool{}
	}
	l.seen[text] = true
}

func (l *labelBuffer) collides(a labelArea) bool {
	for _, b := range l.areas {
		if a.minX <= b.maxX && a.maxX >= b.minX && a.minY <= b.maxY && a.maxY >= b.minY {
			return true
		}
	}
	return false
}
