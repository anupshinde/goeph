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

I needed fast planetary position computation in Go for a research project. Skyfield is great but Python was too slow for my batch workloads. I couldn't find a Go library that did what I needed, so I had Claude build one inspired by Skyfield's approach. It's ~14x faster than the Python equivalent for the same computations (with the default 30-term nutation; ~2-3x faster with full 1365-term nutation matching Skyfield's precision — see [Nutation precision](#nutation-precision)).

I'm publishing this because someone else is probably looking for the same thing I was.

This project was [coded by AI](#ai-disclosure) and validated against Skyfield using [golden tests](#validation-against-skyfield) and [end-to-end CSV comparison](validation/).

---

## What it does

- **Loads JPL DE-series ephemeris files** (tested with DE440s; supports SPK Type 2 and Type 3 segments)
- **Computes body positions** with light-time correction (Sun, Moon, all planets, Pluto, barycenters)
- **Apparent positions** — stellar aberration (full Lorentz) + gravitational deflection (Sun/Jupiter/Saturn)
- **Velocity computation** — Chebyshev polynomial derivatives for body velocities
- **Arbitrary observer** — observe from any body, not just Earth (`ObserveFrom`, `ApparentFrom`)
- **Altitude / Azimuth** — for ground observers at any lat/lon, with hour angle / declination
- **Converts coordinates** between ICRF, ecliptic, RA/Dec, geodetic, ITRF, and galactic frames
- **Handles time scales** (UTC → TT → UT1, leap seconds, delta-T with cubic spline, TDB-TT)
- **Computes sidereal time** (GMST, GAST with nutation correction, Earth Rotation Angle)
- **Transforms geodetic positions** to celestial coordinates (WGS84, nutation, precession) and back (ITRF to geodetic)
- **Computes angular quantities** — separation angle, phase angle, fraction illuminated, position angle, elongation
- **Osculating orbital elements** — Keplerian elements from state vectors (elliptical, parabolic, hyperbolic)
- **Planetary magnitudes** — Mallama & Hilton 2018 phase curves (Mercury through Neptune)
- **Visibility checks** — is_sunlit, is_behind_earth (shadow/limb geometry)
- **Atmospheric refraction** — Bennett's formula for altitude correction
- **Unit types** — `Angle` (degrees, hours, radians, DMS, HMS) and `Distance` (km, AU, meters, light-seconds)
- **Frame rotations** — Galactic (IAU 1958), B1950 (FK4), Ecliptic, ITRF, ICRS-to-J2000 bias; generic `InertialFrame` and `TimeBasedFrame` types
- **Geometric computations** — line-sphere intersection for shadow/limb checks
- **Computes lunar node longitudes** (Meeus formula — not part of Skyfield, added separately)
- **Propagates satellites** via SGP4 (wraps go-satellite) with TEME→ICRF conversion and rise/culmination/set event finding
- **Event search** — generic discrete event finding (bisection) and extrema finding (golden section) for time-series data
- **Almanac** — sunrise/sunset, twilight levels, moon phases, seasons, body risings/settings, meridian transits, oppositions/conjunctions
- **Star positions** — proper motion, parallax, and radial velocity propagation from catalog epoch to any date
- **Kepler orbit propagation** — elliptic, parabolic, and hyperbolic orbits for asteroids and comets
- **Lunar eclipse detection** — finds penumbral, partial, and total lunar eclipses with magnitudes and shadow geometry
- **Stereographic projection** — project sky positions onto a 2D plane for star charts (conformal, preserves angles)
- **Constellation identification** — identify which of the 88 IAU constellations contains a given RA/Dec position (grid-based binary search lookup)

## What it doesn't do

- No solar eclipse computation
- No star catalog file loading (Hipparcos/Tycho-2 parsers)
- No SPK Type 13/21 support
- GMST uses IAU 1982 formula (Skyfield uses IERS 2000 ERA-based) — introduces ~0.3 arcsec/century drift

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

See [`examples/`](examples/) for 24 runnable examples covering the full API, or [`validation/generate_data_go/`](validation/generate_data_go/) for a complete working pipeline that computes positions for all planets, satellites, and ground locations, outputting to CSV.

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
| `spk` | `goeph/spk` | SPK/DAF ephemeris file parser (Type 2 + 3), Chebyshev evaluation, positions (geometric, astrometric, apparent), velocity, arbitrary observer |
| `coord` | `goeph/coord` | ICRF↔ecliptic, RA/Dec, geodetic↔ICRF↔ITRF, galactic, altaz, hour angle/dec, TEME↔ICRF, GMST/GAST/ERA, nutation, precession, aberration, deflection, angles, refraction, visibility, generic frame types |
| `timescale` | `goeph/timescale` | UTC→TT→UT1 conversions, leap second table, delta-T cubic spline (1800-2200), TDB-TT |
| `elements` | `goeph/elements` | Osculating Keplerian orbital elements from state vectors |
| `magnitude` | `goeph/magnitude` | Planetary visual magnitudes (Mallama & Hilton 2018 phase curves) |
| `units` | `goeph/units` | `Angle` and `Distance` types with unit conversions (degrees, hours, radians, DMS/HMS, km, AU, light-seconds) |
| `geometry` | `goeph/geometry` | Line-sphere intersection (for shadow/limb geometry) |
| `search` | `goeph/search` | Generic event search: `FindDiscrete` (bisection), `FindMaxima`/`FindMinima` (golden section) |
| `almanac` | `goeph/almanac` | Sunrise/sunset, twilight, moon phases, seasons, body risings/settings, transits, oppositions/conjunctions |
| `satellite` | `goeph/satellite` | SGP4 satellite propagation, sub-satellite point, TEME→ICRF conversion, rise/culmination/set event finding |
| `star` | `goeph/star` | Star positions with proper motion, parallax, and radial velocity propagation; Galactic Center ICRF direction |
| `kepler` | `goeph/kepler` | Keplerian orbit propagation for asteroids and comets (elliptic, parabolic, hyperbolic) |
| `eclipse` | `goeph/eclipse` | Lunar eclipse detection: penumbral, partial, and total eclipses with magnitudes |
| `projection` | `goeph/projection` | Stereographic projection of sky positions onto a 2D plane for star charts |
| `constellation` | `goeph/constellation` | IAU constellation identification from RA/Dec (88 constellations, grid-based binary search) |
| `lunarnodes` | `goeph/lunarnodes` | Mean lunar node ecliptic longitudes (not from Skyfield; uses Meeus formula) |

---

## Nutation precision

The `coord` package supports two nutation precision modes via `SetNutationPrecision`:

| Mode | Terms | Precision | Speed | Use case |
|------|-------|-----------|-------|----------|
| `NutationStandard` (default) | 30 luni-solar | ~1 arcsec | ~150 ns/call | Batch processing, general astronomy |
| `NutationFull` | 678 luni-solar + 687 planetary | ~0.001 arcsec | ~10.5 μs/call | High-precision single-point, Skyfield parity |

```go
// For high-precision mode (matches Skyfield's default):
coord.SetNutationPrecision(coord.NutationFull)

// To restore fast mode:
coord.SetNutationPrecision(coord.NutationStandard)
```

**When to use which:** Nutation only affects functions that use equator-of-date transforms (`Altaz`, `HourAngleDec`, `GeodeticToICRF`, `TEMEToICRF`, `ITRFFrame`). Position computations (`Observe`, `Apparent`, coordinate conversions) are unaffected by this setting. In both modes, the dominant error sources are light-time correction (~20 arcsec) and GMST formula difference (~0.3 arcsec/century), which dwarf the nutation precision difference. Use `NutationFull` only when you need sub-arcsecond nutation accuracy or exact Skyfield parity.

**Performance impact:** `NutationFull` is ~70x slower per nutation call (~10.5 μs vs ~150 ns). In end-to-end workloads that include altaz/geodetic computations, this translates to ~5-7x slower overall (nutation is one step among many). With `NutationStandard`, goeph is ~14x faster than Skyfield; with `NutationFull`, ~2-3x faster. Skyfield defaults to the full IAU 2000A series (678 luni-solar + 687 planetary terms) but also provides an `iau2000b()` variant (77 luni-solar, 0 planetary) for faster computation. goeph's `NutationFull` matches Skyfield's default for apples-to-apples comparison, while `NutationStandard` (30 terms) is a different truncation than Skyfield's `iau2000b()` but serves a similar speed-vs-precision tradeoff. The ~1 arcsec difference from `NutationStandard` is negligible for most applications.

Call `SetNutationPrecision` once at program startup. It is not safe for concurrent use with different precision needs.

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

1. A Python script (`testdata/generate_golden.py`) runs Skyfield for 3,653 dates at 30-day increments across the full DE440s range (1850–2149), covering 10 bodies, 6 geographic locations, timescale conversions, angular quantities, and more
2. Outputs are saved as JSON golden files at full float64 precision
3. Go tests read the golden files and compare goeph results within documented tolerances
4. CI runs all tests automatically on push/PR via GitHub Actions

| Computation | Measured tolerance | Notes |
|------------|-------------------|-------|
| SPK positions (ICRF) | < 0.01 km | Mercury worst case ~0.002 km (with TDB-TT correction) |
| Velocity | < 0.01 km/day | Chebyshev derivative ~0.0002 km/day (with TDB-TT correction) |
| Apparent positions | < 50 km abs, < 1.5e-5 relative | Light-time in Skyfield's observe() (~20 arcsec, scales with distance) |
| Altitude | < 0.005° | GMST formula difference |
| Azimuth | < 0.02° | GMST formula + minor geometry effects |
| UTC→TT | < 1e-9 days | Same leap second table |
| TT→UT1 | < 1e-6 days (~65 ms) | Cubic spline vs Skyfield's spline (different source knots) |
| GMST | < 1e-3° | goeph uses IAU 1982 (Meeus); Skyfield uses IERS 2000 ERA-based |
| Geodetic→ecliptic | < 0.025° | Light-time in Skyfield's observe() for surface locations |
| ERA | < 1e-8° | Exact same IAU 2000 formula |
| TDB-TT | < 1e-9 s | Same Fairhead & Bretagnon terms |
| Separation angle | < 1e-8° | Same position vectors, same formula |
| Phase angle | < 1e-8° | Exact input vectors from Skyfield; formula-level agreement |
| Elongation | < 1e-10° | Pure modular arithmetic |
| Refraction | < 1e-10° | Same Bennett 1982 formula |
| Lunar nodes | < 1e-8° | Identical Meeus formula |
| Seasons | < 1 day | J2000 ecliptic vs Skyfield's ecliptic of date (~18 hours) |
| Moon phases | < 3 min | Relative longitude cancels frame difference (measured max ~4 sec) |
| Sunrise/sunset | < 3 min | Altaz chain; no ecliptic frame dependency (measured max ~1 sec) |
| Twilight | < 3 min | Altaz chain; extra events from finer step size (measured max ~1 sec) |
| Oppositions | < 3 min | Relative longitude cancels frame difference (measured max ~46 sec) |

See [`testdata/README.md`](testdata/README.md) for the full tolerance breakdown with error sources.

The remaining deviations from Skyfield are primarily due to: (1) light-time correction applied by Skyfield's `observe()` for surface locations (~20 arcsec), and (2) GMST formula differences (IAU 1982 vs IERS 2000 ERA-based, ~0.3 arcsec/century).

In addition to golden tests, the [`validation/`](validation/) directory contains Go and Python data generators that produce identical CSV outputs over a 200-year range, with a comparison script to verify column-by-column accuracy. See:

- [`docs/BENCHMARK_GO_VS_PYTHON.md`](docs/BENCHMARK_GO_VS_PYTHON.md) — detailed timing and accuracy benchmarks (~14x faster with default nutation; ~2-3x with full nutation parity)
- [`docs/PYTHON_SKYFIELD_TO_GO.md`](docs/PYTHON_SKYFIELD_TO_GO.md) — how the Python→Go port was done and the math behind it

---

## Known limitations

1. **SPK Type 2 and 3 only** — rejects other segment types. All JPL DE-series files use Type 2. Some asteroid/comet BSP files use Type 3. Types 13, 21 are not supported.
2. **GMST uses IAU 1982 formula** — differs from Skyfield's IERS 2000 ERA-based GMST by ~0.3 arcsec/century.
3. **No star catalog file loading** — the `Star` type supports proper motion propagation, but there are no Hipparcos/Tycho-2 file parsers yet.
4. **Leap second table frozen at 2017** — new leap seconds require a code update.

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
