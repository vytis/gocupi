package polargraph

import "testing"

func TestUnitParsing(t *testing.T) {
	var tests = []struct {
		a    string
		want SVGUnit
	}{
		{"px", Px},
		{"pt", Pt},
		{"pc", Pc},
		{"cm", Cm},
		{"mm", Mm},
		{"in", In},
	}
	for _, tt := range tests {
		t.Run(tt.a, func(t *testing.T) {
			ans, err := SVGUnitFromString(tt.a)
			if ans != tt.want || err != nil {
				t.Errorf("got %d, want %d", ans, tt.want)
			}
		})
		t.Run(tt.a+"_to_s", func(t *testing.T) {
			ans := StringFromSVGUnit(tt.want)
			if ans != tt.a {
				t.Errorf("got %s, want %s", ans, tt.a)
			}
		})
	}
	t.Run("Check empty", func(t *testing.T) {
		ans, err := SVGUnitFromString("")
		want := Px
		if ans != want || err != nil {
			t.Errorf("got %d, want %d", ans, want)
		}
	})
	t.Run("Check error", func(t *testing.T) {
		_, err := SVGUnitFromString("bad")
		if err == nil {
			t.Errorf("Expected error")
		}
	})
}

func TestUnitConversion(t *testing.T) {
	var tests = []struct {
		a    SVGUnit
		want float64
	}{
		{Px, 1.0},
		{Pt, 1.0 / 0.75},
		{Pc, 16.0},
		{Cm, 96.0 / 2.54},
		{Mm, 96.0 / 254.0},
		{In, 96.0},
	}
	for _, tt := range tests {
		testname := StringFromSVGUnit(tt.a)
		t.Run(testname, func(t *testing.T) {
			ans := PxFromSVGUnit(tt.a)
			if ans != tt.want {
				t.Errorf("got %f, want %f", ans, tt.want)
			}
		})
	}
}
