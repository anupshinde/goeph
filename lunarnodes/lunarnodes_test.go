package lunarnodes

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

type goldenLunarNodes struct {
	Tests []struct {
		TDBJD           float64 `json:"tdb_jd"`
		NorthNodeLonDeg float64 `json:"north_node_lon_deg"`
		SouthNodeLonDeg float64 `json:"south_node_lon_deg"`
	} `json:"tests"`
}

func TestMeanLunarNodes_J2000(t *testing.T) {
	north, south := MeanLunarNodes(j2000JD)
	if math.Abs(north-125.04452) > 0.001 {
		t.Errorf("north at J2000: got %f want ~125.04452", north)
	}
	wantSouth := math.Mod(125.04452+180.0, 360.0)
	if math.Abs(south-wantSouth) > 0.001 {
		t.Errorf("south at J2000: got %f want %f", south, wantSouth)
	}
}

func TestMeanLunarNodes_Opposite(t *testing.T) {
	dates := []float64{2451545.0, 2455000.0, 2460000.0}
	for _, jd := range dates {
		north, south := MeanLunarNodes(jd)
		diff := math.Abs(south - math.Mod(north+180.0, 360.0))
		if diff > 1e-10 {
			t.Errorf("jd=%.1f: south-north != 180°, diff=%f", jd, diff)
		}
	}
}

func TestMeanLunarNodes_Range(t *testing.T) {
	for jd := 2440000.0; jd < 2470000.0; jd += 1000 {
		north, south := MeanLunarNodes(jd)
		if north < 0 || north >= 360 {
			t.Errorf("jd=%.1f: north=%f out of [0,360)", jd, north)
		}
		if south < 0 || south >= 360 {
			t.Errorf("jd=%.1f: south=%f out of [0,360)", jd, south)
		}
	}
}

func TestMeanLunarNodes_Golden(t *testing.T) {
	data, err := os.ReadFile("../testdata/golden_lunarnodes.json")
	if err != nil {
		t.Fatal(err)
	}
	var golden goldenLunarNodes
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}

	const tol = 1e-8
	failures := 0
	for i, tc := range golden.Tests {
		north, south := MeanLunarNodes(tc.TDBJD)
		diffN := math.Abs(north - tc.NorthNodeLonDeg)
		diffS := math.Abs(south - tc.SouthNodeLonDeg)
		if diffN > tol {
			if failures < 10 {
				t.Errorf("test %d north: tdb=%.6f got=%.10f want=%.10f diff=%.2e",
					i, tc.TDBJD, north, tc.NorthNodeLonDeg, diffN)
			}
			failures++
		}
		if diffS > tol {
			if failures < 10 {
				t.Errorf("test %d south: tdb=%.6f got=%.10f want=%.10f diff=%.2e",
					i, tc.TDBJD, south, tc.SouthNodeLonDeg, diffS)
			}
			failures++
		}
	}
	if failures > 0 {
		t.Errorf("%d lunar node failures out of %d tests", failures, len(golden.Tests))
	}
}
