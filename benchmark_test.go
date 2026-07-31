package tinycolor

import "testing"

func BenchmarkNewHex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New("#6699cc").ToHex8()
	}
}

func BenchmarkParseAndConvert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := New("hsla(210, 50%, 60%, 0.75)")
		_ = c.ToRGBString()
		_ = c.ToHSLString()
		_ = c.ToHSVString()
	}
}
