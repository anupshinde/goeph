package coord

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

// goldenEclipticOfDate matches testdata/golden_ecliptic_of_date.json.
type goldenEclipticOfDate struct {
	Tests []struct {
		TTJD      float64    `json:"tt_jd"`
		ICRFKM    [3]float64 `json:"icrf_km"`
		EclLatDeg float64    `json:"ecl_lat_deg"`
		EclLonDeg float64    `json:"ecl_lon_deg"`
	} `json:"tests"`
}

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

// goldenERA matches testdata/golden_era.json.
type goldenERA struct {
	Tests []struct {
		UT1JD  float64 `json:"ut1_jd"`
		ERADeg float64 `json:"era_deg"`
	} `json:"tests"`
}

func TestEarthRotationAngle_J2000(t *testing.T) {
	era := EarthRotationAngle(j2000JD)
	// ERA at J2000 = 0.7790572732640 * 360 = 280.46... degrees
	expected := 0.7790572732640 * 360.0
	expected = math.Mod(expected+math.Mod(j2000JD, 1.0)*360.0, 360.0)
	if math.Abs(era-expected) > 1e-6 {
		t.Errorf("ERA at J2000: got %f, want ~%f", era, expected)
	}
}

func TestEarthRotationAngle_Range(t *testing.T) {
	// ERA should always be in [0, 360)
	for _, jd := range []float64{j2000JD, j2000JD + 0.5, j2000JD - 1000, j2000JD + 50000} {
		era := EarthRotationAngle(jd)
		if era < 0 || era >= 360 {
			t.Errorf("ERA(%.1f) = %f, out of [0, 360)", jd, era)
		}
	}
}

func TestEarthRotationAngle_Golden(t *testing.T) {
	var golden goldenERA
	loadJSON(t, "../testdata/golden_era.json", &golden)

	const tol = 1e-8 // degrees; should match Skyfield exactly (identical formula)
	failures := 0
	for i, tc := range golden.Tests {
		got := EarthRotationAngle(tc.UT1JD)
		diff := got - tc.ERADeg
		if diff > 180 {
			diff -= 360
		} else if diff < -180 {
			diff += 360
		}
		if math.Abs(diff) > tol {
			if failures < 10 {
				t.Errorf("test %d: ut1=%.6f got=%.10f want=%.10f diff=%.2e",
					i, tc.UT1JD, got, tc.ERADeg, diff)
			}
			failures++
		}
	}
	if failures > 0 {
		t.Errorf("%d ERA failures out of %d tests (tol=%.0e°)", failures, len(golden.Tests), tol)
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

func TestMeanObliquityDeg(t *testing.T) {
	// At J2000 epoch, mean obliquity ≈ 23.4393°
	got := MeanObliquityDeg(j2000JD)
	if math.Abs(got-23.4393) > 0.001 {
		t.Errorf("MeanObliquityDeg(J2000) = %.6f, want ~23.4393", got)
	}

	// Should decrease over centuries
	later := MeanObliquityDeg(j2000JD + 36525.0) // 1 century later
	if later >= got {
		t.Errorf("obliquity should decrease: J2000=%.6f, +100y=%.6f", got, later)
	}
}

func TestMeanObliquityDeg_MatchesInternal(t *testing.T) {
	// Verify the exported function matches the internal function
	for _, jd := range []float64{j2000JD, j2000JD - 36525, j2000JD + 18262.5} {
		T := (jd - j2000JD) / 36525.0
		want := meanObliquity(T) * rad2deg
		got := MeanObliquityDeg(jd)
		if got != want {
			t.Errorf("jd=%.1f: MeanObliquityDeg=%.15f, internal=%.15f", jd, got, want)
		}
	}
}

func TestNutationDeg(t *testing.T) {
	dpsi, deps := NutationDeg(j2000JD)
	// Nutation at J2000 should be small (within a few arcseconds = fractions of a degree)
	if math.Abs(dpsi) > 0.01 || math.Abs(deps) > 0.01 {
		t.Errorf("NutationDeg(J2000) = (%.6f, %.6f), want small values", dpsi, deps)
	}
	if dpsi == 0 && deps == 0 {
		t.Error("NutationDeg(J2000) is exactly zero (unexpected)")
	}
}

func TestNutationDeg_MatchesInternal(t *testing.T) {
	for _, jd := range []float64{j2000JD, j2000JD + 10000, j2000JD - 20000} {
		T := (jd - j2000JD) / 36525.0
		wantDpsi, wantDeps := nutationAngles(T)
		gotDpsi, gotDeps := NutationDeg(jd)
		if math.Abs(gotDpsi-wantDpsi*rad2deg) > 1e-15 {
			t.Errorf("jd=%.1f: dpsi mismatch", jd)
		}
		if math.Abs(gotDeps-wantDeps*rad2deg) > 1e-15 {
			t.Errorf("jd=%.1f: deps mismatch", jd)
		}
	}
}

func TestTrueObliquityDeg(t *testing.T) {
	// True obliquity = mean + nutation in obliquity
	jd := j2000JD
	mean := MeanObliquityDeg(jd)
	_, deps := NutationDeg(jd)
	want := mean + deps
	got := TrueObliquityDeg(jd)
	if math.Abs(got-want) > 1e-14 {
		t.Errorf("TrueObliquityDeg(J2000) = %.15f, want %.15f", got, want)
	}
}

func TestTrueObliquityDeg_DiffersFromMean(t *testing.T) {
	jd := j2000JD + 5000
	mean := MeanObliquityDeg(jd)
	trueObl := TrueObliquityDeg(jd)
	if mean == trueObl {
		t.Error("true obliquity should differ from mean (nutation != 0)")
	}
	// Difference should be small (nutation in obliquity < ~10 arcsec ≈ 0.003°)
	if math.Abs(trueObl-mean) > 0.01 {
		t.Errorf("true-mean = %.6f°, too large", trueObl-mean)
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

func TestNutationPrecisionToggle(t *testing.T) {
	// Verify the getter/setter works
	original := GetNutationPrecision()
	defer SetNutationPrecision(original)

	SetNutationPrecision(NutationStandard)
	if GetNutationPrecision() != NutationStandard {
		t.Error("expected NutationStandard")
	}
	SetNutationPrecision(NutationFull)
	if GetNutationPrecision() != NutationFull {
		t.Error("expected NutationFull")
	}
}

func TestNutationModesAgree(t *testing.T) {
	// Both modes should agree to within ~1 arcsec
	original := GetNutationPrecision()
	defer SetNutationPrecision(original)

	for _, T := range []float64{0, 0.5, -1.0, 1.5} {
		SetNutationPrecision(NutationStandard)
		dpsiStd, depsStd := nutationAngles(T)

		SetNutationPrecision(NutationFull)
		dpsiFull, depsFull := nutationAngles(T)

		dpsiDiffArcsec := math.Abs(dpsiStd-dpsiFull) / arcsec2rad
		depsDiffArcsec := math.Abs(depsStd-depsFull) / arcsec2rad

		if dpsiDiffArcsec > 1.5 {
			t.Errorf("T=%.1f: dpsi diff = %.4f arcsec (want < 1.5)", T, dpsiDiffArcsec)
		}
		if depsDiffArcsec > 1.5 {
			t.Errorf("T=%.1f: deps diff = %.4f arcsec (want < 1.5)", T, depsDiffArcsec)
		}
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

func BenchmarkNutationAnglesStandard(b *testing.B) {
	original := GetNutationPrecision()
	defer SetNutationPrecision(original)
	SetNutationPrecision(NutationStandard)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nutationAngles(0.5)
	}
}

func BenchmarkNutationAnglesFull(b *testing.B) {
	original := GetNutationPrecision()
	defer SetNutationPrecision(original)
	SetNutationPrecision(NutationFull)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nutationAngles(0.5)
	}
}

// TestGeodeticToEcliptic_Golden tests the full geodetic→ICRF→ecliptic pipeline
// against Skyfield golden data. This is the end-to-end accuracy test.
// Residual error (~0.02°) is mainly from Skyfield's observe() applying
// light-time correction to surface locations.
func TestGeodeticToEcliptic_Golden(t *testing.T) {
	// Use full nutation for golden test comparison against Skyfield
	original := GetNutationPrecision()
	defer SetNutationPrecision(original)
	SetNutationPrecision(NutationFull)

	var golden goldenLocations
	loadJSON(t, "../testdata/golden_locations.json", &golden)

	const tolerance = 0.025 // degrees (residual from light-time correction in Skyfield's observe())

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

func TestAltaz_Zenith(t *testing.T) {
	// A point directly at the zenith should have altitude ~90°.
	// GeodeticToICRF gives the ICRF direction of a ground point.
	// Altaz of that direction from the same location should be nearly overhead.
	lat, lon := 40.0, -74.0
	jd := j2000JD

	x, y, z := GeodeticToICRF(lat, lon, jd)
	// Scale to some distance (doesn't matter for direction)
	pos := [3]float64{x * 1e6, y * 1e6, z * 1e6}

	alt, az, dist := Altaz(pos, lat, lon, jd)
	_ = az
	if math.Abs(alt-90.0) > 1.0 {
		t.Errorf("zenith altitude = %.4f°, want ~90°", alt)
	}
	if math.Abs(dist-1e6) > 1.0 {
		t.Errorf("distance = %.4f, want 1e6", dist)
	}
}

func TestAltaz_Horizon(t *testing.T) {
	// A point 90° away (in the equatorial plane) should be near the horizon.
	lat, lon := 0.0, 0.0
	jd := j2000JD

	// ICRF direction of (lat=0, lon=90) is roughly 90° away in longitude
	x2, y2, z2 := GeodeticToICRF(0.0, 90.0, jd)
	pos := [3]float64{x2 * 1e6, y2 * 1e6, z2 * 1e6}

	alt, _, _ := Altaz(pos, lat, lon, jd)
	// Should be within a few degrees of the horizon (not exact due to precession/nutation)
	if math.Abs(alt) > 10.0 {
		t.Errorf("horizon point altitude = %.4f°, want near 0°", alt)
	}
}

func TestAltaz_AzimuthRange(t *testing.T) {
	// Azimuth should always be in [0, 360)
	jd := 2451545.0 + 365.25*10.0
	for _, lat := range []float64{-45, 0, 45, 90} {
		for _, lon := range []float64{-180, -90, 0, 90, 180} {
			pos := [3]float64{1e8, 2e8, 3e8}
			alt, az, _ := Altaz(pos, lat, lon, jd)
			_ = alt
			if az < 0 || az >= 360 {
				t.Errorf("lat=%.0f lon=%.0f: az=%.4f outside [0,360)", lat, lon, az)
			}
		}
	}
}

func TestHourAngleDec_OnMeridian(t *testing.T) {
	// An object on the local meridian should have HA near 0° (or 360°).
	// At J2000, the vernal equinox is on the meridian at GAST=0 for lon=0.
	// Use an object at RA=GAST, Dec=0.
	jd := j2000JD
	gast := GAST(jd)

	// Object at RA = GAST (true equinox of date)
	raDeg := gast
	raHours := raDeg / 15.0
	x, y, z := RADecToICRF(raHours, 0)

	// This is approximate since RADecToICRF gives J2000 RA/Dec, not true equinox.
	// The precession/nutation difference at J2000 is very small.
	pos := [3]float64{x, y, z}
	ha, dec := HourAngleDec(pos, 0, jd)
	_ = dec

	// HA should be near 0° (on the meridian)
	haWrapped := ha
	if haWrapped > 180 {
		haWrapped -= 360
	}
	if math.Abs(haWrapped) > 1.0 {
		t.Errorf("on-meridian HA = %.4f°, want ~0°", ha)
	}
}

func BenchmarkGeodeticToICRF(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GeodeticToICRF(40.0, -74.0, 2451545.0)
	}
}

func TestITRFToGeodetic_Roundtrip(t *testing.T) {
	// Test roundtrip: geodetic → ITRF → geodetic
	tests := []struct {
		lat, lon float64
	}{
		{0, 0},
		{45, 90},
		{-45, -90},
		{90, 0},    // north pole
		{-90, 180}, // south pole
		{51.5, -0.1},
		{-33.9, 151.2},
	}
	for _, tc := range tests {
		lat := tc.lat * deg2rad
		lon := tc.lon * deg2rad
		sinLat := math.Sin(lat)
		cosLat := math.Cos(lat)
		N := wgs84A / math.Sqrt(1.0-wgs84E2*sinLat*sinLat)
		x := N * cosLat * math.Cos(lon)
		y := N * cosLat * math.Sin(lon)
		z := N * (1.0 - wgs84E2) * sinLat

		gotLat, gotLon, gotH := ITRFToGeodetic(x, y, z)
		if math.Abs(gotLat-tc.lat) > 1e-10 {
			t.Errorf("lat=%.1f lon=%.1f: gotLat=%.12f, want %.1f", tc.lat, tc.lon, gotLat, tc.lat)
		}
		lonErr := math.Abs(gotLon - tc.lon)
		if lonErr > 180 {
			lonErr = 360 - lonErr
		}
		if lonErr > 1e-10 && math.Abs(tc.lat) < 89.99 { // skip lon check at poles
			t.Errorf("lat=%.1f lon=%.1f: gotLon=%.12f, want %.1f", tc.lat, tc.lon, gotLon, tc.lon)
		}
		if math.Abs(gotH) > 1e-6 { // surface point → height ≈ 0
			t.Errorf("lat=%.1f lon=%.1f: gotH=%.10f km, want ~0", tc.lat, tc.lon, gotH)
		}
	}
}

func TestITRFToGeodetic_Altitude(t *testing.T) {
	// Point 100 km above the equator at prime meridian
	alt := 100.0 // km
	lat := 0.0
	N := wgs84A / math.Sqrt(1.0-wgs84E2*math.Sin(lat)*math.Sin(lat))
	x := N + alt
	y := 0.0
	z := 0.0

	_, _, gotH := ITRFToGeodetic(x, y, z)
	if math.Abs(gotH-alt) > 1e-6 {
		t.Errorf("altitude: got %.10f km, want %.1f km", gotH, alt)
	}
}

func TestIsSunlit_InSunlight(t *testing.T) {
	// Object between Earth and Sun (closer to Earth) — should be sunlit
	sunPos := [3]float64{1.5e8, 0, 0} // Sun at ~1 AU
	objPos := [3]float64{42000, 0, 0}  // GEO orbit, same direction as Sun
	if !IsSunlit(objPos, sunPos) {
		t.Error("object in front of Earth toward Sun should be sunlit")
	}
}

func TestIsSunlit_InShadow(t *testing.T) {
	// Object directly behind Earth from Sun — should be in shadow
	sunPos := [3]float64{1.5e8, 0, 0}
	objPos := [3]float64{-42000, 0, 0} // opposite side of Earth from Sun
	if IsSunlit(objPos, sunPos) {
		t.Error("object behind Earth from Sun should be in shadow")
	}
}

func TestIsSunlit_FarFromShadow(t *testing.T) {
	// Object far above the ecliptic plane — should be sunlit
	sunPos := [3]float64{1.5e8, 0, 0}
	objPos := [3]float64{0, 0, 42000} // above north pole
	if !IsSunlit(objPos, sunPos) {
		t.Error("object far above ecliptic should be sunlit")
	}
}

func TestIsBehindEarth(t *testing.T) {
	observer := [3]float64{42000, 0, 0} // GEO, +X direction
	target := [3]float64{-42000, 0, 0}  // opposite side
	if !IsBehindEarth(observer, target) {
		t.Error("target on opposite side of Earth should be behind Earth")
	}

	// Target same direction as observer but farther — not behind Earth
	target2 := [3]float64{80000, 0, 0}
	if IsBehindEarth(observer, target2) {
		t.Error("target in same direction should not be behind Earth")
	}
}

func TestTEMEToICRF_PreservesMagnitude(t *testing.T) {
	// Rotation should preserve vector magnitude
	posTEME := [3]float64{6778.0, 1234.0, -3456.0} // typical LEO position, km
	jd := 2451545.0 + 365.25*10                     // 10 years from J2000

	posICRF := TEMEToICRF(posTEME, jd)

	magTEME := math.Sqrt(posTEME[0]*posTEME[0] + posTEME[1]*posTEME[1] + posTEME[2]*posTEME[2])
	magICRF := math.Sqrt(posICRF[0]*posICRF[0] + posICRF[1]*posICRF[1] + posICRF[2]*posICRF[2])

	if math.Abs(magICRF-magTEME) > 1e-10 {
		t.Errorf("magnitude changed: TEME=%.10f ICRF=%.10f", magTEME, magICRF)
	}
}

func TestTEMEToICRF_AtJ2000(t *testing.T) {
	// At J2000, precession=identity, nutation is small, eq_eq is small.
	// TEME and ICRF should nearly coincide.
	posTEME := [3]float64{6778.0, 0.0, 0.0}
	posICRF := TEMEToICRF(posTEME, j2000JD)

	// Difference should be very small (only nutation + eq_eq at T=0)
	diff := math.Sqrt(
		(posICRF[0]-posTEME[0])*(posICRF[0]-posTEME[0]) +
			(posICRF[1]-posTEME[1])*(posICRF[1]-posTEME[1]) +
			(posICRF[2]-posTEME[2])*(posICRF[2]-posTEME[2]))
	// At J2000, nutation is ~17 arcsec → ~0.56 km at 6778 km altitude
	if diff > 1.0 {
		t.Errorf("TEME≈ICRF at J2000 but diff=%.6f km", diff)
	}
}

func TestTEMEToICRF_ChangesWithTime(t *testing.T) {
	posTEME := [3]float64{6778.0, 1234.0, -3456.0}
	pos1 := TEMEToICRF(posTEME, j2000JD)
	pos2 := TEMEToICRF(posTEME, j2000JD+365.25*50) // 50 years later

	// Precession should cause a measurable difference
	diff := math.Sqrt(
		(pos1[0]-pos2[0])*(pos1[0]-pos2[0]) +
			(pos1[1]-pos2[1])*(pos1[1]-pos2[1]) +
			(pos1[2]-pos2[2])*(pos1[2]-pos2[2]))
	if diff < 1.0 {
		t.Errorf("TEME→ICRF unchanged after 50 years: diff=%.6f km", diff)
	}
}

func TestInertialFrame_Galactic(t *testing.T) {
	// Galactic InertialFrame should match ICRFToGalactic
	pos := [3]float64{1e8, -5e7, 2e7}
	lat1, lon1 := ICRFToGalactic(pos[0], pos[1], pos[2])
	lat2, lon2 := Galactic.LatLon(pos)

	if math.Abs(lat1-lat2) > 1e-12 || math.Abs(lon1-lon2) > 1e-12 {
		t.Errorf("Galactic frame mismatch: ICRFToGalactic=(%.10f,%.10f) frame=(%.10f,%.10f)",
			lat1, lon1, lat2, lon2)
	}
}

func TestInertialFrame_Ecliptic(t *testing.T) {
	// Ecliptic InertialFrame should match ICRFToEcliptic
	pos := [3]float64{1e8, -5e7, 2e7}
	lat1, lon1 := ICRFToEcliptic(pos[0], pos[1], pos[2])
	lat2, lon2 := Ecliptic.LatLon(pos)

	if math.Abs(lat1-lat2) > 1e-10 || math.Abs(lon1-lon2) > 1e-10 {
		t.Errorf("Ecliptic frame mismatch: ICRFToEcliptic=(%.10f,%.10f) frame=(%.10f,%.10f)",
			lat1, lon1, lat2, lon2)
	}
}

func TestInertialFrame_XYZ(t *testing.T) {
	// XYZ on identity frame should return the same vector
	identity := InertialFrame{
		Name:   "Identity",
		Matrix: [3][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
	}
	pos := [3]float64{1.0, 2.0, 3.0}
	result := identity.XYZ(pos)
	for i := 0; i < 3; i++ {
		if math.Abs(result[i]-pos[i]) > 1e-15 {
			t.Errorf("identity XYZ[%d]: got %f want %f", i, result[i], pos[i])
		}
	}
}

func TestInertialFrame_ZeroVector(t *testing.T) {
	lat, lon := Galactic.LatLon([3]float64{0, 0, 0})
	if lat != 0 || lon != 0 {
		t.Errorf("zero vector LatLon: got (%f, %f), want (0, 0)", lat, lon)
	}
}

func TestTimeBasedFrame_ITRF(t *testing.T) {
	// ITRF frame should rotate with Earth — two times 12h apart should differ
	itrf := ITRFFrame()
	pos := [3]float64{1e8, 0, 0}
	v1 := itrf.XYZ(pos, j2000JD)
	v2 := itrf.XYZ(pos, j2000JD+0.5) // 12 hours later

	dot := v1[0]*v2[0] + v1[1]*v2[1] + v1[2]*v2[2]
	mag := math.Sqrt(v1[0]*v1[0]+v1[1]*v1[1]+v1[2]*v1[2]) *
		math.Sqrt(v2[0]*v2[0]+v2[1]*v2[1]+v2[2]*v2[2])
	cosAngle := dot / mag

	// 12h = 180° rotation, so vectors should be roughly anti-parallel (cos ≈ -1)
	if cosAngle > -0.9 {
		t.Errorf("ITRF 12h apart: cos(angle)=%.4f, want ≈ -1", cosAngle)
	}
}

func TestTimeBasedFrame_ITRF_PreservesMagnitude(t *testing.T) {
	itrf := ITRFFrame()
	pos := [3]float64{6778.0, 1234.0, -3456.0}
	v := itrf.XYZ(pos, j2000JD+365.25*10)

	magIn := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
	magOut := math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
	if math.Abs(magOut-magIn) > 1e-8 {
		t.Errorf("ITRF magnitude changed: %.10f → %.10f", magIn, magOut)
	}
}

func TestTrueEclipticOfDateFrame_Golden(t *testing.T) {
	// Validate TrueEclipticOfDateFrame against Skyfield's ecliptic_latlon(epoch=t).
	// Use full nutation for best accuracy against Skyfield.
	original := GetNutationPrecision()
	defer SetNutationPrecision(original)
	SetNutationPrecision(NutationFull)

	var golden goldenEclipticOfDate
	loadJSON(t, "../testdata/golden_ecliptic_of_date.json", &golden)

	frame := TrueEclipticOfDateFrame()

	var maxLatErr, maxLonErr float64
	for i, tc := range golden.Tests {
		lat, lon := frame.LatLon(tc.ICRFKM, tc.TTJD)

		latErr := math.Abs(lat - tc.EclLatDeg)
		lonErr := math.Abs(lon - tc.EclLonDeg)
		// Handle 360° wraparound
		if lonErr > 180 {
			lonErr = 360 - lonErr
		}

		if latErr > maxLatErr {
			maxLatErr = latErr
		}
		if lonErr > maxLonErr {
			maxLonErr = lonErr
		}

		// Tolerance: 0.001° (~3.6 arcsec) — accounts for minor differences
		// in precession/nutation model details between goeph and Skyfield.
		if latErr > 0.001 || lonErr > 0.001 {
			t.Errorf("test %d (JD=%.1f): lat err=%.6f° lon err=%.6f°",
				i, tc.TTJD, latErr, lonErr)
		}
	}
	t.Logf("max errors: lat=%.6f° lon=%.6f° (%d tests)", maxLatErr, maxLonErr, len(golden.Tests))
}

func TestMeanEclipticOfDateFrame_AtJ2000(t *testing.T) {
	// At J2000.0, mean ecliptic of date should closely match the static
	// J2000 Ecliptic frame (the only difference is the ICRS→J2000 frame
	// bias, which is ~17 mas).
	meanEcl := MeanEclipticOfDateFrame()
	pos := [3]float64{1e8, -5e7, 2e7}

	lat1, lon1 := Ecliptic.LatLon(pos)
	lat2, lon2 := meanEcl.LatLon(pos, j2000JD)

	// Frame bias is ~17 mas ≈ 5e-6 degrees; allow 1e-4° for safety.
	if math.Abs(lat1-lat2) > 1e-4 || math.Abs(lon1-lon2) > 1e-4 {
		t.Errorf("mean ecliptic at J2000 vs static: Δlat=%.6f° Δlon=%.6f°",
			math.Abs(lat1-lat2), math.Abs(lon1-lon2))
	}
}

func TestTrueEclipticOfDateFrame_AtJ2000(t *testing.T) {
	// At J2000.0, true ecliptic should also be close to the static Ecliptic
	// (nutation is small but nonzero — up to ~9 arcsec ≈ 0.0025°).
	trueEcl := TrueEclipticOfDateFrame()
	pos := [3]float64{1e8, -5e7, 2e7}

	lat1, lon1 := Ecliptic.LatLon(pos)
	lat2, lon2 := trueEcl.LatLon(pos, j2000JD)

	if math.Abs(lat1-lat2) > 0.005 || math.Abs(lon1-lon2) > 0.005 {
		t.Errorf("true ecliptic at J2000 vs static: Δlat=%.6f° Δlon=%.6f°",
			math.Abs(lat1-lat2), math.Abs(lon1-lon2))
	}
}

func TestEclipticOfDateFrame_PrecessionGrows(t *testing.T) {
	// Ecliptic precession is ~50 arcsec/year ≈ 0.014°/year.
	// Over 20 years, longitude should shift ~0.28°.
	meanEcl := MeanEclipticOfDateFrame()
	pos := [3]float64{1e8, 0, 0} // on the X-axis (lon≈0 in J2000 ecliptic)

	_, lon0 := meanEcl.LatLon(pos, j2000JD)
	_, lon20 := meanEcl.LatLon(pos, j2000JD+365.25*20)

	shift := lon20 - lon0
	// Expect ~0.28° shift; allow 0.1–0.5° range.
	if shift < 0.1 || shift > 0.5 {
		t.Errorf("20-year longitude precession shift: %.4f°, expected ~0.28°", shift)
	}
}

func TestMeanVsTrueEclipticOfDate_Differ(t *testing.T) {
	// Mean and true should differ due to nutation.
	meanEcl := MeanEclipticOfDateFrame()
	trueEcl := TrueEclipticOfDateFrame()
	pos := [3]float64{1e8, -5e7, 2e7}
	jd := j2000JD + 365.25*10 // 10 years after J2000

	_, lonMean := meanEcl.LatLon(pos, jd)
	_, lonTrue := trueEcl.LatLon(pos, jd)

	diff := math.Abs(lonMean - lonTrue)
	// Nutation in longitude is up to ~17 arcsec ≈ 0.005°; should be nonzero.
	if diff < 1e-6 {
		t.Error("mean and true ecliptic longitudes are identical — nutation not applied")
	}
	// But should not be huge (< 0.01°).
	if diff > 0.01 {
		t.Errorf("mean vs true ecliptic longitude difference too large: %.6f°", diff)
	}
}

func TestEclipticOfDateFrame_PreservesMagnitude(t *testing.T) {
	meanEcl := MeanEclipticOfDateFrame()
	trueEcl := TrueEclipticOfDateFrame()
	pos := [3]float64{6778.0, 1234.0, -3456.0}
	jd := j2000JD + 365.25*10

	magIn := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])

	for _, tc := range []struct {
		name  string
		frame TimeBasedFrame
	}{
		{"MeanEclipticOfDate", meanEcl},
		{"TrueEclipticOfDate", trueEcl},
	} {
		v := tc.frame.XYZ(pos, jd)
		magOut := math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
		if math.Abs(magOut-magIn) > 1e-8 {
			t.Errorf("%s magnitude changed: %.10f → %.10f", tc.name, magIn, magOut)
		}
	}
}

func BenchmarkAltaz(b *testing.B) {
	pos := [3]float64{1.5e8, 0, 0}
	for i := 0; i < b.N; i++ {
		Altaz(pos, 40.0, -74.0, 2451545.0)
	}
}
