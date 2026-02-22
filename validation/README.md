# Validation

End-to-end comparison of the Go (goeph) and Python (Skyfield) celestial data generators.

Both programs compute geocentric ecliptic positions for planets, satellites, ground locations, and lunar nodes over a 200-year range, outputting CSV files with identical column layouts.

## Prerequisites

- **Go**: Go 1.22+ (for `generate_data_go`)
- **Python**: Python with pandas, numpy, skyfield, sgp4, tqdm, pytz (for `generate_data_py`)
- **Ephemeris**: `data/de440s.bsp` must exist at the repo root

## 1. Generate Go output

```bash
cd validation/generate_data_go
go build -o generate_data_go && ./generate_data_go
```

Writes to `/tmp/planetary_data_go.csv`.

## 2. Generate Python output

```bash
cd validation/generate_data_py
python main.py
```

Writes to `/tmp/planetary_data_py.csv`.

## 3. Compare outputs

```bash
cd validation
python compare_outputs.py
```

Or with custom paths:

```bash
python compare_outputs.py /path/to/go.csv /path/to/py.csv
```

Reports per-column and per-category (Planets, Galactic Center, Lunar Nodes, Locations, Satellites) max/mean/median absolute error.

## Expected results

| Category | Max Error | Notes |
|----------|-----------|-------|
| Planets | < 4e-7° | Moon longitude worst case |
| Galactic Center | < 3e-13° | Effectively exact |
| Lunar Nodes | < 5e-13° | Effectively exact |
| Locations | < 0.035° | 30-term nutation + yearly delta-T interpolation vs Skyfield's 687-term + daily IERS data |
| Satellites | up to 180° | SGP4 diverges far from TLE epoch — expected |

See [docs/PYTHON_SKYFIELD_TO_GO.md](../docs/PYTHON_SKYFIELD_TO_GO.md) for implementation details and [docs/BENCHMARK_GO_VS_PYTHON.md](../docs/BENCHMARK_GO_VS_PYTHON.md) for full benchmark results.
