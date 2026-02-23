# goeph — Feature Roadmap & Skyfield Gap Analysis

Reference: [Skyfield](https://github.com/skyfielders/python-skyfield)

This document catalogs every Skyfield feature, what goeph already has, and what remains — organized from easiest to hardest to implement.

---

## What goeph Already Implements

| Feature | Package | Notes |
|---|---|---|
| SPK Type 2 + Type 3 parsing + Clenshaw Chebyshev evaluation | `spk/` | Full DAF/SPK binary parser |
| Light-time corrected observations (`Observe`, `ObserveFrom`) | `spk/` | Up to 10 iterations, arbitrary observer |
| Geometric positions (`GeocentricPosition`) | `spk/` | No light-time correction |
| Apparent positions (`Apparent`, `ApparentFrom`) | `spk/` | Aberration + gravitational deflection |
| Velocity computation (Chebyshev derivative) | `spk/` | `EarthVelocity()`, internal chain |
| Generic segment graph walker | `spk/` | Builds body-to-SSB chains at `Open()` time |
| Temporal stacking (multiple segments per body pair) | `spk/` | Correct segment selected per epoch |
| ICRF to J2000 ecliptic lat/lon | `coord/` | Fixed J2000 mean obliquity |
| RA/Dec (J2000) to ICRF unit vector | `coord/` | |
| GMST / GAST | `coord/` | IAU 1982 Meeus formula |
| WGS84 geodetic to ICRF | `coord/` | Full precession + nutation chain |
| ITRF to geodetic (reverse) | `coord/` | Bowring's method, sub-mm accuracy |
| Altitude / Azimuth | `coord/` | Full ICRF→local horizon chain |
| Hour angle / Declination | `coord/` | True equator of date |
| IAU 2006 precession | `coord/` | Capitaine 2003 parameterization |
| IAU 2000A nutation (dual mode) | `coord/` | 30-term standard (~1 arcsec) or full 1365-term (~0.001 arcsec) via `SetNutationPrecision` |
| Stellar aberration | `coord/` | Full Lorentz transformation |
| Gravitational light deflection | `coord/` | Sun + Jupiter + Saturn deflectors |
| Visibility (is_sunlit, is_behind_earth) | `coord/` | Line-sphere intersection |
| Generic frame types (InertialFrame, TimeBasedFrame) | `coord/` | With predefined Galactic, B1950, Ecliptic, ITRF |
| TEME to ICRF conversion | `coord/`, `satellite/` | For SGP4 satellite positions |
| UTC to TT (hardcoded leap seconds) | `timescale/` | 28 entries, frozen at 2017-01-01 |
| TT to UT1 (hardcoded delta-T, cubic spline) | `timescale/` | 401 entries, 1800-2200 |
| SGP4 satellite propagation | `satellite/` | Wraps `go-satellite`, sub-satellite lat/lon |
| Osculating orbital elements | `elements/` | 19-field Keplerian elements from state vector |
| Planetary magnitudes | `magnitude/` | Mallama & Hilton 2018, Mercury–Neptune |
| Galactic Center ICRF direction | `star/` | Single hardcoded star |
| Mean lunar nodes | `lunarnodes/` | Meeus formula |

### Current Limitations (library-wide)

- ~~Observer is always Earth (399) — no arbitrary observer support~~ ✓ `ObserveFrom()` / `ApparentFrom()` added
- ~~No velocity computation — positions only~~ ✓ `EarthVelocity()` + internal `bodyVelWrtSSB()` added
- ~~No apparent positions (no aberration, no gravitational deflection)~~ ✓ `Apparent()` / `ApparentFrom()` added
- ~~No altaz (altitude/azimuth) for ground observers~~ ✓ `Altaz()` and `HourAngleDec()` added
- No event searching (sunrise/set, moon phases, conjunctions)
- No star catalog beyond single hardcoded Galactic Center
- ~~Only SPK Type 2 segments (no Type 3, 13, 21)~~ ✓ Type 2 and Type 3 now supported
- ~~Nutation limited to 30 of 687 IAU 2000A terms~~ ✓ Full 1365-term nutation available via `NutationFull` mode
- GMST uses IAU 1982 Meeus (not IERS 2000 ERA-based)
- Leap second table frozen at 2017
- ~~Delta-T uses linear interpolation (Skyfield uses cubic spline)~~ ✓ Cubic spline interpolation
- Requesting an unknown body causes panic (not error)

---

## Tier 1: Easy (Pure formulas, no new dependencies or data) — IMPLEMENTED

All Tier 1 features were implemented in PR #2 (`feature/tier1-formulas`), merged to main.

| Feature | Location | Golden Tests | Notes |
|---|---|---|---|
| Separation angle | `coord/angles.go` `SeparationAngle()` | `golden_separation.json` (3,653 tests, tol <1e-8°) | Kahan's numerically stable formula |
| Phase angle | `coord/angles.go` `PhaseAngle()` | `golden_phase.json` (21,918 tests, tol <1e-8°) | Exact input vectors from Skyfield; formula-level agreement |
| Fraction illuminated | `coord/angles.go` `FractionIlluminated()` | (tested via phase angle) | `(1 + cos(phase)) / 2` |
| Position angle | `coord/angles.go` `PositionAngle()` | unit tests | North-to-East, 0-360° range |
| Earth Rotation Angle | `coord/coord.go` `EarthRotationAngle()` | `golden_era.json` (3,653 tests, tol <1e-8°) | IAU 2000 ERA formula |
| Atmospheric refraction | `coord/refraction.go` `Refraction()`, `Refract()` | `golden_refraction.json` (182 tests, tol <1e-10°) | Bennett 1982, uses 0.016667 constant matching Skyfield |
| Galactic frame | `coord/frames.go` `GalacticMatrix` | unit tests (orthogonality, det=+1, galactic center) | IAU 1958 System II |
| B1950 frame | `coord/frames.go` `B1950Matrix` | unit tests (orthogonality) | FK4 mean equator |
| ICRS-to-J2000 bias | `coord/frames.go` `ICRSToJ2000Matrix` | unit tests (near-identity, non-identity) | IERS frame bias constants |
| Elongation | `coord/angles.go` `Elongation()` | `golden_elongation.json` (3,653 tests, tol <1e-10°) | Directional 0-360° |
| Unit types | `units/units.go` `Angle`, `Distance` | unit tests | DMS/HMS decomposition, AU/km/m/light-sec conversions |
| TDB-TT difference | `timescale/timescale.go` `TDBMinusTT()` | `golden_tdbtt.json` (3,653 tests, tol <1e-9s) | Fairhead & Bretagnon 7-term approximation |
| Line-sphere intersection | `geometry/geometry.go` `IntersectLineSphere()` | unit tests | Quadratic formula, NaN for miss |
| ICRF-to-Galactic helper | `coord/frames.go` `ICRFToGalactic()` | unit tests (galactic center, north pole) | Convenience function using GalacticMatrix |

---

## Tier 2: Moderate (Real algorithm work, but self-contained)

No external data files needed, but require meaningful implementation.

### Velocity computation (Chebyshev derivative) — IMPLEMENTED
- **Where**: `spk/spk.go` — `chebyshevDerivative()`, `segVelocity()`, `bodyVelWrtSSB()`, `EarthVelocity()`
- **Golden test**: `golden_velocity.json` (36,530 tests, tol <5 km/day)
- **Notes**: Uses derivative coefficient recurrence, chain rule factor `2.0 / seg.intLen * secPerDay` converts to km/day.

### Stellar aberration — IMPLEMENTED
- **Where**: `coord/aberration.go` — `Aberration(position, velocity, lightTime)`
- **Notes**: Full Lorentz transformation matching Skyfield's `add_aberration()`. Vector helpers in `coord/vec3.go`.

### Gravitational light deflection — IMPLEMENTED
- **Where**: `coord/deflection.go` — `Deflection(position, pe, rmass)`
- **Notes**: PPN formula matching Skyfield's `_compute_deflection()`. Default deflectors: Sun (rmass=1.0), Jupiter (1047.3486), Saturn (3497.898).

### Apparent positions — IMPLEMENTED
- **Where**: `spk/spk.go` — `Apparent(body, tdbJD)`, `ApparentFrom(observer, target, tdbJD)`
- **Golden test**: `golden_apparent.json` (36,530 tests, abs tol <50 km, rel tol <1.5e-5)
- **Notes**: Light-time iteration → deflection (with closest-approach timing) → aberration. Uses `NutationFull` for tightest Skyfield match.

### Arbitrary observer support — IMPLEMENTED
- **Where**: `spk/spk.go` — `ObserveFrom(observer, target, tdbJD)`, `ApparentFrom(observer, target, tdbJD)`
- **Notes**: `Observe()` now calls `ObserveFrom(Earth, body, tdbJD)`. No API breakage.

### Altitude / Azimuth (altaz) — IMPLEMENTED
- **Where**: `coord/altaz.go` — `Altaz(posICRF, latDeg, lonDeg, jdUT1)`
- **Golden test**: `golden_altaz.json` (65,754 tests, alt tol <0.005°, az tol <0.02°)
- **Notes**: Full rotation chain: ICRF → frame bias → precession → nutation → GAST → local horizon. Matches Skyfield's `rotation_at()` + `altaz()` pipeline. Uses `NutationFull` for golden tests.

### Hour angle / Declination (hadec) — IMPLEMENTED
- **Where**: `coord/altaz.go` — `HourAngleDec(posICRF, lonDeg, jdUT1)`
- **Notes**: Uses precession/nutation chain to get true equator RA/Dec, then HA = GAST + lon - RA.

### SPK Type 3 segments — IMPLEMENTED
- **Where**: `spk/spk.go` — `Open()` now accepts Type 2 and Type 3 segments.
- **Notes**: Type 3 has interleaved position+velocity coefficients (`nCoeffs=(rsize-2)/6`). Velocity uses stored coefficients directly rather than derivative computation.

### Reverse geodetic (ITRF to lat/lon/height) — IMPLEMENTED
- **Where**: `coord/geodetic.go` — `ITRFToGeodetic(x, y, z)`
- **Notes**: Bowring's iterative method (3 iterations for sub-mm accuracy). Handles poles and height computation.

### Osculating orbital elements — IMPLEMENTED
- **Where**: `elements/elements.go` — `FromStateVector(posKm, velKmPerSec, muKm3s2)`
- **Notes**: Full Keplerian element set (19 fields): a, e, i, Ω, ω, ν, E, M, n, q, Q, P, and more. Handles circular, elliptical, parabolic, and hyperbolic orbits.

### Planetary magnitudes — IMPLEMENTED
- **Where**: `magnitude/magnitude.go` — `PlanetaryMagnitude()`, `PlanetaryMagnitudeWithGeometry()`
- **Notes**: Mallama & Hilton 2018 phase curves for Mercury through Neptune. Saturn ring tilt and Uranus axial tilt geometry included.

### Generic frame transform types — IMPLEMENTED
- **Where**: `coord/frames.go` — `InertialFrame`, `TimeBasedFrame` types
- **Notes**: `InertialFrame` wraps a static rotation matrix with `XYZ()` and `LatLon()` methods. `TimeBasedFrame` wraps a time-dependent rotation function. Predefined frames: `Galactic`, `B1950`, `Ecliptic`, and `ITRFFrame()`.

### is_sunlit / is_behind_earth — IMPLEMENTED
- **Where**: `coord/visibility.go` — `IsSunlit(posKm, sunPosKm)`, `IsBehindEarth(observerPosKm, targetPosKm)`
- **Notes**: Uses line-sphere intersection with Earth radius 6371 km.

### Satellite ICRF positions (TEME to GCRS) — IMPLEMENTED
- **Where**: `satellite/satellite.go` — `TEMEToICRF(posKmTEME, jdUT1)`, backed by `coord.TEMEToICRF()`
- **Notes**: Rotation chain: Rz(eq_eq) → N^T → P^T. Converts SGP4 TEME output to ICRF/GCRS for interoperability with planetary positions.

### Delta-T cubic spline interpolation — IMPLEMENTED
- **Where**: `timescale/timescale.go` — `DeltaT()` now uses natural cubic spline
- **Notes**: Precomputed second derivatives via Thomas algorithm at init time. Improved TT→UT1 accuracy from ~170 ms to ~65 ms vs Skyfield.

---

## Tier 3: Hard (Significant new subsystems)

Require new algorithmic infrastructure or substantial design.

### Numerical event search — IMPLEMENTED
- **Where**: `search/search.go` — `FindDiscrete()`, `FindMaxima()`, `FindMinima()`
- **Notes**: Bisection for discrete events (~1ms precision), golden section search for extrema (~1s precision). Foundation for all almanac features.

### Almanac: sunrise/sunset, twilight, moon phases, seasons, rise/set/transit, oppositions/conjunctions — IMPLEMENTED
- **Where**: `almanac/almanac.go`
- **Notes**: All almanac functions built on `search.FindDiscrete`. Includes `Seasons`, `MoonPhases`, `SunriseSunset`, `Twilight`, `Risings`, `Settings`, `Transits`, `OppositionsConjunctions`. Golden-tested against Skyfield (seasons ≤1 day, sunrise/sunset ≤5 min, twilight ≤10 min).

### Star catalog support — IMPLEMENTED
- **Where**: `star/star.go` — `Star` type with `PositionAU()`, `PositionKm()`, `RADec()`
- **Notes**: Proper motion, parallax, radial velocity propagation from catalog epoch. Doppler factor k for time dilation. Lazy initialization of ICRF position/velocity vectors.

### Kepler orbit propagation — IMPLEMENTED
- **Where**: `kepler/kepler.go` — `Orbit` type with `PositionAU()`, `PositionKm()`
- **Notes**: Solves Kepler's equation (Newton-Raphson) for elliptic, parabolic (Barker's equation), and hyperbolic orbits. Elements in J2000 ecliptic frame, output in ICRF. Supports both asteroid (a, e, M0) and comet (q, e, Tp) element sets.

### Full IAU 2000A nutation (1365 terms) — IMPLEMENTED
- **Where**: `coord/nutation.go` (API), `coord/nutation_data.go` (coefficient tables), `coord/coord.go` (dispatch)
- **Notes**: Dual-mode via `SetNutationPrecision()`: `NutationStandard` (30 terms, ~150 ns, default) or `NutationFull` (678 luni-solar + 687 planetary = 1365 terms, ~10.5 μs). Tables generated from Skyfield's `nutation.npz`. Golden tests use `NutationFull` for tightest Skyfield match.

### Satellite event finding (rise/culmination/set) — IMPLEMENTED
- **Where**: `satellite/satellite.go` — `FindEvents(sat, latDeg, lonDeg, startJD, endJD, minAltDeg)`
- **Notes**: Returns Rise/Culmination/Set events using `search.FindDiscrete` + `search.FindMaxima`. Topocentric positions via TEME→ICRF + observer subtraction. Step size 1 minute for LEO satellites.

### Lunar eclipse detection — IMPLEMENTED
- **Where**: `eclipse/eclipse.go` — `FindLunarEclipses(eph, startJD, endJD)`
- **Notes**: Finds penumbral, partial, and total lunar eclipses. Computes umbral/penumbral shadow cone radii with Danjon 2% enlargement. Reports eclipse type, magnitudes, and closest approach. Validated against NASA eclipse catalog (2024: 2 eclipses, 2022 Nov: total with magnitude 1.36).

---

## Tier 4: Very Hard (External data files, new parsers, large scope)

These require downloading/parsing external data files or building substantial new subsystems.

### Stereographic projection (sky charts) — IMPLEMENTED
- **Where**: `projection/projection.go` — `NewProjector()`, `Projector.Project()`
- **Notes**: Conformal projection of unit sphere to 2D plane. Pre-computes rotation constants at construction time. Matches Skyfield's SymPy-optimized formula.

### Constellation identification — IMPLEMENTED
- **Where**: `constellation/constellation.go` — `At(raHours, decDeg)`, `Name()`, `Abbreviation()`
- **Notes**: Grid-based lookup with binary search on RA/Dec (O(log n)). 88 IAU constellations, ~48KB embedded boundary data from CDS catalog VI/42. Validated against Skyfield.

### MPC orbit database — DEFERRED
- **What**: Parse Minor Planet Center orbit files to create Kepler orbits for ~1M asteroids.
- **Skyfield**: `load_mpcorb_dataframe()`, `load_comets_dataframe()` in `data/mpc.py`
- **Effort**: File parser + Kepler orbit propagation (Tier 3)
- **External data**: MPC orbit database files

### Hipparcos star catalog — DEFERRED
- **What**: Parse Hipparcos catalog files, bulk-create Star objects with proper motion.
- **Skyfield**: `data.hipparcos.load_dataframe()` in `data/`
- **Effort**: File parser + star propagation (Tier 3)
- **External data**: Hipparcos catalog (~50MB)

### Polar motion / ITRS frame — NOT PLANNED
- **What**: Apply polar motion (x, y pole offsets) to Earth rotation for precise ITRS coordinates. Needs IERS finals2000A data file.
- **Skyfield**: `itrs` frame in `framelib.py`, data from `data/iers.py`
- **Effort**: File parser + data management + frame math
- **External data**: IERS `finals2000A.all` file (~3MB, updated weekly)

### PCK kernel support (planetary surface locations) — NOT PLANNED
- **What**: Parse text + binary PCK files for body orientation parameters (pole RA/Dec/W rotation rate). Build orientation frames for Moon, Mars, etc. Place surface points on non-Earth bodies.
- **Skyfield**: `PlanetaryConstants` in `planetarylib.py`
- **Effort**: New binary parser (DAF format, same as SPK but different segment types) + orientation math + surface point computation
- **External data**: `.tpc` (text) and `.bpc` (binary) PCK kernel files from NAIF

### Tycho-2 star catalog — NOT PLANNED
- **What**: Parse Tycho-2 catalog files, bulk-create Star objects with proper motion.
- **Skyfield**: `data.tycho2.load_dataframe()` in `data/`
- **Effort**: File parser + star propagation (Tier 3)
- **External data**: Tycho-2 catalog (~500MB)

### Data loader / downloader — NOT PLANNED
- **What**: Automatically download ephemeris files, star catalogs, IERS data with caching, expiry checking, and directory management.
- **Skyfield**: `Loader` class in `iokit.py`
- **Effort**: HTTP client + file management + expiry logic
- **Design consideration**: Go libraries typically don't auto-download. Consider making this optional or a separate CLI tool.

---

## Recommended Implementation Order

The highest-value path to "complete Skyfield parity":

### Phase 1: Core position accuracy — DONE
1. ~~**Velocity computation** (Tier 2)~~ ✓
2. ~~**Stellar aberration** (Tier 2)~~ ✓
3. ~~**Gravitational deflection** (Tier 2)~~ ✓
4. ~~**Apparent positions** (Tier 2)~~ ✓

### Phase 2: Observer-centric astronomy — DONE
5. ~~**Altaz** (Tier 2)~~ ✓
6. ~~**Atmospheric refraction** (Tier 1)~~ ✓
7. ~~**Arbitrary observer** (Tier 2)~~ ✓

### Phase 3: Event finding — DONE
8. ~~**Numerical event search** (Tier 3)~~ ✓ `search/` package
9. ~~**Sunrise/sunset** (Tier 3)~~ ✓ `almanac/` package
10. ~~**Moon phases** (Tier 3)~~ ✓ `almanac/` package
11. ~~**Seasons, rise/set of any body** (Tier 3)~~ ✓ `almanac/` package

### Phase 4: Expanded body coverage — DONE
12. ~~**Star catalog support** (Tier 3)~~ ✓ `star/` package expansion
13. ~~**SPK Type 3** (Tier 2)~~ ✓
14. ~~**Kepler orbits** (Tier 3)~~ ✓ `kepler/` package

### Phase 5: Precision improvements — DONE
15. ~~**Full IAU 2000A nutation** (Tier 3) — 687 terms~~ ✓ Dual-mode (30 / 1365 terms)
16. ~~**ERA-based sidereal time** (Tier 1)~~ ✓
17. ~~**Delta-T cubic spline** (Tier 2)~~ ✓

### Phase 6: Nice-to-haves — DONE
18. ~~**Planetary magnitudes** (Tier 2)~~ ✓
19. ~~**Osculating elements** (Tier 2)~~ ✓
20. ~~**Lunar eclipse detection** (Tier 3)~~ ✓ `eclipse/` package
21. ~~**Satellite events** (Tier 3)~~ ✓ `satellite/` FindEvents
22. ~~**Unit types** (Tier 1)~~ ✓
23. ~~**Additional frames** (galactic, B1950) (Tier 1)~~ ✓
24. ~~**Generic frame types** (Tier 2)~~ ✓
25. ~~**is_sunlit / is_behind_earth** (Tier 2)~~ ✓
26. ~~**Reverse geodetic** (Tier 2)~~ ✓
27. ~~**Satellite TEME→ICRF** (Tier 2)~~ ✓

### Phase 7: Final features — DONE
28. ~~**Stereographic projection** (Tier 4)~~ ✓ `projection/` package
29. ~~**Constellation identification** (Tier 4)~~ ✓ `constellation/` package

---

## Verification Strategy

Correctness for a computational library cannot be established by unit tests alone — unit tests verify logic and edge cases, but don't guarantee that the math produces the right numbers. All computational correctness is verified against Skyfield via golden tests.

### Two layers of testing

**1. Golden tests (correctness against Skyfield)**

Every new feature gets golden test data generated by extending `testdata/generate_golden.py`. The golden JSON files are checked into git so CI runs without Python or Skyfield.

- Golden data is the **source of truth** for whether goeph produces correct results.
- Each golden file documents its expected tolerance and the reason for any gap (e.g., 30-term vs 687-term nutation).

**2. Unit tests (logic, edge cases, regressions)**

Standard Go unit tests for:
- Input validation and error handling
- Edge cases (zero vectors, pole positions, boundary dates)
- Regressions from bug fixes
- Internal helper functions

Unit tests verify the code does what we intended. Golden tests verify that what we intended is actually correct.

**3. Examples (documentation and API smoke tests)**

Every new feature must have a runnable example in `examples/` before merging. Examples serve as living documentation and catch API breakage via `go build ./examples/...` in CI. Each example should:
- Be a standalone `main.go` in its own subdirectory
- Have a doc comment explaining the concept
- Produce readable printed output for a minimal scenario (one date, one body, etc.)
- Compile and run successfully from the repo root

### Sampling strategy for golden data

**Primary grid: 30-day increments over the full DE440s range (1850–2149)**

This gives 3,653 test points and is the standard for all features. It is dense enough to catch:
- Systematic drift from approximations (nutation, delta-T, precession)
- Error growth over centuries away from J2000
- Position-dependent patterns across orbits

Fine-grained sampling (e.g., 1-minute increments) is not needed because:
- The underlying physics is smooth (Chebyshev polynomials, continuous rotations)
- If positions agree at day 1 and day 31, they agree at day 15.5 — there are no sub-month oscillations that the algorithms would miss
- It would balloon golden file sizes (43K points/month vs 3,653 total) for no real gain

**Targeted edge-case tests for specific scenarios**

Instead of fine sampling, add surgical test cases for:
- Leap second boundaries (specific dates right at midnight on known leap second days)
- SPK segment temporal stacking boundaries (dates where segments switch)
- Almanac event times (compare specific event times directly, not a grid)
- Any algorithm with piecewise behavior

### Testing compound features at multiple levels

For features built on other features (e.g., apparent positions = light-time + aberration + deflection):

1. **Test intermediates separately** — extract Skyfield's intermediate values in `generate_golden.py`:
   ```python
   astrometric = earth.at(t).observe(mars)
   # astrometric.position.km  → test against goeph Observe()
   apparent = astrometric.apparent()
   # apparent.position.km     → test against goeph Apparent()
   ```
2. **Test the combined result** — with a tolerance that accounts for known gaps (e.g., nutation precision)
3. This pinpoints *where* a disagreement originates rather than debugging a combined result.

### Tolerance documentation

Every golden test must document:
- The measured tolerance vs Skyfield
- The theoretical reason for any gap
- Whether the gap is expected to grow with distance from J2000

See [`testdata/README.md`](testdata/README.md) for the full tolerance table with error sources and sampling strategy. Summary:

- SPK positions: <0.2 km
- Velocity: <5 km/day
- Apparent positions: <50 km abs, <1.5e-5 relative
- Altaz altitude: <0.005°, azimuth: <0.02° (with NutationFull + frame bias)
- UTC→TT: <1e-9 days, TT→UT1: <1e-6 days (~65 ms, cubic spline)
- GMST: <1e-3°, Geodetic→ecliptic: <0.025° (with NutationFull)
- ERA: <1e-8°, TDB-TT: <1e-9 s, Separation: <1e-8°, Phase angle: <1e-8°
- Elongation: <1e-10°, Refraction: <1e-10°, Lunar nodes: <1e-8°

New features will add to this table as they are implemented and validated.
