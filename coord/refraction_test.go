package coord

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

func TestRefraction_Zenith(t *testing.T) {
	r := Refraction(90.0, 10.0, 1013.25)
	if r != 0 {
		t.Errorf("zenith (90°): got %f, want 0", r)
	}
}

func TestRefraction_BelowHorizon(t *testing.T) {
	r := Refraction(-2.0, 10.0, 1013.25)
	if r != 0 {
		t.Errorf("below horizon (-2°): got %f, want 0", r)
	}
}

func TestRefraction_Horizon(t *testing.T) {
	// At the horizon, refraction is about 0.5°
	r := Refraction(0.0, 10.0, 1013.25)
	if r < 0.3 || r > 0.7 {
		t.Errorf("horizon refraction: got %f, want ~0.5", r)
	}
}

func TestRefraction_HighAltitude(t *testing.T) {
	// Refraction decreases with altitude
	r45 := Refraction(45.0, 10.0, 1013.25)
	r10 := Refraction(10.0, 10.0, 1013.25)
	if r45 >= r10 {
		t.Errorf("refraction should decrease with altitude: r(45)=%f r(10)=%f", r45, r10)
	}
}

func TestRefraction_Temperature(t *testing.T) {
	// Higher temperature = less refraction
	rCold := Refraction(10.0, -10.0, 1013.25)
	rHot := Refraction(10.0, 30.0, 1013.25)
	if rCold <= rHot {
		t.Errorf("cold should have more refraction: cold=%f hot=%f", rCold, rHot)
	}
}

func TestRefraction_Pressure(t *testing.T) {
	// Higher pressure = more refraction
	rLow := Refraction(10.0, 10.0, 800.0)
	rHigh := Refraction(10.0, 10.0, 1013.25)
	if rLow >= rHigh {
		t.Errorf("high pressure should have more refraction: low=%f high=%f", rLow, rHigh)
	}
}

func TestRefract_Convergence(t *testing.T) {
	// Refracted altitude should always be >= geometric altitude
	alt := Refract(10.0, 10.0, 1013.25)
	if alt < 10.0 {
		t.Errorf("refracted altitude %f < geometric 10°", alt)
	}
	if alt > 11.0 {
		t.Errorf("refracted altitude %f implausibly large", alt)
	}
}

func TestRefract_NearZenith(t *testing.T) {
	alt := Refract(89.0, 10.0, 1013.25)
	if math.Abs(alt-89.0) > 0.01 {
		t.Errorf("near zenith: refract(89) = %f, want ~89", alt)
	}
}

// goldenRefraction matches testdata/golden_refraction.json.
type goldenRefraction struct {
	Tests []struct {
		AltDeg        float64 `json:"alt_deg"`
		TempC         float64 `json:"temp_c"`
		PressureMbar  float64 `json:"pressure_mbar"`
		RefractionDeg float64 `json:"refraction_deg"`
	} `json:"tests"`
}

func TestRefraction_Golden(t *testing.T) {
	data, err := os.ReadFile("../testdata/golden_refraction.json")
	if err != nil {
		t.Fatal(err)
	}
	var golden goldenRefraction
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}

	const tol = 1e-10 // should match exactly (identical formula)
	failures := 0
	for i, tc := range golden.Tests {
		got := Refraction(tc.AltDeg, tc.TempC, tc.PressureMbar)
		diff := math.Abs(got - tc.RefractionDeg)
		if diff > tol {
			if failures < 10 {
				t.Errorf("test %d: alt=%.1f got=%.15f want=%.15f diff=%.2e",
					i, tc.AltDeg, got, tc.RefractionDeg, diff)
			}
			failures++
		}
	}
	if failures > 0 {
		t.Errorf("%d refraction failures out of %d tests", failures, len(golden.Tests))
	}
}

func BenchmarkRefraction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Refraction(30.0, 10.0, 1013.25)
	}
}

func BenchmarkRefract(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Refract(30.0, 10.0, 1013.25)
	}
}
