# Porting the Celestial Data Generator from Python to Go

## Overview

The Python program in `generate_data_py/` uses the **Skyfield** astronomy library with the **JPL DE440s.bsp** ephemeris file to compute geocentric ecliptic positions of celestial bodies, satellite sub-points, ground location zenith positions, and lunar node longitudes. The output is a CSV file covering a 200-year range centered on a reference date.

The goal was to port this entirely to Go (`generate_data_go/`) using the **goeph** library with **exact numerical parity** against the Python output — meaning we could not use simplified analytical models (like VSOP87) and had to read the same binary DE440s.bsp ephemeris file directly.

---

## The Problem

No production-quality Go library exists that can read JPL SPK/DAF ephemeris files. Python has Skyfield (which wraps `jplephem`), but the Go ecosystem has nothing equivalent. The options were:

1. **Use VSOP87 analytical models** — Fast to implement, but introduces errors of ~1 arcsecond for inner planets and worse for the Moon. Not acceptable for exact parity.
2. **Call Python from Go via FFI/subprocess** — Defeats the purpose of porting.
3. **Write a native SPK/DAF reader in Go** — Complex, but gives exact parity and zero external dependencies.

We chose option 3, which became the **goeph** library.

---

## Approach

### Step 1: Understand the DE440s.bsp Binary Format

The `.bsp` file uses NASA/NAIF's **DAF (Double precision Array File)** container format. Understanding this required inspecting the actual file bytes:

- **File record** (bytes 0–1023): Contains the magic string `"DAF/SPK "`, the number of double-precision (`ND=2`) and integer (`NI=6`) components per summary, and the pointer to the first summary record (`FWARD=62`).
- **Endianness marker** (bytes 88–95): `"LTL-IEEE"` — little-endian IEEE 754.
- **Summary records**: Linked list of 1024-byte records starting at record `FWARD`. Each contains up to ~20 segment summaries. Each summary is 40 bytes: 2 doubles (start/end epoch) + 6 int32s (target body, center body, reference frame, data type, start index, end index).
- **Segment data**: Chebyshev polynomial coefficients stored as arrays of float64 values. The last 4 values in each segment are metadata: `INIT` (initial epoch in TDB seconds past J2000), `INTLEN` (interval length in seconds), `RSIZE` (record size in doubles), and `N` (number of records).

DE440s contains **14 segments**, all SPK **Type 2** (Chebyshev position-only polynomials), covering bodies like Mercury (199 wrt 1), Venus (299 wrt 2), Earth-Moon Barycenter (3 wrt 0), Moon (301 wrt 3), Earth (399 wrt 3), Mars Barycenter (4 wrt 0), and so on.

### Step 2: Implement the SPK Type 2 Evaluator

For a given body at a given time:

1. Convert the requested TDB Julian Date to seconds past J2000.
2. Find the matching segment for the (target, center) pair.
3. Compute which Chebyshev record covers the requested time: `idx = (seconds - INIT) / INTLEN`.
4. Normalize time to [-1, 1] within that record's interval.
5. Evaluate the Chebyshev polynomial for each of x, y, z using the **Clenshaw algorithm**:
   ```
   s2 = 2·s
   w₀ = cₙ₋₁, w₁ = 0
   for i = n-2 down to 1: w₀, w₁ = cᵢ + s2·w₀ - w₁, w₀
   result = c₀ + s·w₀ - w₁
   ```

### Step 3: Body Position Chaining

JPL ephemeris files don't store every body relative to every other body. Instead, they use a hierarchical tree rooted at the Solar System Barycenter (SSB = 0). To get a geocentric position, we chain through the hierarchy:

```
Mercury geocentric = Mercury(199 wrt 1) + MercuryBary(1 wrt 0) - Earth(399 wrt 3) - EMB(3 wrt 0)
Moon geocentric    = Moon(301 wrt 3) - Earth(399 wrt 3)
Sun geocentric     = Sun(10 wrt 0) - Earth(399 wrt 3) - EMB(3 wrt 0)
```

This chaining logic is in `bodyWrtSSB()`, which computes any body's position relative to SSB by walking the segment hierarchy. The geocentric position is then `body_wrt_SSB - earth_wrt_SSB`.

### Step 4: Light-Time Correction

Skyfield's `observe()` method returns an **astrometric position** — the direction where the body was when the light we see now was emitted. This requires iterating:

1. Compute geometric geocentric position.
2. Calculate light-time: `Δt = distance / c` (c = 299,792.458 km/s).
3. Recompute the body's position at `t - Δt`.
4. Repeat until `|Δt_new - Δt_old| < 10⁻¹²` days (typically converges in 2–3 iterations).

### Step 5: Coordinate Transformations

**ICRF to J2000 Ecliptic** (matching Skyfield's `ecliptic_latlon()`):

Skyfield uses the **ECLIPJ2000** inertial frame with a fixed rotation matrix based on the J2000 mean obliquity (84381.448 arcseconds). This is a simple rotation around the X-axis:

```
eₓ =  x
eᵧ =  cos(ε)·y + sin(ε)·z
e_z = -sin(ε)·y + cos(ε)·z
```

Where `cos(ε) = 0.917482062069182` and `sin(ε) = 0.397777155931914`. Then convert to spherical coordinates: `lat = arcsin(e_z/r)`, `lon = atan2(eᵧ, eₓ)`.

These exact values were extracted by tracing through Skyfield's source code to ensure bit-exact parity.

**Geodetic to ICRF** (for ground location zenith vectors):

1. Convert WGS84 geodetic (lat, lon) to ITRF Cartesian using the WGS84 ellipsoid (a = 6378.137 km, f = 1/298.257223563).
2. Rotate from ITRF to true-equinox-of-date frame by GAST (Greenwich Apparent Sidereal Time = GMST + Δψ·cos(ε), where Δψ is the nutation in longitude).
3. Apply the **inverse nutation matrix** (N^T) using IAU 2000A nutation (30-term truncation: Δψ, Δε from the 30 largest luni-solar terms).
4. Apply the **inverse IAU 2006 precession matrix** (P^T) to transform from mean-equinox-of-date to J2000/ICRF.

The full rotation chain is: `ICRF = P^T · N^T · R₃(GAST) · ITRF`.

### Step 6: Other Components

- **Galactic Center**: Fixed J2000 RA/Dec (Sgr A*: 17h45m40.0409s, -29°00'28.118") converted to an ICRF unit vector. Since it's effectively at infinity, no light-time correction is needed.
- **Lunar Nodes**: Direct port of the Meeus formula: `Ω = 125.04452 - 1934.136261·T + 0.0020708·T² + T³/450000`.
- **Satellites**: Three artificial satellites (FastISS, AntiISS, PoleSat) propagated via SGP4 using the `go-satellite` library with modified TLE data.
- **Time Conversion**: UTC → TT (via leap second table + 32.184s) → UT1 (via ΔT table). TDB ≈ TT for ephemeris lookups. See `timescale/` package.

---

## Challenges and Solutions

### Challenge 1: No Go Library for JPL Ephemeris

**Solution**: Wrote a complete DAF/SPK reader from scratch (~270 lines). This involved reading the NAIF DAF specification and inspecting the actual DE440s.bsp file bytes to understand the binary layout. The reader loads all segment data into memory on startup for fast repeated access.

### Challenge 2: Matching Skyfield's Exact Ecliptic Frame

Skyfield's `ecliptic_latlon()` could have used several different ecliptic frames (mean of date, true of date, J2000 inertial). **Solution**: Traced through Skyfield's source code (`framelib.py`, `positionlib.py`) to confirm it uses the ECLIPJ2000 inertial frame with a fixed rotation matrix. Extracted the exact sine/cosine values to 19 decimal places.

### Challenge 3: Ground Location Longitude Error (0.37° → 0.005° → 0.031°)

The ground location pipeline went through three stages of refinement:

**Stage 1 — GMST only (~0.37° error)**: The initial implementation used only GMST to rotate from ITRF to celestial coordinates, putting vectors in the "mean equinox of date" frame — not ICRF/J2000. The accumulated precession from J2000 to 2026 is ~0.36°, matching the observed error exactly.

**Stage 2 — Added precession (~0.005° error near present, 0.77° over 248yr)**: Added the IAU 2006 equinox-based precession matrix (ζ_A, z_A, θ_A) and applied P^T to transform from mean-equinox-of-date to J2000/ICRF. Near the present epoch this reduced error to ~0.005°, but over the full 248-year range errors grew to 0.77° due to a time scale mismatch (not nutation, as initially suspected).

**Stage 3 — Added nutation + ΔT time scale model (~0.031° max over 248yr)**: Two fixes applied:

1. **IAU 2000A nutation (30-term truncation)**: Implemented the 30 largest luni-solar nutation terms (out of Skyfield's 678), computing Δψ and Δε. Added GAST (= GMST + Δψ·cos(ε)) for the Earth rotation angle, and the nutation rotation matrix N^T in the geodetic→ICRF pipeline. The full rotation is now: `ICRF = P^T · N^T · R₃(GAST) · ITRF`. Nutation itself contributes only ~0.002° improvement — the 30-term truncation captures the dominant terms.

2. **ΔT time scale model**: The 0.77° error was from treating UT1 ≈ UTC. Skyfield converts UTC → TT (via leap seconds + 32.184s) → UT1 (via ΔT model). Implemented proper conversion using:
   - Leap second table: 28 entries (1972–2017) from IERS Bulletin C
   - ΔT table: 401 entries at yearly intervals (1800–2200), extracted from Skyfield's IERS data + Morrison-Stephenson-Hohenkerk 2020 model
   - Linear interpolation between yearly entries

   The residual 0.031° comes from yearly ΔT interpolation vs Skyfield's daily IERS data (mean error is 0.005°).

### Challenge 4: Satellite SGP4 Divergence

The Go `go-satellite` library and Python's `sgp4` library implement the same SGP4 algorithm but produce different results when propagating far from the TLE epoch (the TLEs are from early 2024, being propagated to 2026). Even an unmodified ISS shows ~0.8° divergence after 2 years.

**Solution**: Accepted as inherent — SGP4 is only accurate for ~1-2 weeks from epoch anyway, and these are fictitious satellites. The sub-satellite point positions are still physically reasonable.

### Challenge 5: Time Scale Handling

Skyfield distinguishes between UTC, UT1, TT, and TDB time scales:

- **UTC → TT**: `TT = UTC + TAI-UTC + 32.184s`, where TAI-UTC comes from a leap second table (varies from 10s in 1972 to 37s in 2017+).
- **TT → UT1**: `UT1 = TT - ΔT`, where ΔT = TT - UT1 is looked up from a table of historical/predicted values.
- **TDB ≈ TT**: The difference is < 1.7ms (periodic), negligible for our purposes.

The initial implementation used a fixed offset (`UT1 ≈ UTC`, `TT = UTC + 69.184s`). This was accurate near the present (~0.1s error in 2026) but grew to 41s at 1902 and 76s at 2149, causing up to 0.77° location error.

**Solution**: Added the `timescale/` package with:
- A 28-entry leap second table for proper UTC → TT conversion
- A 401-entry ΔT table (yearly, 1800–2200) extracted from Skyfield for proper TT → UT1 conversion
- Binary search for leap seconds, linear interpolation for ΔT

This reduced the 248-year location error from 0.77° to 0.031°. The planet accuracy also improved slightly (proper leap second lookup vs fixed 37s offset).

---

## File Structure

The Go implementation uses the **goeph** library (independent reusable packages) rather than inline code:

```
generate_data_go/
├── main.go              # Entry point: generates time series, writes CSV
├── compute.go           # Core: ComputeRow() for all celestial columns per timestamp
├── go.mod / go.sum      # Module definition (depends on goeph via replace directive)
```

The goeph library provides the core packages:

```
goeph/
├── spk/                 # DAF file parser + SPK Type 2 Chebyshev evaluator + body IDs
├── coord/               # ICRF↔ecliptic, RA/Dec→ICRF, GMST/GAST, precession, nutation, geodetic→ICRF
├── timescale/           # ΔT table, leap second table, UTC→TT→UT1 conversion
├── satellite/           # SGP4/TLE satellite propagation (wraps go-satellite)
├── star/                # Galactic Center (Sgr A*) fixed RA/Dec → ICRF
└── lunarnodes/          # Mean lunar node longitude (Meeus formula)
```

Ground locations (6 sites: Null Island, Chicago, London, Cushing, NY, Mumbai) and satellite TLE definitions are in `compute.go`.

---

## Verification Results

### Near-Present (2026-01-15 00:00:00 UTC)

| Component | Error | Notes |
|-----------|-------|-------|
| Sun | lat: 7e-12°, lon: 1.9e-8° | Essentially exact |
| Moon | lat: 2.7e-8°, lon: 3.9e-7° | Excellent |
| Mercury–Pluto | < 4e-8° | Excellent |
| Galactic Center | < 3e-13° | Exact (fixed coordinates) |
| Lunar Nodes | < 5e-13° | Exact (same formula) |
| Ground Locations | < 0.006° | Nutation + ΔT model |
| Satellites | varies | SGP4 implementation differences |

### Full 248-Year Range (1902–2149, weekly, 12,932 rows)

| Component | Max Error | Mean Error | Notes |
|-----------|-----------|------------|-------|
| Planets | < 3.9e-7° | ~1.6e-7° | Proper leap second lookup |
| Galactic Center | < 2.8e-13° | ~1.8e-13° | Exact |
| Lunar Nodes | < 4.8e-13° | ~5.8e-14° | Exact |
| Ground Locations | **< 0.031°** | ~0.005° | IAU 2000A nutation + ΔT model |
| Satellites | up to 180° | varies | SGP4 meaningless ±124yr from TLE |

**Progression of location accuracy fixes**:

| Fix | Max location error | What changed |
|-----|-------------------|--------------|
| GMST only | ~0.37° | No precession |
| + Precession (P^T) | ~0.005° near present, 0.77° over 248yr | Missing ΔT |
| + Nutation (N^T) + ΔT model | **~0.031° over 248yr** | Core math matches Skyfield |

The residual 0.031° is from yearly ΔT interpolation vs Skyfield's daily IERS data. The core astronomical math (Chebyshev evaluation, coordinate transforms, precession, nutation) matches Skyfield exactly.

---

## Dependencies

- **Go standard library** — binary I/O, math, time, file handling
- **`github.com/joshuaferrara/go-satellite`** — SGP4/TLE satellite propagation (only external dependency)
- **DE440s.bsp** — JPL planetary ephemeris file (shared with the Python version, 31 MB)

---

## Performance

The Go version is **~12x faster** (wall-clock) and uses **~70x less total CPU time** than the Python version across all benchmarks, with no measurable overhead from the nutation computation or ΔT table lookups.

| Scale | Rows | Go | Python | Speedup |
|-------|------|-----|--------|---------|
| Small (44k rows) | 44,642 | 1.5s | 15.7s | 10.7x |
| Full (~10.5M rows) | ~10,500,000 | 3m 15s | 38m 37s | 11.9x |
| 200yr hourly (1.75M rows) | 1,752,266 | 28.1s | 6m 29s | 13.9x |
| 200yr daily (73k rows) | 73,013 | 2.5s | 21.0s | 8.3x |
| 248yr weekly (13k rows) | 12,934 | 1.5s | 7.1s | 4.7x |

See [BENCHMARK_GO_VS_PYTHON.md](BENCHMARK_GO_VS_PYTHON.md) for full benchmark details, raw timing output, and analysis.
