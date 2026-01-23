package main

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"os"
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
}

func defaultParams() Params {
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
	}
}

/* ---------------- SVG Builder ---------------- */

type SVG struct {
	w, h int
	buf  bytes.Buffer
}

func NewSVG(w, h int, bg string) *SVG {
	s := &SVG{w: w, h: h}
	s.buf.WriteString(fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`,
		w, h, w, h,
	))
	s.rect(0, 0, float64(w), float64(h), bg, 1.0)
	return s
}

func (s *SVG) Close() string {
	s.buf.WriteString(`</svg>`)
	return s.buf.String()
}

func (s *SVG) rect(x, y, w, h float64, fill string, op float64) {
	s.buf.WriteString(fmt.Sprintf(
		`<rect x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="%s" opacity="%.3f"/>`,
		x, y, w, h, fill, op,
	))
}

func (s *SVG) circle(cx, cy, r float64, stroke string, sw float64, op float64, fill string) {
	if fill == "" {
		fill = "none"
	}
	s.buf.WriteString(fmt.Sprintf(
		`<circle cx="%.3f" cy="%.3f" r="%.3f" fill="%s" stroke="%s" stroke-width="%.3f" opacity="%.3f"/>`,
		cx, cy, r, fill, stroke, sw, op,
	))
}

func (s *SVG) line(x1, y1, x2, y2 float64, stroke string, sw float64, op float64) {
	s.buf.WriteString(fmt.Sprintf(
		`<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="%s" stroke-width="%.3f" opacity="%.3f"/>`,
		x1, y1, x2, y2, stroke, sw, op,
	))
}

func (s *SVG) path(d string, stroke string, sw float64, op float64, fill string) {
	if fill == "" {
		fill = "none"
	}
	s.buf.WriteString(fmt.Sprintf(
		`<path d="%s" fill="%s" stroke="%s" stroke-width="%.3f" opacity="%.3f"/>`,
		d, fill, stroke, sw, op,
	))
}

/* ---------------- Geometry ---------------- */

type Pt struct{ X, Y float64 }

func add(a, b Pt) Pt         { return Pt{a.X + b.X, a.Y + b.Y} }
func sub(a, b Pt) Pt         { return Pt{a.X - b.X, a.Y - b.Y} }
func mul(a Pt, k float64) Pt { return Pt{a.X * k, a.Y * k} }

func dist(a, b Pt) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Hypot(dx, dy)
}

func polar(r, th float64) Pt {
	return Pt{r * math.Cos(th), r * math.Sin(th)}
}

func rot(p Pt, th float64) Pt {
	c := math.Cos(th)
	s := math.Sin(th)
	return Pt{p.X*c - p.Y*s, p.X*s + p.Y*c}
}

func quantizeAngle(th float64, order int) float64 {
	if order <= 0 {
		return th
	}
	step := 2 * math.Pi / float64(order)
	return math.Round(th/step) * step
}

/* ---------------- Motifs ---------------- */

func randRange(rng *rand.Rand, a, b float64) float64 {
	return a + rng.Float64()*(b-a)
}

func chooseStroke(rng *rand.Rand, p Params) (sw, op float64) {
	sw = randRange(rng, p.StrokeMin, p.StrokeMax)
	op = randRange(rng, p.OpacityMin, p.OpacityMax)
	return
}

func maybeAccent(rng *rand.Rand, p Params) string {
	if rng.Float64() < p.AccentProb {
		return p.AccentColor
	}
	return p.StrokeColor
}

// Flower-of-life-ish: circles on hex lattice within radius.
func layerHexCircleField(svg *SVG, rng *rand.Rand, p Params, center Pt) []Pt {
	var centers []Pt

	// hex basis
	a := Pt{p.CircleGridSpacing, 0}
	b := Pt{p.CircleGridSpacing * 0.5, p.CircleGridSpacing * math.Sin(math.Pi/3)}

	// generate integer lattice coordinates in a bounding box
	maxN := int(math.Ceil(p.BaseR/p.CircleGridSpacing)) + 2
	for i := -maxN; i <= maxN; i++ {
		for j := -maxN; j <= maxN; j++ {
			pt := add(center, add(mul(a, float64(i)), mul(b, float64(j))))
			if dist(pt, center) <= p.BaseR {
				centers = append(centers, pt)
				sw, op := chooseStroke(rng, p)
				svg.circle(pt.X, pt.Y, p.CircleRadius, maybeAccent(rng, p), sw, op, "")
			}
		}
	}
	return centers
}

// Simple "Metatron-like" network: connect to K nearest neighbors.
func layerNearestNetwork(svg *SVG, rng *rand.Rand, p Params, pts []Pt) {
	if p.MetatronK <= 0 || len(pts) < 2 {
		return
	}
	for i := range pts {
		type nb struct {
			j int
			d float64
		}
		var nbs []nb
		for j := range pts {
			if i == j {
				continue
			}
			nbs = append(nbs, nb{j: j, d: dist(pts[i], pts[j])})
		}
		// partial select K smallest (naive sort ok for modest sizes)
		for a := 0; a < len(nbs); a++ {
			for b := a + 1; b < len(nbs); b++ {
				if nbs[b].d < nbs[a].d {
					nbs[a], nbs[b] = nbs[b], nbs[a]
				}
			}
		}
		k := p.MetatronK
		if k > len(nbs) {
			k = len(nbs)
		}
		for t := 0; t < k; t++ {
			j := nbs[t].j
			sw, op := chooseStroke(rng, p)
			svg.line(pts[i].X, pts[i].Y, pts[j].X, pts[j].Y, p.StrokeColor, sw, op*0.8)
		}
	}
}

// Rosette using a polar radius function, rendered as a path.
func layerRosette(svg *SVG, rng *rand.Rand, p Params, center Pt) {
	n := 720 // smoothness
	lobes := float64(p.RosetteCount)
	jit := p.RosetteJ

	var d bytes.Buffer
	for t := 0; t <= n; t++ {
		u := float64(t) / float64(n)
		th := u * 2 * math.Pi

		// quantize for "sacred" feel
		thq := quantizeAngle(th, p.RotOrder)

		// radius modulation + jitter
		r := p.RosetteR0 + (p.RosetteR1-p.RosetteR0)*(0.5+0.5*math.Cos(lobes*thq))
		r *= (1.0 + randRange(rng, -jit, jit))

		pt := add(center, polar(r, thq))
		if t == 0 {
			d.WriteString(fmt.Sprintf("M %.3f %.3f ", pt.X, pt.Y))
		} else {
			d.WriteString(fmt.Sprintf("L %.3f %.3f ", pt.X, pt.Y))
		}
	}
	d.WriteString("Z")

	sw, op := chooseStroke(rng, p)
	svg.path(d.String(), maybeAccent(rng, p), sw, op, "")
}

/* ---------------- Symmetry wrapper ---------------- */

func withSymmetry(rng *rand.Rand, p Params, center Pt, draw func(theta float64, mirror bool)) {
	order := p.RotOrder
	if order <= 0 {
		draw(0, false)
		return
	}
	step := 2 * math.Pi / float64(order)
	for k := 0; k < order; k++ {
		th := float64(k) * step
		draw(th, false)
		if p.Mirror {
			draw(th, true)
		}
	}
}

/* ---------------- Main composition ---------------- */

func main() {
	p := defaultParams()

	// Optional: allow overriding seed via env var for quick testing
	if v := os.Getenv("SEED"); v != "" {
		var s int64
		_, _ = fmt.Sscanf(v, "%d", &s)
		if s != 0 {
			p.Seed = s
		}
	}

	rng := rand.New(rand.NewSource(p.Seed))

	// center with slight jitter
	cx := float64(p.Width)/2 + randRange(rng, -p.CenterJ, p.CenterJ)*float64(p.Width)
	cy := float64(p.Height)/2 + randRange(rng, -p.CenterJ, p.CenterJ)*float64(p.Height)
	center := Pt{cx, cy}

	svg := NewSVG(p.Width, p.Height, p.BgColor)

	// Big boundary rings
	for i := 1; i <= p.RingCount; i++ {
		r := float64(i) * p.RingSpacing
		sw, op := chooseStroke(rng, p)
		svg.circle(center.X, center.Y, r, p.StrokeColor, sw, op*0.8, "")
	}

	// Symmetry: rotate motifs around center for coherence.
	var allCenters []Pt
	withSymmetry(rng, p, center, func(theta float64, mirror bool) {
		// local transform around center
		localCenter := center

		// generate hex field, then rotate its points
		pts := layerHexCircleField(svg, rng, p, localCenter)

		// rotate points for symmetry instance
		for _, pt := range pts {
			v := sub(pt, localCenter)
			if mirror {
				v.Y = -v.Y
			}
			v = rot(v, theta)
			allCenters = append(allCenters, add(localCenter, v))
		}

		// rosette per symmetry instance (light)
		layerRosette(svg, rng, p, localCenter)
	})

	// Network layer (optional)
	if p.DrawMetatron {
		layerNearestNetwork(svg, rng, p, allCenters)
	}

	out := svg.Close()
	_ = os.WriteFile("sacred.svg", []byte(out), 0644)

	fmt.Printf("Wrote sacred.svg (seed=%d)\n", p.Seed)
}
