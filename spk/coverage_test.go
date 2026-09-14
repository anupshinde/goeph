package spk

import (
	"math"
	"testing"
)

// de440s.bsp covers 1849-12-26 to 2150-01-22 (JPL).
const (
	de440sStartJD = 2396752.5
	de440sEndJD   = 2506352.5
)

func TestCoverage_DE440s(t *testing.T) {
	eph := openEph(t)
	for _, body := range []int{Sun, Moon, Earth, Mercury, Venus, MarsBarycenter, PlutoBarycenter} {
		start, end, ok := eph.Coverage(body)
		if !ok || math.Abs(start-de440sStartJD) > 1e-6 || math.Abs(end-de440sEndJD) > 1e-6 {
			t.Errorf("body %d: coverage %.6f–%.6f ok %v, want %.1f–%.1f", body, start, end, ok, de440sStartJD, de440sEndJD)
		}
	}
	if _, _, ok := eph.Coverage(chironID); ok {
		t.Error("Chiron coverage ok without its file")
	}
}

func TestCovers_Edges(t *testing.T) {
	eph := openEph(t)
	const minute = 1.0 / 1440
	cases := []struct {
		jd   float64
		want bool
	}{
		{de440sStartJD - 1, false},
		{de440sStartJD - minute, false},
		{de440sStartJD + minute, true},
		{2451545.0, true},
		{de440sEndJD - minute, true},
		{de440sEndJD + minute, false},
		{de440sEndJD + 1, false},
	}
	for _, c := range cases {
		for _, body := range []int{Earth, Moon, PlutoBarycenter} {
			if got := eph.Covers(body, c.jd); got != c.want {
				t.Errorf("Covers(%d, %.6f) = %v, want %v", body, c.jd, got, c.want)
			}
		}
	}
	if !eph.Covers(SSB, 0) {
		t.Error("SSB not covered")
	}
	if eph.Covers(chironID, 2451545.0) {
		t.Error("Chiron covered without its file")
	}
}

func TestCoverage_StackedFiles(t *testing.T) {
	eph, err := OpenMultiple(append([]string{bspPath}, chironBSPs...)...)
	if err != nil {
		t.Fatal(err)
	}
	// Chiron's files span 1600–2500, but its chain runs through the Sun's
	// de440s segment.
	start, end, ok := eph.Coverage(chironID)
	if !ok || math.Abs(start-de440sStartJD) > 1e-6 || math.Abs(end-de440sEndJD) > 1e-6 {
		t.Errorf("Chiron coverage %.6f–%.6f ok %v, want the de440s span", start, end, ok)
	}
	if !eph.Covers(chironID, 2488069.5) { // 2100-01-01, the file boundary
		t.Error("Chiron not covered at the file boundary")
	}
}

// A gap between stacked segments lies inside Coverage but not Covers.
func TestCovers_Gap(t *testing.T) {
	day := func(jd float64) float64 { return (jd - j2000JD) * secPerDay }
	s := &SPK{
		segMap: map[[2]int][]*segment{
			{5, 0}: {
				{target: 5, center: 0, startSec: day(2451545), endSec: day(2451555)},
				{target: 5, center: 0, startSec: day(2451565), endSec: day(2451575)},
			},
		},
		chains: map[int][]chainLink{5: {{target: 5, center: 0}}},
	}
	start, end, ok := s.Coverage(5)
	if !ok || math.Abs(start-2451545) > 1e-6 || math.Abs(end-2451575) > 1e-6 {
		t.Errorf("coverage %.6f–%.6f ok %v, want 2451545–2451575", start, end, ok)
	}
	for jd, want := range map[float64]bool{2451550: true, 2451560: false, 2451570: true, 2451580: false} {
		if got := s.Covers(5, jd); got != want {
			t.Errorf("Covers(%.1f) = %v, want %v", jd, got, want)
		}
	}
}
