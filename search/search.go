// Package search provides numerical event-finding routines for time-series
// data. It implements generic search primitives that find when a discrete
// function changes value or when a continuous function reaches a local
// extremum.
//
// These routines are the foundation for almanac computations (sunrise/sunset,
// moon phases, seasons, etc.) which live in a separate almanac package.
package search

import "errors"

const (
	// DefaultDiscreteEpsilon is the default convergence threshold for
	// FindDiscrete, approximately 1 millisecond expressed in days.
	DefaultDiscreteEpsilon = 0.001 / 86400.0

	// DefaultExtremaEpsilon is the default convergence threshold for
	// FindMaxima and FindMinima, equal to 1 second expressed in days.
	DefaultExtremaEpsilon = 1.0 / 86400.0

	// invPhi is the inverse golden ratio (sqrt(5)-1)/2, used by goldenSectionMax.
	invPhi = 0.6180339887498949
)

var (
	// ErrInvalidRange is returned when startJD >= endJD.
	ErrInvalidRange = errors.New("search: startJD must be before endJD")

	// ErrInvalidStep is returned when stepDays <= 0.
	ErrInvalidStep = errors.New("search: stepDays must be positive")
)

// DiscreteEvent represents a moment when a discrete function changed value.
type DiscreteEvent struct {
	T        float64 // Julian date when the change occurred
	NewValue int     // value of the function after the change
}

// Extremum represents a local maximum or minimum of a continuous function.
type Extremum struct {
	T     float64 // Julian date of the extremum
	Value float64 // function value at the extremum
}

// FindDiscrete finds times when a discrete function of time changes value.
//
// f is evaluated at coarse intervals of stepDays across [startJD, endJD].
// Where adjacent samples differ, the bracket is refined via bisection to
// within epsilon days.
//
// stepDays must be small enough that no two transitions of f can occur within
// a single step. For example, sunrise/sunset needs stepDays ≈ 0.04 (≈1 hour)
// while moon phases can use stepDays ≈ 7.
//
// If epsilon is 0, DefaultDiscreteEpsilon (~1 ms) is used.
// Returns events sorted by time. Returns nil if no transitions are found.
func FindDiscrete(startJD, endJD, stepDays float64, f func(float64) int, epsilon float64) ([]DiscreteEvent, error) {
	if startJD >= endJD {
		return nil, ErrInvalidRange
	}
	if stepDays <= 0 {
		return nil, ErrInvalidStep
	}
	if epsilon <= 0 {
		epsilon = DefaultDiscreteEpsilon
	}

	// Coarse sampling.
	n := int((endJD-startJD)/stepDays) + 2
	if n < 2 {
		n = 2
	}
	dt := (endJD - startJD) / float64(n-1)

	ts := make([]float64, n)
	vs := make([]int, n)
	for i := 0; i < n; i++ {
		ts[i] = startJD + float64(i)*dt
		vs[i] = f(ts[i])
	}

	// Find brackets where value changes, then bisect each.
	var events []DiscreteEvent
	for i := 0; i < n-1; i++ {
		if vs[i] == vs[i+1] {
			continue
		}
		lo, hi := ts[i], ts[i+1]
		vLo, vHi := vs[i], vs[i+1]

		for hi-lo > epsilon {
			mid := (lo + hi) / 2.0
			vMid := f(mid)
			if vMid == vLo {
				lo = mid
				vLo = vMid
			} else {
				hi = mid
				vHi = vMid
			}
		}
		events = append(events, DiscreteEvent{T: hi, NewValue: vHi})
	}

	return events, nil
}

// FindMaxima finds times of local maxima of a continuous function of time.
//
// f is evaluated at coarse intervals of stepDays across [startJD, endJD].
// Peaks are detected via sign changes in the numerical first difference,
// then refined with golden section search to within epsilon days.
//
// If epsilon is 0, DefaultExtremaEpsilon (1 second) is used.
// Returns maxima sorted by time. Returns nil if no maxima are found.
func FindMaxima(startJD, endJD, stepDays float64, f func(float64) float64, epsilon float64) ([]Extremum, error) {
	if startJD >= endJD {
		return nil, ErrInvalidRange
	}
	if stepDays <= 0 {
		return nil, ErrInvalidStep
	}
	if epsilon <= 0 {
		epsilon = DefaultExtremaEpsilon
	}

	// Sample with one extra step beyond each boundary to detect boundary peaks.
	overshoot := stepDays
	sStart := startJD - overshoot
	sEnd := endJD + overshoot
	n := int((sEnd-sStart)/stepDays) + 3
	if n < 3 {
		n = 3
	}
	dt := (sEnd - sStart) / float64(n-1)

	ts := make([]float64, n)
	ys := make([]float64, n)
	for i := 0; i < n; i++ {
		ts[i] = sStart + float64(i)*dt
		ys[i] = f(ts[i])
	}

	// Detect peaks: points higher than both neighbors.
	var results []Extremum
	for i := 1; i < n-1; i++ {
		if ys[i] > ys[i-1] && ys[i] >= ys[i+1] {
			t, v := goldenSectionMax(ts[i-1], ts[i+1], f, epsilon)
			if t >= startJD && t <= endJD {
				results = append(results, Extremum{T: t, Value: v})
			}
		}
	}

	// Deduplicate maxima closer than epsilon.
	results = dedup(results, epsilon)

	return results, nil
}

// FindMinima finds times of local minima of a continuous function of time.
//
// This is equivalent to finding maxima of -f(t). See FindMaxima for details.
func FindMinima(startJD, endJD, stepDays float64, f func(float64) float64, epsilon float64) ([]Extremum, error) {
	neg := func(t float64) float64 { return -f(t) }
	results, err := FindMaxima(startJD, endJD, stepDays, neg, epsilon)
	if err != nil {
		return nil, err
	}
	for i := range results {
		results[i].Value = -results[i].Value
	}
	return results, nil
}

// goldenSectionMax finds the t in [a, b] that maximizes f(t) to within epsilon,
// using the golden section search algorithm.
func goldenSectionMax(a, b float64, f func(float64) float64, epsilon float64) (float64, float64) {
	// Golden section search for maximum.
	// Maintains bracket [a, b] with two interior probe points.
	c := b - invPhi*(b-a)
	d := a + invPhi*(b-a)
	fc := f(c)
	fd := f(d)

	for b-a > epsilon {
		if fc < fd {
			a = c
			c = d
			fc = fd
			d = a + invPhi*(b-a)
			fd = f(d)
		} else {
			b = d
			d = c
			fd = fc
			c = b - invPhi*(b-a)
			fc = f(c)
		}
	}

	// Return the better of the two probe points.
	if fc > fd {
		return c, fc
	}
	return d, fd
}

// dedup removes consecutive extrema whose times differ by less than epsilon,
// keeping the one with the larger value.
func dedup(results []Extremum, epsilon float64) []Extremum {
	if len(results) <= 1 {
		return results
	}
	out := []Extremum{results[0]}
	for i := 1; i < len(results); i++ {
		prev := &out[len(out)-1]
		if results[i].T-prev.T < epsilon {
			if results[i].Value > prev.Value {
				*prev = results[i]
			}
		} else {
			out = append(out, results[i])
		}
	}
	return out
}
