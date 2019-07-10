package polargraph

import (
	"testing"
)

func TestSomething(t *testing.T) {
	coords := make([]Coordinate, 5)

	coords[0] = Coordinate{X: 1, Y: 2, PenUp: true}
	coords[1] = Coordinate{X: 2, Y: 3, PenUp: false}
	coords[2] = Coordinate{X: 3, Y: 4, PenUp: false}
	coords[3] = Coordinate{X: 4, Y: 5, PenUp: true}
	coords[4] = Coordinate{X: 5, Y: 6, PenUp: false}

	glyphs := MakeGlyphs(coords)

	if len(glyphs) != 2 {
		t.Error("Should be 2 glyphs, found", len(glyphs))
	}

	first := glyphs[0]
	if len(first.Coordinates) != 2 {
		t.Error("Should have 2 coordinates, found", len(first.Coordinates))
	}

	second := glyphs[1]
	if len(second.Coordinates) != 1 {
		t.Error("Should have 1 coordinate, found", len(second.Coordinates))
	}
}