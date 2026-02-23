package magnitude

import (
	"math"
	"testing"
)

func TestMercury(t *testing.T) {
	// Mallama & Hilton test case: phase_angle=1.1677°, r=0.3103 AU, delta=1.3218 AU
	mag := PlanetaryMagnitude(199, 1.1677, 0.3103, 1.3218)
	if math.Abs(mag-(-2.477)) > 0.01 {
		t.Errorf("Mercury magnitude = %f, want ~-2.477", mag)
	}
}

func TestVenus(t *testing.T) {
	// Venus at greatest brilliancy: phase ~124°, r~0.72, delta~0.38
	mag := PlanetaryMagnitude(299, 124.1, 0.7215, 0.3776)
	if math.Abs(mag-(-4.916)) > 0.05 {
		t.Errorf("Venus magnitude = %f, want ~-4.916", mag)
	}
}

func TestJupiter(t *testing.T) {
	// Jupiter at opposition: phase ~0°, r~5.2 AU, delta~4.2 AU
	mag := PlanetaryMagnitude(5, 0.5, 5.2, 4.2)
	// Expected: ~-2.7 to -2.5
	if mag > -2.0 || mag < -3.0 {
		t.Errorf("Jupiter magnitude = %f, want ~-2.5", mag)
	}
}

func TestSaturn(t *testing.T) {
	// Saturn with rings: small phase angle
	mag := saturn(9.015, 8.032, 0.1055, -26.2, -26.2, true)
	if math.Abs(mag-(-0.552)) > 0.1 {
		t.Errorf("Saturn magnitude = %f, want ~-0.552", mag)
	}
}

func TestNeptune(t *testing.T) {
	// Neptune: phase ~0°, r~30 AU, delta~29 AU, year=2009.63
	mag := neptune(30.028, 29.016, 0.0381, 2009.63)
	if math.Abs(mag-7.701) > 0.05 {
		t.Errorf("Neptune magnitude = %f, want ~7.701", mag)
	}
}

func TestNormalizeBodyID(t *testing.T) {
	tests := []struct {
		input, want int
	}{
		{1, 1}, {199, 1},
		{2, 2}, {299, 2},
		{4, 4}, {499, 4},
		{5, 5}, {599, 5},
		{6, 6}, {699, 6},
		{7, 7}, {799, 7},
		{8, 8}, {899, 8},
		{999, 0},
	}
	for _, tc := range tests {
		got := normalizeBodyID(tc.input)
		if got != tc.want {
			t.Errorf("normalizeBodyID(%d) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestUnsupportedBody(t *testing.T) {
	mag := PlanetaryMagnitude(999, 10, 1, 1)
	if !math.IsNaN(mag) {
		t.Errorf("unsupported body should return NaN, got %f", mag)
	}
}
