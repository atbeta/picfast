package service

import (
	"bytes"
	"encoding/json"
	"log/slog"

	"github.com/rwcarlsen/goexif/exif"
)

type ExifData struct {
	CameraMake  string  `json:"camera_make,omitempty"`
	CameraModel string  `json:"camera_model,omitempty"`
	FocalLength string  `json:"focal_length,omitempty"`
	Aperture    string  `json:"aperture,omitempty"`
	ShutterSpeed string `json:"shutter_speed,omitempty"`
	ISO         int     `json:"iso,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	DateTime    string  `json:"date_time,omitempty"`
	Software    string  `json:"software,omitempty"`
	Orientation int     `json:"orientation,omitempty"`
}

func ExtractExif(data []byte) json.RawMessage {
	x, err := exif.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}

	var ed ExifData

	if tag, err := x.Get(exif.Make); err == nil {
		ed.CameraMake, _ = tag.StringVal()
	}
	if tag, err := x.Get(exif.Model); err == nil {
		ed.CameraModel, _ = tag.StringVal()
	}
	if tag, err := x.Get(exif.FNumber); err == nil {
		a, _ := tag.Rat(0)
		if a != nil {
			ed.Aperture = a.String()
		}
	}
	if tag, err := x.Get(exif.FocalLength); err == nil {
		a, _ := tag.Rat(0)
		if a != nil {
			ed.FocalLength = a.String()
		}
	}
	if tag, err := x.Get(exif.ExposureTime); err == nil {
		a, _ := tag.Rat(0)
		if a != nil {
			ed.ShutterSpeed = a.String()
		}
	}
	if tag, err := x.Get(exif.ISOSpeedRatings); err == nil {
		iso, _ := tag.Int(0)
		ed.ISO = iso
	}
	if tag, err := x.Get(exif.Software); err == nil {
		ed.Software, _ = tag.StringVal()
	}
	if tag, err := x.Get(exif.Orientation); err == nil {
		o, _ := tag.Int(0)
		ed.Orientation = o
	}
	if tag, err := x.Get(exif.DateTimeOriginal); err == nil {
		ed.DateTime, _ = tag.StringVal()
	} else if tag, err := x.Get(exif.DateTime); err == nil {
		ed.DateTime, _ = tag.StringVal()
	}

	lat, lon, err := x.LatLong()
	if err == nil {
		ed.Latitude = lat
		ed.Longitude = lon
	}

	result, err := json.Marshal(ed)
	if err != nil {
		slog.Warn("failed to marshal exif", "error", err)
		return nil
	}
	return result
}
