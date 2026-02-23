package coord

// NutationPrecision controls the number of terms used in the IAU 2000A nutation series.
type NutationPrecision int

const (
	// NutationStandard uses the 30 largest luni-solar terms (~1 arcsec precision).
	// This is ~45x faster than NutationFull and sufficient for most applications,
	// since other error sources (light-time ~20 arcsec, GMST formula ~0.3 arcsec/century)
	// dominate the overall accuracy budget.
	NutationStandard NutationPrecision = iota

	// NutationFull uses all 678 luni-solar + 687 planetary terms (~0.001 arcsec precision).
	// Matches Skyfield's default IAU 2000A nutation model. Use for high-precision
	// single-point computations or when exact Skyfield parity is required.
	NutationFull
)

var nutationPrecision = NutationStandard

// SetNutationPrecision sets the nutation precision for the coord package.
// Default is NutationStandard (30 terms, fast).
// Not safe for concurrent use — call once at program startup.
func SetNutationPrecision(p NutationPrecision) {
	nutationPrecision = p
}

// GetNutationPrecision returns the current nutation precision setting.
func GetNutationPrecision() NutationPrecision {
	return nutationPrecision
}
