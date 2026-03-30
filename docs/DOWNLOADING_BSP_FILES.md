# Downloading BSP Ephemeris Files

## Planetary Ephemeris (DE440)

The DE440 planetary ephemeris files are available from NASA/NAIF:

https://naif.jpl.nasa.gov/pub/naif/generic_kernels/spk/planets/

| File | Date Range | Size |
|------|-----------|------|
| `de440s.bsp` | 1849–2150 | ~31 MB |
| `de440.bsp` | 1550–2650 | ~114 MB |

`de440s.bsp` is sufficient for most use cases. The full `de440.bsp` is needed only for dates outside 1849–2150.

## Small Body Ephemeris (Asteroids, Comets, Centaurs)

Small body SPK files are not included in the DE440 planetary ephemeris. They must be downloaded separately from JPL HORIZONS.

### Using the HORIZONS Web Interface

1. Go to https://ssd.jpl.nasa.gov/horizons/app.html
2. Set **Ephemeris Type** → `Small-Body SPK File`
3. Set **Target Body** → search for the object (e.g., `2060 Chiron`)
4. Set **Coordinate Center** → `Sun (body center) [500@10]`
5. Set **Time Specification** → your desired date range
6. Click **Generate Ephemeris** to download the `.bsp` file

**Note:** HORIZONS limits small-body SPK files to a maximum of 200 years per request. For longer ranges, download multiple files and load them together with `spk.OpenMultiple()`.

### Example: Chiron (1600–2500)

Chiron's HORIZONS data is available from ~1600 to ~2500. To cover this full range, two files are needed:

| Request | Start | Stop | NAIF ID |
|---------|-------|------|---------|
| File 1 | 1600-01-01 | 2100-01-01 | 20002060 |
| File 2 | 2100-01-01 | 2500-12-31 | 20002060 |

Loading in Go:

```go
eph, err := spk.OpenMultiple(
    "data/de440s.bsp",
    "data/chiron_1600_2100.bsp",
    "data/chiron_2100_2500.bsp",
)

// Query Chiron using its NAIF ID
pos := eph.Observe(20002060, tdbJD)
```

### Using the HORIZONS API

You can also download SPK files programmatically:

```
https://ssd.jpl.nasa.gov/api/horizons.api?format=json&COMMAND='DES=20002060;'&EPHEM_TYPE=SPK&START_TIME='1600-01-01'&STOP_TIME='1800-01-01'&MAKE_EPHEM=YES&OBJ_DATA=NO
```

The response JSON contains a `spk` field with the base64-encoded BSP file.

Parameters:
- `COMMAND` — target body designation (e.g., `DES=20002060;` for Chiron)
- `START_TIME`, `STOP_TIME` — date range (max 200 years)
- `EPHEM_TYPE=SPK` — request SPK binary format

## NAIF Body IDs

Planetary bodies use standard NAIF IDs (defined in `spk/bodies.go`):

| Body | NAIF ID |
|------|---------|
| Sun | 10 |
| Mercury | 199 |
| Venus | 299 |
| Earth | 399 |
| Moon | 301 |
| Mars Barycenter | 4 |
| Jupiter Barycenter | 5 |
| Saturn Barycenter | 6 |

Small bodies use the convention `2000000 + asteroid_number`:

| Body | Number | NAIF ID |
|------|--------|---------|
| Ceres | 1 | 2000001 |
| Chiron | 2060 | 20002060 |

You can look up NAIF IDs at https://naif.jpl.nasa.gov/pub/naif/toolkit_docs/C/req/naif_ids.html

## SPK Segment Types

The library supports these SPK segment types:

| Type | Description | Used By |
|------|-------------|---------|
| 2 | Chebyshev (position only) | DE440 planetary ephemeris |
| 3 | Chebyshev (position + velocity) | Satellite ephemeris |
| 21 | Extended Modified Difference Arrays | HORIZONS small-body files |
