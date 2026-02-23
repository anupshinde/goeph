package coord

import (
	"math"
	"testing"
)

func TestGalacticMatrix_Orthogonal(t *testing.T) {
	// Verify M * M^T = I (orthogonal rotation matrix)
	var prod [3][3]float64
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				prod[i][j] += GalacticMatrix[i][k] * GalacticMatrix[j][k]
			}
		}
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			want := 0.0
			if i == j {
				want = 1.0
			}
			if math.Abs(prod[i][j]-want) > 1e-14 {
				t.Errorf("GalacticMatrix M*M^T[%d][%d] = %.15e, want %f", i, j, prod[i][j], want)
			}
		}
	}
}

func TestGalacticMatrix_DetPositive(t *testing.T) {
	// det should be +1 for a proper rotation (not reflection)
	m := GalacticMatrix
	det := m[0][0]*(m[1][1]*m[2][2]-m[1][2]*m[2][1]) -
		m[0][1]*(m[1][0]*m[2][2]-m[1][2]*m[2][0]) +
		m[0][2]*(m[1][0]*m[2][1]-m[1][1]*m[2][0])
	if math.Abs(det-1.0) > 1e-14 {
		t.Errorf("det(GalacticMatrix) = %.15f, want 1.0", det)
	}
}

func TestB1950Matrix_Orthogonal(t *testing.T) {
	var prod [3][3]float64
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				prod[i][j] += B1950Matrix[i][k] * B1950Matrix[j][k]
			}
		}
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			want := 0.0
			if i == j {
				want = 1.0
			}
			if math.Abs(prod[i][j]-want) > 1e-14 {
				t.Errorf("B1950Matrix M*M^T[%d][%d] = %.15e, want %f", i, j, prod[i][j], want)
			}
		}
	}
}

func TestICRSToJ2000Matrix_NearIdentity(t *testing.T) {
	// Frame bias is a few milliarcseconds, so the matrix should be very close to identity
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			want := 0.0
			if i == j {
				want = 1.0
			}
			if math.Abs(ICRSToJ2000Matrix[i][j]-want) > 1e-4 {
				t.Errorf("ICRSToJ2000Matrix[%d][%d] = %.15e, want ~%f", i, j, ICRSToJ2000Matrix[i][j], want)
			}
		}
	}
}

func TestICRSToJ2000Matrix_NonIdentity(t *testing.T) {
	// It should NOT be exactly identity
	isIdentity := true
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			want := 0.0
			if i == j {
				want = 1.0
			}
			if ICRSToJ2000Matrix[i][j] != want {
				isIdentity = false
			}
		}
	}
	if isIdentity {
		t.Error("ICRSToJ2000Matrix is exactly identity (expected small bias)")
	}
}

func TestICRFToGalactic_GalacticCenter(t *testing.T) {
	// Galactic center (Sgr A*): RA=17h45m40.0409s, Dec=-29°00'28.118"
	// In ICRF unit vector form:
	x, y, z := RADecToICRF(17.0+45.0/60.0+40.0409/3600.0, -(29.0+0.0/60.0+28.118/3600.0))

	lat, lon := ICRFToGalactic(x, y, z)

	// Should be near l=0°, b=0° (the definition of galactic center)
	if math.Abs(lat) > 0.1 {
		t.Errorf("galactic center lat: got %f, want ~0", lat)
	}
	if math.Abs(lon) > 0.1 && math.Abs(lon-360) > 0.1 {
		t.Errorf("galactic center lon: got %f, want ~0", lon)
	}
}

func TestICRFToGalactic_NorthPole(t *testing.T) {
	// Galactic north pole: RA=12h51m26.28s, Dec=+27°07'41.7"
	x, y, z := RADecToICRF(12.0+51.0/60.0+26.28/3600.0, 27.0+7.0/60.0+41.7/3600.0)

	lat, _ := ICRFToGalactic(x, y, z)

	// Should be near b=90°
	if math.Abs(lat-90.0) > 0.1 {
		t.Errorf("galactic north pole lat: got %f, want ~90", lat)
	}
}

func TestICRFToGalactic_ZeroVector(t *testing.T) {
	lat, lon := ICRFToGalactic(0, 0, 0)
	if lat != 0 || lon != 0 {
		t.Errorf("zero vector: got lat=%f lon=%f", lat, lon)
	}
}
