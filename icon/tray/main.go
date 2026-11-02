// Command tray draws internal/app/tray/tray-template.png, the app's icon in
// the macOS menu bar: the star and leaves of the app's icon, black on
// transparent for macOS to tint, with the gaps between them left clear. It
// draws from polygons, the curves cut into short lines, supersampled, at 2×
// of the menu bar's 18 points.
//
// Run it from the repository root:
//
//	go run ./icon/tray
//
// With -preview dir, it also writes the icon over light and dark menu bars,
// enlarged, to check by eye.
package main

import (
	"flag"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

type pt struct{ x, y float64 }

// The glyph: the star and leaves as icon/main.go draws them, on Icon
// Composer's 1024-point canvas, before it is fit to the square.
var (
	star   = struct{ cx, cy, rx, ry, lean float64 }{497, 503, 240.6, 259.5, -3}
	leaves = []string{
		"M518 246C565 199 684 179 764 204C823 224 845 278 845 338L845 477C808 432 674 423 599 378C550 348 525 298 518 246Z",
		"M488 773C445 820 336 840 276 815C232 795 207 746 207 686L207 532C239 577 326 592 396 621C450 646 480 696 488 773Z",
	}
	// Knocked out of the leaves around the star, as the gaps between them
	// are too thin to show at the menu bar's size.
	gap = 45.0

	// The glyph's longer side, centred in the 18-point square.
	fit = 15.6
)

// steps is the number of lines each curve is cut into.
const steps = 24

// shapes are the star and the leaves as polygons.
var shapes = append([][]pt{starShape()}, leafShapes()...)

func cubic(a, b, c, d pt, t float64) pt {
	u := 1 - t
	w := [4]float64{u * u * u, 3 * u * u * t, 3 * u * t * t, t * t * t}
	return pt{w[0]*a.x + w[1]*b.x + w[2]*c.x + w[3]*d.x, w[0]*a.y + w[1]*b.y + w[2]*c.y + w[3]*d.y}
}

// starShape is the star of icon/main.go's starSVG: each side bows in
// toward the middle, and the vertical points lean right.
func starShape() []pt {
	a, b := star.rx, star.ry
	points := []pt{{-a, 0}, {0, -b}, {a, 0}, {0, b}}
	lean := math.Tan(star.lean * math.Pi / 180)
	var poly []pt
	for i, p := range points {
		q := points[(i+1)%4]
		x, y := p.x+q.x, p.y+q.y
		c1, c2 := pt{0.5 * x, 0.06 * y}, pt{0.04 * x, 0.45 * y}
		if p.x == 0 {
			c1, c2 = c2, c1
		}
		for s := range steps {
			v := cubic(p, c1, c2, q, float64(s)/steps)
			poly = append(poly, pt{star.cx + v.x + v.y*lean, star.cy + v.y})
		}
	}
	return poly
}

// leafShapes reads the leaves' paths, which use only M, C, L and Z.
func leafShapes() [][]pt {
	tok := regexp.MustCompile(`[MCLZ]|-?[\d.]+`)
	var polys [][]pt
	for _, d := range leaves {
		var poly []pt
		var cmd string
		var v []float64
		for _, t := range tok.FindAllString(d, -1) {
			f, err := strconv.ParseFloat(t, 64)
			if err != nil {
				cmd, v = t, v[:0]
				continue
			}
			v = append(v, f)
			switch {
			case (cmd == "M" || cmd == "L") && len(v) == 2:
				poly = append(poly, pt{v[0], v[1]})
				v = v[:0]
			case cmd == "C" && len(v) == 6:
				from := poly[len(poly)-1]
				for s := 1; s <= steps; s++ {
					poly = append(poly, cubic(from, pt{v[0], v[1]}, pt{v[2], v[3]}, pt{v[4], v[5]}, float64(s)/steps))
				}
				v = v[:0]
			}
		}
		polys = append(polys, poly)
	}
	return polys
}

// in tells whether p is in poly, by counting the edges a ray from it to the
// right crosses.
func in(p pt, poly []pt) bool {
	in := false
	for i, a := range poly {
		b := poly[(i+1)%len(poly)]
		if (a.y > p.y) != (b.y > p.y) && p.x < a.x+(p.y-a.y)*(b.x-a.x)/(b.y-a.y) {
			in = !in
		}
	}
	return in
}

// edgeDist is the distance from p to the nearest edge of poly.
func edgeDist(p pt, poly []pt) float64 {
	d := math.Inf(1)
	for i, a := range poly {
		b := poly[(i+1)%len(poly)]
		pa, ba := pt{p.x - a.x, p.y - a.y}, pt{b.x - a.x, b.y - a.y}
		h := math.Max(0, math.Min(1, (pa.x*ba.x+pa.y*ba.y)/(ba.x*ba.x+ba.y*ba.y)))
		d = math.Min(d, math.Hypot(pa.x-ba.x*h, pa.y-ba.y*h))
	}
	return d
}

// inside tells whether p is in the star, or in a leaf and clear of the
// star by the gap.
func inside(p pt) bool {
	if in(p, shapes[0]) {
		return true
	}
	for _, leaf := range shapes[1:] {
		if in(p, leaf) && edgeDist(p, shapes[0]) > gap {
			return true
		}
	}
	return false
}

// frame is the scale and offset from the square to the glyph, which fit
// the glyph's bounds to the square.
func frame() (scale float64, off pt) {
	x0, y0, x1, y1 := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, poly := range shapes {
		for _, v := range poly {
			x0, y0, x1, y1 = math.Min(x0, v.x), math.Min(y0, v.y), math.Max(x1, v.x), math.Max(y1, v.y)
		}
	}
	scale = fit / math.Max(x1-x0, y1-y0)
	return scale, pt{(x0+x1)/2 - 9/scale, (y0+y1)/2 - 9/scale}
}

// render draws the icon px pixels square, each pixel's alpha the share of
// 16×16 samples inside the glyph.
func render(px int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, px, px))
	s := float64(px) / 18
	k, off := frame()
	const n = 16
	for y := 0; y < px; y++ {
		for x := 0; x < px; x++ {
			hit := 0
			for j := 0; j < n; j++ {
				for i := 0; i < n; i++ {
					p := pt{(float64(x)+(float64(i)+0.5)/n)/s/k + off.x, (float64(y)+(float64(j)+0.5)/n)/s/k + off.y}
					if inside(p) {
						hit++
					}
				}
			}
			img.SetNRGBA(x, y, color.NRGBA{0, 0, 0, uint8(math.Round(float64(hit) * 255 / (n * n)))})
		}
	}
	return img
}

// preview puts the icon on a light menu bar and a dark one, tinted as
// macOS tints it, and enlarged with nearest pixels so each one shows.
func preview(icon *image.NRGBA, zoom int) *image.NRGBA {
	w := icon.Bounds().Dx()
	pad := w / 3
	side := w + 2*pad
	out := image.NewNRGBA(image.Rect(0, 0, side*2*zoom, side*zoom))
	grounds := []color.NRGBA{{236, 236, 238, 255}, {40, 40, 44, 255}}
	inks := []float64{0, 255}
	for g := range grounds {
		for y := 0; y < side; y++ {
			for x := 0; x < side; x++ {
				c := grounds[g]
				if ix, iy := x-pad, y-pad; ix >= 0 && iy >= 0 && ix < w && iy < w {
					a := float64(icon.NRGBAAt(ix, iy).A) / 255
					mix := func(v uint8) uint8 { return uint8(float64(v)*(1-a) + inks[g]*a) }
					c = color.NRGBA{mix(c.R), mix(c.G), mix(c.B), 255}
				}
				for dy := 0; dy < zoom; dy++ {
					for dx := 0; dx < zoom; dx++ {
						out.SetNRGBA((g*side+x)*zoom+dx, y*zoom+dy, c)
					}
				}
			}
		}
	}
	return out
}

func save(name string, img image.Image) {
	f, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	out := flag.String("o", "internal/app/tray/tray-template.png", "where to write the icon")
	dir := flag.String("preview", "", "a directory for previews")
	flag.Parse()
	icon := render(36)
	save(*out, icon)
	if *dir != "" {
		save(filepath.Join(*dir, "preview-2x.png"), preview(icon, 8))
		save(filepath.Join(*dir, "preview-1x.png"), preview(render(18), 12))
		save(filepath.Join(*dir, "preview-large.png"), preview(render(360), 1))
	}
}
