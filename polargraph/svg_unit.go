package polargraph

import "fmt"

type SVGUnit int

const (
	Px SVGUnit = iota
	In
	Pt
	Pc
	Cm
	Mm
)

func SVGUnitFromString(value string) (SVGUnit, error) {
	switch value {
	case "":
		return Px, nil
	case "px":
		return Px, nil
	case "pt":
		return Pt, nil
	case "pc":
		return Pc, nil
	case "cm":
		return Cm, nil
	case "mm":
		return Mm, nil
	case "in":
		return In, nil
	}
	return Px, fmt.Errorf("Could not decode unit: %s", value)
}

func PxFromSVGUnit(value SVGUnit) float64 {
	switch value {
	case Px:
		return 1.0
	case Pt:
		return 96.0 / 72.0
	case Pc:
		return 96.0 / 6.0
	case Cm:
		return 96.0 / 2.54
	case Mm:
		return 96.0 / 254.0
	case In:
		return 96.0
	default:
		return 0
	}
}

func StringFromSVGUnit(value SVGUnit) string {
	switch value {
	case Px:
		return "px"
	case Pt:
		return "pt"
	case Pc:
		return "pc"
	case Cm:
		return "cm"
	case Mm:
		return "mm"
	case In:
		return "in"
	default:
		return ""
	}
}
