package polargraph

import "testing"

func TestSizeFromValues(t *testing.T) {
	num100 := SVGNumber{100, Px}
	num50 := SVGNumber{50, Px}
	zero := SVGNumber{0, Px}

	var tests = []struct {
		a       string
		width   string
		height  string
		viewBox []float64
		want    SVGSize
	}{
		{"no viewbox", "100", "100", []float64{}, SVGSize{num100, num100, [4]SVGNumber{zero, zero, num100, num100}}},
		{"viewbox", "100", "100", []float64{100, 100, 50, 50}, SVGSize{num100, num100, [4]SVGNumber{num100, num100, num50, num50}}},
	}
	for _, tt := range tests {
		testname := tt.a
		t.Run(testname, func(t *testing.T) {
			ans, err := SVGSizeFromValues(tt.width, tt.height, tt.viewBox)
			if ans != tt.want || err != nil {
				t.Errorf("got %+v, want %+v, error: %s", ans, tt.want, err)
			}
		})
	}
}

func TestSizeCoordinateFromM(t *testing.T) {
	size := SVGSize{
		SVGNumber{100, Mm},
		SVGNumber{100, Mm},
		[4]SVGNumber{
			SVGNumber{0, Px},
			SVGNumber{0, Px},
			SVGNumber{10, Px},
			SVGNumber{10, Px},
		},
	}
	var tests = []struct {
		a     string
		M     [2]float64
		penUp bool
		want  Coordinate
	}{
		{"zero", [2]float64{0, 0}, false, Coordinate{0, 0, false}},
		{"pen up true", [2]float64{0, 0}, true, Coordinate{0, 0, true}},
		{"1px is 10mm", [2]float64{1, 2}, false, Coordinate{10, 20, false}},
		{"minus coords", [2]float64{-1, 1}, false, Coordinate{-10, 10, false}},
		{"outside coords", [2]float64{50, 50}, false, Coordinate{500, 500, false}},
	}
	for _, tt := range tests {
		testname := tt.a
		t.Run(testname, func(t *testing.T) {
			ans := size.CoordinateFromM(tt.M, tt.penUp)
			if !ans.Equals(tt.want) {
				t.Errorf("got %+v, want %+v", ans, tt.want)
			}
		})
	}

}

func TestSizeFromValuesError(t *testing.T) {
	var tests = []struct {
		a       string
		width   string
		height  string
		viewBox []float64
	}{
		{"bad width", "100mx", "100", []float64{}},
		{"bad height", "100", "100mc", []float64{}},
	}
	for _, tt := range tests {
		testname := tt.a
		t.Run(testname, func(t *testing.T) {
			_, err := SVGSizeFromValues(tt.width, tt.height, tt.viewBox)
			if err == nil {
				t.Errorf("Should not be parsed %s %s", tt.width, tt.height)
			}
		})
	}
}
