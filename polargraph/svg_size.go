package polargraph

import "fmt"

type SVGSize struct {
	width   SVGNumber
	height  SVGNumber
	viewBox [4]SVGNumber
}

func SVGSizeFromValues(width string, height string, viewBox []float64) (size SVGSize, err error) {
	widthNumber, err := SVGNumberFromString(width)
	if err != nil {
		return size, fmt.Errorf("Could not decode width: %s, err: %s", width, err)
	}
	size.width = widthNumber

	heightNumber, err := SVGNumberFromString(height)
	if err != nil {
		return size, fmt.Errorf("Could not decode height: %s, err: %s", height, err)
	}

	size.height = heightNumber

	if len(viewBox) < 4 {
		size.viewBox[2] = widthNumber
		size.viewBox[3] = heightNumber
	} else {
		for i := 0; i < len(viewBox) && i < 4; i++ {
			size.viewBox[i] = SVGNumber{viewBox[i], Px}
		}
	}

	return
}

func (size SVGSize) CoordinateFromM(m [2]float64, penUp bool) Coordinate {
	x, y := SVGNumber{m[0], Px}, SVGNumber{m[1], Px}
	scaleX := size.width.ValueIn(Mm) / size.viewBox[2].ValueIn(Mm)
	scaleY := size.height.ValueIn(Mm) / size.viewBox[3].ValueIn(Mm)
	return Coordinate{X: x.ValueIn(Mm) * scaleX, Y: y.ValueIn(Mm) * scaleY, PenUp: penUp}
}
