package spk

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"testing"
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

	const tol = 0.2 // km tolerance (Mercury body-chain differs most from Skyfield)
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

func TestOpenNonType2(t *testing.T) {
	// Create a minimal SPK-like file with a non-Type-2 segment to exercise that error path.
	// This requires crafting a valid DAF header with a segment that has dataType != 2.
	// We'll write a proper 1024-byte file record + summary record.
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
	// ints: target=10, center=0, frame=1, dataType=3 (NOT type 2), startI=1, endI=100
	intOff := soff + 16 // after 2 doubles
	binary.LittleEndian.PutUint32(buf[intOff:], 10)     // target
	binary.LittleEndian.PutUint32(buf[intOff+4:], 0)    // center
	binary.LittleEndian.PutUint32(buf[intOff+8:], 1)    // frame
	binary.LittleEndian.PutUint32(buf[intOff+12:], 3)   // dataType = 3 (not supported)
	binary.LittleEndian.PutUint32(buf[intOff+16:], 1)   // startI
	binary.LittleEndian.PutUint32(buf[intOff+20:], 100) // endI

	f, err := os.CreateTemp("", "type3spk*.bsp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Write(buf)
	f.Close()

	_, err = Open(f.Name())
	if err == nil {
		t.Fatal("expected error for non-Type-2 SPK segment")
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
