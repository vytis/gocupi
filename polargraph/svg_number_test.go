package polargraph

import "testing"

func TestNumberUnitConversion(t *testing.T) {
	var tests = []struct {
		a SVGNumber
	}{
		{SVGNumber{1, Px}},
		{SVGNumber{1, Pt}},
		{SVGNumber{1, Pc}},
		{SVGNumber{1, Cm}},
		{SVGNumber{1, Mm}},
		{SVGNumber{1, In}},
	}
	for _, tt := range tests {
		testname := StringFromSVGUnit(tt.a.unit)
		t.Run(testname, func(t *testing.T) {
			ans := tt.a.ValueIn(tt.a.unit)
			if ans != tt.a.value {
				t.Errorf("got %f, want %f", ans, tt.a.value)
			}
		})
	}
}

func TestNumberParsing(t *testing.T) {
	var tests = []struct {
		a    string
		want SVGNumber
	}{
		{"1", SVGNumber{1, Px}},
		{"1.5", SVGNumber{1.5, Px}},
		{"1.5px", SVGNumber{1.5, Px}},
		{"1.5pt", SVGNumber{1.5, Pt}},
		{"1.5pc", SVGNumber{1.5, Pc}},
		{"1.5cm", SVGNumber{1.5, Cm}},
		{"1.5mm", SVGNumber{1.5, Mm}},
		{"1.5in", SVGNumber{1.5, In}},
		{"15in", SVGNumber{15, In}},
	}
	for _, tt := range tests {
		testname := tt.a
		t.Run(testname, func(t *testing.T) {
			ans, err := SVGNumberFromString(tt.a)
			if ans != tt.want || err != nil {
				t.Errorf("got %+v, want %+v, error: %s", ans, tt.want, err)
			}
		})
	}
}

func TestNumberParsingError(t *testing.T) {
	var tests = []struct {
		a string
	}{
		{"1..5mm"},
		{"-10.5px"},
		{"1.5ww"},
	}
	for _, tt := range tests {
		testname := tt.a
		t.Run(testname, func(t *testing.T) {
			_, err := SVGNumberFromString(tt.a)
			if err == nil {
				t.Errorf("Should not be parsed %s", tt.a)
			}
		})
	}
}
