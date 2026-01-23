package types

import (
	"GoSacred/lib"
	"fmt"
	"math"
	"math/rand"
)

type Palette struct {
	Bg1, Bg2 string
	Strokes  []string
	Accents  []string
}

type PaletteParams struct {
	Mode    string  // "analogous", "complementary", "triad", "mono", "random"
	HueBase float64 // 0..360; if <0 => random
	HueJit  float64 // degrees
	SatMin  float64 // 0..1
	SatMax  float64
	LumMin  float64 // 0..1
	LumMax  float64
}

func GenPalette(rng *rand.Rand, pp PaletteParams) Palette {
	h := pp.HueBase
	if h < 0 {
		h = rng.Float64() * 360
	}

	var offsets []float64
	switch pp.Mode {
	case "mono":
		offsets = []float64{0}
	case "complementary":
		offsets = []float64{0, 180}
	case "triad":
		offsets = []float64{0, 120, 240}
	case "analogous":
		offsets = []float64{0, 20, -20, 40, -40}
	default: // "random"
		offsets = []float64{0, rng.Float64() * 360}
	}

	// Background: dark gradient variant of base hue
	bgH := math.Mod(h+lib.RandRange(rng, -pp.HueJit, pp.HueJit)+360, 360)
	bg1 := hslHex(bgH, 0.30, 0.06)
	bg2 := hslHex(math.Mod(bgH+10, 360), 0.35, 0.10)

	// Strokes/accents
	strokes := make([]string, 0, 8)
	accents := make([]string, 0, 4)

	for i := 0; i < 8; i++ {
		ho := offsets[rng.Intn(len(offsets))]
		hh := math.Mod(h+ho+lib.RandRange(rng, -pp.HueJit, pp.HueJit)+360, 360)
		s := lib.RandRange(rng, pp.SatMin, pp.SatMax)
		l := lib.RandRange(rng, pp.LumMin, pp.LumMax)
		strokes = append(strokes, hslHex(hh, s, l))
	}
	for i := 0; i < 4; i++ {
		ho := offsets[rng.Intn(len(offsets))]
		hh := math.Mod(h+ho+lib.RandRange(rng, -pp.HueJit, pp.HueJit)+360, 360)
		accents = append(accents, hslHex(hh, 0.75, 0.60))
	}

	return Palette{Bg1: bg1, Bg2: bg2, Strokes: strokes, Accents: accents}
}

// pick one of the stroke colors
func PickStroke(rng *rand.Rand, pal Palette) string {
	if len(pal.Strokes) == 0 {
		return "#eaeaea"
	}
	return pal.Strokes[rng.Intn(len(pal.Strokes))]
}

// sometimes pick an accent, otherwise stroke
func PickStrokeOrAccent(rng *rand.Rand, p Params, pal Palette) string {
	if rng.Float64() < p.AccentProb && len(pal.Accents) > 0 {
		return pal.Accents[rng.Intn(len(pal.Accents))]
	}
	return PickStroke(rng, pal)
}

// HSL -> hex (minimal, accurate enough for art)
func hslHex(h, s, l float64) string {
	r, g, b := hslToRGB(h/360.0, s, l)
	return fmt.Sprintf("#%02x%02x%02x", int(r*255+0.5), int(g*255+0.5), int(b*255+0.5))
}

func hslToRGB(h, s, l float64) (float64, float64, float64) {
	var r, g, b float64
	if s == 0 {
		return l, l, l
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	r = hueToRGB(p, q, h+1.0/3.0)
	g = hueToRGB(p, q, h)
	b = hueToRGB(p, q, h-1.0/3.0)
	return r, g, b
}
func hueToRGB(p, q, t float64) float64 {
	for t < 0 {
		t += 1
	}
	for t > 1 {
		t -= 1
	}
	switch {
	case t < 1.0/6.0:
		return p + (q-p)*6*t
	case t < 1.0/2.0:
		return q
	case t < 2.0/3.0:
		return p + (q-p)*(2.0/3.0-t)*6
	default:
		return p
	}
}
