package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/anupshinde/goeph/coord"
	"github.com/anupshinde/goeph/lunarnodes"
	"github.com/anupshinde/goeph/satellite"
	"github.com/anupshinde/goeph/spk"
	"github.com/anupshinde/goeph/star"
	"github.com/anupshinde/goeph/timescale"
)

// CelestialBody defines a body to compute geocentric ecliptic position for.
type CelestialBody struct {
	Name   string // column prefix (lowercase)
	BodyID int    // NAIF body ID (0 = use custom handler)
}

// Bodies to compute, matching the Python code order.
var celestialBodies = []CelestialBody{
	{Name: "sun", BodyID: spk.Sun},
	{Name: "moon", BodyID: spk.Moon},
	{Name: "mercury", BodyID: spk.Mercury},
	{Name: "venus", BodyID: spk.Venus},
	{Name: "mars", BodyID: spk.MarsBarycenter},
	{Name: "jupiter", BodyID: spk.JupiterBarycenter},
	{Name: "saturn", BodyID: spk.SaturnBarycenter},
	{Name: "uranus", BodyID: spk.UranusBarycenter},
	{Name: "neptune", BodyID: spk.NeptuneBarycenter},
	{Name: "pluto", BodyID: spk.PlutoBarycenter},
	{Name: "gc", BodyID: 0}, // Galactic Center (custom)
}

// Locations for zenith ecliptic position computation.
var locations = []coord.Location{
	{Name: "loc_ni", Lat: 0.0, Lon: 0.0},
	{Name: "loc_chicago", Lat: 41.8674558, Lon: -87.6483924},
	{Name: "loc_london", Lat: 51.5150534, Lon: -0.1016089},
	{Name: "loc_cushing", Lat: 35.9859634, Lon: -96.7954485},
	{Name: "loc_ny", Lat: 40.714469, Lon: -74.0194683},
	{Name: "loc_mumbai", Lat: 19.0602766, Lon: 72.8577106},
}

// Satellite TLE data
const issLine1 = "1 25544U 98067A   24031.54769676  .00006652  00000+0  12801-3 0  9996"
const issLine2 = "2 25544  51.6415  41.8752 0005686  93.3517  58.2198 15.50294839424075"

const poleL1 = "1 99998U 24031A   24031.54769676  .00006652  00000+0  12801-3 0  9996"
const poleL2 = "2 99998  90.0000 180.0000 0005686   0.0000 270.0000 15.99999999    00"

var artisats []satellite.Sat

func initSatellites() {
	// FastISS: 1.618x faster than ISS
	fastMM := fmt.Sprintf("%11.8f", 15.50294839*1.618)
	fastL2 := strings.Replace(issLine2, "15.50294839", fastMM, 1)

	// AntiISS: 0.618x speed
	antiMM := fmt.Sprintf("%11.8f", 15.50294839*0.618)
	antiL2 := strings.Replace(issLine2, "15.50294839", antiMM, 1)

	artisats = []satellite.Sat{
		satellite.NewSat("fastiss", issLine1, fastL2),
		satellite.NewSat("issanti", issLine1, antiL2),
		satellite.NewSat("polesat", poleL1, poleL2),
	}
}

// CelestialRow holds all computed values for a single timestamp.
type CelestialRow struct {
	Time string

	// Body ecliptic positions [lat, lon] for each body
	BodyPositions [][2]float64

	// Satellite sub-points [lat, lon]
	SatPositions [][2]float64

	// Location ecliptic positions [lat, lon]
	LocPositions [][2]float64

	// Lunar nodes
	NorthNodeLon float64
	SouthNodeLon float64
}

// ComputeRow computes all celestial data for a single timestamp.
func ComputeRow(ephemeris *spk.SPK, t time.Time) CelestialRow {
	jdUTC := timescale.TimeToJDUTC(t)
	tdbJD := timescale.UTCToTT(jdUTC) // TDB ≈ TT
	ut1JD := timescale.TTToUT1(tdbJD)

	row := CelestialRow{
		Time: t.Format("2006-01-02 15:04:05+00:00"),
	}

	// Galactic Center ICRF direction (fixed)
	gcX, gcY, gcZ := star.GalacticCenterICRF()

	// Compute body positions
	row.BodyPositions = make([][2]float64, len(celestialBodies))
	for i, body := range celestialBodies {
		var lat, lon float64
		if body.BodyID == 0 {
			lat, lon = coord.ICRFToEcliptic(gcX, gcY, gcZ)
		} else {
			pos := ephemeris.Observe(body.BodyID, tdbJD)
			lat, lon = coord.ICRFToEcliptic(pos[0], pos[1], pos[2])
		}
		row.BodyPositions[i] = [2]float64{lat, lon}
	}

	// Compute satellite sub-points
	row.SatPositions = make([][2]float64, len(artisats))
	for i, sat := range artisats {
		lat, lon := satellite.SubPoint(sat.Sat, t)
		row.SatPositions[i] = [2]float64{lat, lon}
	}

	// Compute location ecliptic positions (zenith direction)
	row.LocPositions = make([][2]float64, len(locations))
	for i, loc := range locations {
		x, y, z := coord.GeodeticToICRF(loc.Lat, loc.Lon, ut1JD)
		lat, lon := coord.ICRFToEcliptic(x, y, z)
		row.LocPositions[i] = [2]float64{lat, lon}
	}

	// Lunar nodes
	row.NorthNodeLon, row.SouthNodeLon = lunarnodes.MeanLunarNodes(tdbJD)

	return row
}

// CSVHeader returns the CSV header matching the Python output format.
func CSVHeader() string {
	header := "Time"
	for _, body := range celestialBodies {
		header += "," + body.Name + "_lat_deg"
		header += "," + body.Name + "_lon_deg"
	}
	for _, sat := range artisats {
		header += "," + sat.Name + "_sub_lat_deg"
		header += "," + sat.Name + "_sub_lon_deg"
	}
	for _, loc := range locations {
		header += "," + loc.Name + "_lat_deg"
		header += "," + loc.Name + "_lon_deg"
	}
	header += ",north_node_lon_deg,south_node_lon_deg"
	return header
}

// CSVRow formats a CelestialRow as a CSV line.
func CSVRow(row CelestialRow) string {
	line := row.Time
	for _, pos := range row.BodyPositions {
		line += "," + formatFloat(pos[0]) + "," + formatFloat(pos[1])
	}
	for _, pos := range row.SatPositions {
		line += "," + formatFloat(pos[0]) + "," + formatFloat(pos[1])
	}
	for _, pos := range row.LocPositions {
		line += "," + formatFloat(pos[0]) + "," + formatFloat(pos[1])
	}
	line += "," + formatFloat(row.NorthNodeLon) + "," + formatFloat(row.SouthNodeLon)
	return line
}

func formatFloat(f float64) string {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return ""
	}
	return fmt.Sprintf("%.17g", f)
}
