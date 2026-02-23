package geometry

import "math"

// IntersectLineSphere computes the distances from the origin to the
// intersections of a line (from the origin through endpoint) with a sphere
// defined by center and radius. Returns (near, far) distances along the line.
//
// If the line does not intersect the sphere, both values are NaN.
// If tangent, both values are equal.
//
// See: http://paulbourke.net/geometry/circlesphere/index.html#linesphere
func IntersectLineSphere(endpoint, center [3]float64, radius float64) (near, far float64) {
	lenE := math.Sqrt(endpoint[0]*endpoint[0] + endpoint[1]*endpoint[1] + endpoint[2]*endpoint[2])
	if lenE == 0 {
		return math.NaN(), math.NaN()
	}

	// Unit direction vector
	dx := endpoint[0] / lenE
	dy := endpoint[1] / lenE
	dz := endpoint[2] / lenE

	// Quadratic formula with a=1 (unit direction vector)
	minusB := 2.0 * (dx*center[0] + dy*center[1] + dz*center[2])
	c := center[0]*center[0] + center[1]*center[1] + center[2]*center[2] - radius*radius
	discriminant := minusB*minusB - 4.0*c

	if discriminant < 0 {
		return math.NaN(), math.NaN()
	}

	dsqrt := math.Sqrt(discriminant)
	near = (minusB - dsqrt) / 2.0
	far = (minusB + dsqrt) / 2.0
	return near, far
}
