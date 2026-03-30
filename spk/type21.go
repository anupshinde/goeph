package spk

// Type 21 (Extended Modified Difference Arrays) support.
// Based on the NAIF SPICE spke21.f algorithm.

const maxTrm = 25 // maximum terms per difference table component

// parseType21Segment builds a segment from raw Type 21 data.
// Layout at end of segment: epoch_table (n entries), epoch_dir (n/100 entries),
// maxDim (1), n (1).
func parseType21Segment(target, center int, startSec, endSec float64, data []float64, nWords int) segment {
	n := int(data[nWords-1])
	maxDim := int(data[nWords-2])
	dlSize := 4*maxDim + 11

	// Epoch table starts after n*dlSize record data
	epochStart := n * dlSize
	epochTable := make([]float64, n)
	copy(epochTable, data[epochStart:epochStart+n])

	return segment{
		target:     target,
		center:     center,
		dataType:   21,
		startSec:   startSec,
		endSec:     endSec,
		maxDim:     maxDim,
		dlSize:     dlSize,
		n:          n,
		epochTable: epochTable,
		data:       data,
	}
}

// type21Position evaluates a Type 21 segment at tdbJD, returning position in km.
func (s *SPK) type21Position(seg *segment, tdbSeconds float64) [3]float64 {
	pos, _ := type21Eval(seg, tdbSeconds)
	return pos
}

// type21Velocity evaluates a Type 21 segment at tdbJD, returning velocity in km/day.
func (s *SPK) type21Velocity(seg *segment, tdbSeconds float64) [3]float64 {
	_, vel := type21Eval(seg, tdbSeconds)
	return vel
}

// type21Eval computes position (km) and velocity (km/day) from a Type 21 segment.
func type21Eval(seg *segment, et float64) (pos [3]float64, vel [3]float64) {
	maxDim := seg.maxDim
	dlSize := seg.dlSize

	// Find the record whose epoch bracket contains et
	recIdx := findType21Record(seg, et)

	// Extract the MDA record
	recStart := recIdx * dlSize
	record := seg.data[recStart : recStart+dlSize]

	// Unpack the record:
	//   record[0]                  = TL (reference epoch)
	//   record[1 .. maxDim]        = G (stepsize vector)
	//   record[maxDim+1..maxDim+6] = REFPOS[x], REFVEL[x], REFPOS[y], REFVEL[y], REFPOS[z], REFVEL[z]
	//   record[maxDim+7..4*maxDim+6] = DT (difference tables, maxDim x 3, column-major)
	//   record[4*maxDim+7]         = KQMAX1 (max integration order + 1)
	//   record[4*maxDim+8..10]     = KQ (integration order per component)

	tl := record[0]
	g := record[1 : maxDim+1]

	var refPos, refVel [3]float64
	refPos[0] = record[maxDim+1]
	refVel[0] = record[maxDim+2]
	refPos[1] = record[maxDim+3]
	refVel[1] = record[maxDim+4]
	refPos[2] = record[maxDim+5]
	refVel[2] = record[maxDim+6]

	// Difference table: maxDim rows x 3 columns, stored column-major
	dtBase := maxDim + 7

	kqmax1 := int(record[4*maxDim+7])
	var kq [3]int
	kq[0] = int(record[4*maxDim+8])
	kq[1] = int(record[4*maxDim+9])
	kq[2] = int(record[4*maxDim+10])

	delta := et - tl
	tp := delta
	mq2 := kqmax1 - 2
	ks := kqmax1 - 1

	// Compute FC and WC coefficients from stepsize vector
	var fc [maxTrm]float64
	var wc [maxTrm]float64
	fc[0] = 1.0

	for j := 1; j <= mq2; j++ {
		fc[j] = tp / g[j-1]
		wc[j-1] = delta / g[j-1]
		tp = delta + g[j-1]
	}

	// Collect reciprocals
	var w [maxTrm + 3]float64
	for j := 1; j <= kqmax1; j++ {
		w[j-1] = 1.0 / float64(j)
	}

	// Compute W(K) terms for position interpolation
	jx := 0
	ks1 := ks - 1

	for ks >= 2 {
		jx++
		for j := 1; j <= jx; j++ {
			w[j+ks-1] = fc[j]*w[j+ks1-1] - wc[j-1]*w[j+ks-1]
		}
		ks = ks1
		ks1--
	}

	// Position interpolation
	for i := 0; i < 3; i++ {
		kqq := kq[i]
		sum := 0.0
		for j := kqq; j >= 1; j-- {
			// DT[j-1, i] in column-major: index = dtBase + i*maxDim + (j-1)
			sum += record[dtBase+i*maxDim+j-1] * w[j+ks-1]
		}
		pos[i] = refPos[i] + delta*(refVel[i]+delta*sum)
	}

	// Recompute W for velocity interpolation
	for j := 1; j <= jx; j++ {
		w[j+ks-1] = fc[j]*w[j+ks1-1] - wc[j-1]*w[j+ks-1]
	}
	ks--

	// Velocity interpolation (result in km/s)
	for i := 0; i < 3; i++ {
		kqq := kq[i]
		sum := 0.0
		for j := kqq; j >= 1; j-- {
			sum += record[dtBase+i*maxDim+j-1] * w[j+ks-1]
		}
		vel[i] = (refVel[i] + delta*sum) * secPerDay // km/s → km/day
	}

	return
}

// findType21Record finds the record index for the given epoch using the epoch table.
func findType21Record(seg *segment, et float64) int {
	n := seg.n

	// Search epoch table for the first entry > et
	for i := 0; i < n; i++ {
		if seg.epochTable[i] > et {
			return i
		}
	}
	// Clamp to last record
	return n - 1
}
