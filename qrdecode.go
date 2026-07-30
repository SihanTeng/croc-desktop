package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"strings"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// decodeQrText extracts the text of the first QR code found in an image
// stream. Screenshots of the croc send view (or terminal) work as long as the
// QR code is visible in them.
func decodeQrText(r io.Reader) (string, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return "", fmt.Errorf("could not read image: %w", err)
	}
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", fmt.Errorf("could not process image: %w", err)
	}
	result, err := qrcode.NewQRCodeReader().Decode(bmp, nil)
	if err != nil {
		return "", fmt.Errorf("no QR code found in the image")
	}
	text := strings.TrimSpace(result.GetText())
	if text == "" {
		return "", fmt.Errorf("QR code is empty")
	}
	return text, nil
}

// DecodeCodeFromFile reads an image file and returns the text of the QR code
// in it (typically the croc code phrase). The image is never shown to the
// user — it is only used to extract the code.
func (a *App) DecodeCodeFromFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	return decodeQrText(f)
}

// DecodeCodeFromBase64 decodes a base64-encoded image (e.g. a pasted
// clipboard screenshot) and returns the text of the QR code in it.
func (a *App) DecodeCodeFromBase64(b64 string) (string, error) {
	// tolerate a data-URL prefix
	if i := strings.Index(b64, ","); i >= 0 && strings.HasPrefix(b64, "data:") {
		b64 = b64[i+1:]
	}
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", fmt.Errorf("could not decode image data: %w", err)
	}
	return decodeQrText(bytes.NewReader(raw))
}

// PickImage opens a native file dialog filtered to image files.
func (a *App) PickImage() (string, error) {
	app := application.Get()
	if app == nil {
		return "", fmt.Errorf("application not ready")
	}
	return app.Dialog.OpenFile().
		CanChooseFiles(true).
		SetTitle("Choose a screenshot of the QR code").
		AddFilter("Images", "*.png;*.jpg;*.jpeg;*.gif;*.bmp;*.webp").
		AddFilter("All files", "*").
		PromptForSingleSelection()
}
