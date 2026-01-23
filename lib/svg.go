package lib

import (
	"bytes"
	"fmt"
)

/* ---------------- SVG Builder ---------------- */

type SVG struct {
	w, h int
	buf  bytes.Buffer
}

func NewSVG(w, h int, bg1, bg2 string) *SVG {
	s := &SVG{w: w, h: h}
	s.buf.WriteString(fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`,
		w, h, w, h,
	))

	// defs: background gradient
	s.buf.WriteString(`<defs>`)
	s.buf.WriteString(`<linearGradient id="bg" x1="0" y1="0" x2="0" y2="1">`)
	s.buf.WriteString(fmt.Sprintf(`<stop offset="0%%" stop-color="%s"/>`, bg1))
	s.buf.WriteString(fmt.Sprintf(`<stop offset="100%%" stop-color="%s"/>`, bg2))
	s.buf.WriteString(`</linearGradient>`)
	s.buf.WriteString(`</defs>`)

	s.Rect(0, 0, float64(w), float64(h), "url(#bg)", 1.0)
	return s
}

func (s *SVG) Close() string {
	s.buf.WriteString(`</svg>`)
	return s.buf.String()
}

func (s *SVG) Rect(x, y, w, h float64, fill string, op float64) {
	s.buf.WriteString(fmt.Sprintf(
		`<rect x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="%s" opacity="%.3f"/>`,
		x, y, w, h, fill, op,
	))
}

func (s *SVG) Circle(cx, cy, r float64, stroke string, sw float64, op float64, fill string) {
	if fill == "" {
		fill = "none"
	}
	s.buf.WriteString(fmt.Sprintf(
		`<circle cx="%.3f" cy="%.3f" r="%.3f" fill="%s" stroke="%s" stroke-width="%.3f" opacity="%.3f"/>`,
		cx, cy, r, fill, stroke, sw, op,
	))
}

func (s *SVG) Line(x1, y1, x2, y2 float64, stroke string, sw float64, op float64) {
	s.buf.WriteString(fmt.Sprintf(
		`<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="%s" stroke-width="%.3f" opacity="%.3f"/>`,
		x1, y1, x2, y2, stroke, sw, op,
	))
}

func (s *SVG) Path(d string, stroke string, sw float64, op float64, fill string) {
	if fill == "" {
		fill = "none"
	}
	s.buf.WriteString(fmt.Sprintf(
		`<path d="%s" fill="%s" stroke="%s" stroke-width="%.3f" opacity="%.3f"/>`,
		d, fill, stroke, sw, op,
	))
}
