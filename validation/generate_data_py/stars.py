from skyfield.api import Star, Angle

# Coordinates for Sgr A* (J2000) from high-precision radio data:
#   RA  = 17h 45m 40.0409s
#   Dec = -29° 00' 28.118"
ra_sgrA  = Angle(hours=(17 + 45/60 + 40.0409/3600))
dec_sgrA = Angle(degrees=(-29 - 0/60 - 28.118/3600))
sgrA_star = Star(
    ra=ra_sgrA,
    dec=dec_sgrA,
    # Proper motion is extremely small, effectively negligible over short timescales.
)

GalacticCenter = sgrA_star # This is a massive blackhole, not a real star
