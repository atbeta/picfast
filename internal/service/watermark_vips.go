//go:build !novips

package service

import (
	"fmt"
	"html"
	"strconv"

	"github.com/davidbyttow/govips/v2/vips"
)

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
