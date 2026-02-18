package main

import (
	"GoSacred/lib"
	"GoSacred/types"
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"os"
)

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

// radialFalloff returns 1.0 at center (d=0) and 0.0 at edge (d>=maxR).
func radialFalloff(d, maxR float64) float64 {
	t := d / maxR
	if t > 1.0 {
		t = 1.0
	}
	if t < 0.0 {
		t = 0.0
	}
	return 1.0 - t
}

// Flower-of-life-ish: circles on hex lattice within radius.
func layerHexCircleField(svg *lib.SVG, rng *rand.Rand, p types.Params, pal types.Palette, center Pt) []Pt {
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
				// Radial falloff: bold at center, fading at edges
				f := radialFalloff(dist(pt, center), p.BaseR)
				// Grid dropout: more likely to skip circles near edges
				if rng.Float64() < p.GridDropout*(1.0-f*f) {
					continue
				}
				centers = append(centers, pt)
				sw, op := chooseStroke(rng, p)
				sw *= 0.3 + 0.7*f
				op *= 0.2 + 0.8*f*f
				svg.Circle(pt.X, pt.Y, p.CircleRadius, types.PickStrokeOrAccent(rng, p, pal), sw, op, "")
			}
		}
	}
	return centers
}

// Simple "Metatron-like" network: connect to K nearest neighbors.
func layerNearestNetwork(svg *lib.SVG, rng *rand.Rand, p types.Params, pts []Pt, center Pt) {
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
			// Radial falloff based on line midpoint
			mid := Pt{(pts[i].X + pts[j].X) / 2, (pts[i].Y + pts[j].Y) / 2}
			f := radialFalloff(dist(mid, center), p.BaseR*1.2)
			sw *= 0.3 + 0.7*f
			op *= 0.2 + 0.8*f*f
			svg.Line(pts[i].X, pts[i].Y, pts[j].X, pts[j].Y, p.StrokeColor, sw, op*0.8)
		}
	}
}

// Rosette using a polar radius function, rendered as a path.
func layerRosette(svg *lib.SVG, rng *rand.Rand, p types.Params, pal types.Palette, center Pt) {
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
		r *= (1.0 + lib.RandRange(rng, -jit, jit))

		pt := add(center, polar(r, thq))
		if t == 0 {
			d.WriteString(fmt.Sprintf("M %.3f %.3f ", pt.X, pt.Y))
		} else {
			d.WriteString(fmt.Sprintf("L %.3f %.3f ", pt.X, pt.Y))
		}
	}
	d.WriteString("Z")

	sw, op := chooseStroke(rng, p)
	svg.Path(d.String(), types.PickStrokeOrAccent(rng, p, pal), sw, op, "")
}

func layerRosetteWithPalette(svg *lib.SVG, rng *rand.Rand, p types.Params, pal types.Palette, center Pt) {
	n := 720 // smoothness; higher = smoother path

	lobes := float64(p.RosetteCount)
	jit := p.RosetteJ

	var d bytes.Buffer
	for t := 0; t <= n; t++ {
		u := float64(t) / float64(n)
		th := u * 2 * math.Pi

		// quantize angle to reinforce "sacred" symmetry
		thq := quantizeAngle(th, p.RotOrder)

		// rosette radius modulation (cosine lobes) + jitter
		// r ranges between RosetteR0 and RosetteR1
		r := p.RosetteR0 + (p.RosetteR1-p.RosetteR0)*(0.5+0.5*math.Cos(lobes*thq))
		r *= (1.0 + lib.RandRange(rng, -jit, jit))

		pt := add(center, polar(r, thq))

		if t == 0 {
			d.WriteString(fmt.Sprintf("M %.3f %.3f ", pt.X, pt.Y))
		} else {
			d.WriteString(fmt.Sprintf("L %.3f %.3f ", pt.X, pt.Y))
		}
	}
	d.WriteString("Z")

	// style
	sw, op := chooseStroke(rng, p)
	color := types.PickStrokeOrAccent(rng, p, pal)

	// draw
	svg.Path(d.String(), color, sw, op, "")
}

func chooseStroke(rng *rand.Rand, p types.Params) (sw, op float64) {
	sw = lib.RandRange(rng, p.StrokeMin, p.StrokeMax)
	op = lib.RandRange(rng, p.OpacityMin, p.OpacityMax)
	return
}

// layerFilledTranslucentShapes draws filled triangles and hexagons at hex
// lattice positions using accent colors at very low opacity (0.02-0.08).
// Overlapping translucent fills create emergent color mixing.
func layerFilledTranslucentShapes(svg *lib.SVG, rng *rand.Rand, p types.Params, pal types.Palette, center Pt) {
	if len(pal.Accents) == 0 {
		return
	}

	// hex basis vectors (same lattice as circle field)
	a := Pt{p.CircleGridSpacing, 0}
	b := Pt{p.CircleGridSpacing * 0.5, p.CircleGridSpacing * math.Sin(math.Pi/3)}

	maxN := int(math.Ceil(p.BaseR/p.CircleGridSpacing)) + 2

	// shape size varies slightly per run for organic feel
	baseSize := p.CircleGridSpacing * lib.RandRange(rng, 0.55, 0.80)

	for i := -maxN; i <= maxN; i++ {
		for j := -maxN; j <= maxN; j++ {
			pt := add(center, add(mul(a, float64(i)), mul(b, float64(j))))
			if dist(pt, center) > p.BaseR*1.05 {
				continue
			}

			// density dropout
			if rng.Float64() < p.FilledShapeDropout {
				continue
			}

			// radial falloff: shapes near center are slightly more opaque
			f := radialFalloff(dist(pt, center), p.BaseR)
			op := lib.RandRange(rng, 0.02, 0.08) * (0.5 + 0.5*f)

			color := pal.Accents[rng.Intn(len(pal.Accents))]

			// small per-shape size jitter
			size := baseSize * lib.RandRange(rng, 0.85, 1.15)

			var pathStr string
			if rng.Float64() < 0.5 {
				// regular hexagon
				pathStr = hexagonPath(pt, size)
			} else {
				// equilateral triangle, alternating orientation by lattice parity
				flip := (i+j)%2 == 0
				pathStr = trianglePath(pt, size, flip)
			}

			// fill only, no stroke
			svg.Path(pathStr, "none", 0, op, color)
		}
	}
}

func hexagonPath(center Pt, size float64) string {
	var d bytes.Buffer
	for k := 0; k < 6; k++ {
		angle := float64(k)*math.Pi/3 + math.Pi/6
		px := center.X + size*math.Cos(angle)
		py := center.Y + size*math.Sin(angle)
		if k == 0 {
			d.WriteString(fmt.Sprintf("M %.3f %.3f", px, py))
		} else {
			d.WriteString(fmt.Sprintf(" L %.3f %.3f", px, py))
		}
	}
	d.WriteString(" Z")
	return d.String()
}

func trianglePath(center Pt, size float64, flip bool) string {
	var d bytes.Buffer
	for k := 0; k < 3; k++ {
		angle := float64(k)*2*math.Pi/3 - math.Pi/2
		if flip {
			angle += math.Pi
		}
		px := center.X + size*math.Cos(angle)
		py := center.Y + size*math.Sin(angle)
		if k == 0 {
			d.WriteString(fmt.Sprintf("M %.3f %.3f", px, py))
		} else {
			d.WriteString(fmt.Sprintf(" L %.3f %.3f", px, py))
		}
	}
	d.WriteString(" Z")
	return d.String()
}

/* ---------------- Symmetry wrapper ---------------- */

func withSymmetry(rng *rand.Rand, p types.Params, pal types.Palette, center Pt, draw func(theta float64, mirror bool)) {
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

/* ---------------- Depth 3d ---------------- */
type Pt3 struct{ X, Y, Z float64 }

func projectPerspective(p Pt3, camZ, fov float64) Pt {
	// Simple pinhole: assumes camera looking toward -Z, centered at origin.
	// camZ > 0. Larger camZ moves camera "back".
	z := camZ - p.Z
	if z < 1e-3 {
		z = 1e-3
	}
	scale := fov / z
	return Pt{p.X * scale, p.Y * scale}
}

func depthStyle(z float64, zMin, zMax float64, baseSW, baseOp float64) (sw, op float64) {
	// Normalize z to 0..1
	t := 0.0
	if zMax > zMin {
		t = (z - zMin) / (zMax - zMin)
	}
	// Near = higher t
	sw = baseSW * (0.7 + 0.9*t)
	op = baseOp * (0.25 + 0.85*t)
	return
}

/* ---------------- Main composition ---------------- */
func main() {
	//p := types.DefaultParams()
	p := types.RandomParams()

	// Optional: allow overriding seed via env var for quick testing
	if v := os.Getenv("SEED"); v != "" {
		var s int64
		_, _ = fmt.Sscanf(v, "%d", &s)
		if s != 0 {
			p.Seed = s
		}
	}

	rng := rand.New(rand.NewSource(p.Seed))

	// 1) Generate palette (driven by seed via rng)
	pp := types.PaletteParams{
		Mode:    lib.ChooseRandom(rng, []string{"mono", "analogous", "complementary", "triad"}),
		HueBase: -1, // <0 => random hue
		HueJit:  lib.RandRange(rng, 4, 18),
		SatMin:  0.15,
		SatMax:  0.95,
		LumMin:  0.15,
		LumMax:  0.92,
	}
	pal := types.GenPalette(rng, pp)

	// 2) Set Params colors from palette (so Params remains “authoritative”)
	p.BgColor = pal.Bg1
	p.StrokeColor = types.PickStroke(rng, pal) // a representative default stroke
	if len(pal.Accents) > 0 {
		p.AccentColor = pal.Accents[0]
	}
	p.AccentProb = lib.RandRange(rng, 0.04, 0.5)

	// 3) Create SVG using the palette background using the gradient background:
	svg := lib.NewSVG(p.Width, p.Height, pal.Bg1, pal.Bg2)

	// Example: rings
	center := Pt{float64(p.Width) / 2, float64(p.Height) / 2}
	maxR := math.Max(p.BaseR, float64(p.RingCount)*p.RingSpacing) * 1.2

	// Glow pass: soft blurred halo behind main geometry
	svg.GroupOpen(`filter="url(#glow)" opacity="0.3"`)
	for i := 1; i <= p.RingCount; i++ {
		r := float64(i) * p.RingSpacing
		f := radialFalloff(r, maxR)
		color := types.PickStroke(rng, pal)
		svg.Circle(center.X, center.Y, r, color, p.StrokeMax*2.5, 0.5*f, "")
	}
	svg.GroupClose()

	// Filled translucent shapes layer painted colour wash beneath wireframes
	layerFilledTranslucentShapes(svg, rng, p, pal, center)

	// Crisp rings with radial falloff
	for i := 1; i <= p.RingCount; i++ {
		r := float64(i) * p.RingSpacing
		f := radialFalloff(r, maxR)
		sw, op := chooseStroke(rng, p)
		color := types.PickStroke(rng, pal)
		svg.Circle(center.X, center.Y, r, color, sw*(0.3+0.7*f), op*0.8*(0.2+0.8*f*f), "")
	}

	// Example: rosette stroke:
	layerRosetteWithPalette(svg, rng, p, pal, center)
	layerRosetteWithPalette(svg, rng, p, pal, center)
	layerRosetteWithPalette(svg, rng, p, pal, center)

	// Big boundary rings with radial falloff
	for i := 1; i <= p.RingCount; i++ {
		r := float64(i) * p.RingSpacing
		f := radialFalloff(r, maxR)
		sw, op := chooseStroke(rng, p)
		svg.Circle(center.X, center.Y, r, p.StrokeColor, sw*(0.3+0.7*f), op*0.8*(0.2+0.8*f*f), "")
	}

	// Symmetry: rotate motifs around center for coherence.
	var allCenters []Pt
	withSymmetry(rng, p, pal, center, func(theta float64, mirror bool) {
		// local transform around center
		localCenter := center
		// generate hex field, then rotate its points
		pts := layerHexCircleField(svg, rng, p, pal, localCenter)
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
		layerRosette(svg, rng, p, pal, localCenter)
	})

	// Network layer (optional)
	if p.DrawMetatron {
		layerNearestNetwork(svg, rng, p, allCenters, center)
	}

	// output
	out := svg.Close()
	_ = os.WriteFile("sacred.svg", []byte(out), 0644)
	fmt.Printf("Wrote sacred.svg (seed=%d)\n", p.Seed)
}
