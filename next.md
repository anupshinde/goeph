This was related to architecting the repo.
Earlier I was calling it "sfgo", but now its called "goeph". The references below are using the old pseudoname "sfgo"

If sfgo were a standalone repo (e.g. github.com/anupshinde/sfgo), I'd structure it like this:


sfgo/
├── go.mod                      # module github.com/anupshinde/sfgo
├── go.sum
├── sfgo.go                     # package sfgo — top-level convenience (Load, etc.)
│
├── spk/                        # SPK ephemeris parser
│   ├── spk.go
│   └── bodies.go
│
├── coord/                      # coordinate transformations
│   └── coord.go
│
├── timescale/                  # time scale conversions (own package)
│   └── timescale.go
│
├── satellite/                  # satellite propagation helpers
│   └── satellite.go
│
├── star/                       # star catalog (galactic center, etc.)
│   └── star.go
│
├── examples/
│   └── celestial_csv/
│       ├── main.go
│       ├── compute.go
│       └── data.go
│
└── testdata/
    └── de440s_excerpt.bsp      # small BSP slice for tests
Key differences from the current directory layout:

Each concern gets its own package — timescale, satellite, star become sub-packages instead of files in the root sfgo package. This avoids a "god package" where everything is sfgo.Something. Users import only what they need:


import "github.com/anupshinde/sfgo/spk"
import "github.com/anupshinde/sfgo/coord"
import "github.com/anupshinde/sfgo/timescale"
Thin root package — sfgo.go would be a convenience layer, maybe a Load() that wraps spk.Open(), or re-exports commonly used types. Or it could just be package docs. Skyfield itself works this way — skyfield.api is the convenience entry point.

Location moves to coord/ — it's a geodetic concept, belongs with coordinate transforms.

bodyWrtSSB becomes generic — the segment graph walker replaces the hardcoded switch, making it a proper library. This is the main functional gap vs Skyfield.

testdata/ — a trimmed BSP file for unit tests so CI doesn't need the full 32MB de440s.bsp.

Own go.mod — independent versioning, no coupling to fractiz-research dependencies.

If you ever extract it to its own repo, the main refactor would be splitting the root package into sub-packages — which is straightforward since the files already group cleanly by concern.
