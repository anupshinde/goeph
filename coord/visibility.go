package coord

import "math"

const earthRadiusKm = 6371.0 // mean radius in km

// IsSunlit returns true if a position (in km, ICRF, relative to Earth center)
// is illuminated by the Sun.
//
// posKm is the object's geocentric position in km (e.g., a satellite).
// sunPosKm is the Sun's geocentric position in km (from SPK.Observe or SPK.Apparent).
//
// Uses geometric shadow test: the object is in shadow if the line from the
// object to the Sun intersects Earth's sphere.
func IsSunlit(posKm, sunPosKm [3]float64) bool {
	// Vector from object to Sun
	toSun := [3]float64{
		sunPosKm[0] - posKm[0],
		sunPosKm[1] - posKm[1],
		sunPosKm[2] - posKm[2],
	}

	// Earth center relative to object position (= -posKm)
	earthCenter := [3]float64{-posKm[0], -posKm[1], -posKm[2]}

	// Check if the line from object toward the Sun intersects Earth
	near, far := intersectLineSphere(toSun, earthCenter, earthRadiusKm)

	// If no intersection, the object is sunlit
	if math.IsNaN(near) {
		return true
	}

	// Intersection must be between object and Sun (t in [0, 1] along toSun)
	// near and far are distances along the unit direction, but we need
	// to normalize by the length of toSun
	sunDist := math.Sqrt(toSun[0]*toSun[0] + toSun[1]*toSun[1] + toSun[2]*toSun[2])
	if sunDist == 0 {
		return false
	}

	// The intersection distances are in the same units as the direction vector
	// Check if any intersection is between 0 and sunDist
	if far < 0 || near > sunDist {
		return true // intersection is behind the object or past the Sun
	}

	return false
}

// IsBehindEarth returns true if the target position is geometrically behind
// Earth as seen from the observer position.
//
// Both positions are geocentric ICRF vectors in km. The target is "behind Earth"
// if the line of sight from observer to target passes through Earth's sphere.
func IsBehindEarth(observerPosKm, targetPosKm [3]float64) bool {
	// Vector from observer to target
	toTarget := [3]float64{
		targetPosKm[0] - observerPosKm[0],
		targetPosKm[1] - observerPosKm[1],
		targetPosKm[2] - observerPosKm[2],
	}

	// Earth center relative to observer
	earthCenter := [3]float64{-observerPosKm[0], -observerPosKm[1], -observerPosKm[2]}

	near, _ := intersectLineSphere(toTarget, earthCenter, earthRadiusKm)
	if math.IsNaN(near) {
		return false
	}

	targetDist := math.Sqrt(toTarget[0]*toTarget[0] + toTarget[1]*toTarget[1] + toTarget[2]*toTarget[2])
	if targetDist == 0 {
		return false
	}

	// Check if intersection is between observer and target
	return near >= 0 && near <= targetDist
}

// intersectLineSphere computes line-sphere intersection.
// endpoint is the direction vector (line goes from origin toward endpoint).
// center is the sphere center relative to the line origin.
// Returns (near, far) distances along the direction. NaN if no intersection.
func intersectLineSphere(endpoint, center [3]float64, radius float64) (near, far float64) {
	lenE := math.Sqrt(endpoint[0]*endpoint[0] + endpoint[1]*endpoint[1] + endpoint[2]*endpoint[2])
	if lenE == 0 {
		return math.NaN(), math.NaN()
	}

	dx := endpoint[0] / lenE
	dy := endpoint[1] / lenE
	dz := endpoint[2] / lenE

	minusB := 2.0 * (dx*center[0] + dy*center[1] + dz*center[2])
	c := center[0]*center[0] + center[1]*center[1] + center[2]*center[2] - radius*radius
	discriminant := minusB*minusB - 4.0*c

	if discriminant < 0 {
		return math.NaN(), math.NaN()
	}

	dsqrt := math.Sqrt(discriminant)
	near = (minusB - dsqrt) / 2.0
	far = (minusB + dsqrt) / 2.0
	return
}
