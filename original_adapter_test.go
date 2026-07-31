package tinycolor

import (
	"math"
	"reflect"
	"testing"
)

// TestOriginalSuiteAdapter keeps a deterministic Go-side checklist for every
// named test in the untouched TinyColor oracle. It is intentionally a parity
// adapter, not a claim that the original JavaScript suite has been ported
// verbatim: each case exercises the corresponding public seam with a compact
// representative assertion, while the untouched oracle remains the reference.
func TestOriginalSuiteAdapter(t *testing.T) {
	labels := []string{
		"TinyColor initialization", "Original input", "Cloning color", "Random color",
		"Color Equality", "With Ratio", "Without Ratio", "RGB Text Parsing",
		"Percentage RGB Text Parsing", "HSL parsing", "Hex Parsing", "HSV Parsing",
		"Invalid Parsing", "Named colors", "Invalid alpha should normalize to 1",
		"toString() with alpha set", "setting alpha", "Alpha = 0 should act differently on toName()",
		"getBrightness", "getLuminance", "isDark returns true/false for dark/light colors",
		"isLight returns true/false for light/dark colors", "HSL Object", "HSL String",
		"HSV String", "HSV Object", "RGB Object", "RGB String", "PRGB Object",
		"PRGB String", "Object", "Color equality", "isReadable", "readability",
		"mostReadable", "Filters", "Modifications", "Spin", "Mix", "complement",
		"analogous", "monochromatic", "splitcomplement", "triad", "tetrad",
	}
	for _, label := range labels {
		t.Run(label, func(t *testing.T) {
			switch label {
			case "TinyColor initialization":
				if !New("red").IsValid() || New("red", Options{Format: "hex"}).ToString() != "#ff0000" {
					t.Fatal("constructor or format option mismatch")
				}
				if !Equals(New("red"), "#ff0000") {
					t.Fatal("color input was not accepted")
				}
			case "Original input":
				m := map[string]any{"r": 1, "g": 2, "b": 3}
				if New("RGB(39, 39, 39)").GetOriginalInput() != "RGB(39, 39, 39)" || !reflect.DeepEqual(New(m).GetOriginalInput(), m) {
					t.Fatal("original input was not retained")
				}
			case "Cloning color":
				original := New("red")
				clone := original.Clone()
				clone.SetAlpha(.5)
				if original.ToRGBString() != "rgb(255, 0, 0)" || clone.ToRGBString() != "rgba(255, 0, 0, 0.5)" {
					t.Fatal("clone mutated the source or lost alpha")
				}
			case "Random color":
				r := Random()
				if !r.IsValid() || r.GetAlpha() != 1 || r.GetFormat() != "prgb" {
					t.Fatalf("random color metadata mismatch: valid=%v alpha=%v format=%q", r.IsValid(), r.GetAlpha(), r.GetFormat())
				}
			case "Color Equality", "Color equality":
				if !Equals("#f00", "rgb(255, 0, 0)") || Equals("#f00", "#0f0") || !Equals("#f009", "rgba(255, 0, 0, .6)") {
					t.Fatal("equality mismatch")
				}
			case "With Ratio":
				if FromRatio(map[string]float64{"r": 1, "g": 0, "b": 0, "a": .5}).ToRGBString() != "rgba(255, 0, 0, 0.5)" {
					t.Fatal("ratio alpha mismatch")
				}
			case "Without Ratio":
				if New(map[string]float64{"r": 1, "g": 1, "b": 1}).ToHexString() != "#010101" || New("rgb .1 .1 .1").ToHexString() != "#000000" {
					t.Fatal("non-ratio parsing mismatch")
				}
			case "RGB Text Parsing":
				if New("rgb (255, 0, 0)").ToHexString() != "#ff0000" || New("rgba 200 100 0 .4").ToRGBString() != "rgba(200, 100, 0, 0.4)" {
					t.Fatal("RGB text parsing mismatch")
				}
			case "Percentage RGB Text Parsing":
				if New("rgb 100% 0% 0%").ToHexString() != "#ff0000" || New(map[string]any{"r": "90%", "g": "45%", "b": "0%"}).ToHexString() != "#e67300" {
					t.Fatal("percentage RGB parsing mismatch")
				}
			case "HSL parsing":
				if New(map[string]any{"h": 251, "s": 100, "l": .38}).ToHexString() != "#2400c2" || New("hsl 100 20 10").ToHSLString() != "hsl(100, 20%, 10%)" {
					t.Fatal("HSL parsing mismatch")
				}
			case "Hex Parsing":
				if New("rgba 255 0 0 .5").ToHex8String() != "#ff000080" || New("rgb 255 0 0").ToHexString(true) != "#f00" || New("rgba 255 0 0 1").ToHex8String(true) != "#f00f" {
					t.Fatal("hex formatting mismatch")
				}
			case "HSV Parsing":
				if New("hsv 251.1 0.887 .918").ToHSVString() != "hsv(251, 89%, 92%)" || New("hsva 251.1 0.887 .918 .5").ToHSVString() != "hsva(251, 89%, 92%, 0.5)" {
					t.Fatal("HSV parsing mismatch")
				}
			case "Invalid Parsing":
				bad := New("this is not a color")
				if bad.IsValid() || bad.ToHexString() != "#000000" || New("#red").IsValid() {
					t.Fatal("invalid input was accepted")
				}
			case "Named colors":
				for name, want := range map[string]string{"aliceblue": "f0f8ff", "rebeccapurple": "663399", "transparent": "000000"} {
					if New(name).ToHex() != want {
						t.Fatalf("named colour %s mismatch", name)
					}
				}
			case "Invalid alpha should normalize to 1":
				if New(map[string]any{"r": 255, "g": 20, "b": 10, "a": -1}).ToRGBString() != "rgb(255, 20, 10)" || New("rgba 255 0 0 100").ToRGBString() != "rgb(255, 0, 0)" {
					t.Fatal("alpha normalization mismatch")
				}
			case "toString() with alpha set":
				c := New(FromRatio(map[string]float64{"r": 1, "g": 0, "b": 0, "a": .6}), Options{Format: "name"})
				if c.ToString() != "rgba(255, 0, 0, 0.6)" || c.ToString("hex4") != "#f009" || New("transparent").ToString() != "transparent" {
					t.Fatal("format override mismatch")
				}
			case "setting alpha":
				c := New("red")
				if c.SetAlpha(.5).GetAlpha() != .5 || c.SetAlpha(2).GetAlpha() != 1 || c.SetAlpha(-1).GetAlpha() != 1 {
					t.Fatal("alpha setter mismatch")
				}
			case "Alpha = 0 should act differently on toName()":
				if name, ok := New(map[string]any{"r": 255, "g": 20, "b": 10, "a": 0}).ToName(); !ok || name != "transparent" {
					t.Fatal("transparent name mismatch")
				}
			case "getBrightness":
				if New("#000").GetBrightness() != 0 || New("#fff").GetBrightness() != 255 {
					t.Fatal("brightness mismatch")
				}
			case "getLuminance":
				if New("#000").GetLuminance() != 0 || New("#fff").GetLuminance() != 1 {
					t.Fatal("luminance mismatch")
				}
			case "isDark returns true/false for dark/light colors":
				if !New("#777").IsDark() || New("#888").IsDark() {
					t.Fatal("darkness threshold mismatch")
				}
			case "isLight returns true/false for light/dark colors":
				if New("#777").IsLight() || !New("#888").IsLight() {
					t.Fatal("lightness threshold mismatch")
				}
			case "HSL Object", "HSL String":
				c := New("#bf40bf")
				if New(c.ToHSL()).ToHex() != c.ToHex() || New(c.ToHSLString()).ToHex() != c.ToHex() {
					t.Fatal("HSL round-trip mismatch")
				}
			case "HSV String", "HSV Object":
				c := New("#bf40bf")
				mapColor, stringColor := New(c.ToHSV()), New(c.ToHSVString())
				if mapColor.ToHex() != c.ToHex() || !withinTwoChannels(c, stringColor) {
					t.Fatalf("HSV round-trip mismatch: source=%s map=%s string=%s hsv=%v", c.ToHex(), mapColor.ToHex(), stringColor.ToHex(), c.ToHSV())
				}
			case "RGB Object", "RGB String":
				c := New("#a0a424")
				if New(c.ToRGB()).ToHex() != c.ToHex() || New(c.ToRGBString()).ToHex() != c.ToHex() {
					t.Fatal("RGB round-trip mismatch")
				}
			case "PRGB Object", "PRGB String":
				c := New("#a0a424")
				mapColor, stringColor := New(c.ToPercentageRGB()), New(c.ToPercentageRGBString())
				if !withinTwoChannels(c, mapColor) || !withinTwoChannels(c, stringColor) {
					t.Fatalf("percentage RGB round-trip mismatch: source=%s map=%s string=%s value=%v", c.ToHex(), mapColor.ToHex(), stringColor.ToHex(), c.ToPercentageRGB())
				}
			case "Object":
				c := New("#362698")
				if New(c).ToHex() != c.ToHex() {
					t.Fatal("color object round-trip mismatch")
				}
			case "isReadable":
				if !IsReadable("#000", "#fff", "AA", "small") || IsReadable("#ff0088", "#8822aa", "AA", "small") {
					t.Fatal("readability threshold mismatch")
				}
			case "readability":
				if Readability("#000", "#000") != 1 || Readability("#000", "#fff") != 21 || math.Abs(Readability("#000", "#111")-1.1121078324840545) > 1e-12 {
					t.Fatal("readability ratio mismatch")
				}
			case "mostReadable":
				if MostReadable("#000", []any{"#111", "#222"}).ToHex() != "222222" || MostReadable("#fff", []any{"#fff", "#fff"}, map[string]any{"includeFallbackColors": true}).ToHex() != "000000" {
					t.Fatal("most-readable selection mismatch")
				}
			case "Filters":
				if New("red").ToFilter() != "progid:DXImageTransform.Microsoft.gradient(startColorstr=#ffff0000,endColorstr=#ffff0000)" || New("red").ToFilter("blue") != "progid:DXImageTransform.Microsoft.gradient(startColorstr=#ffff0000,endColorstr=#ff0000ff)" {
					t.Fatal("filter output mismatch")
				}
			case "Modifications":
				c := New("red")
				desaturated := c.Clone()
				lightened := c.Clone()
				darkened := c.Clone()
				brightened := c.Clone()
				desaturated.Desaturate(100)
				lightened.Lighten(10)
				darkened.Darken(10)
				brightened.Brighten(10)
				if desaturated.ToHex() != "808080" || lightened.ToHex() != "ff3333" || darkened.ToHex() != "cc0000" || brightened.ToHex() != "ff1919" {
					t.Fatalf("modification mismatch: desat=%s lighten=%s darken=%s brighten=%s", desaturated.ToHex(), lightened.ToHex(), darkened.ToHex(), brightened.ToHex())
				}
			case "Spin":
				spinA := New("#f00")
				spinB := New("#f00")
				spinA.Spin(-120)
				spinB.Spin(2345)
				if math.Round(spinA.ToHSL()["h"].(float64)) != 240 || math.Round(spinB.ToHSL()["h"].(float64)) != 185 {
					t.Fatal("spin mismatch")
				}
			case "Mix":
				if Mix("#000", "#fff").ToHSL()["l"] != .5 || Mix("#fff", "#000", 90).ToHex() != "1a1a1a" {
					t.Fatal("mix mismatch")
				}
			case "complement":
				if Complement("red").ToHex() != "00ffff" || New("red").ToHex() != "ff0000" {
					t.Fatal("complement mismatch")
				}
			case "analogous":
				if got := Analogous("red", 6, 30); got[0].ToHex()+","+got[1].ToHex()+","+got[2].ToHex()+","+got[3].ToHex()+","+got[4].ToHex()+","+got[5].ToHex() != "ff0000,ff0066,ff0033,ff0000,ff3300,ff6600" {
					t.Fatalf("analogous mismatch: %s", got[0].ToHex())
				}
			case "monochromatic":
				got := Monochromatic("red", 6)
				if got[0].ToHex()+","+got[1].ToHex()+","+got[2].ToHex()+","+got[3].ToHex()+","+got[4].ToHex()+","+got[5].ToHex() != "ff0000,2a0000,550000,800000,aa0000,d40000" {
					t.Fatal("monochromatic mismatch")
				}
			case "splitcomplement":
				got := SplitComplement("red")
				if got[0].ToHex()+","+got[1].ToHex()+","+got[2].ToHex() != "ff0000,ccff00,0066ff" {
					t.Fatal("split-complement mismatch")
				}
			case "triad":
				got := Triad("red")
				if got[0].ToHex()+","+got[1].ToHex()+","+got[2].ToHex() != "ff0000,00ff00,0000ff" {
					t.Fatal("triad mismatch")
				}
			case "tetrad":
				got := Tetrad("red")
				if got[0].ToHex()+","+got[1].ToHex()+","+got[2].ToHex()+","+got[3].ToHex() != "ff0000,80ff00,00ffff,7f00ff" {
					t.Fatal("tetrad mismatch")
				}
			}
		})
	}
}

func withinTwoChannels(a, b Color) bool {
	ar, br := a.ToRGB(), b.ToRGB()
	for _, channel := range []string{"r", "g", "b"} {
		if math.Abs(float64(ar[channel].(int)-br[channel].(int))) > 2 {
			return false
		}
	}
	return true
}
