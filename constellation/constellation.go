// Package constellation identifies which of the 88 IAU constellations
// contains a given sky position.
//
// The lookup uses a pre-computed grid from the IAU constellation boundary
// data (Roman 1987, CDS catalog VI/42). Binary search on RA and declination
// gives O(log n) lookup time with no linear scanning.
//
// Positions should be given in epoch B1875 equatorial coordinates, which is the
// standard epoch for constellation boundaries. For most practical purposes, J2000
// coordinates work fine — the precession difference is ~1.4° over 125 years, which
// only matters for positions very close to a constellation boundary.
package constellation

import "sort"

// At returns the IAU 3-letter abbreviation of the constellation containing the
// given position. raHours is right ascension in hours [0, 24) and decDeg is
// declination in degrees [-90, 90].
//
// The boundaries use epoch B1875 coordinates (the standard for IAU constellation
// boundaries). J2000 coordinates are acceptable for most purposes.
func At(raHours, decDeg float64) string {
	i := sort.SearchFloat64s(sortedRA[:], raHours)
	j := sort.Search(len(sortedDec), func(k int) bool {
		return sortedDec[k] > decDeg
	})
	if i >= len(grid) {
		i = len(grid) - 1
	}
	if j >= len(grid[0]) {
		j = len(grid[0]) - 1
	}
	return abbreviations[grid[i][j]]
}

// Name returns the full name for a constellation abbreviation.
// Returns empty string if the abbreviation is not recognized.
func Name(abbr string) string {
	name, ok := nameMap[abbr]
	if !ok {
		return ""
	}
	return name
}

// Abbreviation returns the 3-letter IAU abbreviation for a constellation name.
// Returns empty string if the name is not recognized.
func Abbreviation(name string) string {
	abbr, ok := abbrMap[name]
	if !ok {
		return ""
	}
	return abbr
}

// Names returns a copy of all 88 constellation abbreviation-name pairs.
func Names() [][2]string {
	result := make([][2]string, len(constellationNames))
	copy(result, constellationNames[:])
	return result
}

// constellationNames maps abbreviation to full name for all 88 IAU constellations.
var constellationNames = [88][2]string{
	{"And", "Andromeda"},
	{"Ant", "Antlia"},
	{"Aps", "Apus"},
	{"Aql", "Aquila"},
	{"Aqr", "Aquarius"},
	{"Ara", "Ara"},
	{"Ari", "Aries"},
	{"Aur", "Auriga"},
	{"Boo", "Bootes"},
	{"CMa", "Canis Major"},
	{"CMi", "Canis Minor"},
	{"CVn", "Canes Venatici"},
	{"Cae", "Caelum"},
	{"Cam", "Camelopardalis"},
	{"Cap", "Capricornus"},
	{"Car", "Carina"},
	{"Cas", "Cassiopeia"},
	{"Cen", "Centaurus"},
	{"Cep", "Cepheus"},
	{"Cet", "Cetus"},
	{"Cha", "Chamaeleon"},
	{"Cir", "Circinus"},
	{"Cnc", "Cancer"},
	{"Col", "Columba"},
	{"Com", "Coma Berenices"},
	{"CrA", "Corona Australis"},
	{"CrB", "Corona Borealis"},
	{"Crt", "Crater"},
	{"Cru", "Crux"},
	{"Crv", "Corvus"},
	{"Cyg", "Cygnus"},
	{"Del", "Delphinus"},
	{"Dor", "Dorado"},
	{"Dra", "Draco"},
	{"Equ", "Equuleus"},
	{"Eri", "Eridanus"},
	{"For", "Fornax"},
	{"Gem", "Gemini"},
	{"Gru", "Grus"},
	{"Her", "Hercules"},
	{"Hor", "Horologium"},
	{"Hya", "Hydra"},
	{"Hyi", "Hydrus"},
	{"Ind", "Indus"},
	{"LMi", "Leo Minor"},
	{"Lac", "Lacerta"},
	{"Leo", "Leo"},
	{"Lep", "Lepus"},
	{"Lib", "Libra"},
	{"Lup", "Lupus"},
	{"Lyn", "Lynx"},
	{"Lyr", "Lyra"},
	{"Men", "Mensa"},
	{"Mic", "Microscopium"},
	{"Mon", "Monoceros"},
	{"Mus", "Musca"},
	{"Nor", "Norma"},
	{"Oct", "Octans"},
	{"Oph", "Ophiuchus"},
	{"Ori", "Orion"},
	{"Pav", "Pavo"},
	{"Peg", "Pegasus"},
	{"Per", "Perseus"},
	{"Phe", "Phoenix"},
	{"Pic", "Pictor"},
	{"PsA", "Piscis Austrinus"},
	{"Psc", "Pisces"},
	{"Pup", "Puppis"},
	{"Pyx", "Pyxis"},
	{"Ret", "Reticulum"},
	{"Scl", "Sculptor"},
	{"Sco", "Scorpius"},
	{"Sct", "Scutum"},
	{"Ser", "Serpens"},
	{"Sex", "Sextans"},
	{"Sge", "Sagitta"},
	{"Sgr", "Sagittarius"},
	{"Tau", "Taurus"},
	{"Tel", "Telescopium"},
	{"TrA", "Triangulum Australe"},
	{"Tri", "Triangulum"},
	{"Tuc", "Tucana"},
	{"UMa", "Ursa Major"},
	{"UMi", "Ursa Minor"},
	{"Vel", "Vela"},
	{"Vir", "Virgo"},
	{"Vol", "Volans"},
	{"Vul", "Vulpecula"},
}

// nameMap and abbrMap are built at init time.
var (
	nameMap map[string]string
	abbrMap map[string]string
)

func init() {
	nameMap = make(map[string]string, 88)
	abbrMap = make(map[string]string, 88)
	for _, pair := range constellationNames {
		nameMap[pair[0]] = pair[1]
		abbrMap[pair[1]] = pair[0]
	}
}
