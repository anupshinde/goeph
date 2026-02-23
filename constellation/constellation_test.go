package constellation

import "testing"

func TestAt_Skyfield(t *testing.T) {
	// Test cases validated against Skyfield's load_constellation_map().
	tests := []struct {
		name    string
		ra, dec float64
		want    string
	}{
		{"North pole", 0, 90, "UMi"},
		{"South pole", 0, -90, "Oct"},
		{"Origin", 0, 0, "Psc"},
		{"Near Gemini", 6, 20, "Ori"},
		{"Near UMa", 12, 65, "UMa"},
		{"Betelgeuse", 5.92, 7.41, "Ori"},
		{"Sirius", 6.75, -16.72, "CMa"},
		{"Arcturus", 14.26, 19.18, "Boo"},
		{"Vega", 18.62, 38.78, "Lyr"},
		{"Capella", 5.24, 46.0, "Aur"},
		{"M31", 1.17, 35.62, "And"},
		{"Near UMi 2", 16, 80, "UMi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := At(tt.ra, tt.dec)
			if got != tt.want {
				t.Errorf("At(%.2f, %.2f) = %q, want %q", tt.ra, tt.dec, got, tt.want)
			}
		})
	}
}

func TestAt_AllConstellations(t *testing.T) {
	// Ensure At returns only valid abbreviations for a grid of positions.
	valid := make(map[string]bool, 88)
	for _, pair := range constellationNames {
		valid[pair[0]] = true
	}

	for ra := 0.0; ra < 24.0; ra += 1.0 {
		for dec := -90.0; dec <= 90.0; dec += 15.0 {
			abbr := At(ra, dec)
			if !valid[abbr] {
				t.Errorf("At(%.1f, %.1f) = %q, not a valid abbreviation", ra, dec, abbr)
			}
		}
	}
}

func TestName(t *testing.T) {
	tests := []struct {
		abbr, want string
	}{
		{"Ori", "Orion"},
		{"UMa", "Ursa Major"},
		{"Cyg", "Cygnus"},
		{"XXX", ""},
	}
	for _, tt := range tests {
		got := Name(tt.abbr)
		if got != tt.want {
			t.Errorf("Name(%q) = %q, want %q", tt.abbr, got, tt.want)
		}
	}
}

func TestAbbreviation(t *testing.T) {
	tests := []struct {
		name, want string
	}{
		{"Orion", "Ori"},
		{"Ursa Major", "UMa"},
		{"Cygnus", "Cyg"},
		{"Unknown", ""},
	}
	for _, tt := range tests {
		got := Abbreviation(tt.name)
		if got != tt.want {
			t.Errorf("Abbreviation(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestNames_Count(t *testing.T) {
	names := Names()
	if len(names) != 88 {
		t.Errorf("Names() returned %d entries, want 88", len(names))
	}
}

func TestDataIntegrity(t *testing.T) {
	// Verify data arrays have expected sizes.
	if len(sortedRA) != 235 {
		t.Errorf("sortedRA has %d entries, want 235", len(sortedRA))
	}
	if len(sortedDec) != 199 {
		t.Errorf("sortedDec has %d entries, want 199", len(sortedDec))
	}
	if len(grid) != 236 {
		t.Errorf("grid has %d rows, want 236", len(grid))
	}
	if len(grid[0]) != 202 {
		t.Errorf("grid has %d cols, want 202", len(grid[0]))
	}
	if len(abbreviations) != 88 {
		t.Errorf("abbreviations has %d entries, want 88", len(abbreviations))
	}

	// Verify grid values are valid indices.
	for i := range grid {
		for j := range grid[i] {
			idx := grid[i][j]
			if idx < 0 || int(idx) >= len(abbreviations) {
				t.Errorf("grid[%d][%d] = %d, out of range [0, %d)", i, j, idx, len(abbreviations))
			}
		}
	}
}
