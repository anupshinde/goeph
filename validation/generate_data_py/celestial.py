import pandas as pd
import numpy as np
from skyfield.api import load
import locations
import satellites
import lunar_nodes
from stars import GalacticCenter

timescl = load.timescale()

def get_bodies():
    ephem = locations.ephem

    bodies = {
        'Sun':      ephem['sun'],       # geocentric viewpoint
        'Moon':     ephem['moon'],
        'Mercury':  ephem['mercury'],
        'Venus':    ephem['venus'],
        'Mars':     ephem['mars barycenter'],
        'Jupiter':  ephem['jupiter barycenter'],
        'Saturn':   ephem['saturn barycenter'],
        'Uranus':   ephem['uranus barycenter'],
        'Neptune':  ephem['neptune barycenter'],
        'Pluto':    ephem['pluto barycenter'],
        'GC':    GalacticCenter,
    }
    return bodies

locations_to_add = {
    'loc_ni': locations.NULL_ISLAND_loc,
    'loc_chicago': locations.CHICAGO_loc,
    'loc_london': locations.LONDON_loc,
    'loc_cushing': locations.CUSHING_loc,
    'loc_ny': locations.NY_loc,
    'loc_mumbai': locations.MUMBAI_loc,
}

# Add celestial details to a dataframe
def get_celestial_df(timeSeries, addSatellites):
    addLocations=True

    df = pd.DataFrame()
    df["Time"] = timeSeries
    tobjs = timescl.from_datetimes(df.Time)
    sun = locations.ephem['sun']
    earth = locations.ephem['earth']
    bodies = get_bodies()

    # Add geocentric positions for each body
    for body_name, body in bodies.items():
        keyprefix = body_name.lower()
        obs = earth.at(tobjs).observe(body)
         # We could switch to RA/Dec if that improves. This works for now.
        eclip = obs.ecliptic_latlon()
        df[keyprefix+"_lat_deg"] = eclip[0].degrees
        df[keyprefix+"_lon_deg"] = eclip[1]._degrees % 360.0

    if addSatellites:
        # Add positions for satellites
        for body_name, sat in satellites.artisats.items():
            keyprefix = body_name.lower()
            subpoint = sat.at(tobjs).subpoint()
            df[keyprefix+"_sub_lat_deg"] = subpoint.latitude.degrees
            df[keyprefix+"_sub_lon_deg"] = subpoint.longitude.degrees % 360.0

    
    if addLocations:
        for loc_name, loc in locations_to_add.items():
            keyprefix = loc_name.lower()
            obs = earth.at(tobjs).observe(loc)
            eclip = obs.ecliptic_latlon()
            df[keyprefix+"_lat_deg"] = eclip[0].degrees
            df[keyprefix+"_lon_deg"] = eclip[1]._degrees % 360.0

    north_node_lon, south_node_lon = lunar_nodes.mean_lunar_nodes(tobjs)
    df["north_node_lon_deg"] = north_node_lon % 360.0
    df["south_node_lon_deg"] = south_node_lon % 360.0

    return df

