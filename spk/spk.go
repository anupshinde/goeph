package spk

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"sort"
)

const (
	recordLen = 1024
	j2000JD   = 2451545.0
	secPerDay = 86400.0
	// Speed of light in km/day
	cKmPerDay = 299792.458 * secPerDay
)

// SPK holds a parsed SPK/DAF ephemeris file (supports Type 2 segments only).
type SPK struct {
	segments []segment
	segMap   map[[2]int][]*segment // [target, center] → segments (sorted by startSec)
	chains   map[int][]chainLink   // body ID → chain of segment steps to SSB
}

// chainLink represents one hop in a body's chain to SSB.
type chainLink struct {
	target int
	center int
}

type segment struct {
	target   int
	center   int
	startSec float64 // segment start epoch (TDB seconds past J2000) from DAF summary
	endSec   float64 // segment end epoch (TDB seconds past J2000) from DAF summary
	init     float64 // initial epoch (TDB seconds past J2000) from segment metadata
	intLen   float64 // interval length (seconds)
	rsize    int     // record size (doubles per record)
	n        int     // number of records
	nCoeffs  int     // Chebyshev coefficients per component
	data     []float64
}

// Open reads and parses an SPK file. Only Type 2 segments are supported.
func Open(filename string) (*SPK, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read file record (record 1)
	fileRec := make([]byte, recordLen)
	if _, err := f.Read(fileRec); err != nil {
		return nil, fmt.Errorf("reading file record: %w", err)
	}

	locidw := string(fileRec[0:8])
	if locidw != "DAF/SPK " {
		return nil, fmt.Errorf("not an SPK file: got %q", locidw)
	}

	nd := int(binary.LittleEndian.Uint32(fileRec[8:12]))
	ni := int(binary.LittleEndian.Uint32(fileRec[12:16]))
	fward := int(binary.LittleEndian.Uint32(fileRec[76:80]))

	// Summary size: ND doubles + ceil(NI/2) doubles packed as integers
	summaryDoubles := nd + (ni+1)/2
	summaryBytes := summaryDoubles * 8

	spk := &SPK{
		segMap: make(map[[2]int][]*segment),
		chains: make(map[int][]chainLink),
	}

	// Walk summary record chain starting at FWARD
	recNum := fward
	for recNum != 0 {
		offset := int64(recNum-1) * recordLen
		if _, err := f.Seek(offset, 0); err != nil {
			return nil, err
		}
		rec := make([]byte, recordLen)
		if _, err := f.Read(rec); err != nil {
			return nil, err
		}

		nextRec := math.Float64frombits(binary.LittleEndian.Uint64(rec[0:8]))
		nSummaries := int(math.Float64frombits(binary.LittleEndian.Uint64(rec[16:24])))

		pos := 24
		for i := 0; i < nSummaries; i++ {
			summary := rec[pos : pos+summaryBytes]

			// Parse doubles
			startSec := math.Float64frombits(binary.LittleEndian.Uint64(summary[0:8]))
			endSec := math.Float64frombits(binary.LittleEndian.Uint64(summary[8:16]))

			// Parse integers (after ND doubles)
			intOff := nd * 8
			target := int(int32(binary.LittleEndian.Uint32(summary[intOff:])))
			center := int(int32(binary.LittleEndian.Uint32(summary[intOff+4:])))
			// frame at intOff+8
			dataType := int(int32(binary.LittleEndian.Uint32(summary[intOff+12:])))
			startI := int(int32(binary.LittleEndian.Uint32(summary[intOff+16:])))
			endI := int(int32(binary.LittleEndian.Uint32(summary[intOff+20:])))

			if dataType != 2 {
				return nil, fmt.Errorf("unsupported SPK type %d (target=%d, center=%d)", dataType, target, center)
			}

			// Read segment data from file
			nWords := endI - startI + 1
			dataOffset := int64(startI-1) * 8
			if _, err := f.Seek(dataOffset, 0); err != nil {
				return nil, err
			}
			rawData := make([]byte, nWords*8)
			if _, err := f.Read(rawData); err != nil {
				return nil, err
			}

			data := make([]float64, nWords)
			for j := range data {
				data[j] = math.Float64frombits(binary.LittleEndian.Uint64(rawData[j*8 : j*8+8]))
			}

			// Metadata is in the last 4 words
			seg := segment{
				target:   target,
				center:   center,
				startSec: startSec,
				endSec:   endSec,
				init:     data[nWords-4],
				intLen:   data[nWords-3],
				rsize:    int(data[nWords-2]),
				n:        int(data[nWords-1]),
				data:     data[:nWords-4],
			}
			seg.nCoeffs = (seg.rsize - 2) / 3

			spk.segments = append(spk.segments, seg)
			key := [2]int{target, center}
			spk.segMap[key] = append(spk.segMap[key], &spk.segments[len(spk.segments)-1])

			pos += summaryBytes
		}

		if nextRec == 0.0 {
			break
		}
		recNum = int(nextRec)
	}

	// Sort segment slices by startSec for temporal stacking
	for _, segs := range spk.segMap {
		sort.Slice(segs, func(i, j int) bool {
			return segs[i].startSec < segs[j].startSec
		})
	}

	// Build and validate chains from every target body to SSB
	if err := spk.buildChains(); err != nil {
		return nil, err
	}

	return spk, nil
}

// segPosition evaluates a single segment at the given TDB Julian date.
// Returns position in km, ICRF frame.
// If multiple segments cover the same (target, center) pair, picks the
// one whose date range contains the requested epoch.
func (s *SPK) segPosition(target, center int, tdbJD float64) [3]float64 {
	key := [2]int{target, center}
	segs := s.segMap[key]
	if len(segs) == 0 {
		panic(fmt.Sprintf("spk: no segment for target=%d center=%d", target, center))
	}

	seconds := (tdbJD - j2000JD) * secPerDay

	// Find the segment covering this epoch
	seg := findSegment(segs, seconds)

	// Find record index
	idx := int((seconds - seg.init) / seg.intLen)
	if idx < 0 {
		idx = 0
	}
	if idx >= seg.n {
		idx = seg.n - 1
	}

	// Normalized time in [-1, 1]
	offset := seconds - seg.init - float64(idx)*seg.intLen
	tc := 2.0*offset/seg.intLen - 1.0

	// Evaluate Chebyshev for x, y, z
	recStart := idx * seg.rsize
	var pos [3]float64
	for comp := 0; comp < 3; comp++ {
		cStart := recStart + 2 + comp*seg.nCoeffs
		pos[comp] = chebyshev(seg.data[cStart:cStart+seg.nCoeffs], tc)
	}
	return pos
}

// findSegment returns the segment from segs whose [startSec, endSec] range
// contains the given epoch. Falls back to the nearest boundary segment for
// out-of-range epochs (preserves existing clamping behavior).
func findSegment(segs []*segment, seconds float64) *segment {
	if len(segs) == 1 {
		return segs[0]
	}
	for _, seg := range segs {
		if seconds >= seg.startSec && seconds <= seg.endSec {
			return seg
		}
	}
	// Out of range: clamp to first or last segment
	if seconds < segs[0].startSec {
		return segs[0]
	}
	return segs[len(segs)-1]
}

// bodyWrtSSB computes a body's position relative to the Solar System Barycenter
// by summing positions along the pre-computed chain of segments.
func (s *SPK) bodyWrtSSB(body int, tdbJD float64) [3]float64 {
	if body == SSB {
		return [3]float64{}
	}

	chain, ok := s.chains[body]
	if !ok {
		panic(fmt.Sprintf("spk: no chain to SSB for body %d (not in loaded SPK file)", body))
	}

	var pos [3]float64
	for _, link := range chain {
		seg := s.segPosition(link.target, link.center, tdbJD)
		pos = add3(pos, seg)
	}
	return pos
}

// buildChains pre-computes the chain from each target body to SSB (0).
// Returns an error if any chain cannot reach SSB or contains a cycle.
func (s *SPK) buildChains() error {
	for key := range s.segMap {
		target := key[0]
		if _, exists := s.chains[target]; exists {
			continue // already built (could be built as intermediate of another chain)
		}
		if err := s.walkChain(target); err != nil {
			return err
		}
	}
	return nil
}

// walkChain builds the chain from body to SSB and stores it in s.chains.
// Also builds chains for any intermediate bodies encountered along the way.
func (s *SPK) walkChain(body int) error {
	if body == SSB {
		return nil
	}

	// Collect the path, detecting cycles
	var path []chainLink
	visited := make(map[int]bool)
	current := body

	for current != SSB {
		if visited[current] {
			return fmt.Errorf("spk: cycle detected in chain for body %d at body %d", body, current)
		}
		visited[current] = true

		center, found := s.findCenter(current)
		if !found {
			return fmt.Errorf("spk: body %d has no segment (needed in chain for body %d)", current, body)
		}

		path = append(path, chainLink{target: current, center: center})
		current = center
	}

	// Store chains for the target and all intermediates.
	// E.g. for path [{199,1}, {1,0}]:
	//   chains[199] = [{199,1}, {1,0}]
	//   chains[1]   = [{1,0}]
	for i := range path {
		b := path[i].target
		if _, exists := s.chains[b]; !exists {
			s.chains[b] = path[i:]
		}
	}

	return nil
}

// findCenter returns the center body for a given target.
func (s *SPK) findCenter(target int) (int, bool) {
	for key := range s.segMap {
		if key[0] == target {
			return key[1], true
		}
	}
	return 0, false
}

// earthWrtSSB returns Earth's position relative to SSB at tdbJD, in km ICRF.
func (s *SPK) earthWrtSSB(tdbJD float64) [3]float64 {
	return s.bodyWrtSSB(Earth, tdbJD)
}

// GeocentricPosition returns the geometric (no light-time) geocentric position
// of a body in km, ICRF frame.
func (s *SPK) GeocentricPosition(body int, tdbJD float64) [3]float64 {
	earthPos := s.earthWrtSSB(tdbJD)
	bodyPos := s.bodyWrtSSB(body, tdbJD)
	return sub3(bodyPos, earthPos)
}

// Observe computes the astrometric (light-time corrected) geocentric position
// of a body in km, ICRF frame. Matches Skyfield's observe() behavior.
func (s *SPK) Observe(body int, tdbJD float64) [3]float64 {
	earthPos := s.earthWrtSSB(tdbJD)
	bodyPos := s.bodyWrtSSB(body, tdbJD)

	geo := sub3(bodyPos, earthPos)
	dist := length3(geo)

	lightTime := 0.0
	for i := 0; i < 10; i++ {
		newLT := dist / cKmPerDay // light-time in days
		if math.Abs(newLT-lightTime) < 1e-12 {
			break
		}
		lightTime = newLT
		bodyPos = s.bodyWrtSSB(body, tdbJD-lightTime)
		geo = sub3(bodyPos, earthPos)
		dist = length3(geo)
	}

	return geo
}

// chebyshev evaluates a Chebyshev polynomial using the Clenshaw algorithm.
// coeffs are the Chebyshev coefficients, s is the normalized time in [-1, 1].
func chebyshev(coeffs []float64, s float64) float64 {
	n := len(coeffs)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return coeffs[0]
	}

	s2 := 2.0 * s
	w0 := coeffs[n-1]
	w1 := 0.0
	for i := n - 2; i >= 1; i-- {
		w0, w1 = coeffs[i]+s2*w0-w1, w0
	}
	return coeffs[0] + s*w0 - w1
}

func add3(a, b [3]float64) [3]float64 {
	return [3]float64{a[0] + b[0], a[1] + b[1], a[2] + b[2]}
}

func sub3(a, b [3]float64) [3]float64 {
	return [3]float64{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
}

func length3(a [3]float64) float64 {
	return math.Sqrt(a[0]*a[0] + a[1]*a[1] + a[2]*a[2])
}
