package polargraph

import (

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
	return g.end().DistanceTo(other.start())
}

func (g *Glyph) Length() float64 {
	length := 0.0
	for i := 1; i < len(g.Coordinates); i++ {
		prev := g.Coordinates[i - 1]
		cur := g.Coordinates[i]
		length += prev.DistanceTo(cur)
	}
	
	return length
}

func TotalTravelForGlyphs(glyphs []Glyph) float64 {
	total := 0.0
	for i := 0; i < len(glyphs); i++ {
		glyph := glyphs[i]
		total += glyph.Length()
		if i > 0 {
			prev := glyphs[i-1]
			total += prev.DistanceTo(glyph)
		}
	}
	return total
}

func TotalPenUpTravelForGlyphs(glyphs []Glyph) float64 {
	total := 0.0
	for i := 1; i < len(glyphs); i++ {
		glyph := glyphs[i]
		prev := glyphs[i-1]
		total += prev.DistanceTo(glyph)
	}
	return total
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