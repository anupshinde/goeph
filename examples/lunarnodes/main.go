// Example: Mean lunar node positions
//
// The lunar nodes are where the Moon's orbit crosses the ecliptic plane.
// The ascending (north) node regresses westward with an 18.6-year period.
// Eclipses can only occur when the Sun is near a lunar node.
package main

import (
	"fmt"
	"time"

	"github.com/anupshinde/goeph/lunarnodes"
	"github.com/anupshinde/goeph/timescale"
)

func main() {
	// Show lunar node positions over a few years
	fmt.Println("Mean lunar node positions (ecliptic longitude):")
	fmt.Printf("%-12s %15s %15s\n", "Date", "North Node", "South Node")
	fmt.Println("------------ --------------- ---------------")

	for year := 2024; year <= 2028; year++ {
		t := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		jdTT := timescale.UTCToTT(timescale.TimeToJDUTC(t))

		north, south := lunarnodes.MeanLunarNodes(jdTT)

		fmt.Printf("%d-01-01   %12.2f°   %12.2f°\n", year, north, south)
	}

	fmt.Println("\nThe north node moves ~19.3° per year (retrograde),")
	fmt.Println("completing one full cycle in ~18.6 years.")

	// Show monthly positions for 2024
	fmt.Println("\nMonthly north node position for 2024:")
	for month := time.January; month <= time.December; month++ {
		t := time.Date(2024, month, 1, 0, 0, 0, 0, time.UTC)
		jdTT := timescale.UTCToTT(timescale.TimeToJDUTC(t))
		north, _ := lunarnodes.MeanLunarNodes(jdTT)
		fmt.Printf("  %s: %.2f°\n", t.Format("Jan"), north)
	}
}
