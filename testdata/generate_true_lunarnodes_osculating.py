#!/usr/bin/env python3
"""
Generate INDEPENDENT golden test data for true lunar node validation.

Uses Skyfield's osculating_elements_of() to compute the Moon's ascending node
longitude — a DIFFERENT code path from the r×v algorithm in goeph and
generate_true_lunarnodes.py.

Skyfield's approach (with ECLIPJ2000 reference frame):
  1. Rotate Moon geocentric r, v from ICRF to ecliptic frame
  2. h = r' × v' (angular momentum in ecliptic frame)
  3. Ω = atan2(h[0], -h[1]) (longitude of ascending node)

goeph's approach:
  1. h = r × v in ICRF
  2. n = ecliptic_pole × h (node line in ICRF)
  3. Convert n to ecliptic longitude

Both should yield the same result, but the code paths differ: Skyfield uses
its own Distance/Velocity/Angle types, its own cross product, and its own
reference frame rotation. Bugs in coordinate transforms, normalization, or
cross product ordering in either path would cause disagreement.

Usage:
    /opt/anaconda3/bin/python generate_true_lunarnodes_osculating.py

Requires: skyfield (with elementslib)
BSP file: ../data/de440s.bsp
"""

import json
import math
import os
import sys
from datetime import datetime, timedelta, timezone

try:
    from skyfield.api import load
    from skyfield.elementslib import osculating_elements_of
    from skyfield.data.spice import inertial_frames
except ImportError:
    print("skyfield not found. Install with: pip install skyfield")
    sys.exit(1)

BSP_PATH = os.path.join(os.path.dirname(__file__), '..', 'data', 'de440s.bsp')
OUTPUT_DIR = os.path.dirname(__file__)

# Same date range as generate_golden.py
REF_DATE = datetime(2025, 1, 1, 0, 0, 0, tzinfo=timezone.utc)
YEARS_RANGE = 200
DAY_INCREMENT = 30
EARLIEST = datetime(1850, 1, 1, tzinfo=timezone.utc)
LATEST = datetime(2149, 12, 31, tzinfo=timezone.utc)

ECLIPJ2000 = inertial_frames['ECLIPJ2000']


def generate_dates():
    start = REF_DATE - timedelta(days=YEARS_RANGE * 365)
    end = REF_DATE + timedelta(days=YEARS_RANGE * 365)
    if start < EARLIEST:
        start = EARLIEST
    if end > LATEST:
        end = LATEST
    dates = []
    d = start
    while d <= end:
        dates.append(d)
        d += timedelta(days=DAY_INCREMENT)
    return dates


def main():
    print(f"Loading ephemeris: {BSP_PATH}")
    ephem = load(BSP_PATH)
    ts = load.timescale()
    earth = ephem['earth']
    moon = ephem['moon']

    dates = generate_dates()
    print(f"Generating osculating-element true lunar node data for {len(dates)} dates...")

    tests = []
    for i, d in enumerate(dates):
        t = ts.from_datetime(d)
        tt_jd = t.tt

        # Geocentric Moon position via Skyfield's subtraction API
        geocentric = (moon - earth).at(t)

        # Compute osculating elements in the J2000 ecliptic frame
        elements = osculating_elements_of(geocentric, reference_frame=ECLIPJ2000)

        # Longitude of ascending node (ecliptic J2000)
        north_lon = math.degrees(float(elements.longitude_of_ascending_node.radians)) % 360.0
        south_lon = (north_lon + 180.0) % 360.0

        tests.append({
            "tdb_jd": tt_jd,
            "north_node_lon_deg": north_lon,
            "south_node_lon_deg": south_lon,
        })

        if (i + 1) % 1000 == 0:
            print(f"  {i + 1}/{len(dates)} dates processed...")

    print(f"All {len(dates)} dates processed.")

    path = os.path.join(OUTPUT_DIR, "golden_true_lunarnodes_osculating.json")
    with open(path, 'w') as f:
        json.dump({
            "ephemeris": "de440s.bsp",
            "algorithm": "Skyfield osculating_elements_of() with ECLIPJ2000 reference frame",
            "description": "Independent true lunar node longitude from Skyfield's Keplerian element fitting",
            "tests": tests,
        }, f, indent=None, separators=(',', ':'))

    size_mb = os.path.getsize(path) / (1024 * 1024)
    print(f"  golden_true_lunarnodes_osculating.json: {len(tests)} entries, {size_mb:.1f} MB")
    print("Done.")


if __name__ == '__main__':
    main()
