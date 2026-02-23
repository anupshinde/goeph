package search

import (
	"math"
	"testing"
)

// --- helpers ---

func assertDiscreteEvents(t *testing.T, got []DiscreteEvent, wantTimes []float64, wantValues []int, tol float64) {
	t.Helper()
	if len(got) != len(wantTimes) {
		t.Fatalf("got %d events, want %d", len(got), len(wantTimes))
	}
	for i := range got {
		if math.Abs(got[i].T-wantTimes[i]) > tol {
			t.Errorf("event %d: T = %g, want %g (diff %g)", i, got[i].T, wantTimes[i], got[i].T-wantTimes[i])
		}
		if got[i].NewValue != wantValues[i] {
			t.Errorf("event %d: NewValue = %d, want %d", i, got[i].NewValue, wantValues[i])
		}
	}
}

func assertExtrema(t *testing.T, got []Extremum, wantTimes []float64, wantValues []float64, tol float64) {
	t.Helper()
	if len(got) != len(wantTimes) {
		t.Fatalf("got %d extrema, want %d", len(got), len(wantTimes))
	}
	for i := range got {
		if math.Abs(got[i].T-wantTimes[i]) > tol {
			t.Errorf("extremum %d: T = %g, want %g (diff %g)", i, got[i].T, wantTimes[i], got[i].T-wantTimes[i])
		}
		if math.Abs(got[i].Value-wantValues[i]) > tol {
			t.Errorf("extremum %d: Value = %g, want %g (diff %g)", i, got[i].Value, wantValues[i], got[i].Value-wantValues[i])
		}
	}
}

// --- FindDiscrete tests ---

func TestFindDiscrete_SingleStep(t *testing.T) {
	// Step function: 0 before t=5.5, 1 after.
	f := func(t float64) int {
		if t < 5.5 {
			return 0
		}
		return 1
	}
	events, err := FindDiscrete(0, 10, 1.0, f, 0)
	if err != nil {
		t.Fatal(err)
	}
	assertDiscreteEvents(t, events, []float64{5.5}, []int{1}, 1e-6)
}

func TestFindDiscrete_MultipleSteps(t *testing.T) {
	// floor(t/3) gives transitions at 3, 6, 9.
	f := func(t float64) int {
		return int(math.Floor(t / 3.0))
	}
	events, err := FindDiscrete(0, 10, 0.5, f, 0)
	if err != nil {
		t.Fatal(err)
	}
	assertDiscreteEvents(t, events,
		[]float64{3.0, 6.0, 9.0},
		[]int{1, 2, 3},
		1e-6,
	)
}

func TestFindDiscrete_NoEvents(t *testing.T) {
	f := func(t float64) int { return 0 }
	events, err := FindDiscrete(0, 10, 1.0, f, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Errorf("got %d events, want 0", len(events))
	}
}

func TestFindDiscrete_EventNearStart(t *testing.T) {
	f := func(t float64) int {
		if t < 0.001 {
			return 0
		}
		return 1
	}
	events, err := FindDiscrete(0, 10, 0.5, f, 0)
	if err != nil {
		t.Fatal(err)
	}
	assertDiscreteEvents(t, events, []float64{0.001}, []int{1}, 1e-6)
}

func TestFindDiscrete_EventNearEnd(t *testing.T) {
	f := func(t float64) int {
		if t < 9.999 {
			return 0
		}
		return 1
	}
	events, err := FindDiscrete(0, 10, 0.5, f, 0)
	if err != nil {
		t.Fatal(err)
	}
	assertDiscreteEvents(t, events, []float64{9.999}, []int{1}, 1e-6)
}

func TestFindDiscrete_SineSign(t *testing.T) {
	// sign(sin(pi*t)) changes sign at t = 0, 1, 2, 3.
	// Over (0.01, 3.0) we expect transitions near 1, 2, 3 (but 0 is outside).
	// Actually sign changes: 0→1 transition at start is positive,
	// then negative at t=1, positive at t=2, negative at t=3.
	f := func(t float64) int {
		s := math.Sin(math.Pi * t)
		if s >= 0 {
			return 1
		}
		return 0
	}
	events, err := FindDiscrete(0.01, 2.99, 0.1, f, 0)
	if err != nil {
		t.Fatal(err)
	}
	// Expect transitions near t=1 (1→0) and t=2 (0→1).
	assertDiscreteEvents(t, events,
		[]float64{1.0, 2.0},
		[]int{0, 1},
		1e-6,
	)
}

func TestFindDiscrete_MoonPhaseAnalog(t *testing.T) {
	// Simulates moon-phase-like quarters: floor(4*t) mod 4.
	// Over [0, 2] with step 0.1, transitions at 0.25, 0.5, ..., 1.75, 2.0.
	f := func(t float64) int {
		return int(math.Floor(4.0*t)) % 4
	}
	events, err := FindDiscrete(0, 2, 0.1, f, 0)
	if err != nil {
		t.Fatal(err)
	}
	wantTimes := []float64{0.25, 0.5, 0.75, 1.0, 1.25, 1.5, 1.75, 2.0}
	wantValues := []int{1, 2, 3, 0, 1, 2, 3, 0}
	assertDiscreteEvents(t, events, wantTimes, wantValues, 1e-6)
}

func TestFindDiscrete_Precision(t *testing.T) {
	// Step at t = 100.123456789 with a tight epsilon.
	target := 100.123456789
	f := func(t float64) int {
		if t < target {
			return 0
		}
		return 1
	}
	eps := 1e-10
	events, err := FindDiscrete(100, 101, 0.1, f, eps)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	if math.Abs(events[0].T-target) > eps {
		t.Errorf("T = %.15g, want %.15g (diff %g)", events[0].T, target, events[0].T-target)
	}
}

func TestFindDiscrete_TinyRange(t *testing.T) {
	// Range smaller than stepDays — should still work with at least 2 samples.
	f := func(t float64) int {
		if t < 5.0005 {
			return 0
		}
		return 1
	}
	events, err := FindDiscrete(5.0, 5.001, 1.0, f, 0)
	if err != nil {
		t.Fatal(err)
	}
	assertDiscreteEvents(t, events, []float64{5.0005}, []int{1}, 1e-6)
}

func TestFindDiscrete_InvalidRange(t *testing.T) {
	f := func(t float64) int { return 0 }
	_, err := FindDiscrete(10, 5, 1.0, f, 0)
	if err != ErrInvalidRange {
		t.Errorf("got err = %v, want ErrInvalidRange", err)
	}
}

func TestFindDiscrete_InvalidStep(t *testing.T) {
	f := func(t float64) int { return 0 }
	_, err := FindDiscrete(0, 10, -1.0, f, 0)
	if err != ErrInvalidStep {
		t.Errorf("got err = %v, want ErrInvalidStep", err)
	}
}

// --- FindMaxima tests ---

func TestFindMaxima_Sine(t *testing.T) {
	// sin(2*pi*t) has maxima at t = 0.25, 1.25, 2.25.
	f := func(t float64) float64 {
		return math.Sin(2.0 * math.Pi * t)
	}
	maxima, err := FindMaxima(0, 3, 0.2, f, 0)
	if err != nil {
		t.Fatal(err)
	}
	assertExtrema(t, maxima,
		[]float64{0.25, 1.25, 2.25},
		[]float64{1.0, 1.0, 1.0},
		1e-6,
	)
}

func TestFindMaxima_Quadratic(t *testing.T) {
	// -(t-5)^2 + 10 has a single maximum at t=5, value=10.
	f := func(t float64) float64 {
		return -(t-5)*(t-5) + 10
	}
	maxima, err := FindMaxima(0, 10, 1.0, f, 0)
	if err != nil {
		t.Fatal(err)
	}
	assertExtrema(t, maxima, []float64{5.0}, []float64{10.0}, DefaultExtremaEpsilon)
}

func TestFindMaxima_NoMaxima(t *testing.T) {
	// Monotonically increasing — no local maxima.
	f := func(t float64) float64 { return t }
	maxima, err := FindMaxima(0, 10, 1.0, f, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(maxima) != 0 {
		t.Errorf("got %d maxima, want 0", len(maxima))
	}
}

func TestFindMaxima_NearBoundary(t *testing.T) {
	// Maximum at t=0.1, near the left boundary.
	f := func(t float64) float64 {
		return -(t-0.1)*(t-0.1) + 5
	}
	maxima, err := FindMaxima(0, 10, 1.0, f, 0)
	if err != nil {
		t.Fatal(err)
	}
	assertExtrema(t, maxima, []float64{0.1}, []float64{5.0}, 1e-5)
}

func TestFindMaxima_Precision(t *testing.T) {
	// Check that the found maximum is near the true value.
	// Golden section locates the bracket to within epsilon, but the actual
	// peak position within that bracket is limited by floating-point
	// precision of function evaluation: ~sqrt(machEps * |peak|).
	// For -(t-t0)^2 + 100 this limit is ~1.5e-7 days.
	target := 7.123456789
	f := func(t float64) float64 {
		return -(t-target)*(t-target) + 100
	}
	maxima, err := FindMaxima(0, 15, 1.0, f, 1e-9)
	if err != nil {
		t.Fatal(err)
	}
	if len(maxima) != 1 {
		t.Fatalf("got %d maxima, want 1", len(maxima))
	}
	if math.Abs(maxima[0].T-target) > 1e-7 {
		t.Errorf("T = %.15g, want %.15g (diff %g)", maxima[0].T, target, maxima[0].T-target)
	}
	// The function value should be extremely close to the true maximum.
	if math.Abs(maxima[0].Value-100.0) > 1e-13 {
		t.Errorf("Value = %.15g, want 100 (diff %g)", maxima[0].Value, maxima[0].Value-100.0)
	}
}

func TestFindMaxima_InvalidRange(t *testing.T) {
	f := func(t float64) float64 { return t }
	_, err := FindMaxima(10, 5, 1.0, f, 0)
	if err != ErrInvalidRange {
		t.Errorf("got err = %v, want ErrInvalidRange", err)
	}
}

// --- FindMinima tests ---

func TestFindMinima_Sine(t *testing.T) {
	// sin(2*pi*t) has minima at t = 0.75, 1.75, 2.75.
	f := func(t float64) float64 {
		return math.Sin(2.0 * math.Pi * t)
	}
	minima, err := FindMinima(0, 3, 0.2, f, 0)
	if err != nil {
		t.Fatal(err)
	}
	assertExtrema(t, minima,
		[]float64{0.75, 1.75, 2.75},
		[]float64{-1.0, -1.0, -1.0},
		1e-6,
	)
}

func TestFindMinima_Quadratic(t *testing.T) {
	// (t-5)^2 has a single minimum at t=5, value=0.
	f := func(t float64) float64 {
		return (t - 5) * (t - 5)
	}
	minima, err := FindMinima(0, 10, 1.0, f, 0)
	if err != nil {
		t.Fatal(err)
	}
	assertExtrema(t, minima, []float64{5.0}, []float64{0.0}, DefaultExtremaEpsilon)
}
