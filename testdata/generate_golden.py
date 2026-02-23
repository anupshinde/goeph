#!/usr/bin/env python3
"""
Generate golden test data from Skyfield for goeph validation.

Produces JSON files with Skyfield-computed values at 30-day increments
over the full DE440s range (1850-2149).

Usage:
    python3 generate_golden.py

Requires: skyfield, numpy
BSP file: ../data/de440s.bsp (de440s covers 1849-2150)
"""

import json
import os
import sys
import math
import numpy as np
from datetime import datetime, timedelta, timezone

# Use the anaconda python if skyfield isn't found in default
try:
    from skyfield.api import load, Star, Angle, N, W, wgs84
    from skyfield.earthlib import earth_rotation_angle, refraction
    from skyfield.timelib import tdb_minus_tt
    from skyfield.functions import angle_between, length_of
except ImportError:
    print("skyfield not found. Install with: pip install skyfield")
    sys.exit(1)

# --- Configuration ---

BSP_PATH = os.path.join(os.path.dirname(__file__), '..', 'data', 'de440s.bsp')
OUTPUT_DIR = os.path.dirname(__file__)

# Reference date: 2025-01-01 00:00 UTC
REF_DATE = datetime(2025, 1, 1, 0, 0, 0, tzinfo=timezone.utc)
YEARS_RANGE = 200  # -200 to +200 years (clamped to BSP range 1850-2149)
DAY_INCREMENT = 30

# de440s covers 1849-2150, so clamp to safe range
EARLIEST = datetime(1850, 1, 1, tzinfo=timezone.utc)
LATEST = datetime(2149, 12, 31, tzinfo=timezone.utc)

# Bodies to test (name, skyfield key)
BODIES = [
    ("sun", "sun", 10),
    ("moon", "moon", 301),
    ("mercury", "mercury", 199),
    ("venus", "venus", 299),
    ("mars", "mars barycenter", 4),
    ("jupiter", "jupiter barycenter", 5),
    ("saturn", "saturn barycenter", 6),
    ("uranus", "uranus barycenter", 7),
    ("neptune", "neptune barycenter", 8),
    ("pluto", "pluto barycenter", 9),
]

# Locations (name, lat, lon)
LOCATIONS = [
    ("loc_ni", 0.0, 0.0),
    ("loc_cme", 41.8674558, -87.6483924),
    ("loc_lse", 51.5150534, -0.1016089),
    ("loc_cushing", 35.9859634, -96.7954485),
    ("loc_nymex", 40.714469, -74.0194683),
    ("loc_mumbai", 19.0602766, 72.8577106),
]

# Galactic Center (Sgr A*)
GC_RA_HOURS = 17.0 + 45.0/60.0 + 40.0409/3600.0
GC_DEC_DEG = -(29.0 + 0.0/60.0 + 28.118/3600.0)


def generate_dates():
    """Generate dates at 30-day increments, clamped to BSP range."""
    start = REF_DATE - timedelta(days=YEARS_RANGE * 365)
    end = REF_DATE + timedelta(days=YEARS_RANGE * 365)

    # Clamp to BSP range
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


def mean_lunar_node_longitude(tt_jd):
    """Mean North Node ecliptic longitude (degrees) using Meeus formula."""
    T = (tt_jd - 2451545.0) / 36525.0
    omega = 125.04452 - 1934.136261 * T + 0.0020708 * T**2 + T**3 / 450000.0
    return omega % 360.0


def main():
    print(f"Loading ephemeris: {BSP_PATH}")
    ephem = load(BSP_PATH)
    ts = load.timescale()
    earth = ephem['earth']

    # Galactic Center star object
    gc_star = Star(
        ra=Angle(hours=GC_RA_HOURS),
        dec=Angle(degrees=GC_DEC_DEG),
    )

    # Build location objects
    skyfield_locations = []
    for name, lat, lon in LOCATIONS:
        loc = earth + wgs84.latlon(lat, lon)
        skyfield_locations.append((name, lat, lon, loc))

    dates = generate_dates()
    print(f"Generating golden data for {len(dates)} dates...")

    # --- Golden SPK data (body positions) ---
    spk_tests = []
    # --- Golden coord data (ecliptic lat/lon) ---
    ecliptic_tests = []
    # --- Golden timescale data ---
    timescale_tests = []
    # --- Golden location data ---
    location_tests = []
    # --- Golden lunar node data ---
    lunarnode_tests = []
    # --- Golden sidereal data (GMST) ---
    sidereal_tests = []
    # --- Tier 1: ERA, TDB-TT, phase angle, separation, elongation ---
    era_tests = []
    tdbtt_tests = []
    phase_tests = []
    separation_tests = []
    elongation_tests = []

    for i, d in enumerate(dates):
        t = ts.from_datetime(d)
        tt_jd = t.tt
        ut1_jd = t.ut1

        # Timescale: record UTC JD, TT JD, UT1 JD
        # UTC JD: compute from the datetime
        unix_sec = d.timestamp()
        utc_jd = unix_sec / 86400.0 + 2440587.5
        timescale_tests.append({
            "utc_jd": utc_jd,
            "tt_jd": tt_jd,
            "ut1_jd": ut1_jd,
        })

        # GMST
        sidereal_tests.append({
            "ut1_jd": ut1_jd,
            "gmst_deg": t.gmst * 15.0,  # Skyfield returns hours, convert to degrees
        })

        # Body positions
        for body_name, sf_key, naif_id in BODIES:
            body = ephem[sf_key]
            astrometric = earth.at(t).observe(body)
            pos_km = astrometric.position.km
            eclip = astrometric.ecliptic_latlon()
            ecl_lat = eclip[0].degrees
            ecl_lon = eclip[1]._degrees % 360.0

            spk_tests.append({
                "tdb_jd": tt_jd,  # TDB ≈ TT
                "body_id": naif_id,
                "pos_km": [float(pos_km[0]), float(pos_km[1]), float(pos_km[2])],
            })

            ecliptic_tests.append({
                "tdb_jd": tt_jd,
                "body_name": body_name,
                "body_id": naif_id,
                "ecl_lat_deg": float(ecl_lat),
                "ecl_lon_deg": float(ecl_lon),
            })

        # Galactic Center
        gc_obs = earth.at(t).observe(gc_star)
        gc_eclip = gc_obs.ecliptic_latlon()
        ecliptic_tests.append({
            "tdb_jd": tt_jd,
            "body_name": "gc",
            "body_id": 0,
            "ecl_lat_deg": float(gc_eclip[0].degrees),
            "ecl_lon_deg": float(gc_eclip[1]._degrees % 360.0),
        })

        # Locations
        for loc_name, lat, lon, loc_obj in skyfield_locations:
            obs = earth.at(t).observe(loc_obj)
            eclip = obs.ecliptic_latlon()
            location_tests.append({
                "tdb_jd": tt_jd,
                "ut1_jd": ut1_jd,
                "loc_name": loc_name,
                "lat": lat,
                "lon": lon,
                "ecl_lat_deg": float(eclip[0].degrees),
                "ecl_lon_deg": float(eclip[1]._degrees % 360.0),
            })

        # Lunar nodes
        nn_lon = mean_lunar_node_longitude(tt_jd)
        sn_lon = (nn_lon + 180.0) % 360.0
        lunarnode_tests.append({
            "tdb_jd": tt_jd,
            "north_node_lon_deg": nn_lon,
            "south_node_lon_deg": sn_lon,
        })

        # --- Tier 1: ERA ---
        era_val = earth_rotation_angle(ut1_jd) * 360.0  # Skyfield returns turns, convert to degrees
        era_tests.append({
            "ut1_jd": ut1_jd,
            "era_deg": float(era_val),
        })

        # --- Tier 1: TDB-TT ---
        tdbtt_val = tdb_minus_tt(tt_jd)
        tdbtt_tests.append({
            "tt_jd": tt_jd,
            "tdb_minus_tt_sec": float(tdbtt_val),
        })

        # --- Tier 1: Phase angle, fraction illuminated, separation, elongation ---
        # Use Sun, Moon, and a few planets for phase/separation/elongation tests
        sun_body = ephem['sun']
        moon_body = ephem['moon']

        # Observe all bodies from Earth for this date
        sun_astrometric = earth.at(t).observe(sun_body)
        moon_astrometric = earth.at(t).observe(moon_body)

        # Sun and Moon positions in km (observer-to-target vectors)
        sun_pos_km = sun_astrometric.position.km
        moon_pos_km = moon_astrometric.position.km

        # Phase angle for Moon (Sun-Moon-Earth angle)
        moon_phase_angle = moon_astrometric.phase_angle(sun_body).degrees
        moon_frac = moon_astrometric.fraction_illuminated(sun_body)

        phase_tests.append({
            "tdb_jd": tt_jd,
            "body_name": "moon",
            "phase_angle_deg": float(moon_phase_angle),
            "fraction_illuminated": float(moon_frac),
        })

        # Phase angle for planets (Mercury, Venus, Mars, Jupiter, Saturn)
        for body_name, sf_key, naif_id in BODIES:
            if naif_id in (199, 299, 4, 5, 6):  # Mercury, Venus, Mars, Jupiter, Saturn
                body = ephem[sf_key]
                body_astrometric = earth.at(t).observe(body)
                pa = body_astrometric.phase_angle(sun_body).degrees
                fi = body_astrometric.fraction_illuminated(sun_body)
                phase_tests.append({
                    "tdb_jd": tt_jd,
                    "body_name": body_name,
                    "phase_angle_deg": float(pa),
                    "fraction_illuminated": float(fi),
                })

        # Separation angle: Sun-Moon
        sun_pos_au = np.array(sun_astrometric.position.au)
        moon_pos_au = np.array(moon_astrometric.position.au)
        sep_sun_moon = float(np.degrees(angle_between(sun_pos_au, moon_pos_au)))
        separation_tests.append({
            "tdb_jd": tt_jd,
            "body1": "sun",
            "body2": "moon",
            "separation_deg": sep_sun_moon,
        })

        # Moon elongation (ecliptic longitude difference Moon - Sun)
        sun_eclip = sun_astrometric.ecliptic_latlon()
        moon_eclip = moon_astrometric.ecliptic_latlon()
        sun_ecl_lon = float(sun_eclip[1]._degrees % 360.0)
        moon_ecl_lon = float(moon_eclip[1]._degrees % 360.0)
        elong = (moon_ecl_lon - sun_ecl_lon) % 360.0
        elongation_tests.append({
            "tdb_jd": tt_jd,
            "moon_ecl_lon_deg": moon_ecl_lon,
            "sun_ecl_lon_deg": sun_ecl_lon,
            "elongation_deg": float(elong),
        })

        if (i + 1) % 10000 == 0:
            print(f"  {i + 1}/{len(dates)} dates processed...")

    print(f"All {len(dates)} dates processed.")

    # --- Tier 1: Refraction (altitude-based, not date-based) ---
    refraction_tests = []
    temp_C = 10.0
    pressure_mbar = 1013.25
    for alt_deg in [x * 0.5 for x in range(-2, 180)]:  # -1.0 to 89.5 in 0.5° steps
        r = refraction(alt_deg, temp_C, pressure_mbar)
        refraction_tests.append({
            "alt_deg": alt_deg,
            "temp_c": temp_C,
            "pressure_mbar": pressure_mbar,
            "refraction_deg": float(r),
        })
    print(f"  Generated {len(refraction_tests)} refraction test cases")

    # Write golden files
    def write_json(filename, data):
        path = os.path.join(OUTPUT_DIR, filename)
        with open(path, 'w') as f:
            json.dump(data, f, indent=None, separators=(',', ':'))
        size_mb = os.path.getsize(path) / (1024 * 1024)
        print(f"  {filename}: {len(data.get('tests', []))} entries, {size_mb:.1f} MB")

    write_json("golden_spk.json", {
        "ephemeris": "de440s.bsp",
        "description": "Astrometric (light-time corrected) geocentric positions from Skyfield",
        "tests": spk_tests,
    })

    write_json("golden_ecliptic.json", {
        "description": "Ecliptic lat/lon from Skyfield observe().ecliptic_latlon()",
        "tests": ecliptic_tests,
    })

    write_json("golden_timescale.json", {
        "description": "UTC JD -> TT JD -> UT1 JD from Skyfield",
        "tests": timescale_tests,
    })

    write_json("golden_locations.json", {
        "description": "Ground location ecliptic lat/lon from Skyfield",
        "tests": location_tests,
    })

    write_json("golden_lunarnodes.json", {
        "description": "Mean lunar node longitudes (Meeus formula)",
        "tests": lunarnode_tests,
    })

    write_json("golden_sidereal.json", {
        "description": "GMST from Skyfield (t.gmst * 15 -> degrees)",
        "tests": sidereal_tests,
    })

    write_json("golden_era.json", {
        "description": "Earth Rotation Angle from Skyfield (degrees)",
        "tests": era_tests,
    })

    write_json("golden_tdbtt.json", {
        "description": "TDB-TT from Skyfield tdb_minus_tt() (seconds)",
        "tests": tdbtt_tests,
    })

    write_json("golden_phase.json", {
        "description": "Phase angle and fraction illuminated from Skyfield",
        "tests": phase_tests,
    })

    write_json("golden_separation.json", {
        "description": "Separation angles from Skyfield angle_between()",
        "tests": separation_tests,
    })

    write_json("golden_elongation.json", {
        "description": "Moon elongation (ecliptic longitude difference) from Skyfield",
        "tests": elongation_tests,
    })

    write_json("golden_refraction.json", {
        "description": "Atmospheric refraction from Skyfield Bennett formula",
        "tests": refraction_tests,
    })

    print("Done.")


if __name__ == '__main__':
    main()
