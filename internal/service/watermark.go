package service

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

type WatermarkConfig struct {
	Text     string  `json:"text"`
	Position string  `json:"position"`
	FontSize float64 `json:"font_size"`
	Color    string  `json:"color"`
	Opacity  float64 `json:"opacity"`
}

func parseColor(hex string, opacity float64) (r, g, b uint8, a uint8, err error) {
	hex = strings.TrimPrefix(hex, "#")
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
		r, g, b = uint8(rv), uint8(gv), uint8(bv)
		a = uint8(float64(uint8(av)) * opacity)
	default:
		return 255, 255, 255, uint8(opacity * 255), fmt.Errorf("invalid color format")
	}
	return r, g, b, a, nil
}

func calcWatermarkPosition(imgW, imgH, textW, textH int, position string) (x, y int) {
	padding := int(float64(imgH) * 0.02)
	if padding < 10 {
		padding = 10
	}
	switch position {
	case "top-left":
		return padding, padding
	case "top-right":
		return imgW - textW - padding, padding
	case "bottom-left":
		return padding, imgH - textH - padding
	case "center":
		return (imgW - textW) / 2, (imgH - textH) / 2
	default: // bottom-right
		return imgW - textW - padding, imgH - textH - padding
	}
}

// ApplyWatermark overlays text watermark onto the image using govips Text + Composite.
// formatHint should be "png" to keep PNG output; otherwise JPEG is used.
func ApplyWatermark(data []byte, cfg WatermarkConfig, formatHint string, quality int) ([]byte, error) {
	params := vips.NewImportParams()
	params.AutoRotate.Set(true)
	params.FailOnError.Set(true)

	img, err := vips.LoadImageFromBuffer(data, params)
	if err != nil {
		return nil, err
	}
	defer img.Close()

	imgW := img.Width()
	imgH := img.Height()

	// Parse color
	r, g, b, a, err := parseColor(cfg.Color, cfg.Opacity)
	if err != nil {
		r, g, b, a = 255, 255, 255, uint8(cfg.Opacity*255)
	}

	// Build Pango markup for colored text with alpha
	// alpha in hex (00-FF)
	aHex := strconv.FormatInt(int64(a), 16)
	if len(aHex) == 1 {
		aHex = "0" + aHex
	}
	// Use a CJK-friendly font fallback chain and escape text to keep Pango markup valid.
	fontDesc := fmt.Sprintf("Noto Sans CJK SC, Noto Sans CJK, Sans %.0f", cfg.FontSize)
	pangoText := fmt.Sprintf(
		"<span font_desc='%s' foreground='#%02X%02X%02X%s'>%s</span>",
		fontDesc, r, g, b, aHex, html.EscapeString(cfg.Text),
	)

	// Create text image
	textImg, err := vips.Text(&vips.TextParams{
		Text: pangoText,
		DPI:  72,
		RGBA: true,
	})
	if err != nil {
		return nil, err
	}
	defer textImg.Close()

	textW := textImg.Width()
	textH := textImg.Height()

	x, y := calcWatermarkPosition(imgW, imgH, textW, textH, cfg.Position)

	if err := img.Composite(textImg, vips.BlendModeOver, x, y); err != nil {
		return nil, err
	}

	// Export result
	var out []byte
	var exportErr error
	if formatHint == "png" {
		p := vips.NewPngExportParams()
		out, _, exportErr = img.ExportPng(p)
	} else {
		p := vips.NewJpegExportParams()
		if quality > 0 && quality <= 100 {
			p.Quality = quality
		}
		out, _, exportErr = img.ExportJpeg(p)
	}

	if exportErr != nil {
		return nil, exportErr
	}
	return out, nil
}
