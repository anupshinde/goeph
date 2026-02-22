package coord

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

// goldenSidereal matches testdata/golden_sidereal.json.
type goldenSidereal struct {
	Tests []struct {
		UT1JD   float64 `json:"ut1_jd"`
		GMSTDeg float64 `json:"gmst_deg"`
	} `json:"tests"`
}

// goldenLocations matches testdata/golden_locations.json.
type goldenLocations struct {
	Tests []struct {
		TDBJD     float64 `json:"tdb_jd"`
		UT1JD     float64 `json:"ut1_jd"`
		LocName   string  `json:"loc_name"`
		Lat       float64 `json:"lat"`
		Lon       float64 `json:"lon"`
		EclLatDeg float64 `json:"ecl_lat_deg"`
		EclLonDeg float64 `json:"ecl_lon_deg"`
	} `json:"tests"`
}

func loadJSON(t *testing.T, path string, v interface{}) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatal(err)
	}
}

func TestICRFToEcliptic_Zero(t *testing.T) {
	lat, lon := ICRFToEcliptic(0, 0, 0)
	if lat != 0 || lon != 0 {
		t.Errorf("zero vector: got lat=%f lon=%f", lat, lon)
	}
}

func TestICRFToEcliptic_XAxis(t *testing.T) {
	lat, lon := ICRFToEcliptic(1, 0, 0)
	if math.Abs(lat) > 1e-10 || math.Abs(lon) > 1e-10 {
		t.Errorf("x-axis: got lat=%f lon=%f, want 0,0", lat, lon)
	}
}

func TestICRFToEcliptic_Roundtrip(t *testing.T) {
	ex := 0.0
	ey := 1.0
	ez := 0.0
	xICRF := ex
	yICRF := obliquityCos*ey - obliquitySin*ez
	zICRF := obliquitySin*ey + obliquityCos*ez

	lat, lon := ICRFToEcliptic(xICRF, yICRF, zICRF)
	if math.Abs(lat) > 1e-10 {
		t.Errorf("roundtrip lat: got %f want 0", lat)
	}
	if math.Abs(lon-90.0) > 1e-10 {
		t.Errorf("roundtrip lon: got %f want 90", lon)
	}
}

func TestRADecToICRF(t *testing.T) {
	x, y, z := RADecToICRF(0, 0)
	if math.Abs(x-1.0) > 1e-15 || math.Abs(y) > 1e-15 || math.Abs(z) > 1e-15 {
		t.Errorf("RA=0 Dec=0: got (%.15f, %.15f, %.15f)", x, y, z)
	}

	x, y, z = RADecToICRF(6, 0)
	if math.Abs(x) > 1e-15 || math.Abs(y-1.0) > 1e-15 || math.Abs(z) > 1e-15 {
		t.Errorf("RA=6h Dec=0: got (%.15f, %.15f, %.15f)", x, y, z)
	}

	x, y, z = RADecToICRF(0, 90)
	if math.Abs(x) > 1e-15 || math.Abs(y) > 1e-15 || math.Abs(z-1.0) > 1e-15 {
		t.Errorf("RA=0 Dec=90: got (%.15f, %.15f, %.15f)", x, y, z)
	}
}

func TestGMST_J2000(t *testing.T) {
	gmst := GMST(j2000JD)
	if math.Abs(gmst-280.46061837) > 0.001 {
		t.Errorf("GMST at J2000: got %f want ~280.461", gmst)
	}
}

func TestGMST_Golden(t *testing.T) {
	var golden goldenSidereal
	loadJSON(t, "../testdata/golden_sidereal.json", &golden)

	// goeph uses IAU 1982 (Meeus) GMST formula; Skyfield uses IERS 2000 ERA-based.
	// Systematic offset ~1e-4° is expected.
	const tol = 1e-3 // degrees
	failures := 0
	for i, tc := range golden.Tests {
		got := GMST(tc.UT1JD)
		// Normalize to [0, 360) for comparison
		got = math.Mod(got, 360.0)
		if got < 0 {
			got += 360.0
		}
		diff := got - tc.GMSTDeg
		if diff > 180 {
			diff -= 360
		} else if diff < -180 {
			diff += 360
		}
		if math.Abs(diff) > tol {
			if failures < 10 {
				t.Errorf("test %d: ut1=%.6f got=%.10f want=%.10f diff=%.2e",
					i, tc.UT1JD, got, tc.GMSTDeg, diff)
			}
			failures++
		}
	}
	if failures > 0 {
		t.Errorf("%d GMST failures out of %d tests (tol=%.0e°)", failures, len(golden.Tests), tol)
	}
}

func TestGAST(t *testing.T) {
	gast := GAST(j2000JD)
	gmst := GMST(j2000JD)
	diff := gast - gmst
	if diff > 180 {
		diff -= 360
	} else if diff < -180 {
		diff += 360
	}
	if math.Abs(diff) > 0.01 {
		t.Errorf("GAST-GMST difference too large: %f°", diff)
	}
}

func TestNutationAngles(t *testing.T) {
	dpsi, deps := nutationAngles(0)
	dpsiArcsec := dpsi / arcsec2rad
	depsArcsec := deps / arcsec2rad
	if math.Abs(dpsiArcsec) > 30 || math.Abs(depsArcsec) > 30 {
		t.Errorf("nutation at T=0 too large: dpsi=%.3f\" deps=%.3f\"", dpsiArcsec, depsArcsec)
	}
	if dpsiArcsec == 0 && depsArcsec == 0 {
		t.Error("nutation at T=0 is exactly zero (unexpected)")
	}
}

func TestNutationAngles_VaryWithTime(t *testing.T) {
	dpsi0, deps0 := nutationAngles(0)
	dpsi1, deps1 := nutationAngles(1.0) // 1 century later
	if dpsi0 == dpsi1 && deps0 == deps1 {
		t.Error("nutation unchanged after 1 century")
	}
}

func TestFundamentalArgs(t *testing.T) {
	l, lp, F, D, om := fundamentalArgs(0)
	for _, v := range []float64{l, lp, F, D, om} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Fatal("fundamental args returned NaN or Inf")
		}
	}
}

func TestFundamentalArgs_VaryWithTime(t *testing.T) {
	l0, _, _, _, _ := fundamentalArgs(0)
	l1, _, _, _, _ := fundamentalArgs(0.01)
	if l0 == l1 {
		t.Error("fundamental arg l unchanged with different T")
	}
}

func TestMeanObliquity(t *testing.T) {
	eps := meanObliquity(0)
	epsDeg := eps * rad2deg
	if math.Abs(epsDeg-23.4393) > 0.001 {
		t.Errorf("mean obliquity at T=0: got %.4f° want ~23.4393°", epsDeg)
	}
}

func TestMeanObliquity_Decreasing(t *testing.T) {
	eps0 := meanObliquity(0)
	eps1 := meanObliquity(1.0)
	if eps1 >= eps0 {
		t.Error("mean obliquity should decrease over centuries")
	}
}

func TestNutationMatrixTranspose_Identity(t *testing.T) {
	NT := nutationMatrixTranspose(0, 0, meanObliquity(0))
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			want := 0.0
			if i == j {
				want = 1.0
			}
			if math.Abs(NT[i][j]-want) > 1e-10 {
				t.Errorf("NT[%d][%d] = %f, want %f", i, j, NT[i][j], want)
			}
		}
	}
}

func TestNutationMatrixTranspose_NonZero(t *testing.T) {
	dpsi, deps := nutationAngles(0)
	epsM := meanObliquity(0)
	NT := nutationMatrixTranspose(dpsi, deps, epsM)
	if NT[0][1] == 0 || NT[0][2] == 0 {
		t.Error("nutation matrix off-diagonal is zero with nonzero nutation")
	}
}

func TestPrecessionMatrixInverse_T0(t *testing.T) {
	P := precessionMatrixInverse(0)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			want := 0.0
			if i == j {
				want = 1.0
			}
			if math.Abs(P[i][j]-want) > 1e-10 {
				t.Errorf("P[%d][%d] = %.15e, want %f", i, j, P[i][j], want)
			}
		}
	}
}

func TestPrecessionMatrixInverse_Orthogonal(t *testing.T) {
	P := precessionMatrixInverse(1.0)
	var prod [3][3]float64
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				prod[i][j] += P[i][k] * P[j][k]
			}
		}
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			want := 0.0
			if i == j {
				want = 1.0
			}
			if math.Abs(prod[i][j]-want) > 1e-12 {
				t.Errorf("P*P^T[%d][%d] = %.15e, want %f", i, j, prod[i][j], want)
			}
		}
	}
}

func TestGeodeticToICRF_UnitVector(t *testing.T) {
	x, y, z := GeodeticToICRF(0, 0, j2000JD)
	r := math.Sqrt(x*x + y*y + z*z)
	if math.Abs(r-1.0) > 1e-12 {
		t.Errorf("not a unit vector: |r| = %.15f", r)
	}
}

func TestGeodeticToICRF_Pole(t *testing.T) {
	x, y, z := GeodeticToICRF(90, 0, j2000JD)
	r := math.Sqrt(x*x + y*y + z*z)
	x /= r
	y /= r
	z /= r
	if math.Abs(z) < 0.9 {
		t.Errorf("north pole z too small: %.6f", z)
	}
}

func TestGeodeticToICRF_AllGoldenLocations(t *testing.T) {
	var golden goldenLocations
	loadJSON(t, "../testdata/golden_locations.json", &golden)

	for i, tc := range golden.Tests {
		x, y, z := GeodeticToICRF(tc.Lat, tc.Lon, tc.UT1JD)
		r := math.Sqrt(x*x + y*y + z*z)
		if math.Abs(r-1.0) > 1e-10 {
			t.Errorf("test %d: not unit vector, |r|=%.15f", i, r)
			break
		}
	}
}

func TestGeodeticToICRF_DifferentTimes(t *testing.T) {
	x0, y0, z0 := GeodeticToICRF(0, 0, j2000JD)
	x1, y1, z1 := GeodeticToICRF(0, 0, j2000JD+0.5) // 12 hours later
	// Earth rotates, so direction should change
	dot := x0*x1 + y0*y1 + z0*z1
	if math.Abs(dot-1.0) < 1e-6 {
		t.Error("geodetic direction unchanged after 12 hours (Earth should have rotated)")
	}
}

func TestLocationStruct(t *testing.T) {
	loc := Location{Name: "Test", Lat: 40.0, Lon: -74.0}
	if loc.Name != "Test" || loc.Lat != 40.0 || loc.Lon != -74.0 {
		t.Error("Location fields not set correctly")
	}
}

func BenchmarkICRFToEcliptic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ICRFToEcliptic(1e8, -5e7, 2e7)
	}
}

func BenchmarkGAST(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GAST(2451545.0 + float64(i))
	}
}

// TestGeodeticToEcliptic_Golden tests the full geodetic→ICRF→ecliptic pipeline
// against Skyfield golden data. This is the end-to-end accuracy test.
// goeph uses 30-term nutation vs Skyfield's 687 terms.
// The nutation gap grows with distance from J2000. Over a 300-year range,
// max deviation reaches ~0.03° (~113 arcsec). This is a documented limitation.
func TestGeodeticToEcliptic_Golden(t *testing.T) {
	var golden goldenLocations
	loadJSON(t, "../testdata/golden_locations.json", &golden)

	const tolerance = 0.035 // degrees

	var maxLatErr, maxLonErr float64
	failures := 0

	for _, tc := range golden.Tests {
		x, y, z := GeodeticToICRF(tc.Lat, tc.Lon, tc.UT1JD)
		latDeg, lonDeg := ICRFToEcliptic(x, y, z)

		latErr := math.Abs(latDeg - tc.EclLatDeg)
		lonErr := math.Abs(lonDeg - tc.EclLonDeg)
		if lonErr > 180 {
			lonErr = 360 - lonErr
		}

		if latErr > maxLatErr {
			maxLatErr = latErr
		}
		if lonErr > maxLonErr {
			maxLonErr = lonErr
		}

		if latErr > tolerance || lonErr > tolerance {
			failures++
			if failures <= 10 {
				t.Errorf("%s tdb=%.6f: latErr=%.6f° lonErr=%.6f° (want < %.4f°)\n  got:  lat=%.10f lon=%.10f\n  want: lat=%.10f lon=%.10f",
					tc.LocName, tc.TDBJD, latErr, lonErr, tolerance,
					latDeg, lonDeg, tc.EclLatDeg, tc.EclLonDeg)
			}
		}
	}

	if failures > 0 {
		t.Errorf("%d failures out of %d tests", failures, len(golden.Tests))
	}
	t.Logf("GeodeticToEcliptic: tested=%d failed=%d maxLatErr=%.6f° maxLonErr=%.6f°",
		len(golden.Tests), failures, maxLatErr, maxLonErr)
}

func BenchmarkGeodeticToICRF(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GeodeticToICRF(40.0, -74.0, 2451545.0)
	}
}
