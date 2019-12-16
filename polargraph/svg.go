package polargraph

// Reads an SVG file with path data and converts that to a series of Coordinates
// PathParser is based on the canvg javascript code from http://code.google.com/p/canvg/

import (
	"fmt"
	"os"
	"github.com/vytis/svg"
)

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

	if _, werr := fmt.Sscanf(s.Width, "%fmm", &svgWidth); werr != nil {
		panic(fmt.Sprint("Could not decode width:", svgWidth))
	}
	if _, herr := fmt.Sscanf(s.Height, "%fmm", &svgHeight); herr != nil {
		panic(fmt.Sprint("Could not decode height:", svgHeight))
	}

	values, err := s.ViewBoxValues()
	if err != nil {
		panic(err)
	}

	width, height := values[2], values[3]
	scaleX, scaleY := svgWidth / width, svgHeight / height

    for msg := range c {
    	switch msg.Kind {

    	case svg.MoveInstruction:
    		data = append(data, Coordinate{X: msg.M[0] * scaleX, Y: msg.M[1] * scaleY, PenUp: true})
    	case svg.CircleInstruction:
    		fmt.Println("SVG: Circle not supported",)
    	case svg.CurveInstruction:
    		fmt.Println("SVG: Curve not supported",)
    	case svg.LineInstruction:
    		data = append(data, Coordinate{X: msg.M[0] * scaleX, Y: msg.M[1] * scaleY, PenUp: false})
    	case svg.HLineInstruction:
    		fmt.Println("SVG: HLine not supported",)
    	case svg.CloseInstruction:
    		fmt.Println("SVG: Close not supported",)
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
