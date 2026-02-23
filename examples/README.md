# Examples

Runnable examples demonstrating goeph features. Each example is a standalone `main.go` that can be run from the repository root.

## Running

```bash
# Run any example from the repo root (needed for data/de440s.bsp path)
go run ./examples/positions/
go run ./examples/apparent/
```

## Examples

### Core positions and observations

| Example | Description |
|---|---|
| [positions](positions/) | Load a BSP ephemeris and get planet positions (core workflow) |
| [apparent](apparent/) | Apparent positions with aberration and gravitational deflection |
| [velocity](velocity/) | Body velocities via Chebyshev polynomial derivatives |
| [observefrom](observefrom/) | Observe from arbitrary bodies (not just Earth) |

### Coordinate transforms

| Example | Description |
|---|---|
| [coordinates](coordinates/) | Convert positions to RA/Dec and ecliptic coordinates |
| [geodetic](geodetic/) | Observer on Earth's surface, zenith direction |
| [altaz](altaz/) | Altitude and azimuth for a ground observer |
| [galactic](galactic/) | Convert positions to galactic coordinates |

### Time and angles

| Example | Description |
|---|---|
| [timescales](timescales/) | UTC/TT/UT1 time conversion chain, TDB-TT difference |
| [sidereal](sidereal/) | GMST, GAST, and Earth Rotation Angle |
| [separation](separation/) | Angular separation between the Sun and Moon |
| [phase](phase/) | Phase angle and fraction illuminated for planets |
| [elongation](elongation/) | Moon elongation from Sun and lunar phase names |
| [refraction](refraction/) | Atmospheric refraction correction at various altitudes |

### Orbital mechanics and photometry

| Example | Description |
|---|---|
| [elements](elements/) | Osculating orbital elements from state vectors |
| [magnitude](magnitude/) | Planetary visual magnitudes (Mallama & Hilton 2018) |

### Configuration

| Example | Description |
|---|---|
| [nutation](nutation/) | Compare NutationStandard vs NutationFull precision modes |

### Other

| Example | Description |
|---|---|
| [units](units/) | Working with Angle and Distance types |
| [satellite](satellite/) | SGP4 satellite propagation from TLE |
| [lunarnodes](lunarnodes/) | Mean lunar node ecliptic longitudes |
