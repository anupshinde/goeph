package spk

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/anupshinde/goeph/coord"
)

const bspPath = "../data/de440s.bsp"

// goldenSPK matches the JSON structure in testdata/golden_spk.json.
type goldenSPK struct {
	Tests []struct {
		TDBJD  float64    `json:"tdb_jd"`
		BodyID int        `json:"body_id"`
		PosKm  [3]float64 `json:"pos_km"`
	} `json:"tests"`
}

func loadGoldenSPK(t *testing.T) goldenSPK {
	t.Helper()
	data, err := os.ReadFile("../testdata/golden_spk.json")
	if err != nil {
		t.Fatal(err)
	}
	var g goldenSPK
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatal(err)
	}
	return g
}

func openEph(t *testing.T) *SPK {
	t.Helper()
	eph, err := Open(bspPath)
	if err != nil {
		t.Fatal(err)
	}
	return eph
}

func TestOpen(t *testing.T) {
	eph := openEph(t)
	if len(eph.segments) == 0 {
		t.Fatal("expected segments, got none")
	}
	if len(eph.segMap) == 0 {
		t.Fatal("expected segMap entries, got none")
	}
	if len(eph.chains) == 0 {
		t.Fatal("expected chains, got none")
	}
}

func TestOpenInvalidPath(t *testing.T) {
	_, err := Open("/nonexistent/file.bsp")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestOpenInvalidFile(t *testing.T) {
	f, err := os.CreateTemp("", "notspk*.bsp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Write(make([]byte, 2048))
	f.Close()

	_, err = Open(f.Name())
	if err == nil {
		t.Fatal("expected error for invalid SPK file")
	}
}

func TestObserveGolden(t *testing.T) {
	eph := openEph(t)
	golden := loadGoldenSPK(t)

	const tol = 0.01 // km tolerance (with TDB-TT correction; Mercury worst case ~0.002 km)
	failures := 0
	for i, tc := range golden.Tests {
		pos := eph.Observe(tc.BodyID, tc.TDBJD)
		for j := 0; j < 3; j++ {
			diff := math.Abs(pos[j] - tc.PosKm[j])
			if diff > tol {
				if failures < 10 {
					t.Errorf("test %d: body=%d tdb=%.6f axis=%d: got %.6f want %.6f diff=%.6f",
						i, tc.BodyID, tc.TDBJD, j, pos[j], tc.PosKm[j], diff)
				}
				failures++
			}
		}
	}
	if failures > 0 {
		t.Errorf("%d total axis failures out of %d tests (tol=%.0e km)", failures, len(golden.Tests), tol)
	}
}

func TestGeocentricPosition(t *testing.T) {
	eph := openEph(t)
	tdbJD := 2451545.0
	geo := eph.GeocentricPosition(Sun, tdbJD)
	obs := eph.Observe(Sun, tdbJD)

	dist := math.Sqrt(geo[0]*geo[0] + geo[1]*geo[1] + geo[2]*geo[2])
	if dist < 1e6 {
		t.Errorf("Sun distance too small: %.0f km", dist)
	}

	diff := math.Sqrt(
		(geo[0]-obs[0])*(geo[0]-obs[0]) +
			(geo[1]-obs[1])*(geo[1]-obs[1]) +
			(geo[2]-obs[2])*(geo[2]-obs[2]))
	if diff < 1.0 || diff > 1e5 {
		t.Errorf("light-time correction diff out of range: %.3f km", diff)
	}
}

func TestBodyWrtSSB_AllBodies(t *testing.T) {
	eph := openEph(t)
	tdbJD := 2451545.0

	bodies := []int{
		MercuryBarycenter, VenusBarycenter, EarthMoonBary,
		MarsBarycenter, JupiterBarycenter, SaturnBarycenter,
		UranusBarycenter, NeptuneBarycenter, PlutoBarycenter,
		Sun, Moon, Earth, Mercury, Venus,
	}

	for _, body := range bodies {
		pos := eph.bodyWrtSSB(body, tdbJD)
		dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
		if dist == 0 {
			t.Errorf("body %d: zero distance from SSB", body)
		}
	}
}

func TestBodyWrtSSB_UnsupportedPanics(t *testing.T) {
	eph := openEph(t)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for unsupported body")
		}
	}()
	eph.bodyWrtSSB(999, 2451545.0)
}

func TestSegPosition_MissingSegmentPanics(t *testing.T) {
	eph := openEph(t)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for missing segment")
		}
	}()
	eph.segPosition(999, 888, 2451545.0)
}

func TestChebyshev(t *testing.T) {
	if v := chebyshev([]float64{5.0}, 0.7); v != 5.0 {
		t.Errorf("single coeff: got %f want 5.0", v)
	}
	if v := chebyshev(nil, 0.5); v != 0.0 {
		t.Errorf("nil coeffs: got %f want 0.0", v)
	}
	v := chebyshev([]float64{3.0, 2.0}, 0.5)
	want := 3.0 + 2.0*0.5
	if math.Abs(v-want) > 1e-15 {
		t.Errorf("two coeffs: got %f want %f", v, want)
	}
	v = chebyshev([]float64{1.0, 2.0, 3.0}, 0.5)
	want = 1.0 + 2.0*0.5 + 3.0*(2.0*0.25-1.0)
	if math.Abs(v-want) > 1e-14 {
		t.Errorf("three coeffs: got %f want %f", v, want)
	}
}

func TestAdd3(t *testing.T) {
	r := add3([3]float64{1, 2, 3}, [3]float64{4, 5, 6})
	if r != [3]float64{5, 7, 9} {
		t.Errorf("add3: got %v", r)
	}
}

func TestSub3(t *testing.T) {
	r := sub3([3]float64{4, 5, 6}, [3]float64{1, 2, 3})
	if r != [3]float64{3, 3, 3} {
		t.Errorf("sub3: got %v", r)
	}
}

func TestLength3(t *testing.T) {
	v := length3([3]float64{3, 4, 0})
	if math.Abs(v-5.0) > 1e-15 {
		t.Errorf("length3: got %f want 5.0", v)
	}
}

func TestSegPosition_BoundaryClamp(t *testing.T) {
	eph := openEph(t)
	pos := eph.bodyWrtSSB(Sun, 2300000.0)
	dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
	if dist == 0 {
		t.Error("clamped early date returned zero position")
	}

	pos = eph.bodyWrtSSB(Sun, 2550000.0)
	dist = math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
	if dist == 0 {
		t.Error("clamped late date returned zero position")
	}
}

func TestOpenUnsupportedType(t *testing.T) {
	// Create a minimal SPK-like file with an unsupported segment type to exercise that error path.
	buf := make([]byte, 3*recordLen)
	copy(buf[0:8], "DAF/SPK ")
	// ND=2, NI=6 (standard SPK)
	binary.LittleEndian.PutUint32(buf[8:12], 2)
	binary.LittleEndian.PutUint32(buf[12:16], 6)
	// FWARD=2 (summary records start at record 2)
	binary.LittleEndian.PutUint32(buf[76:80], 2)

	// Record 2: summary record
	off := recordLen
	// next record = 0 (no more summary records)
	// prev record = 0
	// nSummaries = 1
	binary.LittleEndian.PutUint64(buf[off+16:off+24], math.Float64bits(1.0))

	// First summary at offset 24
	soff := off + 24
	// 2 doubles (start_sec, end_sec) + 6 ints packed as 3 doubles
	// ints: target=10, center=0, frame=1, dataType=13 (unsupported), startI=1, endI=100
	intOff := soff + 16 // after 2 doubles
	binary.LittleEndian.PutUint32(buf[intOff:], 10)     // target
	binary.LittleEndian.PutUint32(buf[intOff+4:], 0)    // center
	binary.LittleEndian.PutUint32(buf[intOff+8:], 1)    // frame
	binary.LittleEndian.PutUint32(buf[intOff+12:], 13)  // dataType = 13 (unsupported)
	binary.LittleEndian.PutUint32(buf[intOff+16:], 1)   // startI
	binary.LittleEndian.PutUint32(buf[intOff+20:], 100) // endI

	f, err := os.CreateTemp("", "type13spk*.bsp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Write(buf)
	f.Close()

	_, err = Open(f.Name())
	if err == nil {
		t.Fatal("expected error for unsupported SPK segment type")
	}
}

func TestChainBuilding(t *testing.T) {
	eph := openEph(t)

	tests := []struct {
		body    int
		name    string
		wantLen int // number of links in chain
	}{
		{Sun, "Sun", 1},               // 10 → 0
		{MercuryBarycenter, "MBary", 1}, // 1 → 0
		{Mercury, "Mercury", 2},       // 199 → 1 → 0
		{Venus, "Venus", 2},           // 299 → 2 → 0
		{Moon, "Moon", 2},             // 301 → 3 → 0
		{Earth, "Earth", 2},           // 399 → 3 → 0
		{MarsBarycenter, "MarsBary", 1}, // 4 → 0
	}

	for _, tc := range tests {
		chain, ok := eph.chains[tc.body]
		if !ok {
			t.Errorf("%s (body %d): no chain found", tc.name, tc.body)
			continue
		}
		if len(chain) != tc.wantLen {
			t.Errorf("%s (body %d): chain length = %d, want %d",
				tc.name, tc.body, len(chain), tc.wantLen)
		}
	}
}

func TestChainReachesSSB(t *testing.T) {
	eph := openEph(t)
	for body, chain := range eph.chains {
		if len(chain) == 0 {
			t.Errorf("body %d: empty chain", body)
			continue
		}
		lastLink := chain[len(chain)-1]
		if lastLink.center != SSB {
			t.Errorf("body %d: chain does not reach SSB; last center = %d", body, lastLink.center)
		}
	}
}

// goldenVelocity matches the JSON structure in testdata/golden_velocity.json.
type goldenVelocity struct {
	Tests []struct {
		TDBJD    float64    `json:"tdb_jd"`
		BodyID   int        `json:"body_id"`
		VelKmDay [3]float64 `json:"vel_km_day"`
	} `json:"tests"`
}

func loadGoldenVelocity(t *testing.T) goldenVelocity {
	t.Helper()
	data, err := os.ReadFile("../testdata/golden_velocity.json")
	if err != nil {
		t.Fatal(err)
	}
	var g goldenVelocity
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatal(err)
	}
	return g
}

// goldenApparent matches the JSON structure in testdata/golden_apparent.json.
type goldenApparent struct {
	Tests []struct {
		TDBJD  float64    `json:"tdb_jd"`
		BodyID int        `json:"body_id"`
		PosKm  [3]float64 `json:"pos_km"`
	} `json:"tests"`
}

func loadGoldenApparent(t *testing.T) goldenApparent {
	t.Helper()
	data, err := os.ReadFile("../testdata/golden_apparent.json")
	if err != nil {
		t.Fatal(err)
	}
	var g goldenApparent
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestVelocityGolden(t *testing.T) {
	eph := openEph(t)
	golden := loadGoldenVelocity(t)

	// Skyfield's astrometric.velocity is the rate of change of the astrometric
	// position vector: body_vel(t - lightTime) - earth_vel(t).
	// We use observe() to get the light time, then compute light-time corrected velocity.
	const tol = 0.01 // km/day tolerance (with TDB-TT correction; measured max ~0.0002 km/day)
	failures := 0
	for i, tc := range golden.Tests {
		_, lightTime := eph.observe(Earth, tc.BodyID, tc.TDBJD)
		bodyVel := eph.bodyVelWrtSSB(tc.BodyID, tc.TDBJD-lightTime)
		earthVel := eph.bodyVelWrtSSB(Earth, tc.TDBJD)
		geoVel := sub3(bodyVel, earthVel)

		for j := 0; j < 3; j++ {
			diff := math.Abs(geoVel[j] - tc.VelKmDay[j])
			if diff > tol {
				if failures < 10 {
					t.Errorf("test %d: body=%d tdb=%.6f axis=%d: got %.3f want %.3f diff=%.3f km/day",
						i, tc.BodyID, tc.TDBJD, j, geoVel[j], tc.VelKmDay[j], diff)
				}
				failures++
			}
		}
	}
	if failures > 0 {
		t.Errorf("%d total axis failures out of %d tests (tol=%.0f km/day)", failures, len(golden.Tests), tol)
	}
}

func TestApparentGolden(t *testing.T) {
	eph := openEph(t)
	golden := loadGoldenApparent(t)

	// Apparent positions differ from Skyfield due to:
	// - Underlying astrometric positions differ by up to 0.2 km
	// - Aberration (~20 arcsec rotation) amplifies direction errors at large distances
	// - goeph uses 30-term nutation vs Skyfield's 687-term (affects Earth velocity direction)
	//
	// The angular error from 30-term nutation grows with centuries from J2000 and
	// produces absolute km errors proportional to distance. We use a combined
	// tolerance: max(absTol, relTol * distance) where relTol ~5e-6 (~1 arcsec).
	const absTol = 50.0  // km — covers nearby bodies (Moon, Sun)
	const relTol = 1.5e-5 // fractional — covers ~3 arcsec angular error for distant bodies far from J2000
	failures := 0
	maxDiff := 0.0
	maxRelDiff := 0.0
	for i, tc := range golden.Tests {
		pos := eph.Apparent(tc.BodyID, tc.TDBJD)
		dist := math.Sqrt(tc.PosKm[0]*tc.PosKm[0] + tc.PosKm[1]*tc.PosKm[1] + tc.PosKm[2]*tc.PosKm[2])
		tol := absTol
		if rt := relTol * dist; rt > tol {
			tol = rt
		}
		for j := 0; j < 3; j++ {
			diff := math.Abs(pos[j] - tc.PosKm[j])
			if diff > maxDiff {
				maxDiff = diff
			}
			if dist > 0 {
				rd := diff / dist
				if rd > maxRelDiff {
					maxRelDiff = rd
				}
			}
			if diff > tol {
				if failures < 10 {
					t.Errorf("test %d: body=%d tdb=%.6f axis=%d: got %.6f want %.6f diff=%.6f tol=%.1f",
						i, tc.BodyID, tc.TDBJD, j, pos[j], tc.PosKm[j], diff, tol)
				}
				failures++
			}
		}
	}
	t.Logf("max apparent position diff: %.3f km (max relative: %.2e)", maxDiff, maxRelDiff)
	if failures > 0 {
		t.Errorf("%d total axis failures out of %d tests", failures, len(golden.Tests))
	}
}

func TestChebyshevDerivative(t *testing.T) {
	// f(x) = 5.0 (constant) → f'(x) = 0
	if v := chebyshevDerivative([]float64{5.0}, 0.5); v != 0.0 {
		t.Errorf("constant: got %f want 0.0", v)
	}
	// nil → 0
	if v := chebyshevDerivative(nil, 0.5); v != 0.0 {
		t.Errorf("nil: got %f want 0.0", v)
	}
	// f(x) = 3 + 2*T1(x) = 3 + 2x → f'(x) = 2
	v := chebyshevDerivative([]float64{3.0, 2.0}, 0.5)
	if math.Abs(v-2.0) > 1e-15 {
		t.Errorf("linear: got %f want 2.0", v)
	}
	// f(x) = 1 + 2*T1(x) + 3*T2(x) = 1 + 2x + 3*(2x^2-1) = -2 + 2x + 6x^2
	// f'(x) = 2 + 12x
	v = chebyshevDerivative([]float64{1.0, 2.0, 3.0}, 0.5)
	want := 2.0 + 12.0*0.5 // = 8.0
	if math.Abs(v-want) > 1e-14 {
		t.Errorf("quadratic at 0.5: got %f want %f", v, want)
	}
	// Same polynomial at x = -0.3
	v = chebyshevDerivative([]float64{1.0, 2.0, 3.0}, -0.3)
	want = 2.0 + 12.0*(-0.3) // = -1.6
	if math.Abs(v-want) > 1e-14 {
		t.Errorf("quadratic at -0.3: got %f want %f", v, want)
	}
	// f(x) = 1 + 2*T1 + 3*T2 + 4*T3 = 1 + 2x + 3*(2x^2-1) + 4*(4x^3-3x)
	// f(x) = -2 - 10x + 6x^2 + 16x^3
	// f'(x) = -10 + 12x + 48x^2
	v = chebyshevDerivative([]float64{1.0, 2.0, 3.0, 4.0}, 0.5)
	want = -10.0 + 12.0*0.5 + 48.0*0.25 // = -10 + 6 + 12 = 8
	if math.Abs(v-want) > 1e-13 {
		t.Errorf("cubic at 0.5: got %f want %f", v, want)
	}
}

func TestEarthVelocity_Sanity(t *testing.T) {
	eph := openEph(t)
	tdbJD := 2451545.0 // J2000.0

	vel := eph.EarthVelocity(tdbJD)
	speed := length3(vel) // km/day

	// Earth's orbital speed is ~29.78 km/s ≈ 2,572,992 km/day
	speedKmPerSec := speed / secPerDay
	if speedKmPerSec < 25 || speedKmPerSec > 35 {
		t.Errorf("Earth speed: %.2f km/s, expected ~29.78 km/s", speedKmPerSec)
	}
}

func TestVelocityAllBodies(t *testing.T) {
	eph := openEph(t)
	tdbJD := 2451545.0

	bodies := []struct {
		id       int
		name     string
		minKmSec float64 // minimum expected speed in km/s
		maxKmSec float64
	}{
		{Sun, "Sun", 0.001, 2.0},          // Sun moves slowly around SSB
		{Mercury, "Mercury", 30, 60},       // Mercury: fast, eccentric
		{Venus, "Venus", 30, 40},           // Venus: ~35 km/s
		{Earth, "Earth", 25, 35},           // Earth: ~30 km/s
		{Moon, "Moon", 25, 36},             // Moon: similar to Earth + ~1 km/s
		{MarsBarycenter, "Mars", 20, 30},   // Mars: ~24 km/s
	}

	for _, tc := range bodies {
		vel := eph.bodyVelWrtSSB(tc.id, tdbJD)
		speed := length3(vel) / secPerDay // km/s
		if speed < tc.minKmSec || speed > tc.maxKmSec {
			t.Errorf("%s: speed %.2f km/s outside [%.0f, %.0f]",
				tc.name, speed, tc.minKmSec, tc.maxKmSec)
		}
	}
}

func TestObserveFromMatchesObserve(t *testing.T) {
	eph := openEph(t)
	tdbJD := 2451545.0

	bodies := []int{Sun, Moon, Mercury, Venus, MarsBarycenter}
	for _, body := range bodies {
		obs := eph.Observe(body, tdbJD)
		from := eph.ObserveFrom(Earth, body, tdbJD)
		for j := 0; j < 3; j++ {
			if obs[j] != from[j] {
				t.Errorf("body %d axis %d: Observe=%.6f ObserveFrom=%.6f",
					body, j, obs[j], from[j])
			}
		}
	}
}

func TestApparentVsObserve(t *testing.T) {
	eph := openEph(t)
	tdbJD := 2451545.0

	bodies := []int{Sun, Moon, Mercury, Venus, MarsBarycenter}
	for _, body := range bodies {
		obs := eph.Observe(body, tdbJD)
		app := eph.Apparent(body, tdbJD)

		// Apparent should differ from astrometric due to aberration + deflection
		diff := length3(sub3(app, obs))
		obsDist := length3(obs)

		// Aberration shifts directions by ~20 arcsec ≈ 1e-4 radians.
		// At 1 AU (~1.5e8 km), that's ~15,000 km offset.
		// Diff should be nonzero but small relative to distance (< 0.1%).
		if diff == 0 {
			t.Errorf("body %d: apparent == astrometric (no correction applied)", body)
		}
		if diff > obsDist*1e-3 {
			t.Errorf("body %d: apparent-astrometric diff %.1f km too large (dist=%.0f km)",
				body, diff, obsDist)
		}
	}
}

func TestApparentFromMatchesApparent(t *testing.T) {
	eph := openEph(t)
	tdbJD := 2451545.0

	for _, body := range []int{Sun, Moon, MarsBarycenter} {
		app := eph.Apparent(body, tdbJD)
		from := eph.ApparentFrom(Earth, body, tdbJD)
		for j := 0; j < 3; j++ {
			if app[j] != from[j] {
				t.Errorf("body %d axis %d: Apparent=%.6f ApparentFrom=%.6f",
					body, j, app[j], from[j])
			}
		}
	}
}

func BenchmarkObserve(b *testing.B) {
	eph, err := Open(bspPath)
	if err != nil {
		b.Fatal(err)
	}
	tdbJD := 2451545.0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eph.Observe(Sun, tdbJD)
	}
}

func BenchmarkApparent(b *testing.B) {
	eph, err := Open(bspPath)
	if err != nil {
		b.Fatal(err)
	}
	tdbJD := 2451545.0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eph.Apparent(Sun, tdbJD)
	}
}

// goldenAltaz matches the JSON structure in testdata/golden_altaz.json.
type goldenAltaz struct {
	Tests []struct {
		TDBJD   float64 `json:"tdb_jd"`
		UT1JD   float64 `json:"ut1_jd"`
		BodyID  int     `json:"body_id"`
		LocName string  `json:"loc_name"`
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
		AltDeg  float64 `json:"alt_deg"`
		AzDeg   float64 `json:"az_deg"`
		DistKm  float64 `json:"dist_km"`
	} `json:"tests"`
}

func loadGoldenAltaz(t *testing.T) goldenAltaz {
	t.Helper()
	data, err := os.ReadFile("../testdata/golden_altaz.json")
	if err != nil {
		t.Fatal(err)
	}
	var g goldenAltaz
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestAltazGolden(t *testing.T) {
	eph := openEph(t)
	golden := loadGoldenAltaz(t)

	// The altaz rotation chain (precession + nutation + GAST + local horizon)
	// differs from Skyfield primarily due to 30-term vs 687-term nutation.
	// This affects Earth rotation angle, producing angular errors that grow
	// with centuries from J2000. Expected tolerance: ~0.1° near J2000,
	// growing to ~1° at the extremes of the date range.
	const altTol = 1.0 // degrees
	const azTol = 1.0  // degrees
	altFailures := 0
	azFailures := 0
	maxAltErr := 0.0
	maxAzErr := 0.0

	for i, tc := range golden.Tests {
		pos := eph.Apparent(tc.BodyID, tc.TDBJD)
		alt, az, _ := coord.Altaz(pos, tc.Lat, tc.Lon, tc.UT1JD)

		altErr := math.Abs(alt - tc.AltDeg)
		// Azimuth wraps around 360°
		azErr := math.Abs(az - tc.AzDeg)
		if azErr > 180 {
			azErr = 360 - azErr
		}

		if altErr > maxAltErr {
			maxAltErr = altErr
		}
		if azErr > maxAzErr {
			maxAzErr = azErr
		}

		if altErr > altTol {
			if altFailures < 5 {
				t.Errorf("test %d: body=%d loc=%s tdb=%.3f: alt got=%.4f want=%.4f err=%.4f°",
					i, tc.BodyID, tc.LocName, tc.TDBJD, alt, tc.AltDeg, altErr)
			}
			altFailures++
		}
		// Skip azimuth check when altitude is near ±90° (azimuth is undefined at zenith/nadir)
		if math.Abs(tc.AltDeg) < 85.0 && azErr > azTol {
			if azFailures < 5 {
				t.Errorf("test %d: body=%d loc=%s tdb=%.3f: az got=%.4f want=%.4f err=%.4f°",
					i, tc.BodyID, tc.LocName, tc.TDBJD, az, tc.AzDeg, azErr)
			}
			azFailures++
		}
	}

	t.Logf("altaz golden: %d tests, maxAltErr=%.4f° maxAzErr=%.4f°", len(golden.Tests), maxAltErr, maxAzErr)
	if altFailures > 0 {
		t.Errorf("%d altitude failures out of %d tests (tol=%.1f°)", altFailures, len(golden.Tests), altTol)
	}
	if azFailures > 0 {
		t.Errorf("%d azimuth failures out of %d tests (tol=%.1f°)", azFailures, len(golden.Tests), azTol)
	}
}
