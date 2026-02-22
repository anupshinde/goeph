from skyfield.api import load, Angle, Star, EarthSatellite
tscale = load.timescale()

# ------------------------------------------------------
# Define ISS and Derived Satellites
# ------------------------------------------------------

tle_line1 = '1 25544U 98067A   24031.54769676  .00006652  00000+0  12801-3 0  9996'
tle_line2 = '2 25544  51.6415  41.8752 0005686  93.3517  58.2198 15.50294839424075'
iss = EarthSatellite(tle_line1, tle_line2, 'ISS', tscale)

# ISSFast: 1.618x faster than ISS
iss_fast = EarthSatellite(tle_line1, tle_line2.replace("15.50294839", str(15.50294839 * 1.618)), 'ISSFast', tscale)

# AntiISS: Slower speed than ISS and in the opposite direction
iss_anti = EarthSatellite(tle_line1, tle_line2.replace("15.50294839", str(15.50294839 * 0.618)), 'AntiISS', tscale)

# PoleSat: Fixed at 0° longitude, 90-minute orbit (polar orbit)
# The first one had problems, keeping this for reference in the future
# pole_tle1 = '1 99998U 24031A   24031.54769676  .00006652  00000+0  12801-3 0  9996'
# pole_tle2 = '2 99998  90.0000  0.0000 0005686  0.0000  0.0000 16.00000000000000'
pole_tle1 = '1 99998U 24031A   24031.54769676  .00006652  00000+0  12801-3 0  9996'
pole_tle2 = '2 99998  90.0000  180.0000 0005686  0.0000  270.0000 15.99999999'
pole_sat = EarthSatellite(pole_tle1, pole_tle2, 'PoleSat', tscale)


artisats = {
    # "ISS": iss, # This has some problems returning NaN values.
    "FastISS": iss_fast,
    "ISSAnti": iss_anti,
    "PoleSAT": pole_sat,
}