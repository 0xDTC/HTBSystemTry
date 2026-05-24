package tray

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// generateIcon builds the tray icon at runtime: a dark square with a Hack The
// Box-green "H". This avoids shipping (and hardcoding a path to) an image asset.
func generateIcon() []byte {
	const size = 32
	bg := color.NRGBA{R: 0x11, G: 0x1A, B: 0x27, A: 0xFF} // HTB dark navy
	fg := color.NRGBA{R: 0x9F, G: 0xEF, B: 0x00, A: 0xFF} // HTB green
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.SetNRGBA(x, y, bg)
		}
	}

	// Draw a bold "H": two vertical bars and a central crossbar.
	fill := func(x0, y0, x1, y1 int) {
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				img.SetNRGBA(x, y, fg)
			}
		}
	}
	fill(8, 6, 12, 26)
	fill(20, 6, 24, 26)
	fill(12, 14, 20, 18)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}
	return buf.Bytes()
}
