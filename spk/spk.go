package spk

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
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
	segMap   map[[2]int]*segment // [target, center] → segment
}

type segment struct {
	target   int
	center   int
	init     float64 // initial epoch (TDB seconds past J2000)
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

	spk := &SPK{segMap: make(map[[2]int]*segment)}

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
			// startSec := math.Float64frombits(binary.LittleEndian.Uint64(summary[0:8]))
			// endSec := math.Float64frombits(binary.LittleEndian.Uint64(summary[8:16]))

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
				target:  target,
				center:  center,
				init:    data[nWords-4],
				intLen:  data[nWords-3],
				rsize:   int(data[nWords-2]),
				n:       int(data[nWords-1]),
				data:    data[:nWords-4],
			}
			seg.nCoeffs = (seg.rsize - 2) / 3

			spk.segments = append(spk.segments, seg)
			spk.segMap[[2]int{target, center}] = &spk.segments[len(spk.segments)-1]

			pos += summaryBytes
		}

		if nextRec == 0.0 {
			break
		}
		recNum = int(nextRec)
	}

	return spk, nil
}

// segPosition evaluates a single segment at the given TDB Julian date.
// Returns position in km, ICRF frame.
func (s *SPK) segPosition(target, center int, tdbJD float64) [3]float64 {
	seg := s.segMap[[2]int{target, center}]
	if seg == nil {
		panic(fmt.Sprintf("spk: no segment for target=%d center=%d", target, center))
	}

	seconds := (tdbJD - j2000JD) * secPerDay

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

// bodyWrtSSB computes a body's position relative to the Solar System Barycenter.
func (s *SPK) bodyWrtSSB(body int, tdbJD float64) [3]float64 {
	switch body {
	case 1, 2, 3, 4, 5, 6, 7, 8, 9, 10:
		// These have direct segments wrt SSB
		return s.segPosition(body, 0, tdbJD)
	case Mercury: // 199 → Mercury Bary (1) → SSB
		bary := s.segPosition(1, 0, tdbJD)
		off := s.segPosition(199, 1, tdbJD)
		return add3(bary, off)
	case Venus: // 299 → Venus Bary (2) → SSB
		bary := s.segPosition(2, 0, tdbJD)
		off := s.segPosition(299, 2, tdbJD)
		return add3(bary, off)
	case Moon: // 301 → EMB (3) → SSB
		emb := s.segPosition(3, 0, tdbJD)
		off := s.segPosition(301, 3, tdbJD)
		return add3(emb, off)
	case Earth: // 399 → EMB (3) → SSB
		emb := s.segPosition(3, 0, tdbJD)
		off := s.segPosition(399, 3, tdbJD)
		return add3(emb, off)
	default:
		panic(fmt.Sprintf("spk: unsupported body %d", body))
	}
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
