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

func TestDistance(t *testing.T) {
	g1_cords := make([]Coordinate, 2)
	g1_cords[0] = Coordinate{X: 0, Y: 5, PenUp: false}
	g1_cords[1] = Coordinate{X: 5, Y: 5, PenUp: false}

	g1 := Glyph{ Coordinates: g1_cords}

	g2_cords := make([]Coordinate, 2)
	g2_cords[0] = Coordinate{X: 5, Y: 2, PenUp: false}
	g2_cords[1] = Coordinate{X: 10, Y: 2, PenUp: false}

	g2 := Glyph{ Coordinates: g2_cords}

	distance := g1.DistanceTo(g2)

	if distance != 3.0 {
		t.Error("Distance is wrong:", distance)
	}
}

func TestLength(t *testing.T) {
	g1_cords := make([]Coordinate, 2)
	g1_cords[0] = Coordinate{X: 0, Y: 5, PenUp: false}
	g1_cords[1] = Coordinate{X: 5, Y: 5, PenUp: false}

	g1 := Glyph{ Coordinates: g1_cords}

	l1 := g1.Length()

	if l1 != 5.0 {
		t.Error("Length is wrong:", l1)
	}

	g2_cords := make([]Coordinate, 3)
	g2_cords[0] = Coordinate{X: 5, Y: 2, PenUp: false}
	g2_cords[1] = Coordinate{X: 10, Y: 2, PenUp: false}
	g2_cords[2] = Coordinate{X: 10, Y: 12, PenUp: false}


	g2 := Glyph{ Coordinates: g2_cords}

	l2 := g2.Length()

	if l2 != 15.0 {
		t.Error("Length is wrong:", l2)
	}
}

func TestTotalTravel(t *testing.T) {
	g1_cords := make([]Coordinate, 2)
	g1_cords[0] = Coordinate{X: 0, Y: 5, PenUp: false}
	g1_cords[1] = Coordinate{X: 5, Y: 5, PenUp: false}

	g1 := Glyph{ Coordinates: g1_cords}

	g2_cords := make([]Coordinate, 2)
	g2_cords[0] = Coordinate{X: 5, Y: 2, PenUp: false}
	g2_cords[1] = Coordinate{X: 10, Y: 2, PenUp: false}

	g2 := Glyph{ Coordinates: g2_cords}

	glyphs := make([]Glyph, 2)
	glyphs[0] = g1
	glyphs[1] = g2

	total := TotalTravelForGlyphs(glyphs)
	shouldBe := g1.Length() + g1.DistanceTo(g2) + g2.Length()
	if total != shouldBe {
		t.Error("Total travel is wrong:", total)
	}

}

func TestPenUpTravel(t *testing.T) {
	g1_cords := make([]Coordinate, 2)
	g1_cords[0] = Coordinate{X: 0, Y: 5, PenUp: false}
	g1_cords[1] = Coordinate{X: 5, Y: 5, PenUp: false}

	g1 := Glyph{ Coordinates: g1_cords}

	g2_cords := make([]Coordinate, 2)
	g2_cords[0] = Coordinate{X: 5, Y: 2, PenUp: false}
	g2_cords[1] = Coordinate{X: 10, Y: 2, PenUp: false}

	g2 := Glyph{ Coordinates: g2_cords}

	g3_cords := make([]Coordinate, 2)
	g3_cords[0] = Coordinate{X: 10, Y: 7, PenUp: false}
	g3_cords[1] = Coordinate{X: 5, Y: 7, PenUp: false}

	g3 := Glyph{ Coordinates: g3_cords}


	glyphs := make([]Glyph, 3)
	glyphs[0] = g1
	glyphs[1] = g2
	glyphs[2] = g3

	total := TotalPenUpTravelForGlyphs(glyphs)
	shouldBe := g1.DistanceTo(g2) + g2.DistanceTo(g3)
	if total != shouldBe {
		t.Error("Total travel is wrong:", total)
	}

}

