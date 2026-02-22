import os
from skyfield.api import load
from skyfield.api import N, W, wgs84

# Uses de440s.bsp from the repo's data/ folder,
# or download from https://naif.jpl.nasa.gov/pub/naif/generic_kernels/spk/planets/
bsp_path = os.path.join(os.path.dirname(__file__), '..', '..', 'data')
ephem = load(os.path.join(bsp_path, 'de440s.bsp'))

earth = ephem['earth']



MUMBAI_LAT = 19.0602766
MUMBAI_LON = 72.8577106

LONDON_LAT=51.5150534
LONDON_LON=-0.1016089

CUSHING_LAT=35.9859634 # CUSHING, OKLAHOMA
CUSHING_LON=-96.7954485

NY_LAT=40.714469 # NEW YORK 
NY_LON=-74.0194683

CHICAGO_LAT=41.8674558 # CHICAGO 
CHICAGO_LON=-87.6483924

NULL_ISLAND_loc = earth + wgs84.latlon(0 * N, 0 * W)

CHICAGO_loc = earth + wgs84.latlon(CHICAGO_LAT, CHICAGO_LON)   # CHICAGO # 41.8674558,-87.6483924

LONDON_loc = earth + wgs84.latlon(LONDON_LAT, LONDON_LON)   # LONDON # 51.5150534,-0.1016089

CUSHING_loc = earth + wgs84.latlon(CUSHING_LAT, CUSHING_LON)   # CUSHING # 35.9859634,-96.7954485

NY_loc = earth + wgs84.latlon(NY_LAT, NY_LON)   # NY # 40.714469,-74.0194683

MUMBAI_loc = earth + wgs84.latlon(MUMBAI_LAT, MUMBAI_LON)   # MUMBAI # 19.0602766,72.8577106
