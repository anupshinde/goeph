# goeph

[![Tests](https://github.com/anupshinde/goeph/actions/workflows/test.yml/badge.svg)](https://github.com/anupshinde/goeph/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/anupshinde/goeph)](https://goreportcard.com/report/github.com/anupshinde/goeph)

A fast Go library for computing planetary positions from JPL ephemeris files.

---

## Origin & attribution

goeph is a Go library inspired by **[Skyfield](https://github.com/skyfielders/python-skyfield)**, the excellent Python library by Brandon Rhodes for computing positions of stars, planets, and satellites.

- **Skyfield** is the reference implementation. All goeph outputs are validated against Skyfield.
- goeph is **not affiliated with, endorsed by, or a replacement for** the Skyfield project.
- The mathematical foundations (IAU nutation/precession models, Chebyshev polynomial evaluation, SPK binary format) are from published standards (IERS Conventions, JPL/NAIF documentation).

If you need authoritative astronomy or Skyfield's full feature set, use [Skyfield](https://github.com/skyfielders/python-skyfield) directly.

---

## Why this exists

I needed fast planetary position computation in Go for a research project. Skyfield is great but Python was too slow for my batch workloads. I couldn't find a Go library that did what I needed, so I had Claude build one inspired by Skyfield's approach. It's ~14x faster than the Python equivalent for the same computations.

I'm publishing this because someone else is probably looking for the same thing I was.

This project was [coded by AI](#ai-disclosure) and validated against Skyfield using [golden tests](#validation-against-skyfield) and [end-to-end CSV comparison](validation/).

---

## What it does

- **Loads JPL DE-series ephemeris files** (tested with DE440s; should work with any SPK containing Type 2 segments)
- **Computes body positions** with light-time correction (Sun, Moon, all planets, Pluto, barycenters)
- **Converts coordinates** between ICRF, ecliptic, RA/Dec, geodetic, and galactic frames
- **Handles time scales** (UTC → TT → UT1, leap seconds, delta-T, TDB-TT)
- **Computes sidereal time** (GMST, GAST with nutation correction, Earth Rotation Angle)
- **Transforms geodetic positions** to celestial coordinates (WGS84, nutation, precession)
- **Computes angular quantities** — separation angle, phase angle, fraction illuminated, position angle, elongation
- **Atmospheric refraction** — Bennett's formula for altitude correction
- **Unit types** — `Angle` (degrees, hours, radians, DMS, HMS) and `Distance` (km, AU, meters, light-seconds)
- **Frame rotations** — Galactic (IAU 1958), B1950 (FK4), ICRS-to-J2000 bias matrices
- **Geometric computations** — line-sphere intersection for shadow/limb checks
- **Computes lunar node longitudes** (Meeus formula — not part of Skyfield, added separately)
- **Propagates satellites** via SGP4 (wraps go-satellite)

## What it doesn't do

- No apparent positions (no relativistic deflection or aberration)
- No altaz computation
- No sunrise/sunset, moon phases, or event searching
- No star catalog loading
- No velocity computation
- No SPK Type 3/13/21 support
- Nutation uses top 30 IAU 2000A terms (not all 687) — sufficient for sub-arcsecond work, not micro-arcsecond

Contributions for any of these are welcome — see "Project status & support" below.

---

## Quick start

```go
package main

import (
    "fmt"
    "time"

    "github.com/anupshinde/goeph/coord"
    "github.com/anupshinde/goeph/spk"
    "github.com/anupshinde/goeph/timescale"
)

func main() {
    // Load the included ephemeris file
    eph, err := spk.Open("data/de440s.bsp")
    if err != nil {
        panic(err)
    }

    // Pick a time
    t := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
    jdUTC := timescale.TimeToJDUTC(t)
    tdbJD := timescale.UTCToTT(jdUTC) // TDB ≈ TT for our purposes

    // Get Moon's position (light-time corrected, ICRF frame, km)
    pos := eph.Observe(spk.Moon, tdbJD)

    // Convert to ecliptic latitude/longitude
    lat, lon := coord.ICRFToEcliptic(pos[0], pos[1], pos[2])
    fmt.Printf("Moon ecliptic: lat=%.6f° lon=%.6f°\n", lat, lon)

    // Get Mars position
    mars := eph.Observe(spk.MarsBarycenter, tdbJD)
    marsLat, marsLon := coord.ICRFToEcliptic(mars[0], mars[1], mars[2])
    fmt.Printf("Mars ecliptic: lat=%.6f° lon=%.6f°\n", marsLat, marsLon)
}
```

See [`examples/`](examples/) for 13 runnable examples covering the full API, or [`validation/generate_data_go/`](validation/generate_data_go/) for a complete working pipeline that computes positions for all planets, satellites, and ground locations, outputting to CSV.

---

## Installation

```bash
go get github.com/anupshinde/goeph
```

An ephemeris file (`de440s.bsp`, ~32 MB) is included in `data/`. You can also download others from NASA:
- [de440s.bsp](https://naif.jpl.nasa.gov/pub/naif/generic_kernels/spk/planets/de440s.bsp) (~32 MB, 1849-2150) — included & tested
- [de421.bsp](https://naif.jpl.nasa.gov/pub/naif/generic_kernels/spk/planets/de421.bsp) (~17 MB, 1900-2050) — untested
- [de440.bsp](https://naif.jpl.nasa.gov/pub/naif/generic_kernels/spk/planets/de440.bsp) (~115 MB, 1550-2650) — untested

---

## Packages

| Package | Import | What it does |
|---------|--------|-------------|
| `spk` | `goeph/spk` | SPK/DAF ephemeris file parser, Chebyshev polynomial evaluation, light-time corrected positions |
| `coord` | `goeph/coord` | ICRF↔ecliptic, RA/Dec↔ICRF, geodetic↔ICRF, galactic, GMST/GAST/ERA, nutation (IAU 2000A), precession (IAU 2006), separation/phase/position angles, elongation, refraction |
| `timescale` | `goeph/timescale` | UTC→TT→UT1 conversions, leap second table, delta-T table (1800-2200), TDB-TT |
| `units` | `goeph/units` | `Angle` and `Distance` types with unit conversions (degrees, hours, radians, DMS/HMS, km, AU, light-seconds) |
| `geometry` | `goeph/geometry` | Line-sphere intersection (for shadow/limb geometry) |
| `satellite` | `goeph/satellite` | SGP4 satellite propagation, sub-satellite point computation |
| `star` | `goeph/star` | Fixed star coordinates (Galactic Center) |
| `lunarnodes` | `goeph/lunarnodes` | Mean lunar node ecliptic longitudes (not from Skyfield; uses Meeus formula) |

---

## Supported bodies

All bodies available in DE-series ephemeris files:

| Body | ID | Notes |
|------|----|-------|
| Sun | 10 | |
| Moon | 301 | |
| Mercury | 199 | Via Mercury Barycenter chain |
| Venus | 299 | Via Venus Barycenter chain |
| Mars Barycenter | 4 | |
| Jupiter Barycenter | 5 | |
| Saturn Barycenter | 6 | |
| Uranus Barycenter | 7 | |
| Neptune Barycenter | 8 | |
| Pluto Barycenter | 9 | |
| Earth | 399 | |
| Earth-Moon Barycenter | 3 | |
| Planet Barycenters | 1-9 | Direct segments vs SSB |

---

## Validation against Skyfield

goeph outputs are verified against Skyfield (Python) using a golden-test approach:

1. A Python script (`testdata/generate_golden.py`) runs Skyfield for 3,653 dates at 30-day increments across the full DE440s range (1850–2149), covering all 10 bodies, 6 geographic locations, timescale conversions, GMST, and lunar nodes
2. Outputs are saved as JSON golden files at full float64 precision
3. Go tests read the golden files and compare goeph results within documented tolerances
4. CI runs all tests automatically on push/PR via GitHub Actions

| Computation | Measured tolerance | Notes |
|------------|-------------------|-------|
| SPK positions (ICRF) | < 0.2 km | Mercury worst case (barycenter chain); most bodies < 0.01 km |
| UTC→TT conversion | < 1e-9 days | Same leap second table |
| TT→UT1 conversion | < 2e-6 days | Linear interpolation vs Skyfield's spline for delta-T |
| GMST | < 1e-3° | goeph uses IAU 1982 (Meeus); Skyfield uses IERS 2000 ERA-based |
| Geodetic→ecliptic | < 0.035° | 30-term nutation gap grows with distance from J2000 |
| Lunar nodes | < 1e-8° | Identical Meeus formula |
| ERA | < 1e-8° | Exact same IAU 2000 formula |
| TDB-TT | < 1e-9 s | Same Fairhead & Bretagnon terms |
| Separation angle | < 1e-8° | Same position vectors, same formula |
| Phase angle | < 0.5° | Barycentric vs astrometric vector reconstruction |
| Elongation | < 1e-10° | Pure modular arithmetic |
| Refraction | < 1e-10° | Same Bennett 1982 formula |

The nutation gap (30 vs 687 IAU 2000A terms) is the main deviation from Skyfield. It produces ~1 arcsecond error near J2000, growing to ~0.03° (~113 arcsec) at the extremes of the 300-year test range. For sub-arcsecond work near the present, this is fine; for micro-arcsecond precision or dates far from J2000, the full model would need to be ported.

In addition to golden tests, the [`validation/`](validation/) directory contains Go and Python data generators that produce identical CSV outputs over a 200-year range, with a comparison script to verify column-by-column accuracy. See:

- [`docs/BENCHMARK_GO_VS_PYTHON.md`](docs/BENCHMARK_GO_VS_PYTHON.md) — detailed timing and accuracy benchmarks (~14x faster than Python/Skyfield)
- [`docs/PYTHON_SKYFIELD_TO_GO.md`](docs/PYTHON_SKYFIELD_TO_GO.md) — how the Python→Go port was done and the math behind it

---

## Known limitations

1. **SPK Type 2 only** — rejects non-Type-2 segments. All JPL DE-series files use Type 2. Asteroid/comet BSP files (Type 3, 13, 21) are not supported.
2. **Geocentric only** — `Observe()` and `GeocentricPosition()` use Earth as the observer. Arbitrary observer bodies are not yet supported.
3. **No apparent positions** — light-time correction is applied, but not gravitational deflection or aberration (~1.7 arcsec near Sun).
4. **No velocity** — only positions are computed.
5. **30-term nutation** — sub-arcsecond, not micro-arcsecond.

---

## Project status & support

This is a **personal research artifact published as-is.**

- No guaranteed support, response time, or roadmap.
- I use this myself. If it breaks for my use case, I'll fix it.
- PRs are accepted if they:
  - Pass golden tests
  - Reduce error vs Skyfield, or improve performance without increasing error
  - Don't expand scope without strong justification

If you need a maintained, production-grade library, fork it and own it for your use case.

---

## Forks & alternatives

If you build something better — more complete, faster, more accurate — I'm happy to link to it. Open a PR or issue with a link to your project.

---

## AI disclosure

This project was **coded by [Claude Opus 4.6](https://claude.ai)** (Anthropic), directed by a human and verified with tests compared to [Skyfield](https://github.com/skyfielders/python-skyfield) (Python) output. The initial implementation was completed rapidly with AI — in a single day — which is why the golden test harness and end-to-end validation are critical.

- The Go implementation of Skyfield-inspired algorithms was done by Claude
- Benchmarking and output comparison against Skyfield was done collaboratively
- The golden test strategy ensures correctness is verifiable regardless of how the code was written

Most of the code itself is not reviewed manually, unless when it was absolutely needed — the outputs were measured for correctness over very long periods. The golden test harness ensures correctness is measurable, not assumed.

The code is the code. The tests are the tests. Judge it by its outputs.

> This README is also maintained by AI.

---

## License

MIT — same as Skyfield.

---

## Acknowledgements

- **[Skyfield](https://github.com/skyfielders/python-skyfield)** by Brandon Rhodes — the reference implementation this project validates against. Excellent library; use it if Python works for your use case.
- **[JPL/NAIF](https://naif.jpl.nasa.gov/)** — for the SPK ephemeris format and DE-series planetary ephemeris data.
- **[Claude Opus 4.6](https://claude.ai)** (Anthropic) — AI that wrote the Go implementation.
- **[go-satellite](https://github.com/joshuaferrara/go-satellite)** — SGP4 propagation library used for satellite computations.
