package tinycolor

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

// names.json is extracted from the untouched TinyColor oracle at kickoff.
// It is embedded so the private candidate stays standalone.
//
//go:embed names.json
var namesJSON []byte

var namedColors = loadNamedColors()

var (
	functionInputRE = regexp.MustCompile(`^(rgba?|hsla?|hsva?)\s*\((.*)\)$`)
	spaceInputRE    = regexp.MustCompile(`^(rgba?|hsla?|hsva?)\s+(.+)$`)
)

func loadNamedColors() map[string]string {
	var m map[string]string
	if err := json.Unmarshal(namesJSON, &m); err != nil {
		panic(err)
	}
	return m
}

// Options mirrors the options that affect the oracle's constructor.
type Options struct {
	Format       string
	GradientType string
}

// Color is the private Go port's value representation. Methods that mirror
// TinyColor's modifying methods update the receiver and return it.
type Color struct {
	r, g, b      float64
	a            float64
	original     any
	format       string
	gradientType string
	valid        bool
}

// New constructs a Color from CSS-like text or a map with r/g/b, h/s/l or
// h/s/v fields. Invalid input deliberately normalizes to opaque black, as the
// original library does.
func New(input any, opts ...Options) Color {
	opt := Options{}
	if len(opts) > 0 {
		opt = opts[0]
	}
	c := Color{original: input, a: 1, format: opt.Format, gradientType: opt.GradientType}
	if parsed, ok := parseInput(input); ok {
		c.r, c.g, c.b, c.a, c.format, c.valid = parsed.r, parsed.g, parsed.b, parsed.a, parsed.format, true
		if opt.Format != "" {
			c.format = opt.Format
		}
		return c
	}
	c.r, c.g, c.b, c.a, c.valid = 0, 0, 0, 1, false
	return c
}

// FromRatio accepts 0..1 channel ratios before normal construction.
func FromRatio(input map[string]float64, opts ...Options) Color {
	m := map[string]any{}
	for k, v := range input {
		if k == "a" {
			m[k] = v
		} else {
			m[k] = fmt.Sprintf("%.12g%%", clamp01(v)*100)
		}
	}
	return New(m, opts...)
}

type parsedColor struct {
	r, g, b, a float64
	format     string
}

func parseInput(input any) (parsedColor, bool) {
	switch v := input.(type) {
	case Color:
		return parsedColor{v.r, v.g, v.b, v.a, v.format}, v.valid
	case *Color:
		if v == nil {
			return parsedColor{}, false
		}
		return parsedColor{v.r, v.g, v.b, v.a, v.format}, v.valid
	case string:
		return parseString(v)
	case map[string]any:
		return parseMap(v)
	case map[string]float64:
		m := make(map[string]any, len(v))
		for k, x := range v {
			m[k] = x
		}
		return parseMap(m)
	default:
		return parsedColor{}, false
	}
}

func parseString(raw string) (parsedColor, bool) {
	s := strings.ToLower(strings.TrimSpace(raw))
	if s == "transparent" {
		return parsedColor{0, 0, 0, 0, "name"}, true
	}
	if hex, ok := namedColors[s]; ok {
		return parseHex(hex, "name")
	}
	if strings.HasPrefix(s, "#") {
		return parseHex(strings.TrimPrefix(s, "#"), "hex")
	}
	if len(s) == 3 || len(s) == 4 || len(s) == 6 || len(s) == 8 {
		if _, err := strconv.ParseUint(s, 16, 32); err == nil {
			return parseHex(s, "hex")
		}
	}
	m := functionInputRE.FindStringSubmatch(s)
	if len(m) != 3 {
		noParen := spaceInputRE.FindStringSubmatch(s)
		if len(noParen) != 3 {
			return parsedColor{}, false
		}
		m = noParen
	}
	parts := splitComponents(m[2])
	if len(parts) < 3 || len(parts) > 4 {
		return parsedColor{}, false
	}
	alpha := 1.0
	if len(parts) == 4 {
		var ok bool
		alpha, ok = parseAlpha(parts[3])
		if !ok {
			alpha = 1
		}
	}
	switch m[1] {
	case "rgb", "rgba":
		pr, ok1 := parseChannel(parts[0], 255)
		pg, ok2 := parseChannel(parts[1], 255)
		pb, ok3 := parseChannel(parts[2], 255)
		if !ok1 || !ok2 || !ok3 {
			return parsedColor{}, false
		}
		format := "rgb"
		if strings.Contains(parts[0], "%") || strings.Contains(parts[1], "%") || strings.Contains(parts[2], "%") {
			format = "prgb"
		}
		return parsedColor{pr, pg, pb, alpha, format}, true
	case "hsl", "hsla":
		h, ok1 := parseNumber(parts[0])
		sat, ok2 := parseUnit(parts[1])
		l, ok3 := parseUnit(parts[2])
		if !ok1 || !ok2 || !ok3 {
			return parsedColor{}, false
		}
		r, g, b := hslToRGB(h/360, sat, l)
		return parsedColor{r, g, b, alpha, "hsl"}, true
	case "hsv", "hsva":
		h, ok1 := parseNumber(parts[0])
		sat, ok2 := parseUnit(parts[1])
		v, ok3 := parseUnit(parts[2])
		if !ok1 || !ok2 || !ok3 {
			return parsedColor{}, false
		}
		r, g, b := hsvToRGB(h/360, sat, v)
		return parsedColor{r, g, b, alpha, "hsv"}, true
	}
	return parsedColor{}, false
}

func parseHex(s, format string) (parsedColor, bool) {
	if len(s) == 3 || len(s) == 4 {
		var b strings.Builder
		for _, ch := range s {
			b.WriteRune(ch)
			b.WriteRune(ch)
		}
		s = b.String()
	}
	if len(s) != 6 && len(s) != 8 {
		return parsedColor{}, false
	}
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return parsedColor{}, false
	}
	if len(s) == 6 {
		return parsedColor{float64(n >> 16), float64((n >> 8) & 255), float64(n & 255), 1, format}, true
	}
	return parsedColor{float64(n >> 24), float64((n >> 16) & 255), float64((n >> 8) & 255), float64(n&255) / 255, format}, true
}

func parseMap(m map[string]any) (parsedColor, bool) {
	if r, rok := valueNumber(m["r"]); rok {
		g, gok := valueNumber(m["g"])
		b, bok := valueNumber(m["b"])
		if gok && bok {
			a := 1.0
			if av, ok := valueNumber(m["a"]); ok {
				a = boundAlpha(av)
			}
			format := "rgb"
			if strings.Contains(fmt.Sprint(m["r"]), "%") || strings.Contains(fmt.Sprint(m["g"]), "%") || strings.Contains(fmt.Sprint(m["b"]), "%") {
				format = "prgb"
			}
			return parsedColor{channelValue(r, m["r"]), channelValue(g, m["g"]), channelValue(b, m["b"]), a, format}, true
		}
	}
	if h, hok := valueNumber(m["h"]); hok {
		s, sok := valueNumber(m["s"])
		l, lok := valueNumber(m["l"])
		v, vok := valueNumber(m["v"])
		a := 1.0
		if av, ok := valueNumber(m["a"]); ok {
			a = boundAlpha(av)
		}
		if sok && lok {
			if math.Abs(s) > 1 || strings.Contains(fmt.Sprint(m["s"]), "%") {
				s /= 100
			}
			if math.Abs(l) > 1 || strings.Contains(fmt.Sprint(m["l"]), "%") {
				l /= 100
			}
			r, g, b := hslToRGB(h/360, s, l)
			return parsedColor{r, g, b, a, "hsl"}, true
		}
		if sok && vok {
			if math.Abs(s) > 1 || strings.Contains(fmt.Sprint(m["s"]), "%") {
				s /= 100
			}
			if math.Abs(v) > 1 || strings.Contains(fmt.Sprint(m["v"]), "%") {
				v /= 100
			}
			r, g, b := hsvToRGB(h/360, s, v)
			return parsedColor{r, g, b, a, "hsv"}, true
		}
	}
	return parsedColor{}, false
}

func valueNumber(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case string:
		x = strings.TrimSpace(strings.TrimSuffix(x, "%"))
		n, err := strconv.ParseFloat(x, 64)
		return n, err == nil
	default:
		return 0, false
	}
}

func channelValue(n float64, raw any) float64 {
	if strings.Contains(fmt.Sprint(raw), "%") {
		return clamp01(n/100) * 255
	}
	if n >= 0 && n <= 1 && (n != math.Trunc(n) || n == 0) {
		return n * 255
	}
	return clamp(n, 0, 255)
}

func splitComponents(s string) []string {
	s = strings.ReplaceAll(s, ",", " ")
	return strings.Fields(s)
}

func parseNumber(s string) (float64, bool) {
	n, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(s, "%")), 64)
	return n, err == nil
}

func parseUnit(s string) (float64, bool) {
	n, ok := parseNumber(s)
	if !ok {
		return 0, false
	}
	if strings.Contains(s, "%") || math.Abs(n) > 1 {
		n /= 100
	}
	return clamp01(n), true
}

func parseChannel(s string, max float64) (float64, bool) {
	n, ok := parseNumber(s)
	if !ok {
		return 0, false
	}
	if strings.Contains(s, "%") {
		return clamp01(n/100) * max, true
	}
	return clamp(n, 0, max), true
}

func parseAlpha(s string) (float64, bool) {
	n, ok := parseNumber(s)
	if !ok {
		return 1, false
	}
	return boundAlpha(n), true
}

// boundAlpha follows the oracle's alpha contract: only finite values in
// [0,1] are accepted; invalid, negative, and over-range values normalize to
// opaque (1), while an explicit zero remains transparent.
func boundAlpha(a float64) float64 {
	if math.IsNaN(a) || math.IsInf(a, 0) || a < 0 || a > 1 {
		return 1
	}
	return a
}
func clamp01(x float64) float64 { return clamp(x, 0, 1) }
func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

func rgbToHSL(r, g, b float64) (h, s, l float64) {
	r, g, b = r/255, g/255, b/255
	max, min := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
	l = (max + min) / 2
	if max == min {
		return 0, 0, l
	}
	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	return h / 6, s, l
}

func hslToRGB(h, s, l float64) (r, g, b float64) {
	h = math.Mod(h, 1)
	if h < 0 {
		h += 1
	}
	s, l = clamp01(s), clamp01(l)
	if s == 0 {
		return l * 255, l * 255, l * 255
	}
	q := l * (1 + s)
	if l >= 0.5 {
		q = l + s - l*s
	}
	p := 2*l - q
	return hueToRGB(p, q, h+1.0/3.0) * 255, hueToRGB(p, q, h) * 255, hueToRGB(p, q, h-1.0/3.0) * 255
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t++
	}
	if t > 1 {
		t--
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

func rgbToHSV(r, g, b float64) (h, s, v float64) {
	r, g, b = r/255, g/255, b/255
	max, min := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
	d := max - min
	v = max
	if max != 0 {
		s = d / max
	}
	if d == 0 {
		return 0, s, v
	}
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	return h / 6, s, v
}

func hsvToRGB(h, s, v float64) (r, g, b float64) {
	h = math.Mod(h, 1)
	if h < 0 {
		h += 1
	}
	s, v = clamp01(s), clamp01(v)
	i := math.Floor(h * 6)
	f := h*6 - i
	p, q, t := v*(1-s), v*(1-f*s), v*(1-(1-f)*s)
	switch int(i) % 6 {
	case 0:
		return v * 255, t * 255, p * 255
	case 1:
		return q * 255, v * 255, p * 255
	case 2:
		return p * 255, v * 255, t * 255
	case 3:
		return p * 255, q * 255, v * 255
	case 4:
		return t * 255, p * 255, v * 255
	default:
		return v * 255, p * 255, q * 255
	}
}

func roundByte(x float64) int { return int(math.Floor(x + 0.5)) }
func jsRound(x float64) int {
	if x < 0 {
		magnitude := -x
		whole, fraction := math.Modf(magnitude)
		if fraction == 0.5 {
			return -int(whole)
		}
	}
	return int(math.Floor(x + 0.5))
}
func roundAlpha(a float64) float64 {
	return math.Round(clamp01(a)*100) / 100
}
func hexByte(x float64) string { return fmt.Sprintf("%02x", roundByte(clamp(x, 0, 255))) }

func (c Color) IsValid() bool         { return c.valid }
func (c Color) GetFormat() string     { return c.format }
func (c Color) GetOriginalInput() any { return c.original }
func (c Color) GetAlpha() float64     { return c.a }
func (c Color) ToRGB() map[string]any {
	return map[string]any{"r": roundByte(c.r), "g": roundByte(c.g), "b": roundByte(c.b), "a": c.a}
}
func (c Color) ToHSL() map[string]any {
	h, s, l := rgbToHSL(c.r, c.g, c.b)
	return map[string]any{"h": h * 360, "s": s, "l": l, "a": c.a}
}
func (c Color) ToHSV() map[string]any {
	h, s, v := rgbToHSV(c.r, c.g, c.b)
	return map[string]any{"h": h * 360, "s": s, "v": v, "a": c.a}
}
func (c Color) ToHex(allow3 ...bool) string {
	full := hexByte(c.r) + hexByte(c.g) + hexByte(c.b)
	if len(allow3) > 0 && allow3[0] {
		return shortHex(full)
	}
	return full
}
func (c Color) ToHexString(allow3 ...bool) string { return "#" + c.ToHex(allow3...) }
func (c Color) ToHex8(allow4 ...bool) string {
	full := c.ToHex() + hexByte(c.a*255)
	if len(allow4) > 0 && allow4[0] {
		return shortHex8(full)
	}
	return full
}
func (c Color) ToHex8String(allow4 ...bool) string { return "#" + c.ToHex8(allow4...) }
func (c Color) GetBrightness() float64             { return (c.r*299 + c.g*587 + c.b*114) / 1000 }
func (c Color) IsDark() bool                       { return c.GetBrightness() < 128 }
func (c Color) IsLight() bool                      { return !c.IsDark() }
func (c Color) GetLuminance() float64 {
	channel := func(x float64) float64 {
		x /= 255
		if x <= 0.03928 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(c.r) + 0.7152*channel(c.g) + 0.0722*channel(c.b)
}

func (c *Color) SetAlpha(a float64) *Color { c.a = boundAlpha(a); return c }

func (c Color) ToRGBString() string {
	if c.a == 1 {
		return fmt.Sprintf("rgb(%d, %d, %d)", roundByte(c.r), roundByte(c.g), roundByte(c.b))
	}
	return fmt.Sprintf("rgba(%d, %d, %d, %s)", roundByte(c.r), roundByte(c.g), roundByte(c.b), formatAlpha(c.a))
}
func (c Color) ToPercentageRGBString() string {
	pr := roundByte(clamp01(c.r/255) * 100)
	pg := roundByte(clamp01(c.g/255) * 100)
	pb := roundByte(clamp01(c.b/255) * 100)
	if c.a == 1 {
		return fmt.Sprintf("rgb(%d%%, %d%%, %d%%)", pr, pg, pb)
	}
	return fmt.Sprintf("rgba(%d%%, %d%%, %d%%, %s)", pr, pg, pb, formatAlpha(c.a))
}
func (c Color) ToPercentageRGB() map[string]any {
	return map[string]any{
		"r": fmt.Sprintf("%d%%", roundByte(clamp01(c.r/255)*100)),
		"g": fmt.Sprintf("%d%%", roundByte(clamp01(c.g/255)*100)),
		"b": fmt.Sprintf("%d%%", roundByte(clamp01(c.b/255)*100)),
		"a": c.a,
	}
}
func (c Color) ToHSLString() string {
	h, s, l := rgbToHSL(c.r, c.g, c.b)
	hi, si, li := roundByte(h*360), roundByte(s*100), roundByte(l*100)
	if c.a == 1 {
		return fmt.Sprintf("hsl(%d, %d%%, %d%%)", hi, si, li)
	}
	return fmt.Sprintf("hsla(%d, %d%%, %d%%, %s)", hi, si, li, formatAlpha(c.a))
}
func (c Color) ToHSVString() string {
	h, s, v := rgbToHSV(c.r, c.g, c.b)
	hi, si, vi := roundByte(h*360), roundByte(s*100), roundByte(v*100)
	if c.a == 1 {
		return fmt.Sprintf("hsv(%d, %d%%, %d%%)", hi, si, vi)
	}
	return fmt.Sprintf("hsva(%d, %d%%, %d%%, %s)", hi, si, vi, formatAlpha(c.a))
}

func formatAlpha(a float64) string { return strconv.FormatFloat(roundAlpha(a), 'f', -1, 64) }

func (c Color) ToName() (string, bool) {
	if c.a == 0 {
		return "transparent", true
	}
	if c.a < 1 {
		return "", false
	}
	hex := shortHex(c.ToHex())
	for name, value := range namedColors {
		if value == hex {
			return name, true
		}
	}
	return "", false
}

func shortHex(hex string) string {
	if len(hex) != 6 || hex[0] != hex[1] || hex[2] != hex[3] || hex[4] != hex[5] {
		return hex
	}
	return string([]byte{hex[0], hex[2], hex[4]})
}

func shortHex8(hex string) string {
	if len(hex) != 8 || hex[0] != hex[1] || hex[2] != hex[3] || hex[4] != hex[5] || hex[6] != hex[7] {
		return hex
	}
	return string([]byte{hex[0], hex[2], hex[4], hex[6]})
}

func (c Color) String() string { return c.ToString() }
func (c Color) ToString(formats ...string) string {
	formatSet := len(formats) > 0
	format := c.format
	if formatSet {
		format = formats[0]
	}
	hasAlpha := c.a < 1 && c.a >= 0
	needsAlphaFormat := !formatSet && hasAlpha && (format == "hex" || format == "hex6" || format == "hex3" || format == "hex4" || format == "hex8" || format == "name")
	if needsAlphaFormat {
		if format == "name" && c.a == 0 {
			return "transparent"
		}
		return c.ToRGBString()
	}
	switch format {
	case "rgb":
		return c.ToRGBString()
	case "prgb":
		return c.ToPercentageRGBString()
	case "hex", "hex6":
		return c.ToHexString()
	case "hex3":
		return c.ToHexString(true)
	case "hex4":
		return c.ToHex8String(true)
	case "hex8":
		return c.ToHex8String()
	case "hsl":
		return c.ToHSLString()
	case "hsv":
		return c.ToHSVString()
	case "name":
		if n, ok := c.ToName(); ok {
			return n
		}
	}
	return c.ToHexString()
}

func (c Color) ToFilter(second ...any) string {
	argb := func(x Color) string {
		return hexByte(x.a*255) + hexByte(x.r) + hexByte(x.g) + hexByte(x.b)
	}
	start := "#" + argb(c)
	end := start
	if len(second) > 0 {
		end = "#" + argb(New(second[0]))
	}
	prefix := ""
	if c.gradientType != "" {
		prefix = "GradientType = 1, "
	}
	return "progid:DXImageTransform.Microsoft.gradient(" + prefix + "startColorstr=" + start + ",endColorstr=" + end + ")"
}

func (c Color) Clone() Color { return New(c.ToString()) }

func (c *Color) apply(fn func(Color) Color) *Color {
	next := fn(*c)
	c.r, c.g, c.b, c.a = next.r, next.g, next.b, next.a
	return c
}
func (c *Color) Lighten(amount ...float64) *Color {
	return c.apply(func(x Color) Color { return modifyHSL(x, 0, 0, firstOr(amount, 10)/100) })
}
func (c *Color) Darken(amount ...float64) *Color {
	return c.apply(func(x Color) Color { return modifyHSL(x, 0, 0, -firstOr(amount, 10)/100) })
}
func (c *Color) Saturate(amount ...float64) *Color {
	return c.apply(func(x Color) Color { return modifyHSL(x, 0, firstOr(amount, 10)/100, 0) })
}
func (c *Color) Desaturate(amount ...float64) *Color {
	return c.apply(func(x Color) Color { return modifyHSL(x, 0, -firstOr(amount, 10)/100, 0) })
}
func (c *Color) Greyscale() *Color { return c.Desaturate(100) }
func (c *Color) Brighten(amount ...float64) *Color {
	n := jsRound(-255 * firstOr(amount, 10) / 100)
	c.r, c.g, c.b = clamp(c.r-float64(n), 0, 255), clamp(c.g-float64(n), 0, 255), clamp(c.b-float64(n), 0, 255)
	return c
}
func (c *Color) Spin(amount ...float64) *Color {
	h, s, l := rgbToHSL(c.r, c.g, c.b)
	h = math.Mod(h+firstOr(amount, 0)/360+1, 1)
	c.r, c.g, c.b = hslToRGB(h, s, l)
	return c
}
func firstOr(xs []float64, d float64) float64 {
	if len(xs) > 0 {
		return xs[0]
	}
	return d
}
func modifyHSL(c Color, dh, ds, dl float64) Color {
	h, s, l := rgbToHSL(c.r, c.g, c.b)
	h = math.Mod(h+dh+1, 1)
	s = clamp01(s + ds)
	l = clamp01(l + dl)
	c.r, c.g, c.b = hslToRGB(h, s, l)
	return c
}

func Mix(c1, c2 any, amount ...float64) Color {
	a := firstOr(amount, 50) / 100
	x, y := New(c1), New(c2)
	return Color{r: x.r + (y.r-x.r)*a, g: x.g + (y.g-x.g)*a, b: x.b + (y.b-x.b)*a, a: x.a + (y.a-x.a)*a, valid: x.valid && y.valid, format: "rgb"}
}
func Random() Color {
	return FromRatio(map[string]float64{"r": rand.Float64(), "g": rand.Float64(), "b": rand.Float64()})
}
func Equals(c1, c2 any) bool {
	x, y := New(c1), New(c2)
	return x.r == y.r && x.g == y.g && x.b == y.b && x.a == y.a
}
func Complement(c any) Color {
	x := New(c)
	h, s, l := rgbToHSL(x.r, x.g, x.b)
	x.r, x.g, x.b = hslToRGB(h+0.5, s, l)
	return x
}
func Readability(c1, c2 any) float64 {
	x, y := New(c1), New(c2)
	l1, l2 := x.GetLuminance(), y.GetLuminance()
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}
func IsReadable(c1, c2 any, level, size string) bool {
	ratio := Readability(c1, c2)
	if level == "AAA" && size == "normal" {
		return ratio >= 7
	}
	if (level == "AA" || level == "AAA") && size == "large" {
		return ratio >= 3
	}
	return ratio >= 4.5
}

func MostReadable(base any, candidates []any, args ...map[string]any) Color {
	best := New("black")
	bestScore := 0.0
	for _, candidate := range candidates {
		score := Readability(base, candidate)
		if score > bestScore {
			bestScore = score
			best = New(candidate)
		}
	}
	includeFallback := false
	level, size := "AA", "small"
	if len(args) > 0 {
		if v, ok := args[0]["includeFallbackColors"].(bool); ok {
			includeFallback = v
		}
		if v, ok := args[0]["level"].(string); ok {
			level = v
		}
		if v, ok := args[0]["size"].(string); ok {
			size = v
		}
	}
	if includeFallback && !IsReadable(base, best, level, size) {
		black, white := New("black"), New("white")
		if Readability(base, white) > Readability(base, black) {
			return white
		}
		return black
	}
	return best
}

func Analogous(c any, results, slices int) []Color {
	if results <= 0 {
		results = 6
	}
	if slices <= 0 {
		slices = 30
	}
	x := New(c)
	h, s, l := rgbToHSL(x.r, x.g, x.b)
	part := 1 / float64(slices)
	start := math.Mod(h-part*float64(results/2)+1, 1)
	out := []Color{x}
	for i := 1; i < results; i++ {
		r, g, b := hslToRGB(start+part*float64(i), s, l)
		out = append(out, Color{r: r, g: g, b: b, a: x.a, valid: x.valid, format: "hsl"})
	}
	return out
}
func Monochromatic(c any, results int) []Color {
	if results <= 0 {
		results = 6
	}
	x := New(c)
	h, s, v := rgbToHSV(x.r, x.g, x.b)
	out := make([]Color, 0, results)
	for i := 0; i < results; i++ {
		value := v
		if i > 0 {
			value = math.Mod(v+float64(i)/float64(results), 1)
			// The JavaScript oracle normalizes HSV inputs through bound01,
			// which floors fractional values to four decimal places before
			// converting back to RGB. Preserve that observable parity here.
			value = math.Floor(value*10000) / 10000
		}
		r, g, b := hsvToRGB(h, s, value)
		out = append(out, Color{r: r, g: g, b: b, a: x.a, valid: x.valid, format: "hsv"})
	}
	return out
}
func SplitComplement(c any) []Color {
	x := New(c)
	h, s, l := rgbToHSL(x.r, x.g, x.b)
	out := []Color{x}
	for _, d := range []float64{72, 216} {
		r, g, b := hslToRGB(h+d/360, s, l)
		out = append(out, Color{r: r, g: g, b: b, a: x.a, valid: x.valid, format: "hsl"})
	}
	return out
}
func Triad(c any) []Color  { return Polyad(c, 3) }
func Tetrad(c any) []Color { return Polyad(c, 4) }
func Polyad(c any, n int) []Color {
	x := New(c)
	h, s, l := rgbToHSL(x.r, x.g, x.b)
	out := []Color{x}
	for i := 1; i < n; i++ {
		r, g, b := hslToRGB(h+float64(i)/float64(n), s, l)
		out = append(out, Color{r: r, g: g, b: b, a: x.a, valid: x.valid, format: "hsl"})
	}
	return out
}
