package vmap

import (
	"math"
	"sort"
)

// point is a pixel coordinate in canvas space.
type point struct {
	x, y int
}

// canvas is a port of mapscii's Canvas: painting operations on top of a
// brailleBuffer. Polygon filling uses even-odd scanline filling instead
// of earcut triangulation — identical dot coverage for terminal
// resolution, without the triangulation dependency.
type canvas struct {
	width, height int
	buffer        *brailleBuffer
}

func newCanvas(width, height int) *canvas {
	return &canvas{width: width, height: height, buffer: newBrailleBuffer(width, height)}
}

func (c *canvas) frame() string { return c.buffer.frame() }

func (c *canvas) text(text string, x, y int, color uint8) {
	c.buffer.writeText(text, x, y, color, false)
}

func (c *canvas) setBackground(color uint8) {
	c.buffer.setGlobalBackground(color)
}

func (c *canvas) background(x, y int, color uint8) {
	c.buffer.setBackground(x, y, color)
}

func (c *canvas) polyline(points []point, color uint8, width int) {
	for i := 1; i < len(points); i++ {
		c.line(points[i-1].x, points[i-1].y, points[i].x, points[i].y, width, color)
	}
}

// line is a port of Canvas._line: plain Bresenham for width 1, and Alois
// Zingl's thick-line variant for wider strokes.
func (c *canvas) line(x0, y0, x1, y1, width int, color uint8) {
	w := math.Max(0, float64(width-1))
	if w == 0 {
		c.bresenham(x0, y0, x1, y1, color)
		return
	}

	dx := abs(x1 - x0)
	sx := 1
	if x0 >= x1 {
		sx = -1
	}
	dy := abs(y1 - y0)
	sy := 1
	if y0 >= y1 {
		sy = -1
	}
	err := dx - dy
	ed := 1.0
	if dx+dy != 0 {
		ed = math.Sqrt(float64(dx*dx) + float64(dy*dy))
	}
	w = (w + 1) / 2

	for {
		c.buffer.setPixel(x0, y0, color)
		e2 := err
		x2 := x0
		if 2*e2 >= -dx {
			e2 += dy
			y2 := y0
			for float64(e2) < ed*w && (y1 != y2 || dx > dy) {
				y2 += sy
				c.buffer.setPixel(x0, y2, color)
				e2 += dx
			}
			if x0 == x1 {
				break
			}
			e2 = err
			err -= dy
			x0 += sx
		}
		if 2*e2 <= dy {
			e2 = dx - e2
			for float64(e2) < ed*w && (x1 != x2 || dx < dy) {
				x2 += sx
				c.buffer.setPixel(x2, y0, color)
				e2 += dy
			}
			if y0 == y1 {
				break
			}
			err += dx
			y0 += sy
		}
	}
}

func (c *canvas) bresenham(x0, y0, x1, y1 int, color uint8) {
	dx := abs(x1 - x0)
	sx := 1
	if x0 >= x1 {
		sx = -1
	}
	dy := -abs(y1 - y0)
	sy := 1
	if y0 >= y1 {
		sy = -1
	}
	err := dx + dy
	for {
		c.buffer.setPixel(x0, y0, color)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

// polygon fills rings (outer ring plus holes) using even-odd scanline
// filling, then strokes the ring outlines so edges stay crisp — matching
// the coverage mapscii gets from its earcut triangle rasterization.
func (c *canvas) polygon(rings [][]point, color uint8) bool {
	if len(rings) == 0 || len(rings[0]) < 3 {
		return false
	}

	minY, maxY := c.height-1, 0
	type edge struct{ x0, y0, x1, y1 int }
	var edges []edge
	for _, ring := range rings {
		if len(ring) < 3 {
			continue
		}
		for i := 0; i < len(ring); i++ {
			a := ring[i]
			b := ring[(i+1)%len(ring)]
			if a.y != b.y {
				edges = append(edges, edge{a.x, a.y, b.x, b.y})
			}
			if a.y < minY {
				minY = a.y
			}
			if a.y > maxY {
				maxY = a.y
			}
		}
	}

	if minY < 0 {
		minY = 0
	}
	if maxY >= c.height {
		maxY = c.height - 1
	}

	var xs []float64
	for y := minY; y <= maxY; y++ {
		xs = xs[:0]
		fy := float64(y)
		for _, e := range edges {
			y0, y1 := float64(e.y0), float64(e.y1)
			if (y0 <= fy) == (y1 <= fy) {
				continue
			}
			t := (fy - y0) / (y1 - y0)
			xs = append(xs, float64(e.x0)+t*float64(e.x1-e.x0))
		}
		sort.Float64s(xs)
		for i := 0; i+1 < len(xs); i += 2 {
			left := int(math.Ceil(xs[i]))
			right := int(math.Floor(xs[i+1]))
			if left < 0 {
				left = 0
			}
			if right >= c.width {
				right = c.width - 1
			}
			for x := left; x <= right; x++ {
				c.buffer.setPixel(x, y, color)
			}
		}
	}

	for _, ring := range rings {
		if len(ring) < 3 {
			continue
		}
		closed := append(ring, ring[0])
		c.polyline(closed, color, 1)
	}
	return true
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
