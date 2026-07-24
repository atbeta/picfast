package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/atbeta/picfast/internal/events"
	"github.com/atbeta/picfast/internal/sqlc"
)

const maxOCRTextLen = 4096

// DummyOCRProcessor simulates an OCR task on uploaded images
func DummyOCRProcessor(db *sqlc.Queries, event sqlc.OutboxEvent) {
	if event.Type != events.TypeImageUploaded {
		return
	}

	var data events.ImageUploadedData
	if err := json.Unmarshal(event.Payload, &data); err != nil {
		slog.Warn("ocr: failed to unmarshal event payload", "error", err)
		return
	}

	if data.Mimetype == "image/gif" {
		return
	}

	time.Sleep(2 * time.Second)

	origin := data.OriginName
	if len(origin) > 200 {
		origin = origin[:200]
	}
	ocrText := "Extracted text for " + origin
	if len(ocrText) > maxOCRTextLen {
		ocrText = ocrText[:maxOCRTextLen]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the image ID from the key
	img, err := db.GetImageByKey(ctx, data.Key)
	if err != nil {
		slog.Warn("ocr: image not found", "key", data.Key, "error", err)
		return
	}

	// Update the image with the extracted OCR text
	err = db.UpdateImageOCR(ctx, sqlc.UpdateImageOCRParams{
		ID:      img.ID,
		OcrText: ocrText,
	})
	if err != nil {
		slog.Warn("ocr: failed to update image ocr text", "id", img.ID, "error", err)
		return
	}

	slog.Info("ocr: dummy extraction complete", "key", data.Key)
}
