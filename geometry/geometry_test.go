package geometry

import (
	"math"
	"testing"
)

func TestIntersectLineSphere_Through(t *testing.T) {
	// Line along X-axis, sphere centered at (5,0,0) with radius 1
	endpoint := [3]float64{1, 0, 0}
	center := [3]float64{5, 0, 0}
	near, far := IntersectLineSphere(endpoint, center, 1.0)

	if math.Abs(near-4.0) > 1e-12 {
		t.Errorf("near: got %f, want 4.0", near)
	}
	if math.Abs(far-6.0) > 1e-12 {
		t.Errorf("far: got %f, want 6.0", far)
	}
}

func TestIntersectLineSphere_Tangent(t *testing.T) {
	// Line along X-axis, sphere at (5,1,0) with radius 1 (tangent from below)
	endpoint := [3]float64{1, 0, 0}
	center := [3]float64{5, 1, 0}
	near, far := IntersectLineSphere(endpoint, center, 1.0)

	if math.Abs(near-far) > 1e-12 {
		t.Errorf("tangent: near=%f far=%f, should be equal", near, far)
	}
	if math.Abs(near-5.0) > 1e-12 {
		t.Errorf("tangent point: got %f, want 5.0", near)
	}
}

func TestIntersectLineSphere_Miss(t *testing.T) {
	// Line along X-axis, sphere at (0,5,0) with radius 1 (misses)
	endpoint := [3]float64{1, 0, 0}
	center := [3]float64{0, 5, 0}
	near, far := IntersectLineSphere(endpoint, center, 1.0)

	if !math.IsNaN(near) || !math.IsNaN(far) {
		t.Errorf("miss: got near=%f far=%f, want NaN", near, far)
	}
}

func TestIntersectLineSphere_ZeroEndpoint(t *testing.T) {
	endpoint := [3]float64{0, 0, 0}
	center := [3]float64{5, 0, 0}
	near, far := IntersectLineSphere(endpoint, center, 1.0)

	if !math.IsNaN(near) || !math.IsNaN(far) {
		t.Errorf("zero endpoint: got near=%f far=%f, want NaN", near, far)
	}
}

func TestIntersectLineSphere_OriginInsideSphere(t *testing.T) {
	// Origin is inside the sphere
	endpoint := [3]float64{1, 0, 0}
	center := [3]float64{0, 0, 0}
	near, far := IntersectLineSphere(endpoint, center, 5.0)

	// One intersection behind (negative), one in front
	if near >= 0 {
		t.Errorf("near should be negative (behind origin): got %f", near)
	}
	if far <= 0 {
		t.Errorf("far should be positive (in front of origin): got %f", far)
	}
	if math.Abs(near+5.0) > 1e-12 {
		t.Errorf("near: got %f, want -5.0", near)
	}
	if math.Abs(far-5.0) > 1e-12 {
		t.Errorf("far: got %f, want 5.0", far)
	}
}

func TestIntersectLineSphere_DiagonalLine(t *testing.T) {
	// Line from origin through (1,1,0), sphere at origin with radius 1
	endpoint := [3]float64{1, 1, 0}
	center := [3]float64{0, 0, 0}
	near, far := IntersectLineSphere(endpoint, center, 1.0)

	if math.Abs(near+1.0) > 1e-12 {
		t.Errorf("near: got %f, want -1.0", near)
	}
	if math.Abs(far-1.0) > 1e-12 {
		t.Errorf("far: got %f, want 1.0", far)
	}
}
