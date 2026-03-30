#!/usr/bin/env python3
"""
Generate golden test data for ecliptic-of-date frame transforms.

Produces a JSON file with Skyfield-computed ecliptic-of-date lat/lon
from known ICRF position vectors at various TT Julian dates.

Usage:
    python3 generate_ecliptic_of_date.py

Requires: skyfield, numpy
"""

import json
import math
import os
import sys
import numpy as np

try:
    from skyfield.api import load
except ImportError:
    print("skyfield not found. Install with: pip install skyfield")
    sys.exit(1)

OUTPUT_DIR = os.path.dirname(__file__)

# Fixed ICRF test vectors (km) — arbitrary directions and magnitudes
TEST_VECTORS = [
    [1e8, 0, 0],
    [0, 1e8, 0],
    [0, 0, 1e8],
    [1e8, -5e7, 2e7],
    [-3e7, 8e7, -1e7],
    [1.496e8, 0, 0],       # ~1 AU along X
    [5e7, 5e7, 5e7],       # equal components
    [-1e8, -1e8, 0],       # in XY plane, third quadrant
]

# TT Julian dates spanning 1900–2100
TT_JDS = [
    2415020.5,   # 1900-01-01
    2433282.5,   # 1950-01-01
    2451545.0,   # 2000-01-01 (J2000.0)
    2455197.5,   # 2010-01-01
    2458849.5,   # 2020-01-01
    2460676.5,   # 2025-01-01
    2461041.5,   # 2025-12-31
    2469807.5,   # 2050-01-01
    2488070.0,   # 2100-01-01
]


def main():
    ts = load.timescale()
    tests = []

    for jd_tt in TT_JDS:
        t = ts.tt_jd(jd_tt)

        for vec in TEST_VECTORS:
            pos_au = np.array(vec) / 149597870.7  # km to AU

            # Build a mock position and use Skyfield's ecliptic_latlon(epoch=t)
            # to get true ecliptic of date coordinates.
            from skyfield.positionlib import ICRF
            position = ICRF(position_au=pos_au, t=t, center=399)

            # epoch=t gives ecliptic of date (true ecliptic + equinox of date)
            eclip = position.ecliptic_latlon(epoch=t)
            ecl_lat = float(eclip[0].degrees)
            ecl_lon = float(eclip[1]._degrees % 360.0)

            tests.append({
                "tt_jd": jd_tt,
                "icrf_km": vec,
                "ecl_lat_deg": ecl_lat,
                "ecl_lon_deg": ecl_lon,
            })

    output = {
        "description": "True ecliptic-of-date lat/lon from Skyfield ICRF.ecliptic_latlon(epoch=t)",
        "tests": tests,
    }

    path = os.path.join(OUTPUT_DIR, "golden_ecliptic_of_date.json")
    with open(path, 'w') as f:
        json.dump(output, f, indent=2)
    print(f"Wrote {len(tests)} test cases to {path}")


if __name__ == "__main__":
    main()
