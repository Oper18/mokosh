package thumb

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var watermarkFont *opentype.Font

func init() {
	f, err := opentype.Parse(goregular.TTF)
	if err != nil {
		log.Errorf("thumb: failed to parse watermark font: %s", err)
		return
	}

	watermarkFont = f
}

// WatermarkFile opens the image at fileName, draws a single large diagonal text watermark, and returns JPEG-encoded bytes.
func WatermarkFile(fileName, text string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("thumb: watermark text is empty")
	}

	if watermarkFont == nil {
		return nil, fmt.Errorf("thumb: watermark font not available")
	}

	img, err := imaging.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("thumb: open %s: %w", fileName, err)
	}

	marked, err := applyWatermark(img, text)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err = imaging.Encode(&buf, marked, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
		return nil, fmt.Errorf("thumb: watermark encode: %w", err)
	}

	return buf.Bytes(), nil
}

// applyWatermark draws a single semi-transparent text watermark horizontally centered in the middle of the image.
func applyWatermark(img image.Image, text string) (image.Image, error) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	// Font size: ~8% of the shorter dimension, capped at reasonable limits.
	minDim := w
	if h < minDim {
		minDim = h
	}
	fontSize := float64(minDim) * 0.08
	if fontSize < 16 {
		fontSize = 16
	} else if fontSize > 120 {
		fontSize = 120
	}

	face, err := opentype.NewFace(watermarkFont, &opentype.FaceOptions{
		Size: fontSize,
		DPI:  72,
	})
	if err != nil {
		return nil, fmt.Errorf("thumb: watermark face: %w", err)
	}
	defer face.Close()

	textW := font.MeasureString(face, text).Ceil()
	lineH := face.Metrics().Height.Ceil()

	dst := image.NewNRGBA(bounds)
	draw.Draw(dst, bounds, img, bounds.Min, draw.Src)

	// Draw text once, horizontally centered, vertically centered.
	x := bounds.Min.X + (w-textW)/2
	y := bounds.Min.Y + (h+lineH)/2

	col := image.NewUniform(color.NRGBA{R: 255, G: 255, B: 255, A: 60})
	(&font.Drawer{Dst: dst, Src: col, Face: face, Dot: fixed.P(x, y)}).DrawString(text)

	return dst, nil
}
