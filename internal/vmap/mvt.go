package vmap

import (
	"fmt"
	"math"
)

// Minimal Mapbox Vector Tile (MVT) protobuf decoder — the subset of
// vector_tile.proto v2 that mapscii consumes via @mapbox/vector-tile:
// layers with names, extents, features (type, tags, geometry commands).

type mvtLayer struct {
	name     string
	extent   float64
	keys     []string
	values   []any
	features []mvtFeature
}

type mvtFeature struct {
	geomType int // 1=Point, 2=LineString, 3=Polygon
	tags     []uint64
	geometry []uint64
}

// properties resolves a feature's tag pairs against the layer's key/value
// tables, like VectorTileFeature in @mapbox/vector-tile.
func (l *mvtLayer) properties(f *mvtFeature) map[string]any {
	props := make(map[string]any, len(f.tags)/2+1)
	for i := 0; i+1 < len(f.tags); i += 2 {
		k := f.tags[i]
		v := f.tags[i+1]
		if k < uint64(len(l.keys)) && v < uint64(len(l.values)) {
			props[l.keys[k]] = l.values[v]
		}
	}
	return props
}

// loadGeometry decodes the command stream into pixel rings, mirroring
// VectorTileFeature.loadGeometry.
func (f *mvtFeature) loadGeometry() [][]point {
	var lines [][]point
	var line []point
	x, y := 0, 0
	i := 0
	for i < len(f.geometry) {
		cmdInt := f.geometry[i]
		i++
		cmd := cmdInt & 7
		count := int(cmdInt >> 3)
		switch cmd {
		case 1: // MoveTo
			for range count {
				if len(line) > 0 {
					lines = append(lines, line)
				}
				line = nil
				x += zigzag(f.geometry[i])
				y += zigzag(f.geometry[i+1])
				i += 2
				line = append(line, point{x, y})
			}
		case 2: // LineTo
			for range count {
				x += zigzag(f.geometry[i])
				y += zigzag(f.geometry[i+1])
				i += 2
				line = append(line, point{x, y})
			}
		case 7: // ClosePath
			if len(line) > 0 {
				line = append(line, line[0])
			}
		default:
			return lines
		}
	}
	if len(line) > 0 {
		lines = append(lines, line)
	}
	return lines
}

func zigzag(v uint64) int {
	return int(v>>1) ^ -int(v&1)
}

// decodeMVT parses an uncompressed vector tile protobuf.
func decodeMVT(data []byte) (map[string]*mvtLayer, error) {
	layers := map[string]*mvtLayer{}
	r := &pbReader{buf: data}
	for !r.done() {
		field, wire, err := r.tag()
		if err != nil {
			return nil, err
		}
		if field == 3 && wire == 2 { // repeated Layer
			raw, err := r.bytes()
			if err != nil {
				return nil, err
			}
			layer, err := decodeLayer(raw)
			if err != nil {
				return nil, err
			}
			layers[layer.name] = layer
		} else if err := r.skip(wire); err != nil {
			return nil, err
		}
	}
	return layers, nil
}

func decodeLayer(data []byte) (*mvtLayer, error) {
	layer := &mvtLayer{extent: 4096}
	r := &pbReader{buf: data}
	for !r.done() {
		field, wire, err := r.tag()
		if err != nil {
			return nil, err
		}
		switch field {
		case 1: // name
			b, err := r.bytes()
			if err != nil {
				return nil, err
			}
			layer.name = string(b)
		case 2: // features
			b, err := r.bytes()
			if err != nil {
				return nil, err
			}
			f, err := decodeFeature(b)
			if err != nil {
				return nil, err
			}
			layer.features = append(layer.features, f)
		case 3: // keys
			b, err := r.bytes()
			if err != nil {
				return nil, err
			}
			layer.keys = append(layer.keys, string(b))
		case 4: // values
			b, err := r.bytes()
			if err != nil {
				return nil, err
			}
			v, err := decodeValue(b)
			if err != nil {
				return nil, err
			}
			layer.values = append(layer.values, v)
		case 5: // extent
			v, err := r.varint()
			if err != nil {
				return nil, err
			}
			layer.extent = float64(v)
		default:
			if err := r.skip(wire); err != nil {
				return nil, err
			}
		}
	}
	return layer, nil
}

func decodeFeature(data []byte) (mvtFeature, error) {
	var f mvtFeature
	r := &pbReader{buf: data}
	for !r.done() {
		field, wire, err := r.tag()
		if err != nil {
			return f, err
		}
		switch {
		case field == 2 && wire == 2: // packed tags
			vals, err := r.packedVarints()
			if err != nil {
				return f, err
			}
			f.tags = append(f.tags, vals...)
		case field == 3: // type
			v, err := r.varint()
			if err != nil {
				return f, err
			}
			if v <= 3 {
				f.geomType = int(v) // out-of-spec types stay 0 = unknown
			}
		case field == 4 && wire == 2: // packed geometry
			vals, err := r.packedVarints()
			if err != nil {
				return f, err
			}
			f.geometry = append(f.geometry, vals...)
		default:
			if err := r.skip(wire); err != nil {
				return f, err
			}
		}
	}
	return f, nil
}

// decodeValue parses a Value message; numbers become float64 so style
// filter comparisons behave like JavaScript's.
func decodeValue(data []byte) (any, error) {
	r := &pbReader{buf: data}
	var out any
	for !r.done() {
		field, wire, err := r.tag()
		if err != nil {
			return nil, err
		}
		switch field {
		case 1: // string
			b, err := r.bytes()
			if err != nil {
				return nil, err
			}
			out = string(b)
		case 2: // float
			v, err := r.fixed32()
			if err != nil {
				return nil, err
			}
			out = float64(v)
		case 3: // double
			v, err := r.fixed64()
			if err != nil {
				return nil, err
			}
			out = v
		case 4, 5: // int, uint
			v, err := r.varint()
			if err != nil {
				return nil, err
			}
			// #nosec G115 -- int64 fields carry the value's two's-complement
			// bits in the varint; the wrapping conversion is the decode.
			out = float64(int64(v))
		case 6: // sint
			v, err := r.varint()
			if err != nil {
				return nil, err
			}
			out = float64(zigzag(v))
		case 7: // bool
			v, err := r.varint()
			if err != nil {
				return nil, err
			}
			out = v != 0
		default:
			if err := r.skip(wire); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

// pbReader is a minimal protobuf wire-format reader.
type pbReader struct {
	buf []byte
	pos int
}

func (r *pbReader) done() bool { return r.pos >= len(r.buf) }

func (r *pbReader) tag() (field int, wire int, err error) {
	v, err := r.varint()
	if err != nil {
		return 0, 0, err
	}
	return int(v >> 3), int(v & 7), nil
}

func (r *pbReader) varint() (uint64, error) {
	var v uint64
	var shift uint
	for {
		if r.pos >= len(r.buf) {
			return 0, fmt.Errorf("truncated varint")
		}
		b := r.buf[r.pos]
		r.pos++
		v |= uint64(b&0x7f) << shift
		if b < 0x80 {
			return v, nil
		}
		shift += 7
		if shift > 63 {
			return 0, fmt.Errorf("varint overflow")
		}
	}
}

func (r *pbReader) bytes() ([]byte, error) {
	n, err := r.varint()
	if err != nil {
		return nil, err
	}
	if n > math.MaxInt32 {
		return nil, fmt.Errorf("bytes field too large")
	}
	end := r.pos + int(n)
	if end > len(r.buf) {
		return nil, fmt.Errorf("truncated bytes field")
	}
	b := r.buf[r.pos:end]
	r.pos = end
	return b, nil
}

func (r *pbReader) packedVarints() ([]uint64, error) {
	b, err := r.bytes()
	if err != nil {
		return nil, err
	}
	sub := &pbReader{buf: b}
	var out []uint64
	for !sub.done() {
		v, err := sub.varint()
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func (r *pbReader) fixed32() (float32, error) {
	if r.pos+4 > len(r.buf) {
		return 0, fmt.Errorf("truncated fixed32")
	}
	bits := uint32(r.buf[r.pos]) | uint32(r.buf[r.pos+1])<<8 |
		uint32(r.buf[r.pos+2])<<16 | uint32(r.buf[r.pos+3])<<24
	r.pos += 4
	return math.Float32frombits(bits), nil
}

func (r *pbReader) fixed64() (float64, error) {
	if r.pos+8 > len(r.buf) {
		return 0, fmt.Errorf("truncated fixed64")
	}
	var bits uint64
	for i := 0; i < 8; i++ {
		bits |= uint64(r.buf[r.pos+i]) << (8 * i)
	}
	r.pos += 8
	return math.Float64frombits(bits), nil
}

func (r *pbReader) skip(wire int) error {
	switch wire {
	case 0:
		_, err := r.varint()
		return err
	case 1:
		r.pos += 8
	case 2:
		_, err := r.bytes()
		return err
	case 5:
		r.pos += 4
	default:
		return fmt.Errorf("unsupported wire type %d", wire)
	}
	if r.pos > len(r.buf) {
		return fmt.Errorf("truncated field")
	}
	return nil
}
