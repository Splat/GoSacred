package types

import (
	"math"
	"math/rand"
	"time"
)

type Params struct {
	Seed int64

	Width, Height int
	Margin        float64

	RotOrder int     // rotational symmetry order, e.g. 6, 8, 12
	Mirror   bool    // optional reflection symmetry
	CenterJ  float64 // 0..1 jitter of center

	BaseR       float64 // base radius for main system
	RingCount   int
	RingSpacing float64

	CircleGridSpacing float64 // spacing for flower-of-life lattice
	CircleRadius      float64 // radius for each lattice circle

	DrawMetatron bool
	MetatronK    int // connect each node to K nearest nodes (simple approximation)

	RosetteCount int     // number of rosette lobes
	RosetteR0    float64 // inner radius
	RosetteR1    float64 // outer radius
	RosetteJ     float64 // jitter

	StrokeMin  float64
	StrokeMax  float64
	OpacityMin float64
	OpacityMax float64

	BgColor     string
	StrokeColor string
	AccentColor string
	AccentProb  float64

	GridDropout float64 // probability of skipping hex grid circles (0-0.45)

	FilledShapeDropout float64 // probability of skipping a filled translucent shape (0-1)
}

func DefaultParams() Params {
	return Params{
		Seed: time.Now().UnixNano(),

		Width:  1200,
		Height: 1200,
		Margin: 40,

		RotOrder: 12,
		Mirror:   false,
		CenterJ:  0.02,

		BaseR:       520,
		RingCount:   7,
		RingSpacing: 60,

		CircleGridSpacing: 85,
		CircleRadius:      55,

		DrawMetatron: true,
		MetatronK:    3,

		RosetteCount: 24,
		RosetteR0:    140,
		RosetteR1:    440,
		RosetteJ:     0.02,

		StrokeMin:  0.7,
		StrokeMax:  2.2,
		OpacityMin: 0.10,
		OpacityMax: 0.65,

		BgColor:     "#0b0b10",
		StrokeColor: "#f2f2f2",
		AccentColor: "#c7a86b",
		AccentProb:  0.08,

		GridDropout: 0.25,

		FilledShapeDropout: 0.40,
	}
}

/*
RandomParams generates random parameters for the drawing outside specified
size constraints. Recommend 1200, 1200, 40 for messing around
*/
func RandomParams() Params {
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))

	size := choose(r, canvasSizes)
	sym := choose(r, symmetryOrders)

	baseR := randFloat(r, float64(size)*0.30, float64(size)*0.48)
	ratio := choose(r, sacredRatios)

	ringSpacing := randFloat(r, 40, 90) * ratio
	ringCount := randInt(r, 4, 9)

	gridSpacing := randFloat(r, 60, 110) * ratio
	circleRadius := gridSpacing * randFloat(r, 0.45, 0.65)

	return Params{
		Seed: seed,

		Width:  size,
		Height: size,
		Margin: randFloat(r, 20, 60),

		RotOrder: sym,
		Mirror:   chance(r, 0.35),
		CenterJ:  randFloat(r, 0.0, 0.035),

		BaseR:       baseR,
		RingCount:   ringCount,
		RingSpacing: ringSpacing,

		CircleGridSpacing: gridSpacing,
		CircleRadius:      circleRadius,

		DrawMetatron: chance(r, 0.65),
		MetatronK:    randInt(r, 2, 4),

		RosetteCount: sym * randInt(r, 1, 3),
		RosetteR0:    baseR * randFloat(r, 0.20, 0.35),
		RosetteR1:    baseR * randFloat(r, 0.75, 1.05),
		RosetteJ:     randFloat(r, 0.0, 0.03),

		StrokeMin: randFloat(r, 0.5, 0.9),
		StrokeMax: randFloat(r, 1.8, 3.0),

		OpacityMin: randFloat(r, 0.05, 0.15),
		OpacityMax: randFloat(r, 0.45, 0.85),

		// These will usually be overridden by palette generation,
		// but safe defaults matter
		// TODO: leverage palette modes
		BgColor:     "#0b0b10",
		StrokeColor: "#eaeaea",
		AccentColor: "#c7a86b",
		AccentProb:  randFloat(r, 0.04, 0.12),

		GridDropout: randFloat(r, 0.15, 0.45),

		FilledShapeDropout: randFloat(r, 0.25, 0.60),
	}
}

func randFloat(r *rand.Rand, min, max float64) float64 {
	return min + r.Float64()*(max-min)
}

func randInt(r *rand.Rand, min, max int) int {
	return min + r.Intn(max-min+1)
}

func choose[T any](r *rand.Rand, items []T) T {
	return items[r.Intn(len(items))]
}

func chance(r *rand.Rand, p float64) bool {
	return r.Float64() < p
}

var symmetryOrders = []int{3, 4, 6, 8, 10, 12, 16}
var canvasSizes = []int{900, 1000, 1200, 1600, 2000}

var paletteModes = []string{
	"mono",
	"analogous",
	"complementary",
	"triad",
}

var sacredRatios = []float64{
	1.0,
	math.Sqrt2,
	math.Sqrt(3),
	(1 + math.Sqrt(5)) / 2, // φ
}
