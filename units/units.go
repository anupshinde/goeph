package units

import "math"

// AUToKm is the IAU 2012 nominal astronomical unit in kilometers.
const AUToKm = 149597870.7

const (
	deg2rad = math.Pi / 180.0
	rad2deg = 180.0 / math.Pi
)

// --- Angle ---

// Angle represents an angular measurement.
type Angle struct {
	rad float64
}

// NewAngle creates an Angle from radians.
func NewAngle(radians float64) Angle { return Angle{rad: radians} }

// AngleFromDegrees creates an Angle from degrees.
func AngleFromDegrees(deg float64) Angle { return Angle{rad: deg * deg2rad} }

// AngleFromHours creates an Angle from hours of right ascension.
func AngleFromHours(hours float64) Angle { return Angle{rad: hours * math.Pi / 12.0} }

// Radians returns the angle in radians.
func (a Angle) Radians() float64 { return a.rad }

// Degrees returns the angle in degrees.
func (a Angle) Degrees() float64 { return a.rad * rad2deg }

// Hours returns the angle in hours of right ascension.
func (a Angle) Hours() float64 { return a.rad * 12.0 / math.Pi }

// Arcminutes returns the angle in arcminutes.
func (a Angle) Arcminutes() float64 { return a.Degrees() * 60.0 }

// Arcseconds returns the angle in arcseconds.
func (a Angle) Arcseconds() float64 { return a.Degrees() * 3600.0 }

// DMS decomposes the angle into sign, integer degrees, integer arcminutes,
// and fractional arcseconds. Sign is +1 or -1.
func (a Angle) DMS() (sign float64, deg, min int, sec float64) {
	total := a.Degrees()
	sign = 1.0
	if total < 0 {
		sign = -1.0
		total = -total
	}
	deg = int(total)
	remainder := (total - float64(deg)) * 60.0
	min = int(remainder)
	sec = (remainder - float64(min)) * 60.0
	return
}

// HMS decomposes the angle (as right ascension) into sign, integer hours,
// integer minutes, and fractional seconds. Sign is +1 or -1.
func (a Angle) HMS() (sign float64, hours, min int, sec float64) {
	total := a.Hours()
	sign = 1.0
	if total < 0 {
		sign = -1.0
		total = -total
	}
	hours = int(total)
	remainder := (total - float64(hours)) * 60.0
	min = int(remainder)
	sec = (remainder - float64(min)) * 60.0
	return
}

// --- Distance ---

// Distance represents a distance measurement.
type Distance struct {
	km float64
}

// NewDistance creates a Distance from kilometers.
func NewDistance(km float64) Distance { return Distance{km: km} }

// DistanceFromAU creates a Distance from astronomical units.
func DistanceFromAU(au float64) Distance { return Distance{km: au * AUToKm} }

// DistanceFromMeters creates a Distance from meters.
func DistanceFromMeters(m float64) Distance { return Distance{km: m / 1000.0} }

// Km returns the distance in kilometers.
func (d Distance) Km() float64 { return d.km }

// AU returns the distance in astronomical units.
func (d Distance) AU() float64 { return d.km / AUToKm }

// M returns the distance in meters.
func (d Distance) M() float64 { return d.km * 1000.0 }

// LightSeconds returns the distance in light-seconds.
func (d Distance) LightSeconds() float64 { return d.km / 299792.458 }
