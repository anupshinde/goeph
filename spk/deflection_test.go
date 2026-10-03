package spk

import (
	"math"
	"testing"

	"github.com/anupshinde/goeph/coord"
)

const (
	auKmTest     = 1.495978707e8
	sunRadKmTest = 695700.0
	gsTest       = 1.32712440017987e20 // GM_sun, m^3/s^2
	cMPerSecTest = 299792458.0
)

// sepArcsec returns the angle between two vectors in arcseconds, via the
// chord between unit vectors: unlike acos this keeps full precision for the
// very small angles deflection produces.
func sepArcsec(a, b [3]float64) float64 {
	na, nb := length3(a), length3(b)
	var c float64
	for i := 0; i < 3; i++ {
		d := a[i]/na - b[i]/nb
		c += d * d
	}
	return 2 * math.Asin(math.Min(1, math.Sqrt(c)/2)) * 180 / math.Pi * 3600
}

// aberratedPosition returns the astrometric position carrying aberration but
// no deflection: Apparent without its first step.
func aberratedPosition(eph *SPK, body int, tdbJD float64) [3]float64 {
	astrometric, lightTime := eph.observe(Earth, body, tdbJD)
	return coord.Aberration(astrometric, eph.EarthVelocity(tdbJD), lightTime)
}

// deflectionOnly isolates the deflection a call to Apparent applied, by
// comparing it with the same position carrying aberration alone.
func deflectionOnly(t *testing.T, eph *SPK, body int, tdbJD float64) (arcsec float64, aberrated, apparent [3]float64) {
	t.Helper()
	aberrated = aberratedPosition(eph, body, tdbJD)
	apparent = eph.Apparent(body, tdbJD)
	return sepArcsec(apparent, aberrated), aberrated, apparent
}

// Near solar conjunction the ray grazes the Sun and the deflection is at its
// largest. For a ray passing the Sun at impact parameter b, with the Sun a
// distance D_LS from the target and D_S from the observer along the line of
// sight, weak-field lensing bends it by
//
//	δ = 4·GM/(c²·b) · D_LS/D_S
//
// which is independent of the expression under test. The three dates below
// have the planet within a couple of degrees of the Sun, the ray passing
// 2 to 6 solar radii out.
func TestApparentDeflectionAtSolarConjunction(t *testing.T) {
	eph := openEph(t)
	for _, c := range []struct {
		name string
		body int
		jd   float64
	}{
		{"Mars", MarsBarycenter, 2460265.0},        // 2023-11-16, b ≈ 2.0 R_sun
		{"Jupiter", JupiterBarycenter, 2460449.25}, // 2024-05-18, b ≈ 2.8 R_sun
		{"Saturn", SaturnBarycenter, 2460369.5},    // 2024-02-28, b ≈ 6.0 R_sun
	} {
		got, aberrated, apparent := deflectionOnly(t, eph, c.body, c.jd)

		astrometric, _ := eph.observe(Earth, c.body, c.jd)
		sun, _ := eph.observe(Earth, Sun, c.jd)
		chi := sepArcsec(astrometric, sun) / 3600 * math.Pi / 180
		eDist := length3(sun)                   // observer to Sun
		impact := eDist * math.Sin(chi)         // b
		target := length3(astrometric)          // D_S
		lens := length3(sub3(astrometric, sun)) // D_LS

		if r := impact / sunRadKmTest; r < 1.5 || r > 7 {
			t.Fatalf("%s: impact parameter %.2f solar radii, want the ray just outside the limb", c.name, r)
		}
		want := 4 * gsTest / (cMPerSecTest * cMPerSecTest * impact * 1000) * (lens / target) * 180 / math.Pi * 3600
		if math.Abs(got-want) > 2e-3*want {
			t.Errorf("%s: deflection %.6f″, want %.6f″ (%+.1e)", c.name, got, want, got/want-1)
		}

		// Light bends towards the Sun, so the planet appears further from it:
		// measured against the same aberrated Sun, the whole of the deflection
		// shows up as added elongation.
		abSun := aberratedPosition(eph, Sun, c.jd)
		before, after := sepArcsec(aberrated, abSun), sepArcsec(apparent, abSun)
		if after-before < 0.99*got {
			t.Errorf("%s: elongation %.4f″ → %.4f″ (%+.6f″ of %.6f″), want the deflection to push it outwards",
				c.name, before, after, after-before, got)
		}
	}
}

// A deflector behind the observer, nearer than the target, must bend the
// incoming ray by almost nothing: the ray travels away from it the whole way.
// On 2003-08-04 Jupiter sits roughly opposite Neptune as seen from Earth, and
// the Sun is behind the observer for much of a 48-hour sweep.
func TestApparentDeflectorBehindObserver(t *testing.T) {
	eph := openEph(t)
	const (
		start = 2452855.5 // 2003-08-04 TDB
		step  = 1.0 / 144 // 10 minutes
	)
	var worst, worstStep, prev float64
	for i := 0; i <= 288; i++ {
		got, _, _ := deflectionOnly(t, eph, NeptuneBarycenter, start+float64(i)*step)
		if got > worst {
			worst = got
		}
		if i > 0 {
			if d := math.Abs(got - prev); d > worstStep {
				worstStep = d
			}
		}
		prev = got
	}
	// Neptune sits at a wide elongation throughout, so the whole correction is
	// a few tens of microarcseconds and changes smoothly.
	if worst > 1e-3 {
		t.Errorf("worst deflection %.6f″ over the sweep, want ≲0.001″", worst)
	}
	if worstStep > 1e-4 {
		t.Errorf("deflection jumps by %.6f″ between 10-minute samples, want a smooth run", worstStep)
	}
}

// Deflection is applied at all, and stays within the grazing-limb bound for
// rays that clear the Sun.
func TestApparentDeflectionIsApplied(t *testing.T) {
	eph := openEph(t)
	const jd = 2460265.0
	if got, _, _ := deflectionOnly(t, eph, MarsBarycenter, jd); got < 0.4 {
		t.Errorf("deflection near solar conjunction %.6f″, want ~0.54″", got)
	}
	for _, body := range []int{Mercury, Venus, MarsBarycenter, JupiterBarycenter, SaturnBarycenter, NeptuneBarycenter} {
		for _, d := range []float64{0, 97.3, 206.5, 311.1} {
			got, _, _ := deflectionOnly(t, eph, body, 2460000.5+d)
			astrometric, _ := eph.observe(Earth, body, 2460000.5+d)
			sun, _ := eph.observe(Earth, Sun, 2460000.5+d)
			chi := sepArcsec(astrometric, sun) / 3600 * math.Pi / 180
			if length3(sun)*math.Sin(chi) < sunRadKmTest {
				continue // ray passes behind the solar disk
			}
			if got > 1.76 {
				t.Errorf("body %d at %+.1f d: deflection %.6f″ exceeds the grazing limit", body, d, got)
			}
		}
	}
}
