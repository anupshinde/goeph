package coord

import (
	"math"
	"testing"
)

const (
	auKm      = 1.495978707e8 // astronomical unit in km
	sunRadKm  = 695700.0      // solar radius in km
	radToArcs = 180 / math.Pi * 3600
)

// deflectionGeometry places the deflector at the origin, the observer at
// distance eKm from it, and the target at distance dKm from the observer in a
// direction making the angle chi (radians) with the direction to the
// deflector. It returns the observer-to-target and deflector-to-observer
// vectors Deflection takes.
func deflectionGeometry(eKm, dKm, chi float64) (position, deflectorToObserver [3]float64) {
	// Observer on the +x axis, so the deflector lies along -x from it.
	deflectorToObserver = [3]float64{eKm, 0, 0}
	position = [3]float64{-dKm * math.Cos(chi), dKm * math.Sin(chi), 0}
	return position, deflectorToObserver
}

// separationFromDeflector returns the angle (radians) between the target
// direction and the direction to the deflector, as seen by the observer.
func separationFromDeflector(position, deflectorToObserver [3]float64) float64 {
	toDeflector := scale3(-1, deflectorToObserver)
	cos := dot3(position, toDeflector) / (length3(position) * length3(toDeflector))
	return math.Acos(math.Max(-1, math.Min(1, cos)))
}

// For a source far behind the deflector the deflection is
//
//	δ = 2·GM/(c²·E) · (1 + cos χ)/sin χ
//
// with E the observer's distance from the deflector and χ the angle between
// the deflector and the target as seen by the observer. The expression the
// code evaluates must reduce to this.
func TestDeflection_DistantSourceClosedForm(t *testing.T) {
	const dKm = 1e6 * auKm // far enough that deflector-to-target ≈ observer-to-target
	for _, chiDeg := range []float64{0.2665, 1, 5, 30, 90, 150, 179} {
		chi := chiDeg * math.Pi / 180
		position, pe := deflectionGeometry(auKm, dKm, chi)
		d := Deflection(position, pe, 1.0)

		got := length3(d) / length3(position) * radToArcs
		want := 2 * gs / (cMPerSec * cMPerSec * auKm * 1000) * (1 + math.Cos(chi)) / math.Sin(chi) * radToArcs
		if math.Abs(got-want) > 1e-4*want {
			t.Errorf("chi %.4f°: deflection %.6f″, want %.6f″", chiDeg, got, want)
		}
	}
}

// A ray grazing the Sun's limb is deflected by the classical 1.75″.
func TestDeflection_GrazingTheSunsLimb(t *testing.T) {
	chi := math.Asin(sunRadKm / auKm)
	position, pe := deflectionGeometry(auKm, 1e6*auKm, chi)
	got := length3(Deflection(position, pe, 1.0)) / length3(position) * radToArcs
	if math.Abs(got-1.7518) > 0.005 {
		t.Errorf("grazing deflection %.4f″, want 1.7518″", got)
	}
	// It falls off as 1/b: twice the impact parameter, half the deflection.
	chi2 := math.Asin(2 * sunRadKm / auKm)
	position2, pe2 := deflectionGeometry(auKm, 1e6*auKm, chi2)
	got2 := length3(Deflection(position2, pe2, 1.0)) / length3(position2) * radToArcs
	if r := got / got2; math.Abs(r-2) > 0.01 {
		t.Errorf("deflection ratio at 1 and 2 solar radii = %.4f, want 2", r)
	}
}

// Light bends towards the mass, so the image moves away from the deflector:
// the deflected direction always sits further from the deflector than the
// astrometric one. This holds at every geometry, which is what makes it a
// check on the direction of the deflector-to-observer vector the caller
// passes in.
func TestDeflection_MovesImageAwayFromDeflector(t *testing.T) {
	for _, chiDeg := range []float64{0.3, 2, 20, 90, 120, 179} {
		chi := chiDeg * math.Pi / 180
		for _, dKm := range []float64{1.5 * auKm, 5 * auKm, 30 * auKm, 1e6 * auKm} {
			position, pe := deflectionGeometry(auKm, dKm, chi)
			before := separationFromDeflector(position, pe)
			after := separationFromDeflector(add3(position, Deflection(position, pe, 1.0)), pe)
			if after <= before {
				t.Errorf("chi %.1f° target %.1f AU: separation %.9f° → %.9f°, want it to grow",
					chiDeg, dKm/auKm, before*180/math.Pi, after*180/math.Pi)
			}
			if d := (after - before) * radToArcs; d > 1.76 {
				t.Errorf("chi %.1f° target %.1f AU: deflection %.4f″ exceeds the grazing limit",
					chiDeg, dKm/auKm, d)
			}
		}
	}
}

// A deflector behind the observer bends the incoming ray by almost nothing:
// the (1 + cos χ)/sin χ factor vanishes as χ approaches 180°. The ray only
// nears the deflector's centre — where the deflection does diverge and the
// near-collinear guard takes over — when the deflector is in front, between
// the observer and the target.
func TestDeflection_BehindTheObserverIsNegligible(t *testing.T) {
	for _, chiDeg := range []float64{179, 179.9, 179.99, 179.999999} {
		chi := chiDeg * math.Pi / 180
		for _, eKm := range []float64{0.1 * auKm, auKm, 5 * auKm} {
			position, pe := deflectionGeometry(eKm, 30*auKm, chi)
			if d := length3(Deflection(position, pe, 1.0)) / length3(position) * radToArcs; d > 0.001 {
				t.Errorf("chi %.6f° deflector at %.2f AU: deflection %.6f″, want ≲0.001″", chiDeg, eKm/auKm, d)
			}
		}
	}
}

// The target itself standing in front of the deflector's centre is the real
// singularity: the guard returns no correction there rather than a divergent
// one.
func TestDeflection_NearCollinearGuard(t *testing.T) {
	for _, chi := range []float64{0, 1e-9, math.Pi} {
		position, pe := deflectionGeometry(auKm, 30*auKm, chi)
		if d := Deflection(position, pe, 1.0); d != [3]float64{} {
			t.Errorf("chi %g: correction %v, want none", chi, d)
		}
	}
	if d := Deflection([3]float64{}, [3]float64{auKm, 0, 0}, 1.0); d != [3]float64{} {
		t.Errorf("zero position: correction %v, want none", d)
	}
}

// Jupiter and Saturn deflect in proportion to their mass.
func TestDeflection_ScalesWithReciprocalMass(t *testing.T) {
	position, pe := deflectionGeometry(5*auKm, 1e6*auKm, 0.01)
	sun := length3(Deflection(position, pe, 1.0))
	jup := length3(Deflection(position, pe, 1047.3486))
	if r := sun / jup; math.Abs(r-1047.3486) > 1e-6*1047.3486 {
		t.Errorf("Sun/Jupiter deflection ratio %.6f, want 1047.3486", r)
	}
}
