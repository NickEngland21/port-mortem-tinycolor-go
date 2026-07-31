package tinycolor

import "testing"

func TestKickoffDifferentialFixtures(t *testing.T) {
	cases := []struct {
		name, input, valid, format, hex8, string string
	}{
		{"red", "#ff0000", "true", "hex", "ff0000ff", "#ff0000"},
		{"rgba", "rgba(10,20,30,0.5)", "true", "rgb", "0a141e80", "rgba(10, 20, 30, 0.5)"},
		{"hsl", "hsl(120,100%,50%)", "true", "hsl", "00ff00ff", "hsl(120, 100%, 50%)"},
		{"hsla", "hsla(240,100%,50%,0.25)", "true", "hsl", "0000ff40", "hsla(240, 100%, 50%, 0.25)"},
		{"name", "rebeccapurple", "true", "name", "663399ff", "#663399"},
		{"percentage", "rgb(100%,0%,50%)", "true", "prgb", "ff0080ff", "rgb(100%, 0%, 50%)"},
		{"transparent", "transparent", "true", "name", "00000000", "transparent"},
		{"invalid", "not-a-color", "false", "", "000000ff", "#000000"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := New(tc.input)
			if c.IsValid() != (tc.valid == "true") || c.GetFormat() != tc.format || c.ToHex8() != tc.hex8 || c.ToString() != tc.string {
				t.Fatalf("got valid=%v format=%q hex8=%q string=%q", c.IsValid(), c.GetFormat(), c.ToHex8(), c.ToString())
			}
		})
	}
}

func TestObjectAndHSVFixtures(t *testing.T) {
	rgb := New(map[string]any{"r": 255, "g": 128, "b": 0, "a": 0.75})
	if rgb.ToHex8() != "ff8000bf" || rgb.ToString() != "rgba(255, 128, 0, 0.75)" {
		t.Fatalf("rgb object mismatch: %s %s", rgb.ToHex8(), rgb.ToString())
	}
	hsv := New(map[string]any{"h": 360, "s": 1, "v": 1})
	if hsv.ToHex8() != "ff0000ff" || hsv.ToString() != "hsv(0, 100%, 100%)" {
		t.Fatalf("hsv object mismatch: %s %s", hsv.ToHex8(), hsv.ToString())
	}
}

func TestAllOracleNamedColours(t *testing.T) {
	for name, value := range namedColors {
		c := New(name)
		expected, ok := parseHex(value, "name")
		if !ok || c.ToHex() != (&Color{r: expected.r, g: expected.g, b: expected.b}).ToHex() {
			t.Fatalf("named colour %s: got %s want %s", name, c.ToHex(), value)
		}
	}
}

func TestLegacySeparatorsAndHexShortcuts(t *testing.T) {
	if New("rgb 200 100 0").ToHexString() != "#c86400" {
		t.Fatal("space-separated rgb failed")
	}
	if New("rgba 255 0 0 0.5").ToHex8String() != "#ff000080" {
		t.Fatal("space-separated rgba failed")
	}
	if New("rgb 255 0 0").ToHexString(true) != "#f00" {
		t.Fatal("short hex failed")
	}
	if New("rgba 255 0 0 1").ToHex8String(true) != "#f00f" {
		t.Fatal("short hex8 failed")
	}
	if New("rgb .1 .1 .1").ToHexString() != "#000000" {
		t.Fatal("fractional rgb normalization failed")
	}
}

func TestRatioAndExplicitFormatOverrides(t *testing.T) {
	if FromRatio(map[string]float64{"r": 1, "g": 1, "b": 1}).ToHexString() != "#ffffff" {
		t.Fatal("ratio white failed")
	}
	if FromRatio(map[string]float64{"r": 1, "g": 0, "b": 0, "a": 0.5}).ToRGBString() != "rgba(255, 0, 0, 0.5)" {
		t.Fatal("ratio alpha failed")
	}
	c := FromRatio(map[string]float64{"r": 1, "g": 0, "b": 0, "a": 0.6}, Options{Format: "name"})
	if c.ToString() != "rgba(255, 0, 0, 0.6)" || c.ToString("hex3") != "#f00" || c.ToString("hex8") != "#ff000099" || c.ToString("name") != "#ff0000" {
		t.Fatal("format override failed")
	}
	transparent := FromRatio(map[string]float64{"r": 1, "g": 0, "b": 0, "a": 0}, Options{Format: "name"})
	if transparent.ToString() != "transparent" {
		t.Fatal("transparent format failed")
	}
}

func TestOracleCombinationAndUtilityCases(t *testing.T) {
	bright := New("#6699cc")
	if bright.Brighten(10).ToHex() != "7fb2e5" {
		t.Fatal("brighten parity failed")
	}
	if Complement("red").ToHex() != "00ffff" {
		t.Fatal("complement parity failed")
	}
	if got := []string{SplitComplement("red")[0].ToHex(), SplitComplement("red")[1].ToHex(), SplitComplement("red")[2].ToHex()}; got[0] != "ff0000" || got[1] != "ccff00" || got[2] != "0066ff" {
		t.Fatalf("split complement parity failed: %v", got)
	}
	if got := Analogous("red", 6, 30); got[0].ToHex() != "ff0000" || got[1].ToHex() != "ff0066" || got[5].ToHex() != "ff6600" {
		t.Fatalf("analogous parity failed: %v %v %v", got[0].ToHex(), got[1].ToHex(), got[5].ToHex())
	}
	if got := Monochromatic("red", 6); got[0].ToHex() != "ff0000" || got[1].ToHex() != "2a0000" || got[5].ToHex() != "d40000" {
		t.Fatalf("monochromatic parity failed: %v %v %v", got[0].ToHex(), got[1].ToHex(), got[5].ToHex())
	}
	if Mix("red", "blue", 50).ToHex8() != "800080ff" {
		t.Fatal("mix parity failed")
	}
	if New("rgba(1,2,3,.5)").ToFilter() != "progid:DXImageTransform.Microsoft.gradient(startColorstr=#80010203,endColorstr=#80010203)" {
		t.Fatal("filter parity failed")
	}
	if Readability("black", "white") != 21 || !IsReadable("black", "white", "AA", "small") {
		t.Fatal("readability parity failed")
	}
}

func TestMutationAndCombinations(t *testing.T) {
	c := New("#6699cc")
	if c.Lighten(10).ToHex() != "8cb2d9" {
		t.Fatalf("lighten mismatch: %s", c.ToHex())
	}
	if len(Triad("red")) != 3 || len(Tetrad("red")) != 4 || len(Analogous("red", 6, 30)) != 6 {
		t.Fatal("combination cardinality mismatch")
	}
	if Readability("black", "white") < 20 {
		t.Fatal("readability mismatch")
	}
}
