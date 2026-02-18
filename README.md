# GoSacred

A generative art engine written in Go that produces sacred geometry compositions as SVG files. Each run creates a unique, seed-reproducible artwork combining ancient geometric patterns with parametric randomization and color theory.

## What Is Sacred Geometry?

Sacred geometry is the study of geometric patterns, shapes, and proportions that recur throughout nature, art, and architecture across cultures and millennia. Practitioners and artists observe that certain forms — the spiral of a nautilus shell, the hexagonal lattice of a honeycomb, the branching of a river delta — share deep mathematical relationships that feel inherently harmonious.

At its core, sacred geometry explores a small set of fundamental constructs:

- **The Flower of Life** — overlapping circles arranged on a hexagonal lattice, found carved into temples from ancient Egypt to medieval Europe. It encodes the geometry of close-packing spheres and the relationships between the Platonic solids.
- **Metatron's Cube** — a figure derived from the Flower of Life by connecting every circle's center to its neighbors, revealing the hidden edges of all five Platonic solids within a flat drawing.
- **The Golden Ratio (phi)** — the proportion approximately equal to 1.618, appearing in spiral phyllotaxis (sunflower seed heads), the proportions of the Parthenon, and Renaissance painting compositions.
- **Rotational Symmetry** — the property that a figure looks the same after rotation by some fraction of a full turn. Mandalas, rose windows, and snowflakes all exhibit this, with orders of 3, 4, 6, 8, or 12 being the most common in human art and crystallography alike.

Geometric art takes these principles and uses them as a generative vocabulary. Rather than drawing representational subjects, the artist composes with circles, polygons, lattices, spirals, and symmetry operations — letting mathematical relationships produce visual harmony. The results can feel meditative, architectural, organic, or cosmic depending on palette, density, and layering choices.

GoSacred sits in this tradition: it uses hexagonal lattices, nearest-neighbor networks, rosette curves, concentric rings, and rotational symmetry as its building blocks, then layers randomization on top so that every output is a one-of-a-kind composition that still obeys the underlying geometric harmonies.

## How It Works

The generator follows a layered composition pipeline:

1. **Parameter generation** — either default or fully randomized, controlling canvas size, symmetry order, spacing, stroke ranges, and color mode. Sacred ratios (1, sqrt(2), sqrt(3), phi) are applied to scale-sensitive values.
2. **Palette generation** — an HSL-based system produces background gradients, 8 stroke colors, and 4 accent colors in one of several harmony modes (mono, complementary, triad, analogous).
3. **Layer rendering** — each layer draws into a shared SVG buffer in painter's-algorithm order:
   - Gradient background
   - Concentric rings
   - Rosette curves (lobed radial flowers)
   - Hex circle field (Flower of Life pattern, applied with rotational symmetry)
   - Nearest-neighbor network (Metatron's Cube connections)
4. **SVG output** — the final buffer is written to `sacred.svg`.

Every numeric decision (stroke width, opacity, color selection, jitter) is driven by a seeded RNG, so passing the same seed reproduces the same artwork exactly.

## Running

```bash
go run .               # generate with a random seed
SEED=42 go run .       # generate with a specific seed for reproducibility
```

Output is written to `sacred.svg` in the project root.

## Project Structure

```
.
├── main.go                 # composition pipeline and geometry primitives
├── lib/
│   ├── svg.go              # SVG element builder (circle, line, path, rect)
│   └── common_helpers.go   # RandRange, ChooseRandom utilities
├── types/
│   ├── params.go           # Params struct, DefaultParams, RandomParams
│   └── palette.go          # HSL palette generation and color picking
└── sacred.svg              # latest generated output
```

## Feature Development Roadmap

The current output is functional but tends toward flat, uniform compositions. The following enhancements are planned to add depth, texture, and compositional variety:

1. `DONE` Radial Depth Falloff - Make opacity and stroke width decay as a function of distance from center. Elements near the core render bold and opaque; elements at the periphery become gossamer-thin and ghostly. Quadratic falloff for more dramatic depth.
2. `DONE` Glow Filter Layer - Use SVG `<filter>` elements with Gaussian blur to create a soft glow behind key geometry. Duplicate core elements into a blurred group at reduced opacity, then draw the crisp layer on top for a luminous, neon-like atmosphere.
3. `DONE`Random Grid Dropout - Instead of drawing every circle in the hex field, randomly skip 20-40% of them with spatially varying probability. Break the wallpaper-repeat uniformity and create organic density variation across the composition.
4. `DONE` Filled Translucent Shapes - Add a layer of filled triangles or hexagons at very low opacity (0.02-0.08) using accent colors. Overlapping translucent fills create emergent color mixing and give the composition a painted, layered feel.
5. Dot Scatter - Place small filled circles along hex grid points and rosette paths with radius varying by distance from center. Adds fine-grained texture and breaks up the purely wireframe aesthetic.
6. Spiral Layer - Introduce logarithmic spirals based on the golden ratio. Draw 2-6 spirals rotated by the current symmetry order with decreasing opacity along the path for a dissolving, organic effect.
7. Stroke Dash Patterns - Apply SVG `stroke-dasharray` patterns to select network lines and ring segments. Varied dash rhythms (long-short-dot-short) make line work feel more hand-drawn and less mechanical.
8. 3D Perspective Projection - Wire up the existing `projectPerspective()` and `depthStyle()` functions that are currently unused. Assign Z values to hex grid points based on sine waves or distance functions, then project to 2D with depth-mapped stroke and opacity for dome or bowl effects.
9. Multiple Focal Points - Generate 2-3 overlapping mandalas at different positions and scales — a large primary composition with smaller satellite compositions. Shared palette, independent symmetry orders.
10. Rosette Variation - Make each rosette layer dramatically different rather than similarly parameterized. Vary lobe count independently, use contrasting inner/outer radii, and give each its own opacity treatment — one nearly invisible and enormous, another small and bold.