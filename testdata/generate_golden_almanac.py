#!/usr/bin/env python3
"""
Generate golden test data for almanac features from Skyfield.

Produces JSON files with Skyfield-computed event times for:
- Seasons (equinoxes and solstices)
- Moon phases
- Sunrise/sunset for a mid-latitude observer

Usage:
    python3 generate_golden_almanac.py

Requires: skyfield, numpy
BSP file: ../data/de440s.bsp
"""

import json
import os
import sys

try:
    from skyfield.api import load, wgs84
    from skyfield import almanac
except ImportError:
    print("skyfield not found. Install with: pip install skyfield")
    sys.exit(1)

BSP_PATH = os.path.join(os.path.dirname(__file__), '..', 'data', 'de440s.bsp')
OUTPUT_DIR = os.path.dirname(__file__)


def write_json(filename, data):
    path = os.path.join(OUTPUT_DIR, filename)
    with open(path, 'w') as f:
        json.dump(data, f, indent=None, separators=(',', ':'))
    size_kb = os.path.getsize(path) / 1024
    count = len(data.get('tests', []))
    print(f"  {filename}: {count} entries, {size_kb:.1f} KB")


def main():
    print(f"Loading ephemeris: {BSP_PATH}")
    ephem = load(BSP_PATH)
    ts = load.timescale()

    # --- Seasons: 2000-2050 ---
    print("Generating seasons data (2000-2050)...")
    t0 = ts.utc(2000, 1, 1)
    t1 = ts.utc(2050, 12, 31)
    times, values = almanac.find_discrete(t0, t1, almanac.seasons(ephem))
    season_tests = []
    for t, v in zip(times, values):
        season_tests.append({
            "tt_jd": float(t.tt),
            "season": int(v),
        })
    print(f"  Found {len(season_tests)} season events")
    write_json("golden_seasons.json", {
        "description": "Equinox and solstice times from Skyfield almanac.seasons(). season: 0=spring, 1=summer, 2=autumn, 3=winter.",
        "tests": season_tests,
    })

    # --- Moon phases: 2000-2050 ---
    print("Generating moon phase data (2000-2050)...")
    times, values = almanac.find_discrete(t0, t1, almanac.moon_phases(ephem))
    phase_tests = []
    for t, v in zip(times, values):
        phase_tests.append({
            "tt_jd": float(t.tt),
            "phase": int(v),
        })
    print(f"  Found {len(phase_tests)} moon phase events")
    write_json("golden_moon_phases.json", {
        "description": "Moon phase transition times from Skyfield almanac.moon_phases(). phase: 0=new, 1=first_quarter, 2=full, 3=last_quarter.",
        "tests": phase_tests,
    })

    # --- Sunrise/sunset: New York (40.7128°N, 74.0060°W), year 2024 ---
    print("Generating sunrise/sunset data (NYC, 2024)...")
    loc = wgs84.latlon(40.7128, -74.0060)
    observer = ephem['earth'] + loc
    t0_ss = ts.utc(2024, 1, 1)
    t1_ss = ts.utc(2024, 12, 31)
    f = almanac.sunrise_sunset(ephem, loc)
    times, values = almanac.find_discrete(t0_ss, t1_ss, f)
    sunrise_tests = []
    for t, v in zip(times, values):
        sunrise_tests.append({
            "tt_jd": float(t.tt),
            "is_sunrise": int(v),  # 1=sunrise, 0=sunset
        })
    print(f"  Found {len(sunrise_tests)} sunrise/sunset events")
    write_json("golden_sunrise_sunset.json", {
        "description": "Sunrise/sunset times from Skyfield almanac.sunrise_sunset() for NYC (40.7128N, 74.0060W), year 2024. is_sunrise: 1=sunrise, 0=sunset.",
        "lat": 40.7128,
        "lon": -74.0060,
        "tests": sunrise_tests,
    })

    # --- Twilight: New York, January 2024 (one month) ---
    print("Generating twilight data (NYC, Jan 2024)...")
    t0_tw = ts.utc(2024, 1, 1)
    t1_tw = ts.utc(2024, 2, 1)
    f_tw = almanac.dark_twilight_day(ephem, loc)
    times, values = almanac.find_discrete(t0_tw, t1_tw, f_tw)
    twilight_tests = []
    for t, v in zip(times, values):
        twilight_tests.append({
            "tt_jd": float(t.tt),
            "level": int(v),
        })
    print(f"  Found {len(twilight_tests)} twilight events")
    write_json("golden_twilight.json", {
        "description": "Twilight transition times from Skyfield almanac.dark_twilight_day() for NYC, Jan 2024. level: 0=night, 1=astronomical, 2=nautical, 3=civil, 4=day.",
        "lat": 40.7128,
        "lon": -74.0060,
        "tests": twilight_tests,
    })

    # --- Oppositions/Conjunctions: Mars, 2000-2050 ---
    print("Generating Mars oppositions/conjunctions (2000-2050)...")
    t0_oc = ts.utc(2000, 1, 1)
    t1_oc = ts.utc(2050, 12, 31)
    f_oc = almanac.oppositions_conjunctions(ephem, ephem['mars barycenter'])
    times, values = almanac.find_discrete(t0_oc, t1_oc, f_oc)
    opp_tests = []
    for t, v in zip(times, values):
        opp_tests.append({
            "tt_jd": float(t.tt),
            "value": int(v),
        })
    print(f"  Found {len(opp_tests)} opposition/conjunction events")
    write_json("golden_oppositions.json", {
        "description": "Mars opposition/conjunction times from Skyfield almanac.oppositions_conjunctions(). value: 0=conjunction, 1=opposition (Skyfield convention).",
        "body": "mars_barycenter",
        "body_id": 4,
        "tests": opp_tests,
    })

    print("Done.")


if __name__ == '__main__':
    main()
