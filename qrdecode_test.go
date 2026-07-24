package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/skip2/go-qrcode"
)

func TestDecodeQrTextRoundtrip(t *testing.T) {
	const code = "4821-falcon-river-quartz"
	pngBytes, err := qrcode.Encode(code, qrcode.Medium, 256)
	if err != nil {
		t.Fatal(err)
	}

	app := &App{}

	got, err := app.DecodeCodeFromBase64(base64.StdEncoding.EncodeToString(pngBytes))
	if err != nil {
		t.Fatalf("DecodeCodeFromBase64: %v", err)
	}
	if got != code {
		t.Fatalf("got %q, want %q", got, code)
	}

	// data-URL prefix tolerated
	got, err = app.DecodeCodeFromBase64("data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes))
	if err != nil {
		t.Fatalf("DecodeCodeFromBase64 with prefix: %v", err)
	}
	if got != code {
		t.Fatalf("got %q, want %q", got, code)
	}
}

func TestDecodeQrTextNoQr(t *testing.T) {
	// a plain blank image contains no QR code
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.White)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	app := &App{}
	_, err := app.DecodeCodeFromBase64(base64.StdEncoding.EncodeToString(buf.Bytes()))
	if err == nil {
		t.Fatal("expected an error for an image without a QR code")
	}
	if !strings.Contains(err.Error(), "no QR code") {
		t.Fatalf("unexpected error: %v", err)
	}
}
