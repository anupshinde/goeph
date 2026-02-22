
# ------------------------------------------------------
# Mean lunar node longitude
# ------------------------------------------------------
def mean_lunar_node_longitude(time):
    """Return the *mean* North Node ecliptic longitude (degrees), geocentric."""
    # (Meeus "Astronomical Algorithms" formula)
    jd_tt = time.tt  # Skyfield's 'time.tt' => Julian Date in Terrestrial Time
    T = (jd_tt - 2451545.0) / 36525.0

    omega = (
        125.04452
        - 1934.136261 * T
        + 0.0020708 * (T**2)
        + (T**3) / 450000.0
    )
    return omega % 360.0

# ------------------------------------------------------
# Mean lunar nodes
# ------------------------------------------------------
def mean_lunar_nodes(time):
    """Return (north_node_lon, south_node_lon) in degrees, geocentric ecliptic."""
    n = mean_lunar_node_longitude(time)
    s = (n + 180.0) % 360.0
    return (n, s)
