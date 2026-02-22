package satellite

import (
	"math"
	"time"

	gosatellite "github.com/joshuaferrara/go-satellite"
)

// Sat holds a named satellite for propagation.
type Sat struct {
	Name string
	Sat  gosatellite.Satellite
}

// NewSat creates a Sat from TLE lines using WGS84 gravity model.
func NewSat(name, line1, line2 string) Sat {
	return Sat{
		Name: name,
		Sat:  gosatellite.TLEToSat(line1, line2, gosatellite.GravityWGS84),
	}
}

// SubPoint returns the sub-satellite point (geographic lat/lon in degrees).
func SubPoint(s gosatellite.Satellite, t time.Time) (latDeg, lonDeg float64) {
	year := t.Year()
	month := int(t.Month())
	day := t.Day()
	hour := t.Hour()
	min := t.Minute()
	sec := t.Second()

	pos, _ := gosatellite.Propagate(s, year, month, day, hour, min, sec)
	jd := gosatellite.JDay(year, month, day, hour, min, sec)
	gmst := gosatellite.ThetaG_JD(jd)

	_, _, latLong := gosatellite.ECIToLLA(pos, gmst)
	ll := gosatellite.LatLongDeg(latLong)

	lonDeg = math.Mod(ll.Longitude+360.0, 360.0)
	return ll.Latitude, lonDeg
}
