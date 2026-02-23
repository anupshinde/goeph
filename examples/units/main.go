// Example: Working with Angle and Distance types
//
// The units package provides Angle and Distance types with convenient
// conversion methods. Angle supports degrees, hours, radians, DMS, and
// HMS formats. Distance supports km, AU, meters, and light-seconds.
package main

import (
	"fmt"
	"math"

	"github.com/anupshinde/goeph/units"
)

func main() {
	// Angle conversions
	fmt.Println("=== Angle Conversions ===\n")

	// Sirius: RA = 6h 45m 8.9s, Dec = -16° 42' 58"
	ra := units.AngleFromHours(6.0 + 45.0/60.0 + 8.9/3600.0)
	dec := units.AngleFromDegrees(-(16.0 + 42.0/60.0 + 58.0/3600.0))

	fmt.Println("Sirius:")
	fmt.Printf("  RA = %.6f hours = %.4f°\n", ra.Hours(), ra.Degrees())
	_, raH, raM, raS := ra.HMS()
	fmt.Printf("  RA = %dh %dm %.1fs\n", raH, raM, raS)

	decSign, decD, decAM, decAS := dec.DMS()
	sign := "+"
	if decSign < 0 {
		sign = "-"
	}
	fmt.Printf("  Dec = %s%d° %d' %.1f\"\n", sign, decD, decAM, decAS)
	fmt.Printf("  Dec = %.4f° = %.6f rad\n\n", dec.Degrees(), dec.Radians())

	// Full circle
	full := units.AngleFromDegrees(360)
	fmt.Printf("Full circle: %.0f° = %.4f rad = %.0f hours = %.0f arcmin = %.0f arcsec\n\n",
		full.Degrees(), full.Radians(), full.Hours(),
		full.Arcminutes(), full.Arcseconds())

	// Distance conversions
	fmt.Println("=== Distance Conversions ===\n")

	// Earth-Sun distance
	earthSun := units.DistanceFromAU(1.0)
	fmt.Printf("1 AU = %.1f km = %.0f m = %.2f light-seconds\n",
		earthSun.Km(), earthSun.M(), earthSun.LightSeconds())

	// Moon distance
	moonDist := units.NewDistance(384400)
	fmt.Printf("Moon distance: %.0f km = %.6f AU = %.3f light-seconds\n",
		moonDist.Km(), moonDist.AU(), moonDist.LightSeconds())

	// Jupiter distance (~5.2 AU)
	jupiterDist := units.DistanceFromAU(5.2)
	fmt.Printf("Jupiter (~5.2 AU): %.0f km = %.1f light-seconds (%.1f light-minutes)\n",
		jupiterDist.Km(), jupiterDist.LightSeconds(), jupiterDist.LightSeconds()/60)

	// Speed of light check
	lightDist := units.NewDistance(299792.458)
	fmt.Printf("\nSpeed of light: %.3f km = %.6f light-seconds (should be 1.0)\n",
		lightDist.Km(), lightDist.LightSeconds())

	// Small angle
	oneArcsec := units.AngleFromDegrees(1.0 / 3600.0)
	fmt.Printf("\n1 arcsecond = %.6e° = %.6e rad = %.1f arcseconds\n",
		oneArcsec.Degrees(), oneArcsec.Radians(), oneArcsec.Arcseconds())

	// Pi radians
	piRad := units.NewAngle(math.Pi)
	fmt.Printf("π radians = %.4f° = %.4f hours\n", piRad.Degrees(), piRad.Hours())
}
