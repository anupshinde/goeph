# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

goeph is a Go library for computing planetary positions from JPL SPK ephemeris files. It is a port of core functionality from Python's [Skyfield](https://github.com/skyfielders/python-skyfield) library. All outputs are validated against Skyfield using golden tests.

## Build & Test Commands

```bash
go build ./...                    # build all packages
go test ./...                     # run all tests (no tests exist yet)
go test -run TestName ./spk/      # run a single test in a package
go vet ./...                      # static analysis
```

The project requires Go 1.22+. Single external dependency: `github.com/joshuaferrara/go-satellite`.

## Architecture

The library is organized as independent packages with no circular dependencies:

- **`spk/`** — Core package. Parses JPL SPK/DAF binary ephemeris files (Type 2 segments only). Evaluates Chebyshev polynomials via Clenshaw algorithm. Provides `Observe()` for light-time corrected geocentric positions and `GeocentricPosition()` for geometric positions. All positions are in km, ICRF frame. Body lookups use a hardcoded `bodyWrtSSB()` switch (not a generic segment graph walker like Skyfield).
- **`coord/`** — Coordinate transforms: ICRF↔ecliptic, RA/Dec↔ICRF, geodetic↔ICRF. Includes IAU 2006 precession, IAU 2000A nutation (top 30 terms only), GMST/GAST sidereal time, and WGS84 geodetic conversion.
- **`timescale/`** — Time scale chain: UTC→TT (via hardcoded leap second table + 32.184s) → UT1 (via hardcoded delta-T table, 1800–2200). Converts `time.Time` to Julian dates.
- **`satellite/`** — Thin wrapper around `go-satellite` for SGP4 propagation.
- **`star/`** — Fixed star coordinates (currently only Galactic Center).
- **`lunarnodes/`** — Mean lunar node ecliptic longitude computation (Meeus formula).

### Data flow for a typical computation

1. `timescale.TimeToJDUTC()` → `timescale.UTCToTT()` to get TDB Julian date (TDB≈TT)
2. `spk.Open()` loads a `.bsp` file, `spk.Observe(bodyID, tdbJD)` returns light-time corrected ICRF km
3. `coord.ICRFToEcliptic()` or other converters produce final coordinates

### Key design decisions

- Nutation uses only 30 of 687 IAU 2000A terms (~1 arcsecond precision, not micro-arcsecond)
- `bodyWrtSSB()` panics on unknown body IDs — only supports bodies listed in `spk/bodies.go`
- Leap second and delta-T tables are hardcoded arrays, not loaded from external files
- No velocity computation, no apparent positions (no aberration/deflection)

## Validation

Golden tests compare goeph output against Skyfield (Python) at full float64 precision. Expected tolerances: SPK positions <1e-6 km, ecliptic coords <1e-10°, time conversions exact or <1e-12 days.

## Examples

`examples/celestial_csv/` contains a complete pipeline that computes positions for all planets, satellites, and ground locations, outputting CSV.
