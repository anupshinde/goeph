// Example: Atmospheric refraction correction
//
// Shows how the atmosphere bends light, making objects appear higher than
// they geometrically are. The effect is strongest near the horizon (~0.57°)
// and negligible at high altitudes.
//
// Refraction() gives the correction for an observed (apparent) altitude.
// Refract() converts a true (geometric) altitude to apparent altitude.
package main

import (
	"fmt"

	"github.com/anupshinde/goeph/coord"
)

func main() {
	// Standard conditions: 10°C, 1013.25 mbar (sea level)
	tempC := 10.0
	pressure := 1013.25

	fmt.Printf("Atmospheric refraction (T=%.0f°C, P=%.2f mbar)\n\n", tempC, pressure)

	// Refraction correction at various altitudes
	fmt.Printf("%-15s %15s %15s\n", "True Alt", "Refraction", "Apparent Alt")
	fmt.Println("--------------- --------------- ---------------")

	altitudes := []float64{0, 1, 5, 10, 20, 30, 45, 60, 80}
	for _, alt := range altitudes {
		refraction := coord.Refraction(alt, tempC, pressure)
		apparent := coord.Refract(alt, tempC, pressure)
		fmt.Printf("%12.1f°   %12.4f°   %12.4f°\n", alt, refraction, apparent)
	}

	// The sunrise case: Sun appears on horizon when it's geometrically
	// about 0.57° below it
	fmt.Println("\nSunrise effect:")
	horizonRefraction := coord.Refraction(0, tempC, pressure)
	fmt.Printf("  At horizon (0° apparent): refraction = %.4f° (%.1f arcmin)\n",
		horizonRefraction, horizonRefraction*60)
	fmt.Printf("  The Sun appears to rise when it is still %.1f arcmin below the horizon\n",
		horizonRefraction*60)
}
