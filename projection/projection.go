// Package projection provides stereographic projection of sky positions onto
// a 2D plane for plotting star charts.
//
// Stereographic projection maps the celestial sphere onto a flat plane while
// preserving angles (conformal). It is centered at an arbitrary direction and
// projects unit vectors to (x, y) coordinates. Objects near the center map
// close to (0, 0); the center itself maps exactly to the origin.
package projection

import "math"

// Projector projects 3D unit vectors onto a 2D plane using stereographic
// projection centered at a given direction.
type Projector struct {
	// Pre-computed constants for the projection.
	xc, yc, zc float64 // center unit vector
	t0          float64 // 1 / sqrt(xc² + yc²)
	t2          float64 // sqrt(1 - zc²)
	t3          float64 // t0 * t2
	t6base      float64 // t0 * zc
}

// NewProjector creates a stereographic projection centered at the given
// direction. The center is specified as a 3D vector (not necessarily unit
// length); it will be normalized internally.
//
// The projection maps the center to (0, 0). Nearby positions map to small
// (x, y) values; positions 90° from center map to distance 1.0.
func NewProjector(centerX, centerY, centerZ float64) *Projector {
	// Normalize to unit vector.
	r := math.Sqrt(centerX*centerX + centerY*centerY + centerZ*centerZ)
	xc := centerX / r
	yc := centerY / r
	zc := centerZ / r

	// Pre-compute subexpressions (matching Skyfield's SymPy-optimized form).
	t0 := 1.0 / math.Sqrt(xc*xc+yc*yc)
	t2 := math.Sqrt(1.0 - zc*zc)
	t3 := t0 * t2

	return &Projector{
		xc: xc, yc: yc, zc: zc,
		t0: t0, t2: t2, t3: t3,
		t6base: t0 * zc,
	}
}

// Project maps a 3D position to 2D stereographic coordinates.
// The input vector does not need to be unit length; it is normalized internally.
//
// Returns (x, y) where x increases to the right and y increases upward.
// The center of projection maps to (0, 0).
func (p *Projector) Project(x, y, z float64) (px, py float64) {
	// Normalize input to unit vector.
	r := math.Sqrt(x*x + y*y + z*z)
	x /= r
	y /= r
	z /= r

	// Stereographic projection with pre-rotated center (Skyfield's optimized form).
	t1 := x * p.xc
	t4 := y * p.yc
	t5 := 1.0 / (t1*p.t3 + p.t3*t4 + z*p.zc + 1.0)

	px = p.t0 * t5 * (x*p.yc - p.xc*y)
	py = -t5 * (t1*p.t6base - p.t2*z + t4*p.t6base)
	return px, py
}
