package x264

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestEncode(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0))

	opts := &Options{
		Width:     640,
		Height:    480,
		FrameRate: 25,
		Tune:      "zerolatency",
		Preset:    "veryfast",
		Profile:   "baseline",
		LogLevel:  LogDebug,
	}

	enc, err := NewEncoder(buf, opts)
	if err != nil {
		t.Fatal(err)
	}

	img := NewYCbCr(image.Rect(0, 0, opts.Width, opts.Height))
	draw.Draw(img, img.Bounds(), image.Black, image.ZP, draw.Src)

	for i := 0; i < opts.Width/2; i++ {
		img.Set(i, opts.Height/2, color.RGBA{255, 0, 0, 255})

		err = enc.Encode(img)
		if err != nil {
			t.Error(err)
		}
	}

	err = enc.Flush()
	if err != nil {
		t.Error(err)
	}

	err = enc.Close()
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile(filepath.Join(os.TempDir(), "test.264"), buf.Bytes(), 0644)
	if err != nil {
		t.Error(err)
	}
}

func TestEncodeFlush(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0))

	opts := &Options{
		Width:     640,
		Height:    480,
		FrameRate: 25,
		Tune:      "film",
		Preset:    "fast",
		Profile:   "high",
		LogLevel:  LogDebug,
	}

	enc, err := NewEncoder(buf, opts)
	if err != nil {
		t.Fatal(err)
	}

	img := NewYCbCr(image.Rect(0, 0, opts.Width, opts.Height))
	draw.Draw(img, img.Bounds(), image.Black, image.ZP, draw.Src)

	for i := 0; i < opts.Width/2; i++ {
		img.Set(i, opts.Height/2, color.RGBA{255, 0, 0, 255})

		err = enc.Encode(img)
		if err != nil {
			t.Error(err)
		}
	}

	err = enc.Flush()
	if err != nil {
		t.Error(err)
	}

	err = enc.Close()
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile(filepath.Join(os.TempDir(), "test.high.264"), buf.Bytes(), 0644)
	if err != nil {
		t.Error(err)
	}
}
