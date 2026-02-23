# Golden Test Data

All computational correctness in goeph is validated against [Skyfield](https://github.com/skyfielders/python-skyfield) via golden tests. The JSON files in this directory are the source of truth.

## Generating golden data

```bash
cd testdata
/opt/anaconda3/bin/python generate_golden.py
```

Requires Python 3 with `skyfield` installed. The script loads `../data/de440s.bsp` and writes all `golden_*.json` files. These are checked into git so CI runs without Python or Skyfield.

## Sampling strategy

**Primary grid:** 3,653 dates at 30-day increments across the full DE440s range (1850-2149).

This is dense enough to catch systematic drift from approximations (nutation, delta-T, precession) and error growth over centuries away from J2000. Fine-grained sampling is unnecessary because the underlying physics is smooth (Chebyshev polynomials, continuous rotations).

**Bodies tested:** 10 bodies for position/velocity/apparent (Sun, Moon, Mercury, Venus, Mars, Jupiter, Saturn, Uranus, Neptune, Pluto barycenters). 3 bodies for altaz (Sun, Moon, Mars) across 6 observer locations.

## Tolerance summary

All tolerances are measured max error vs Skyfield across the full date range.

### Positions and velocities

| Golden file | Test | Tolerance | Measured max | Units | Error source |
|---|---|---|---|---|---|
| `golden_spk.json` | SPK positions | 0.01 | ~0.002 | km | Mercury barycenter chain (with TDB-TT correction) |
| `golden_velocity.json` | Velocity | 0.01 | ~0.0002 | km/day | Chebyshev derivative (with TDB-TT correction) |
| `golden_apparent.json` | Apparent positions | max(50, 1.5e-5 * dist) | ~27,000 km abs | km | 30-term nutation (~3 arcsec angular error, scales with distance) |

### Time scales

| Golden file | Test | Tolerance | Measured max | Units | Error source |
|---|---|---|---|---|---|
| `golden_timescale.json` | UTC to TT | 1e-9 | ~1e-10 | days | Identical leap second table |
| `golden_timescale.json` | TT to UT1 | 1e-6 | 7.5e-7 | days (~65 ms) | Cubic spline vs Skyfield's spline (different source knots) |
| `golden_tdbtt.json` | TDB-TT | 1e-9 | ~1e-10 | seconds | Identical Fairhead & Bretagnon formula |

### Sidereal time and Earth rotation

| Golden file | Test | Tolerance | Measured max | Units | Error source |
|---|---|---|---|---|---|
| `golden_era.json` | Earth Rotation Angle | 1e-8 | ~1e-9 | degrees | Identical IAU 2000 formula |
| `golden_sidereal.json` | GMST | 1e-3 | ~5e-4 | degrees | IAU 1982 Meeus vs Skyfield's IERS 2000 ERA-based |

### Coordinate transforms

| Golden file | Test | Tolerance | Measured max | Units | Error source |
|---|---|---|---|---|---|
| `golden_locations.json` | Geodetic to ecliptic | 0.035 | ~0.033 | degrees | 30-term vs 687-term nutation; gap grows with centuries from J2000 |
| `golden_altaz.json` | Altitude | 1.0 | 0.011 | degrees | 30-term nutation propagates into Earth rotation |
| `golden_altaz.json` | Azimuth | 1.0 | 0.78 (0.057 for \|alt\|<80°) | degrees | Near-zenith singularity amplifies small errors; away from zenith error is <0.06° |

### Angular quantities

| Golden file | Test | Tolerance | Measured max | Units | Error source |
|---|---|---|---|---|---|
| `golden_separation.json` | Separation angle | 1e-8 | ~1e-9 | degrees | Same position vectors, same formula |
| `golden_phase.json` | Phase angle | 1e-8 | ~6e-14 | degrees | Exact input vectors from Skyfield; formula-level agreement |
| `golden_elongation.json` | Elongation | 1e-10 | ~1e-11 | degrees | Pure modular arithmetic |
| `golden_refraction.json` | Refraction | 1e-10 | ~1e-11 | degrees | Identical Bennett 1982 formula |

### Other

| Golden file | Test | Tolerance | Measured max | Units | Error source |
|---|---|---|---|---|---|
| `golden_lunarnodes.json` | Lunar node longitude | 1e-8 | ~1e-9 | degrees | Identical Meeus formula |

## Dominant error source

The single largest source of disagreement with Skyfield is **30-term nutation** (goeph) vs **687-term IAU 2000A nutation** (Skyfield). This introduces ~1 arcsecond of angular error near J2000 that:

- Grows to several arcseconds at the edges of the date range (1850, 2149)
- Cascades through the precession/nutation/Earth-rotation chain, affecting geodetic transforms, altaz, and apparent positions
- Produces km-level position errors proportional to target distance (a 3 arcsec error at Saturn's distance of ~1.4 billion km is ~20,000 km)

This is a deliberate tradeoff: 30 terms keeps the nutation table small and computation fast, at the cost of ~1 arcsecond precision. Implementing full 687-term nutation (Tier 3 in the roadmap) would reduce most tolerances by 3 orders of magnitude.

## Golden file format

All files are JSON with a top-level `tests` array. Each test entry includes the input parameters and expected output values from Skyfield. See `generate_golden.py` for the exact schema of each file.
