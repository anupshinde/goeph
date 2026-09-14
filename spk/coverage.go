package spk

// Coverage and Covers report where the loaded segments hold data. Outside
// that span the position functions still return a value — the nearest
// segment's Chebyshev polynomial evaluated past its end — which is an
// extrapolation, not an ephemeris position. Callers that must not use such
// values check Covers first.

// Covers reports whether the loaded segments cover body's position relative
// to the Solar System Barycenter at tdbJD: every segment in the body's chain
// holds data for that instant. It is false for a body not in the loaded files.
//
// Observe and Apparent read the target at the light-time-retarded instant
// and the Earth at tdbJD, so both need covering.
func (s *SPK) Covers(body int, tdbJD float64) bool {
	if body == SSB {
		return true
	}
	chain, ok := s.chains[body]
	if !ok {
		return false
	}
	seconds := (tdbJD-j2000JD)*secPerDay + tdbMinusTT(tdbJD)
	for _, link := range chain {
		covered := false
		for _, seg := range s.segMap[[2]int{link.target, link.center}] {
			if seconds >= seg.startSec && seconds <= seg.endSec {
				covered = true
				break
			}
		}
		if !covered {
			return false
		}
	}
	return true
}

// Coverage returns the span (TDB Julian dates) over which the loaded segments
// cover body's position relative to the Solar System Barycenter: from the
// latest chain start to the earliest chain end. When several files are
// stacked for one body the span runs from the first file's start to the last
// file's end, so a gap between files lies inside it; Covers is exact.
// ok is false for a body not in the loaded files.
func (s *SPK) Coverage(body int) (startJD, endJD float64, ok bool) {
	chain, ok := s.chains[body]
	if !ok {
		return 0, 0, false
	}
	var startSec, endSec float64
	for i, link := range chain {
		segs := s.segMap[[2]int{link.target, link.center}]
		first, last := segs[0].startSec, segs[0].endSec
		for _, seg := range segs[1:] {
			first = min(first, seg.startSec)
			last = max(last, seg.endSec)
		}
		if i == 0 {
			startSec, endSec = first, last
			continue
		}
		startSec = max(startSec, first)
		endSec = min(endSec, last)
	}
	return secondsToJD(startSec), secondsToJD(endSec), true
}

// secondsToJD inverts the JD → segment-seconds conversion of segPosition.
func secondsToJD(seconds float64) float64 {
	jd := j2000JD + seconds/secPerDay
	return jd - tdbMinusTT(jd)/secPerDay
}
