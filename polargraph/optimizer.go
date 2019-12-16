package polargraph

import (
	"math"
	"fmt"
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

// Not real distance, but much faster to calculate
func (g *Glyph) SeparationFrom(other Glyph) float64 {
	return g.end().SeparationFrom(other.start())
}

func (g *Glyph) SeparationFromReversed(other Glyph) float64 {
	return g.end().SeparationFrom(other.end())
}

func (g *Glyph) DistanceTo(other Glyph) float64 {
	return g.end().DistanceTo(other.start())
}

func (g *Glyph) DistanceToReversed(other Glyph) float64 {
	return g.end().DistanceTo(other.end())
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

func (g *Glyph) Equals(other Glyph) bool {
	if len(g.Coordinates) != len(other.Coordinates) {
		return false
	}

	for i := 0; i < len(g.Coordinates); i++ {
		if g.Coordinates[i] != other.Coordinates[i] {
			return false
		}

	}

	return true
}

func (g *Glyph) Reversed() Glyph {
	reversed := make([]Coordinate, len(g.Coordinates))

	for i := 0 ; i < len(g.Coordinates); i++ {
		this := g.Coordinates[i]
		if this.PenUp {
			this.PenUp = false
		}
		opp := len(g.Coordinates)-1-i
		reversed[opp] = this
	}

	reversed[0].PenUp = true

	return Glyph{Coordinates: reversed}
}

func (g *Glyph) CanBeMergedWith(other Glyph) bool {
	return g.end().Same(other.start())
}

func (g *Glyph) MergeWith(other Glyph) Glyph {
	otherCoords := other.Coordinates
	otherCoords[0].PenUp = false
	theseCoords := make([]Coordinate, len(g.Coordinates))
	copy(theseCoords, g.Coordinates)
	coordinates := append(theseCoords, otherCoords...)
	return Glyph{ Coordinates: coordinates}
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

func OptimizeTravel(input []Coordinate) (output []Coordinate) {
	glyphs := MakeGlyphs(input)
	optimizedGlyphs := ReorderGlyphs(glyphs)
	output = MakeCoordinates(optimizedGlyphs)

	return output
}

func MakeGlyphs(coordinates []Coordinate) (glyphs []Glyph) {
	glyphs = make([]Glyph, 0)

	// First coordinate is always moving with pen up
	penUp := 0
	for i := 1; i < len(coordinates); i++ {
		coordinate := coordinates[i]

		if coordinate.PenUp {
			// Slice from previous pen up to current
			glyphCoordinates := coordinates[penUp:i]
			glyph := Glyph{Coordinates: glyphCoordinates}
			glyphs = append(glyphs, glyph)
			penUp = i
		} 
	}

	// Last one is until the end
	glyph := Glyph{Coordinates: coordinates[penUp:]}
	glyphs = append(glyphs, glyph)

	for j := 0; j < len(glyphs); j++ {
		if glyphs[j].start().PenUp == false {
			panic(fmt.Sprint("Glyph at ", j, " starts without pen up: ", glyphs[j]))
		}
	}

	return glyphs
}

func ReorderGlyphs(glyphs []Glyph) (sorted []Glyph) {
	sorted = make([]Glyph, 0)
	if len(glyphs) == 0 {
		return
	}

	penUpDistanceBefore := TotalPenUpTravelForGlyphs(glyphs)

	fmt.Println("Reordering, starting penUp distance:", penUpDistanceBefore)
	sorted = append(sorted, glyphs[0])
	glyphs = glyphs[1:]

	for (len(glyphs) > 0) {

		// Take last glyph from sorted ones
		glyph := sorted[len(sorted) - 1]
		glyph_end := glyph.end()
		
		distance, index, reversed := math.MaxFloat64, -1, false

		// Find closest glyph 
		for i := 0; i < len(glyphs); i++ {

			start := glyphs[i].start()
			end := glyphs[i].end()
			d := glyph_end.SeparationFrom(start)
			if d < distance {
				distance, index, reversed = d, i, false
			}
			r := glyph_end.SeparationFrom(end)
			if r < distance {
				distance, index, reversed = r, i, true
			}
		}
		
		closest := glyphs[index]

		var next Glyph
		// Check if we need to reverse it
		if reversed {
			next = closest.Reversed()
		} else {
			next = closest
		}

		if next.start().PenUp == false {
			if reversed {
				panic(fmt.Sprint("Reversed glyph starts without pen up"))
			} else {
				panic(fmt.Sprint("Glyph at ", index, " starts without pen up: ", next, " Sorted: ", len(sorted)))
			}
		}


		// Merge with last or just add to list
		if glyph.CanBeMergedWith(next) {
			merged := glyph.MergeWith(next)
			if merged.start().PenUp == false {
				panic(fmt.Sprint("glyph starts without pen up. glyph: ",glyph ," next: ", next, " merged: ", merged))
			}

			sorted[len(sorted) - 1] = merged
		} else {
			sorted = append(sorted, next)
		}

		// Remove found glyph
		glyphs = append(glyphs[:index], glyphs[index+1:]...)
	}

	penUpDistanceAfter := TotalPenUpTravelForGlyphs(sorted)

	fmt.Println("Done, penUp distance:", penUpDistanceAfter, "reduced to", (float64(penUpDistanceAfter) / float64(penUpDistanceBefore)) * 100, "%" )

	return sorted
}

func MakeCoordinates(glyphs []Glyph) (coordinates []Coordinate) {
	coordinates = make([]Coordinate, 0)
	if len(glyphs) == 0 {
		return
	}

	for i := 0; i < len(glyphs); i++ {
		glyph := glyphs[i]
		if glyph.start().PenUp == false {
			panic(fmt.Sprint("Glyph at ", i, " starts without pen up"))
		}
		for j := 1; j < len(glyph.Coordinates); j++ {
			if glyph.Coordinates[j].PenUp == true {

				panic(fmt.Sprint("Coord at ", j, " is pen up"))
			}
		}
		coordinates = append(coordinates, glyph.Coordinates...)
	}

	last := coordinates[len(coordinates)-1]
	last.PenUp = true
	coordinates = append(coordinates, last)

	return
}
