package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/anupshinde/goeph/spk"
)

func main() {
	// Load ephemeris — uses de440s.bsp from the repo's data/ folder,
	// or download from https://naif.jpl.nasa.gov/pub/naif/generic_kernels/spk/planets/
	bspPath := filepath.Join("..", "..", "data", "de440s.bsp")
	fmt.Println("Loading ephemeris:", bspPath)
	ephemeris, err := spk.Open(bspPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading ephemeris: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Ephemeris loaded.")

	// Initialize satellites
	initSatellites()

	// Reference date and time range (±100 years)
	referenceDate := time.Date(2026, 1, 19, 0, 0, 0, 0, time.UTC)
	startDate := referenceDate.Add(-100 * 365 * 24 * time.Hour)
	endDate := referenceDate.Add(100 * 365 * 24 * time.Hour)

	times := generateTimeSeries(startDate, endDate)
	fmt.Printf("Time range: %s to %s\n", startDate.Format(time.RFC3339), endDate.Format(time.RFC3339))
	fmt.Printf("Total timestamps: %d\n", len(times))

	// Process and write CSV
	processAndWrite(ephemeris, times, "/tmp/planetary_data_go.csv")

	fmt.Println("Done.")
}

func generateTimeSeries(start, end time.Time) []time.Time {
	startDate := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	endDate := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)

	var times []time.Time
	t := startDate
	for !t.After(endDate) {
		times = append(times, t)
		t = t.Add(1 * time.Hour)
	}
	return times
}

func processAndWrite(ephemeris *spk.SPK, times []time.Time, outputFile string) {
	os.MkdirAll(filepath.Dir(outputFile), 0755)
	os.Remove(outputFile)

	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating %s: %v\n", outputFile, err)
		os.Exit(1)
	}
	defer f.Close()

	fmt.Fprintln(f, CSVHeader())

	batchSize := 100000
	total := len(times)
	processed := 0

	for processed < total {
		end := processed + batchSize
		if end > total {
			end = total
		}

		for i := processed; i < end; i++ {
			row := ComputeRow(ephemeris, times[i])
			fmt.Fprintln(f, CSVRow(row))
		}

		processed = end
		pct := float64(processed) / float64(total) * 100
		fmt.Printf("  %s: %d/%d (%.1f%%)\n", outputFile, processed, total, pct)
	}
}
