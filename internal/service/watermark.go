package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type WatermarkConfig struct {
	Text     string  `json:"text"`
	Position string  `json:"position"`
	FontSize float64 `json:"font_size"`
	Color    string  `json:"color"`
	Opacity  float64 `json:"opacity"`
}

// ErrWatermarkUnsupported marks builds without text watermark support — the
// pure-Go backend has no font renderer.
var ErrWatermarkUnsupported = errors.New("watermark unsupported by this build")

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
