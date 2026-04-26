package service

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type WatermarkConfig struct {
	Text     string  `json:"text"`
	Position string  `json:"position"`
	FontSize float64 `json:"font_size"`
	Color    string  `json:"color"`
	Opacity  float64 `json:"opacity"`
}

func parseColor(hex string, opacity float64) (color.Color, error) {
	hex = strings.TrimPrefix(hex, "#")
	var r, g, b, a uint8
	switch len(hex) {
	case 6:
		rv, _ := strconv.ParseUint(hex[0:2], 16, 8)
		gv, _ := strconv.ParseUint(hex[2:4], 16, 8)
		bv, _ := strconv.ParseUint(hex[4:6], 16, 8)
		r, g, b = uint8(rv), uint8(gv), uint8(bv)
		a = uint8(opacity * 255)
	case 8:
		rv, _ := strconv.ParseUint(hex[0:2], 16, 8)
		gv, _ := strconv.ParseUint(hex[2:4], 16, 8)
		bv, _ := strconv.ParseUint(hex[4:6], 16, 8)
		av, _ := strconv.ParseUint(hex[6:8], 16, 8)
		r, g, b, a = uint8(rv), uint8(gv), uint8(bv), uint8(av)
		a = uint8(float64(a) * opacity)
	default:
		return color.White, fmt.Errorf("invalid color format")
	}
	return color.RGBA{r, g, b, a}, nil
}

func calcWatermarkPosition(imgW, imgH, textW, textH int, position string) (x, y int) {
	padding := int(float64(imgH) * 0.02)
	if padding < 10 {
		padding = 10
	}
	switch position {
	case "top-left":
		return padding, padding + textH
	case "top-right":
		return imgW - textW - padding, padding + textH
	case "bottom-left":
		return padding, imgH - padding
	case "center":
		return (imgW - textW) / 2, (imgH + textH) / 2
	default: // bottom-right
		return imgW - textW - padding, imgH - padding
	}
}

// ApplyWatermark overlays text watermark onto the image.
// formatHint should be "png" to keep PNG output; otherwise JPEG is used.
func ApplyWatermark(data []byte, cfg WatermarkConfig, formatHint string, quality int) ([]byte, error) {
	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Parse font
	ttf, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	face, err := opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    cfg.FontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}

	// Parse color with opacity
	c, err := parseColor(cfg.Color, cfg.Opacity)
	if err != nil {
		c = color.RGBA{255, 255, 255, uint8(cfg.Opacity * 255)}
	}

	// Convert to RGBA for drawing
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Measure text
	d := &font.Drawer{Face: face}
	textWidth := d.MeasureString(cfg.Text).Round()
	textHeight := int(cfg.FontSize)

	// Calculate position
	x, y := calcWatermarkPosition(width, height, textWidth, textHeight, cfg.Position)

	// Draw text
	d = &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	d.DrawString(cfg.Text)

	// Encode result
	buf := new(bytes.Buffer)
	if formatHint == "png" {
		if err := png.Encode(buf, rgba); err != nil {
			return nil, err
		}
	} else {
		q := quality
		if q <= 0 || q > 100 {
			q = 85
		}
		if err := jpeg.Encode(buf, rgba, &jpeg.Options{Quality: q}); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
