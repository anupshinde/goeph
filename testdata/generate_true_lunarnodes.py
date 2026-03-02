#!/usr/bin/env python3
"""
Generate golden test data for true lunar node validation.

Computes the Moon's instantaneous ascending node longitude using the same
r×v algorithm that goeph implements:
  1. Get Moon geocentric position and velocity from Skyfield/DE440s
  2. h = r × v (orbit normal)
  3. k = ecliptic pole in ICRF
  4. n = k × h (node line)
  5. Ascending node longitude = ecliptic longitude of n

Usage:
    python3 generate_true_lunarnodes.py

Requires: skyfield, numpy
BSP file: ../data/de440s.bsp
"""

import json
import os
import sys
import numpy as np
from datetime import datetime, timedelta, timezone

try:
    from skyfield.api import load
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

# J2000 mean obliquity (same constants as goeph coord package)
OBLIQUITY_SIN = 0.3977771559319137062
OBLIQUITY_COS = 0.9174820620691818140


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

    # Ecliptic pole in ICRF: (0, -sin(eps), cos(eps))
    k = np.array([0.0, -OBLIQUITY_SIN, OBLIQUITY_COS])

    dates = generate_dates()
    print(f"Generating true lunar node data for {len(dates)} dates...")

    tests = []
    for i, d in enumerate(dates):
        t = ts.from_datetime(d)
        tt_jd = t.tt

        # Geocentric position (geometric, no light-time)
        # Skyfield: earth.at(t).position.km gives barycentric position of Earth
        # moon.at(t).position.km gives barycentric position of Moon
        earth_pos = earth.at(t).position.km
        moon_pos = moon.at(t).position.km
        r = moon_pos - earth_pos  # geocentric Moon position (km)

        earth_vel = earth.at(t).velocity.km_per_s * 86400  # km/day
        moon_vel = moon.at(t).velocity.km_per_s * 86400    # km/day
        v = moon_vel - earth_vel  # geocentric Moon velocity (km/day)

        # Orbit normal: h = r × v
        h = np.cross(r, v)
        h_unit = h / np.linalg.norm(h)

        # Node line: n = k × h
        n = np.cross(k, h_unit)

        # Convert node direction to ecliptic longitude
        # ICRF → ecliptic rotation (around X-axis by obliquity)
        ex = n[0]
        ey = OBLIQUITY_COS * n[1] + OBLIQUITY_SIN * n[2]
        # ez not needed for longitude

        north_lon = np.degrees(np.arctan2(ey, ex)) % 360.0
        south_lon = (north_lon + 180.0) % 360.0

        tests.append({
            "tdb_jd": tt_jd,
            "north_node_lon_deg": float(north_lon),
            "south_node_lon_deg": float(south_lon),
        })

        if (i + 1) % 1000 == 0:
            print(f"  {i + 1}/{len(dates)} dates processed...")

    print(f"All {len(dates)} dates processed.")

    path = os.path.join(OUTPUT_DIR, "golden_true_lunarnodes.json")
    with open(path, 'w') as f:
        json.dump({
            "ephemeris": "de440s.bsp",
            "description": "True lunar node longitude from r×v algorithm using Skyfield/DE440s",
            "tests": tests,
        }, f, indent=None, separators=(',', ':'))

    size_mb = os.path.getsize(path) / (1024 * 1024)
    print(f"  golden_true_lunarnodes.json: {len(tests)} entries, {size_mb:.1f} MB")
    print("Done.")


if __name__ == '__main__':
    main()
