"""Compare Go and Python celestial CSV outputs column by column.

Usage:
    python compare_outputs.py [go_csv] [py_csv]

Defaults to /tmp/planetary_data_go.csv and /tmp/planetary_data_py.csv.
"""

import sys
import pandas as pd
import numpy as np

DEFAULT_GO = "/tmp/planetary_data_go.csv"
DEFAULT_PY = "/tmp/planetary_data_py.csv"

# Column categories for grouped reporting
PLANETS = {"sun", "moon", "mercury", "venus", "mars", "jupiter", "saturn", "uranus", "neptune", "pluto"}
GC = {"gc"}
SATELLITES = {"fastiss", "issanti", "polesat"}
LOCATIONS = {"loc_ni", "loc_chicago", "loc_london", "loc_cushing", "loc_ny", "loc_mumbai"}
NODES = {"north_node_lon_deg", "south_node_lon_deg"}


def categorize(col):
    prefix = col.rsplit("_sub_lat_deg", 1)[0].rsplit("_sub_lon_deg", 1)[0].rsplit("_lat_deg", 1)[0].rsplit("_lon_deg", 1)[0]
    if prefix in PLANETS:
        return "Planets"
    if prefix in GC:
        return "Galactic Center"
    if prefix in SATELLITES:
        return "Satellites"
    if prefix in LOCATIONS:
        return "Locations"
    if col in NODES:
        return "Lunar Nodes"
    return "Other"


def main():
    go_path = sys.argv[1] if len(sys.argv) > 1 else DEFAULT_GO
    py_path = sys.argv[2] if len(sys.argv) > 2 else DEFAULT_PY

    print(f"Go CSV:     {go_path}")
    print(f"Python CSV: {py_path}")
    print()

    go_df = pd.read_csv(go_path)
    py_df = pd.read_csv(py_path)

    print(f"Go rows: {len(go_df)}, Python rows: {len(py_df)}")
    print(f"Go columns: {len(go_df.columns)}, Python columns: {len(py_df.columns)}")
    print()

    # Align on Time column
    merged = pd.merge(go_df, py_df, on="Time", suffixes=("_go", "_py"))
    print(f"Matched rows (by Time): {len(merged)}")
    print(f"Unmatched Go rows: {len(go_df) - len(merged)}")
    print(f"Unmatched Py rows: {len(py_df) - len(merged)}")
    print()

    if len(merged) == 0:
        print("ERROR: No matching timestamps. Check Time column format.")
        sys.exit(1)

    # Find common numeric columns (exclude Time)
    go_cols = set(c.replace("_go", "") for c in merged.columns if c.endswith("_go"))
    py_cols = set(c.replace("_py", "") for c in merged.columns if c.endswith("_py"))
    common = sorted(go_cols & py_cols)

    go_only = go_cols - py_cols
    py_only = py_cols - go_cols
    if go_only:
        print(f"Columns only in Go: {sorted(go_only)}")
    if py_only:
        print(f"Columns only in Python: {sorted(py_only)}")

    # Per-column errors
    results = []
    for col in common:
        go_vals = pd.to_numeric(merged[col + "_go"], errors="coerce")
        py_vals = pd.to_numeric(merged[col + "_py"], errors="coerce")

        # Skip rows where either is NaN
        mask = go_vals.notna() & py_vals.notna()
        if mask.sum() == 0:
            results.append((col, categorize(col), 0, np.nan, np.nan, np.nan))
            continue

        diff = (go_vals[mask] - py_vals[mask]).abs()
        results.append((
            col,
            categorize(col),
            int(mask.sum()),
            diff.max(),
            diff.mean(),
            diff.median(),
        ))

    # Print per-column detail
    print(f"\n{'='*90}")
    print(f"{'Column':<30} {'Category':<16} {'Rows':>7} {'Max Error':>12} {'Mean Error':>12} {'Median Error':>12}")
    print(f"{'='*90}")

    for col, cat, n, mx, mn, md in results:
        if np.isnan(mx):
            print(f"{col:<30} {cat:<16} {n:>7} {'N/A':>12} {'N/A':>12} {'N/A':>12}")
        else:
            print(f"{col:<30} {cat:<16} {n:>7} {mx:>12.2e} {mn:>12.2e} {md:>12.2e}")

    # Group summary
    print(f"\n{'='*90}")
    print(f"{'Category':<20} {'Max Error':>12} {'Mean Error':>12} {'Median Error':>12}")
    print(f"{'='*90}")

    categories = ["Planets", "Galactic Center", "Lunar Nodes", "Locations", "Satellites"]
    for cat in categories:
        cat_results = [(mx, mn, md) for _, c, _, mx, mn, md in results if c == cat and not np.isnan(mx)]
        if not cat_results:
            print(f"{cat:<20} {'N/A':>12} {'N/A':>12} {'N/A':>12}")
            continue
        maxes, means, medians = zip(*cat_results)
        print(f"{cat:<20} {max(maxes):>12.2e} {np.mean(means):>12.2e} {np.median(medians):>12.2e}")

    print()


if __name__ == "__main__":
    main()
