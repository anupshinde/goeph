# Examples

Runnable examples demonstrating goeph features. Each example is a standalone `main.go` that can be run from the repository root.

## Running

```bash
# Run any example from the repo root (needed for data/de440s.bsp path)
go run ./examples/positions/
go run ./examples/coordinates/
```

## Examples

| Example | Description |
|---|---|
| [positions](positions/) | Load a BSP ephemeris and get planet positions (core workflow) |
| [coordinates](coordinates/) | Convert positions to RA/Dec and ecliptic coordinates |
| [geodetic](geodetic/) | Observer on Earth's surface, zenith direction |
| [timescales](timescales/) | UTC/TT/UT1 time conversion chain, TDB-TT difference |
| [sidereal](sidereal/) | GMST, GAST, and Earth Rotation Angle |
| [separation](separation/) | Angular separation between the Sun and Moon |
| [phase](phase/) | Phase angle and fraction illuminated for planets |
| [elongation](elongation/) | Moon elongation from Sun and lunar phase names |
| [refraction](refraction/) | Atmospheric refraction correction at various altitudes |
| [galactic](galactic/) | Convert positions to galactic coordinates |
| [units](units/) | Working with Angle and Distance types |
| [satellite](satellite/) | SGP4 satellite propagation from TLE |
| [lunarnodes](lunarnodes/) | Mean lunar node ecliptic longitudes |
