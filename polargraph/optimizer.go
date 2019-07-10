package polargraph

import (
	"math"
)

type Glyph struct {
	Coordinates []Coordinate
}

func (g *Glyph) start() Coordinate {
	return g.Coordinates[0]
}

func (g *Glyph) end() Coordinate {
	return g.Coordinates[len(g.Coordinates) - 1]
}

func (g *Glyph) DistanceTo(other Glyph) float64 {
	return math.Sqrt( math.Pow(g.end().X - other.start().X, 2) + math.Pow(g.end().Y - other.start().Y, 2))
}

func MakeGlyphs(coordinates []Coordinate) (glyphs []Glyph) {
	glyphs = make([]Glyph, 0)

	// First coordinate is always moving with pen up
	penUp := 0
	for i := 1; i < len(coordinates); i++ {
		coordinate := coordinates[i]

		if coordinate.PenUp {
			// Slice from previous pen up to current, not including both sides
			glyphCoordinates := coordinates[penUp + 1:i]
			glyph := Glyph{Coordinates: glyphCoordinates}
			glyphs = append(glyphs, glyph)
			penUp = i
		} 
	}

	// Last one is until the end
	glyph := Glyph{Coordinates: coordinates[penUp + 1:]}
	glyphs = append(glyphs, glyph)


	return glyphs
}