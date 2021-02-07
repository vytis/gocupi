package polargraph

import (
	"fmt"
	"regexp"
	"strconv"
)

type SVGNumber struct {
	value float64
	unit  SVGUnit
}

func SVGNumberFromString(value string) (number SVGNumber, err error) {
	r := regexp.MustCompile(`^(\d+\.?\d*)(\w*)$`)
	match := r.FindStringSubmatch(value)

	if len(match) < 3 {
		err = fmt.Errorf("Could not decode number: %s", value)
		return
	}

	parsedNumber, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		err = fmt.Errorf("Could not decode number: %s [%s]", value, err)
		return
	}
	number.value = parsedNumber
	unit, unitErr := SVGUnitFromString(match[2])
	if unitErr != nil {
		err = unitErr
		return
	}
	number.unit = unit
	return
}

func (number SVGNumber) ValueIn(unit SVGUnit) float64 {
	return number.value * PxFromSVGUnit(number.unit) / PxFromSVGUnit(unit)
}
