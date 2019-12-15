package polargraph

import (
	"testing"
)

func TestMakeGlyphs(t *testing.T) {
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
	if len(first.Coordinates) != 3 {
		t.Error("Should have 3 coordinates, found", len(first.Coordinates))
	}

	second := glyphs[1]
	if len(second.Coordinates) != 2 {
		t.Error("Should have 2 coordinates, found", len(second.Coordinates))
	}
}

func TestLine(t *testing.T) {
	coords := make([]Coordinate, 2)

	coords[0] = Coordinate{X: 1, Y: 2, PenUp: true}
	coords[1] = Coordinate{X: 2, Y: 3, PenUp: false}

	glyphs := MakeGlyphs(coords)
	if len(glyphs) != 1 {
		t.Error("Should be 1 glyph, found", len(glyphs))
	}

	first := glyphs[0]
	if len(first.Coordinates) != 2 {
		t.Error("Should have 2 coordinates, found", len(first.Coordinates))
	}
}

func TestDistance(t *testing.T) {
	g1_cords := make([]Coordinate, 2)
	g1_cords[0] = Coordinate{X: 0, Y: 5, PenUp: true}
	g1_cords[1] = Coordinate{X: 5, Y: 5, PenUp: false}

	g1 := Glyph{ Coordinates: g1_cords}

	g2_cords := make([]Coordinate, 2)
	g2_cords[0] = Coordinate{X: 5, Y: 2, PenUp: true}
	g2_cords[1] = Coordinate{X: 10, Y: 2, PenUp: false}

	g2 := Glyph{ Coordinates: g2_cords}

	distance := g1.DistanceTo(g2)

	if distance != 3.0 {
		t.Error("Distance is wrong:", distance)
	}
}

func TestReversedDistance(t *testing.T) {
	g1_cords := make([]Coordinate, 2)
	g1_cords[0] = Coordinate{X: 0, Y: 5, PenUp: true}
	g1_cords[1] = Coordinate{X: 5, Y: 5, PenUp: false}

	g1 := Glyph{ Coordinates: g1_cords}

	g2_cords := make([]Coordinate, 2)
	g2_cords[0] = Coordinate{X: 5, Y: 5, PenUp: true}
	g2_cords[1] = Coordinate{X: 10, Y: 5, PenUp: false}

	g2 := Glyph{ Coordinates: g2_cords}

	distance := g1.DistanceToReversed(g2)

	if distance != 5.0 {
		t.Error("Distance is wrong:", distance)
	}
}

func TestLength(t *testing.T) {
	g1_cords := make([]Coordinate, 2)
	g1_cords[0] = Coordinate{X: 0, Y: 5, PenUp: true}
	g1_cords[1] = Coordinate{X: 5, Y: 5, PenUp: false}

	g1 := Glyph{ Coordinates: g1_cords}

	l1 := g1.Length()

	if l1 != 5.0 {
		t.Error("Length is wrong:", l1)
	}

	g2_cords := make([]Coordinate, 3)
	g2_cords[0] = Coordinate{X: 5, Y: 2, PenUp: true}
	g2_cords[1] = Coordinate{X: 10, Y: 2, PenUp: false}
	g2_cords[2] = Coordinate{X: 10, Y: 12, PenUp: false}


	g2 := Glyph{ Coordinates: g2_cords}

	l2 := g2.Length()

	if l2 != 15.0 {
		t.Error("Length is wrong:", l2)
	}
}

func TestEquals(t *testing.T) {
	g1_cords := make([]Coordinate, 2)
	g1_cords[0] = Coordinate{X: 0, Y: 5, PenUp: true}
	g1_cords[1] = Coordinate{X: 5, Y: 5, PenUp: false}

	g1 := Glyph{ Coordinates: g1_cords}

	g2_cords := make([]Coordinate, 2)
	g2_cords[0] = Coordinate{X: 5, Y: 2, PenUp: true}
	g2_cords[1] = Coordinate{X: 10, Y: 2, PenUp: false}

	g2 := Glyph{ Coordinates: g2_cords}

	g3_cords := make([]Coordinate, 2)
	g3_cords[0] = Coordinate{X: 0, Y: 5, PenUp: true}
	g3_cords[1] = Coordinate{X: 5, Y: 5, PenUp: false}

	g3 := Glyph{ Coordinates: g3_cords}

	if !g1.Equals(g1) {
		t.Error("Should be equal to self")
	}

	if g1.Equals(g2) {
		t.Error("Should not be equal")
	}

	if !g1.Equals(g3) {
		t.Error("Should be equal to other")
	}
}


func TestTotalTravel(t *testing.T) {
	g1_cords := make([]Coordinate, 2)
	g1_cords[0] = Coordinate{X: 0, Y: 5, PenUp: true}
	g1_cords[1] = Coordinate{X: 5, Y: 5, PenUp: false}

	g1 := Glyph{ Coordinates: g1_cords}

	g2_cords := make([]Coordinate, 2)
	g2_cords[0] = Coordinate{X: 5, Y: 2, PenUp: true}
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
	g1_cords[0] = Coordinate{X: 0, Y: 5, PenUp: true}
	g1_cords[1] = Coordinate{X: 5, Y: 5, PenUp: false}

	g1 := Glyph{ Coordinates: g1_cords}

	g2_cords := make([]Coordinate, 2)
	g2_cords[0] = Coordinate{X: 5, Y: 2, PenUp: true}
	g2_cords[1] = Coordinate{X: 10, Y: 2, PenUp: false}

	g2 := Glyph{ Coordinates: g2_cords}

	g3_cords := make([]Coordinate, 2)
	g3_cords[0] = Coordinate{X: 10, Y: 7, PenUp: true}
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

func TestReversed(t *testing.T) {
	g1_cords := make([]Coordinate, 3)
	g1_cords[0] = Coordinate{X: 0, Y: 5, PenUp: true}
	g1_cords[1] = Coordinate{X: 5, Y: 5, PenUp: false}
	g1_cords[2] = Coordinate{X: 5, Y: 10, PenUp: false}

	g1 := Glyph{ Coordinates: g1_cords}

	reversed := g1.Reversed()
	if !reversed.Coordinates[0].Same(g1.Coordinates[2]) {
		t.Error("Not reversed!, reversed:", reversed,"original:", g1)
	}

}

func TestReorderEmpty(t *testing.T) {
	reordered := ReorderGlyphs(make([]Glyph, 0))
	if len(reordered) > 0 {
		t.Error("Failed reordering:", reordered)
	}
}

func TestCheckMergeGlyphs(t *testing.T) {
	g1_cords := make([]Coordinate, 3)
	g1_cords[0] = Coordinate{X: 0, Y: 5, PenUp: true}
	g1_cords[1] = Coordinate{X: 5, Y: 5, PenUp: false}
	g1_cords[2] = Coordinate{X: 5, Y: 10, PenUp: false}
	g1 := Glyph{ Coordinates: g1_cords}

	g2_cords := make([]Coordinate, 2)
	g2_cords[0] = Coordinate{X: 5, Y: 10, PenUp: true}
	g2_cords[1] = Coordinate{X: 10, Y: 2, PenUp: false}

	g2 := Glyph{ Coordinates: g2_cords}

	if g1.CanBeMergedWith(g2) == false {
		t.Error("Can be merged:", g1, g2)
	}

	if g2.CanBeMergedWith(g1) {
		t.Error("Cannot be merged:", g2, g1)
	}
}

func TestMergeGlyphs(t *testing.T) {
	g1_cords := make([]Coordinate, 3)
	g1_cords[0] = Coordinate{X: 0, Y: 5, PenUp: true}
	g1_cords[1] = Coordinate{X: 5, Y: 5, PenUp: false}
	g1_cords[2] = Coordinate{X: 5, Y: 10, PenUp: false}
	g1 := Glyph{ Coordinates: g1_cords}

	g2_cords := make([]Coordinate, 2)
	g2_cords[0] = Coordinate{X: 5, Y: 10, PenUp: true}
	g2_cords[1] = Coordinate{X: 10, Y: 2, PenUp: false}

	g2 := Glyph{ Coordinates: g2_cords}

	merged := g1.MergeWith(g2)

	all_coords := append(g1_cords, g2_cords...)
	shouldBe := Glyph{Coordinates: all_coords}

	if !merged.Equals(shouldBe) {
		t.Error("Not merged correctly", g1, g2)
	}
}

func TestReorderOne(t *testing.T) {
	g1_cords := make([]Coordinate, 3)
	g1_cords[0] = Coordinate{X: 1, Y: 11, PenUp: true}
	g1_cords[1] = Coordinate{X: 3, Y: 11, PenUp: false}
	g1_cords[2] = Coordinate{X: 4, Y: 10, PenUp: false}

	g1 := Glyph{ Coordinates: g1_cords}

	glyphs := make([]Glyph, 1)
	glyphs[0] = g1

	reordered := ReorderGlyphs(glyphs)

	if len(reordered) !=1 {
		t.Error("Failed reordering:", reordered)
	}
}

func TestReorder(t *testing.T) {
	g1_cords := make([]Coordinate, 3)
	g1_cords[0] = Coordinate{X: 1, Y: 11, PenUp: true}
	g1_cords[1] = Coordinate{X: 3, Y: 11, PenUp: false}
	g1_cords[2] = Coordinate{X: 4, Y: 10, PenUp: false}

	g1 := Glyph{ Coordinates: g1_cords}

	g2_cords := make([]Coordinate, 2)
	g2_cords[0] = Coordinate{X: 3, Y: 7, PenUp: true}
	g2_cords[1] = Coordinate{X: 6, Y: 7, PenUp: false}

	g2 := Glyph{ Coordinates: g2_cords}

	g3_cords := make([]Coordinate, 3)
	g3_cords[0] = Coordinate{X: 4, Y: 9, PenUp: true}
	g3_cords[1] = Coordinate{X: 7, Y: 9, PenUp: false}
	g3_cords[2] = Coordinate{X: 7, Y: 7, PenUp: false}

	g3 := Glyph{ Coordinates: g3_cords}


	glyphs := make([]Glyph, 3)
	glyphs[0] = g1
	glyphs[1] = g2
	glyphs[2] = g3


	reordered := ReorderGlyphs(glyphs)

	if len(reordered) != 3 {
		t.Error("Wrong glyph count!")
	}

	if !reordered[0].Equals(g1) {
		t.Error("First glyph wrong:", reordered[0])
	}

	if !reordered[1].Equals(g3) {
		t.Error("Second glyph wrong:", reordered[1])
	}

	if !reordered[2].Equals(g2.Reversed()) {
		t.Error("Third glyph wrong:", reordered[2])
	}

}

func TestReorderMerge(t *testing.T) {
	g1_cords := make([]Coordinate, 3)
	g1_cords[0] = Coordinate{X: 1, Y: 11, PenUp: true}
	g1_cords[1] = Coordinate{X: 3, Y: 11, PenUp: false}
	g1_cords[2] = Coordinate{X: 4, Y: 10, PenUp: false}

	g1 := Glyph{ Coordinates: g1_cords}

	g2_cords := make([]Coordinate, 2)
	g2_cords[0] = Coordinate{X: 3, Y: 7, PenUp: true}
	g2_cords[1] = Coordinate{X: 7, Y: 7, PenUp: false}

	g2 := Glyph{ Coordinates: g2_cords}

	g3_cords := make([]Coordinate, 3)
	g3_cords[0] = Coordinate{X: 4, Y: 9, PenUp: true}
	g3_cords[1] = Coordinate{X: 7, Y: 9, PenUp: false}
	g3_cords[2] = Coordinate{X: 7, Y: 7, PenUp: false}

	g3 := Glyph{ Coordinates: g3_cords}

	g4_cords := make([]Coordinate, 2)
	g4_cords[0] = Coordinate{X: 3, Y: 7, PenUp: true}
	g4_cords[1] = Coordinate{X: 3, Y: 4, PenUp: false}

	g4 := Glyph{ Coordinates: g4_cords}

	glyphs := make([]Glyph, 4)
	glyphs[0] = g1
	glyphs[1] = g2
	glyphs[2] = g3
	glyphs[3] = g4


	reordered := ReorderGlyphs(glyphs)

	if len(reordered) != 2 {
		t.Error("Wrong glyph count! should be 2, but have", len(reordered))
	}

	if !reordered[0].Equals(g1) {
		t.Error("First glyph wrong:", reordered[0])
	}

	merged := g3.MergeWith(g2.Reversed())
	final := merged.MergeWith(g4)

	if !reordered[1].Equals(final) {
		t.Error("Second glyph wrong:", reordered[1])
	}
}

func TestMakeCoordinates(t *testing.T) {
	g1_cords := make([]Coordinate, 2)
	g1_cords[0] = Coordinate{X: 0, Y: 5, PenUp: true}
	g1_cords[1] = Coordinate{X: 5, Y: 5, PenUp: false}

	g1 := Glyph{ Coordinates: g1_cords}

	g2_cords := make([]Coordinate, 2)
	g2_cords[0] = Coordinate{X: 5, Y: 2, PenUp: true}
	g2_cords[1] = Coordinate{X: 10, Y: 2, PenUp: false}

	g2 := Glyph{ Coordinates: g2_cords}

	glyphs := make([]Glyph, 2)
	glyphs[0] = g1
	glyphs[1] = g2

	coordinates := MakeCoordinates(glyphs)

	shouldBe := make([]Coordinate, 5)
	// First glyph
	shouldBe[0] = Coordinate{X: 0, Y: 5, PenUp: true}
	shouldBe[1] = Coordinate{X: 5, Y: 5, PenUp: false}

	// Second glyph
	shouldBe[2] = Coordinate{X: 5, Y: 2, PenUp: true}
	shouldBe[3] = Coordinate{X: 10, Y: 2, PenUp: false}

	// Pen up, we are done
	shouldBe[4] = Coordinate{X: 10, Y: 2, PenUp: true}

	if len(coordinates) != len(shouldBe) {
		t.Error("Wrong number of coordinates, should be", len(shouldBe), "is", len(coordinates))
		return
	}

	for i := 0; i < len(coordinates); i++ {
		if !coordinates[i].Equals(shouldBe[i]) {
			t.Error("Wrong coordinates at index", i, coordinates[i], "should be", shouldBe[i])
		}
	}
}

func TestMultipleLines(t *testing.T) {

}
