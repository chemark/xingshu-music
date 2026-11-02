// Command icon draws the app icons: the layers of AppIcon.icon, the Icon
// Composer document that macOS 26 and later render in Liquid Glass, and
// resources/icon.png, the same icon as macOS draws it, on the grid of older
// macOS icons, for older macOS and Windows; and from AppIcon-dev.icon, cream
// where the app's is dark, resources/icon-dev.png, the icon of mygo dev
// builds.
//
// Run it from the repository root on a Mac with Xcode 26 or later:
//
//	go run ./icon
//
// The icon.json of each document, its background, glass and dark-mode
// colors, is edited by hand or in Icon Composer.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const ictool = "/Applications/Xcode.app/Contents/Applications/Icon Composer.app/Contents/Executables/ictool"

// icons are the documents drawn, with the PNG rendered from each and the
// color of its star.
var icons = []struct {
	document, output, star string
}{
	// Dark, with a cream star.
	{"icon/AppIcon.icon", "resources/icon.png", "#F8EFE4"},
	// Cream, with a dark one.
	{"icon/AppIcon-dev.icon", "resources/icon-dev.png", "#12151C"},
}

// A star for 星 between two leaves for 树, on Icon Composer's 1024-point
// full-bleed canvas: the star a little left of and above the middle, its
// vertical points leaning right, and the leaves curving away from its sides
// with a gap between.
var (
	star   = struct{ cx, cy, rx, ry, lean float64 }{497, 503, 240.6, 259.5, -3}
	leaves = []string{
		"M518 246C565 199 684 179 764 204C823 224 845 278 845 338L845 477C808 432 674 423 599 378C550 348 525 298 518 246Z",
		"M488 773C445 820 336 840 276 815C232 795 207 746 207 686L207 532C239 577 326 592 396 621C450 646 480 696 488 773Z",
	}
)

func main() {
	for _, ic := range icons {
		layers := map[string]string{
			"star.svg":   starSVG(ic.star),
			"leaves.svg": leavesSVG(),
		}
		for name, s := range layers {
			if err := os.WriteFile(filepath.Join(ic.document, "Assets", name), []byte(s), 0o644); err != nil {
				log.Fatal(err)
			}
		}
		if err := render(ic.document, ic.output); err != nil {
			log.Fatal(err)
		}
	}
}

func svg(body string) string {
	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024">` + "\n" + body + "\n</svg>\n"
}

func n(v float64) string {
	return strings.TrimSuffix(strings.TrimRight(fmt.Sprintf("%.2f", v), "0"), ".")
}

// starSVG draws a four-pointed star, each side a curve from one point to the
// next that bows in toward the middle: it leaves a point along its axis and
// comes into the next along that one's.
func starSVG(fill string) string {
	a, b := star.rx, star.ry
	points := [][2]float64{{-a, 0}, {0, -b}, {a, 0}, {0, b}}
	d := "M" + n(-a) + " 0"
	for i, p := range points {
		q := points[(i+1)%4]
		// The corner of the bounding box between the two points.
		x, y := p[0]+q[0], p[1]+q[1]
		c1, c2 := [2]float64{0.5 * x, 0.06 * y}, [2]float64{0.04 * x, 0.45 * y}
		if p[0] == 0 {
			c1, c2 = [2]float64{0.04 * x, 0.45 * y}, [2]float64{0.5 * x, 0.06 * y}
		}
		d += fmt.Sprintf("C%s %s %s %s %s %s", n(c1[0]), n(c1[1]), n(c2[0]), n(c2[1]), n(q[0]), n(q[1]))
	}
	return svg(fmt.Sprintf(`<path transform="translate(%s %s) skewX(%s)" d="%sZ" fill="%s"/>`,
		n(star.cx), n(star.cy), n(star.lean), d, fill))
}

func leavesSVG() string {
	var b strings.Builder
	b.WriteString(`<g fill="#69984F">`)
	for _, d := range leaves {
		fmt.Fprintf(&b, "\n"+`<path d="%s"/>`, d)
	}
	b.WriteString("\n</g>")
	return svg(b.String())
}

// render has ictool draw document as macOS does, and lays it on the
// 1024-pixel grid of macOS icons at output: an 824-pixel body, centered,
// over a soft shadow.
func render(document, output string) error {
	const canvas, body = 1024, 824
	dir, err := os.MkdirTemp("", "icon")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	drawn := filepath.Join(dir, "icon.png")
	size := fmt.Sprint(body)
	if out, err := exec.Command(ictool, document, "--export-image", "--output-file", drawn,
		"--platform", "macOS", "--rendition", "Default", "--width", size, "--height", size, "--scale", "1").CombinedOutput(); err != nil {
		return fmt.Errorf("ictool (it comes with Xcode 26 or later): %v\n%s", err, out)
	}
	// ictool draws in Display P3, and the PNG this writes carries no
	// profile, which readers take as sRGB: convert, or the reds go dull.
	if out, err := exec.Command("sips", "--matchTo", "/System/Library/ColorSync/Profiles/sRGB Profile.icc", drawn).CombinedOutput(); err != nil {
		return fmt.Errorf("sips: %v\n%s", err, out)
	}
	f, err := os.Open(drawn)
	if err != nil {
		return err
	}
	defer f.Close()
	src, err := png.Decode(f)
	if err != nil {
		return err
	}

	at := image.Pt((canvas-body)/2, (canvas-body)/2)
	dst := image.NewNRGBA(image.Rect(0, 0, canvas, canvas))
	shadow := image.NewAlpha(dst.Bounds())
	draw.Draw(shadow, src.Bounds().Add(at).Add(image.Pt(0, 12)), src, src.Bounds().Min, draw.Src)
	blur(shadow, 10)
	draw.DrawMask(dst, dst.Bounds(), image.NewUniform(color.NRGBA{A: 72}), image.Point{}, shadow, image.Point{}, draw.Over)
	draw.Draw(dst, src.Bounds().Add(at), src, src.Bounds().Min, draw.Over)

	out, err := os.Create(output)
	if err != nil {
		return err
	}
	if err := png.Encode(out, dst); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// blur approximates a Gaussian blur of the given radius with three box
// blurs in each direction.
func blur(m *image.Alpha, radius int) {
	w, h := m.Rect.Dx(), m.Rect.Dy()
	buf := make([]uint8, len(m.Pix))
	pass := func(src, dst []uint8, count, length, stride, step int) {
		for line := range count {
			base := line * stride
			sum := 0
			for i := -radius; i <= radius; i++ {
				sum += int(src[base+min(max(i, 0), length-1)*step])
			}
			for i := range length {
				dst[base+i*step] = uint8(sum / (2*radius + 1))
				sum += int(src[base+min(i+radius+1, length-1)*step]) - int(src[base+max(i-radius, 0)*step])
			}
		}
	}
	for range 3 {
		pass(m.Pix, buf, h, w, m.Stride, 1)
		pass(buf, m.Pix, w, h, 1, m.Stride)
	}
}
