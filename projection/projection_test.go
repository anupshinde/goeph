package projection

import (
	"math"
	"testing"
)

func TestProjectCenter(t *testing.T) {
	// Center should project to (0, 0).
	p := NewProjector(0, 0, 1) // center at north pole
	x, y := p.Project(0, 0, 1)
	if math.Abs(x) > 1e-14 || math.Abs(y) > 1e-14 {
		t.Errorf("center: got (%g, %g), want (0, 0)", x, y)
	}
}

func TestProject_NorthPoleCenter(t *testing.T) {
	// Matches Skyfield: center at RA=0, Dec=90° (north pole = 0,0,1)
	p := NewProjector(0, 0, 1)

	tests := []struct {
		name       string
		ra, dec    float64 // hours, degrees
		wantX      float64
		wantY      float64
		tol        float64
	}{
		{"1° from pole", 0, 89, 0.0, -0.008727, 1e-5},
		{"10° from pole RA=6h", 6, 80, -0.087489, 0.0, 1e-5},
		{"45° from pole RA=12h", 12, 45, 0.0, 0.414214, 1e-5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y, z := raDecToXYZ(tt.ra, tt.dec)
			px, py := p.Project(x, y, z)
			if math.Abs(px-tt.wantX) > tt.tol || math.Abs(py-tt.wantY) > tt.tol {
				t.Errorf("got (%.6f, %.6f), want (%.6f, %.6f)", px, py, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestProject_ArbitraryCenter(t *testing.T) {
	// Center at RA=12h, Dec=45° (matching Skyfield test values)
	cx, cy, cz := raDecToXYZ(12, 45)
	p := NewProjector(cx, cy, cz)

	tests := []struct {
		name    string
		ra, dec float64
		wantX   float64
		wantY   float64
		tol     float64
	}{
		{"center", 12, 45, 0.0, 0.0, 1e-10},
		{"1° north", 12, 46, 0.0, 0.008727, 1e-5},
		{"1h east", 13, 45, -0.092293, 0.008592, 1e-5},
		{"45° south", 12, 0, 0.0, -0.414214, 1e-5},
		{"north pole", 0, 90, 0.0, 0.414214, 1e-5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y, z := raDecToXYZ(tt.ra, tt.dec)
			px, py := p.Project(x, y, z)
			if math.Abs(px-tt.wantX) > tt.tol || math.Abs(py-tt.wantY) > tt.tol {
				t.Errorf("got (%.6f, %.6f), want (%.6f, %.6f)", px, py, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestProject_Conformal(t *testing.T) {
	// Stereographic projection is conformal: it preserves angles locally.
	// Two orthogonal directions from a point should remain orthogonal after projection.
	cx, cy, cz := raDecToXYZ(12, 45)
	p := NewProjector(cx, cy, cz)

	// Base point slightly off-center
	bx, by, bz := raDecToXYZ(12.5, 46)
	bpx, bpy := p.Project(bx, by, bz)

	// Point moved in RA direction
	ax, ay, az := raDecToXYZ(12.51, 46)
	apx, apy := p.Project(ax, ay, az)

	// Point moved in Dec direction
	dx, dy, dz := raDecToXYZ(12.5, 46.1)
	dpx, dpy := p.Project(dx, dy, dz)

	// Vectors from base to each
	v1x, v1y := apx-bpx, apy-bpy
	v2x, v2y := dpx-bpx, dpy-bpy

	// Dot product should be near zero (orthogonal)
	dot := v1x*v2x + v1y*v2y
	mag1 := math.Sqrt(v1x*v1x + v1y*v1y)
	mag2 := math.Sqrt(v2x*v2x + v2y*v2y)
	cosAngle := dot / (mag1 * mag2)

	if math.Abs(cosAngle) > 0.01 {
		t.Errorf("non-conformal: cos(angle) = %.6f, want ~0", cosAngle)
	}
}

// raDecToXYZ converts RA (hours) and Dec (degrees) to a unit ICRF vector.
func raDecToXYZ(raHours, decDeg float64) (x, y, z float64) {
	ra := raHours * 15.0 * math.Pi / 180.0 // hours to radians
	dec := decDeg * math.Pi / 180.0
	cosDec := math.Cos(dec)
	x = cosDec * math.Cos(ra)
	y = cosDec * math.Sin(ra)
	z = math.Sin(dec)
	return
}
