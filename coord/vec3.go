package coord

import "math"

func dot3(a, b [3]float64) float64 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

func length3(a [3]float64) float64 {
	return math.Sqrt(a[0]*a[0] + a[1]*a[1] + a[2]*a[2])
}

func scale3(s float64, a [3]float64) [3]float64 {
	return [3]float64{s * a[0], s * a[1], s * a[2]}
}

func sub3(a, b [3]float64) [3]float64 {
	return [3]float64{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
}

func add3(a, b [3]float64) [3]float64 {
	return [3]float64{a[0] + b[0], a[1] + b[1], a[2] + b[2]}
}
