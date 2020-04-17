package polargraph

// Reads an SVG file with path data and converts that to a series of Coordinates
// PathParser is based on the canvg javascript code from http://code.google.com/p/canvg/

import (
	"fmt"
	"os"

	"github.com/vytis/svg"
)

func convertToMM(value float64, unit string) (inMM float64) {
	switch unit {
	case "mm":
		return value
	default:
		return value * 25.4 / 96.0
	}
}

// read a file
func ParseSvgFile(fileName string) (data []Coordinate, svgWidth float64, svgHeight float64) {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	data = make([]Coordinate, 0)

	s, err := svg.ParseSvgFromReader(file, "Some", 1)

	if err != nil {
		panic(err)
	}

	c, _ := s.ParseDrawingInstructions()

	widthUnits := ""
	readWidth, werr := fmt.Sscanf(s.Width, "%f%s", &svgWidth, &widthUnits)
	if readWidth == 0 && werr != nil {
		panic(fmt.Sprint("Could not decode width:", svgWidth))
	}
	svgWidth = convertToMM(svgWidth, widthUnits)

	heightUnits := ""
	readHeight, herr := fmt.Sscanf(s.Height, "%f%s", &svgHeight, &heightUnits)
	if readHeight == 0 && herr != nil {
		panic(fmt.Sprint("Could not decode height:", svgHeight))
	}
	svgHeight = convertToMM(svgHeight, heightUnits)

	fmt.Println("W:", svgWidth, widthUnits, " H:", svgHeight, heightUnits)
	var scaleX, scaleY float64
	values, err := s.ViewBoxValues()
	if err == nil {
		width, height := convertToMM(values[2], "px"), convertToMM(values[3], "px")
		scaleX, scaleY = svgWidth/width, svgHeight/height
	} else {
		scaleX, scaleY = 1, 1
	}

	for msg := range c {
		switch msg.Kind {

		case svg.MoveInstruction:
			data = append(data, Coordinate{X: convertToMM(msg.M[0], "px") * scaleX, Y: convertToMM(msg.M[1], "px") * scaleY, PenUp: true})
		case svg.CircleInstruction:
			fmt.Println("SVG: Circle not supported")
		case svg.CurveInstruction:
			fmt.Println("SVG: Curve not supported")
		case svg.LineInstruction:
			data = append(data, Coordinate{X: convertToMM(msg.M[0], "px") * scaleX, Y: convertToMM(msg.M[1], "px") * scaleY, PenUp: false})
		case svg.HLineInstruction:
			fmt.Println("SVG: HLine not supported")
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
