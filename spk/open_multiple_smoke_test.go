package spk

import (
	"math"
	"testing"
)

var chironBSPs = []string{
	"../data/chiron_1600_2100.bsp",
	"../data/chiron_2100_2500.bsp",
}

const chironID = 20002060

func TestOpenMultipleSingleFile(t *testing.T) {
	// OpenMultiple with one file should produce identical results to Open.
	eph, err := OpenMultiple(bspPath)
	if err != nil {
		t.Fatal(err)
	}
	ephSingle := openEph(t)

	if len(eph.segments) != len(ephSingle.segments) {
		t.Fatalf("segment count: got %d, want %d", len(eph.segments), len(ephSingle.segments))
	}

	// Spot-check positions match
	tdbJD := 2451545.0 // J2000
	for _, body := range []int{Mercury, Venus, MarsBarycenter, JupiterBarycenter, SaturnBarycenter} {
		pos := eph.Observe(body, tdbJD)
		want := ephSingle.Observe(body, tdbJD)
		for j := 0; j < 3; j++ {
			if pos[j] != want[j] {
				t.Errorf("body %d axis %d: got %v, want %v", body, j, pos[j], want[j])
			}
		}
	}
}

func TestOpenMultipleWithChiron(t *testing.T) {
	// Load planetary ephemeris + all small body SPK files together.
	files := append([]string{bspPath}, chironBSPs...)
	eph, err := OpenMultiple(files...)
	if err != nil {
		t.Fatal(err)
	}

	// Verify planets still work
	tdbJD := 2451545.0
	ephSingle := openEph(t)
	pos := eph.Observe(Mercury, tdbJD)
	want := ephSingle.Observe(Mercury, tdbJD)
	for j := 0; j < 3; j++ {
		if pos[j] != want[j] {
			t.Errorf("Mercury axis %d: got %v, want %v", j, pos[j], want[j])
		}
	}

	// Verify small body returns a non-zero position at J2000
	chiron := eph.Observe(chironID, tdbJD)
	dist := math.Sqrt(chiron[0]*chiron[0] + chiron[1]*chiron[1] + chiron[2]*chiron[2])
	if dist == 0 {
		t.Fatal("position is zero at J2000")
	}
	t.Logf("geocentric at J2000: [%.3f, %.3f, %.3f] km (dist=%.0f km)", chiron[0], chiron[1], chiron[2], dist)

	// Distance should be in a reasonable range (~7-20 AU = ~1e9 to ~3e9 km)
	if dist < 1e9 || dist > 4e9 {
		t.Errorf("distance %.0f km outside expected range", dist)
	}
}

func TestOpenMultipleChironBoundary(t *testing.T) {
	// Load adjacent files individually and together, verify they agree at the boundary.
	// Boundary is at ~2100, JD 2488069.5 = 2100-01-01
	boundaryJD := 2488069.5

	eph1, err := OpenMultiple(bspPath, "../data/chiron_1600_2100.bsp")
	if err != nil {
		t.Fatal(err)
	}
	eph2, err := OpenMultiple(bspPath, "../data/chiron_2100_2500.bsp")
	if err != nil {
		t.Fatal(err)
	}
	ephBoth, err := OpenMultiple(bspPath,
		"../data/chiron_1600_2100.bsp",
		"../data/chiron_2100_2500.bsp",
	)
	if err != nil {
		t.Fatal(err)
	}

	pos1 := eph1.Observe(chironID, boundaryJD)
	pos2 := eph2.Observe(chironID, boundaryJD)
	posBoth := ephBoth.Observe(chironID, boundaryJD)

	// All three should agree within sub-km precision
	for j := 0; j < 3; j++ {
		diff12 := math.Abs(pos1[j] - pos2[j])
		if diff12 > 1.0 {
			t.Errorf("axis %d: file1=%.6f file2=%.6f diff=%.6f km", j, pos1[j], pos2[j], diff12)
		}
		diffBoth := math.Abs(posBoth[j] - pos1[j])
		if diffBoth > 1.0 {
			t.Errorf("axis %d: combined=%.6f file1=%.6f diff=%.6f km", j, posBoth[j], pos1[j], diffBoth)
		}
	}
}

func TestOpenMultipleNoFiles(t *testing.T) {
	_, err := OpenMultiple()
	if err == nil {
		t.Fatal("expected error for no files")
	}
}

func TestOpenMultipleOneInvalid(t *testing.T) {
	_, err := OpenMultiple(bspPath, "/nonexistent/file.bsp")
	if err == nil {
		t.Fatal("expected error when one file is invalid")
	}
}
