package polargraph

// Reads an SVG file with path data and converts that to a series of Coordinates
// PathParser is based on the canvg javascript code from http://code.google.com/p/canvg/

import (
	"fmt"
	"os"

	"github.com/rustyoz/svg"
)

// read a file
func ParseSvgFile(fileName string) (data []Coordinate, svgWidth float64, svgHeight float64) {
	file, err := os.Open(fileName)
	if err != nil {
		panic(fmt.Errorf("Could not open SVG at %s", fileName))
	}

	data = make([]Coordinate, 0)

	s, err := svg.ParseSvgFromReader(file, "Some", 1)

	if err != nil {
		panic(fmt.Errorf("Could not parse SVG, err: %s", err))
	}

	c, _ := s.ParseDrawingInstructions()
	viewBox, _ := s.ViewBoxValues()
	size, err := SVGSizeFromValues(s.Width, s.Height, viewBox)
	if err != nil {
		panic(fmt.Errorf("Could not decode SVGSize, err: %s", err))
	}

	svgWidth = size.width.ValueIn(Mm)
	svgHeight = size.height.ValueIn(Mm)

	fmt.Println("W:", size.width.ValueIn(Mm), "mm", " H:", size.height.ValueIn(Mm), "mm")

	for msg := range c {
		switch msg.Kind {
		case svg.MoveInstruction:
			values := [2]float64{msg.M[0], msg.M[1]}
			coordinate := size.CoordinateFromM(values, true)
			data = append(data, coordinate)
		case svg.CircleInstruction:
			fmt.Println("SVG: Circle not supported")
		case svg.CurveInstruction:
			fmt.Println("SVG: Curve not supported")
		case svg.LineInstruction:
			values := [2]float64{msg.M[0], msg.M[1]}
			coordinate := size.CoordinateFromM(values, false)
			data = append(data, coordinate)
		case svg.CloseInstruction:
			fmt.Println("SVG: Close not supported")
		case svg.PaintInstruction:
			// fmt.Println("Paint: ignoring")

		default:
			fmt.Println("Other:", msg.Kind)
			panic("A")
		}

	}

	return
}

func GenerateSvgPath(data Coordinates, plotCoords chan<- Coordinate) {
	defer close(plotCoords)

	minPoint, maxPoint := data.Extents()

	imageSize := maxPoint.Minus(minPoint)

	fmt.Println("SVG Min:", minPoint, "Max:", maxPoint)

	if imageSize.X > (Settings.DrawingSurfaceMaxX_MM-Settings.DrawingSurfaceMinX_MM) || imageSize.Y > (Settings.DrawingSurfaceMaxY_MM-Settings.DrawingSurfaceMinY_MM) {
		panic(fmt.Sprint(
			"SVG coordinates extend past drawable surface, as defined in setup. Svg size was: ",
			imageSize,
			" And settings bounds are, X: ", Settings.DrawingSurfaceMaxX_MM, " - ", Settings.DrawingSurfaceMinX_MM,
			" Y: ", Settings.DrawingSurfaceMaxY_MM, " - ", Settings.DrawingSurfaceMinY_MM))
	}

	plotCoords <- Coordinate{X: 0, Y: 0, PenUp: true}
	firstPoint := data[0]
	plotCoords <- Coordinate{X: firstPoint.X, Y: firstPoint.Y, PenUp: true}

	for index := 0; index < len(data); index++ {
		curTarget := data[index]
		plotCoords <- curTarget
	}

	plotCoords <- Coordinate{X: 0, Y: 0, PenUp: true}

}
