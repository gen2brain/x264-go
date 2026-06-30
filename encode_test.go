package x264

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"io"
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
	draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)

	for i := 0; i < opts.Width/2; i++ {
		img.Set(i, opts.Height/2, color.RGBA{R: 255, A: 255})

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

	err = os.WriteFile(filepath.Join(os.TempDir(), "test.264"), buf.Bytes(), 0644)
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
	draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)

	for i := 0; i < opts.Width/2; i++ {
		img.Set(i, opts.Height/2, color.RGBA{R: 255, A: 255})

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

	err = os.WriteFile(filepath.Join(os.TempDir(), "test.high.264"), buf.Bytes(), 0644)
	if err != nil {
		t.Error(err)
	}
}

func TestEncodeAbr(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0))

	opts := &Options{
		Width:         640,
		Height:        480,
		FrameRate:     25,
		Tune:          "zerolatency",
		Preset:        "veryfast",
		Profile:       "baseline",
		RateControl:   "abr",
		Bitrate:       500,
		RateMax:       600,
		VbvBufferSize: 600,
		LogLevel:      LogDebug,
	}

	enc, err := NewEncoder(buf, opts)
	if err != nil {
		t.Fatal(err)
	}

	img := NewYCbCr(image.Rect(0, 0, opts.Width, opts.Height))
	draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)

	for i := 0; i < opts.Width/2; i++ {
		img.Set(i, opts.Height/2, color.RGBA{R: 255, A: 255})

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

	err = os.WriteFile(filepath.Join(os.TempDir(), "test.abr.264"), buf.Bytes(), 0644)
	if err != nil {
		t.Error(err)
	}
}

func TestEncodeAbrNoBitrate(t *testing.T) {
	opts := &Options{
		Width:       640,
		Height:      480,
		FrameRate:   25,
		Preset:      "veryfast",
		Tune:        "zerolatency",
		RateControl: "abr",
		LogLevel:    LogError,
	}

	_, err := NewEncoder(bytes.NewBuffer(nil), opts)
	if err == nil {
		t.Fatal("expected error for abr without Bitrate, got nil")
	}
}

func TestEncodeCrf(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0))

	opts := &Options{
		Width:        640,
		Height:       480,
		FrameRate:    25,
		Tune:         "zerolatency",
		Preset:       "veryfast",
		Profile:      "baseline",
		RateControl:  "crf",
		RateConstant: 18,
		LogLevel:     LogDebug,
	}

	enc, err := NewEncoder(buf, opts)
	if err != nil {
		t.Fatal(err)
	}

	img := NewYCbCr(image.Rect(0, 0, opts.Width, opts.Height))
	draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)

	for i := 0; i < opts.Width/2; i++ {
		img.Set(i, opts.Height/2, color.RGBA{R: 255, A: 255})

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

	err = os.WriteFile(filepath.Join(os.TempDir(), "test.crf.264"), buf.Bytes(), 0644)
	if err != nil {
		t.Error(err)
	}
}

func TestEncodeMixedTypesNoAlias(t *testing.T) {
	opts := &Options{
		Width:     64,
		Height:    64,
		FrameRate: 25,
		Preset:    "ultrafast",
		Tune:      "zerolatency",
		LogLevel:  LogError,
	}

	enc, err := NewEncoder(bytes.NewBuffer(nil), opts)
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()

	yc := NewYCbCr(image.Rect(0, 0, opts.Width, opts.Height))
	draw.Draw(yc, yc.Bounds(), image.Black, image.Point{}, draw.Src)
	yc.Y[0] = 200
	yPtr := &yc.Y[0]

	if err = enc.Encode(yc); err != nil {
		t.Fatal(err)
	}

	// Encoding a different frame type must not rebind or mutate the caller's YCbCr.
	rgba := image.NewRGBA(image.Rect(0, 0, opts.Width, opts.Height))
	if err = enc.Encode(rgba); err != nil {
		t.Fatal(err)
	}

	if &yc.Y[0] != yPtr {
		t.Error("caller YCbCr.Y backing array was rebound by the encoder")
	}
	if yc.Y[0] != 200 {
		t.Errorf("caller YCbCr.Y was mutated: got %d, want 200", yc.Y[0])
	}
}

func BenchmarkEncodeRGBA(b *testing.B) {
	opts := &Options{
		Width:     640,
		Height:    480,
		FrameRate: 25,
		Preset:    "ultrafast",
		Tune:      "zerolatency",
		LogLevel:  LogNone,
	}

	enc, err := NewEncoder(io.Discard, opts)
	if err != nil {
		b.Fatal(err)
	}
	defer enc.Close()

	img := image.NewRGBA(image.Rect(0, 0, opts.Width, opts.Height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{R: 80, G: 160, B: 240, A: 255}}, image.Point{}, draw.Src)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(img); err != nil {
			b.Fatal(err)
		}
	}
}

// countNALType counts Annex-B NAL units of the given type (byte after a start code, low 5 bits).
func countNALType(stream []byte, nalType byte) int {
	n := 0
	for i := 0; i+3 < len(stream); i++ {
		if stream[i] == 0 && stream[i+1] == 0 && stream[i+2] == 1 {
			if stream[i+3]&0x1f == nalType {
				n++
			}
			i += 2
		}
	}
	return n
}

func TestEncodeKeyframes(t *testing.T) {
	const idr = 5 // H.264 IDR slice NAL type

	encode := func(o *Options, frames int) []byte {
		buf := bytes.NewBuffer(nil)
		enc, err := NewEncoder(buf, o)
		if err != nil {
			t.Fatal(err)
		}
		img := NewYCbCr(image.Rect(0, 0, o.Width, o.Height))
		draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)
		for i := 0; i < frames; i++ {
			img.Set(i%o.Width, o.Height/2, color.RGBA{R: 255, A: 255})
			if err := enc.Encode(img); err != nil {
				t.Error(err)
			}
		}
		if err := enc.Flush(); err != nil {
			t.Error(err)
		}
		if err := enc.Close(); err != nil {
			t.Error(err)
		}
		return buf.Bytes()
	}

	base := func() *Options {
		return &Options{Width: 320, Height: 240, FrameRate: 30, Preset: "ultrafast", Tune: "zerolatency", Profile: "baseline", LogLevel: LogNone}
	}

	// Asserts relationships, not absolute counts (sliced-threads emits several IDR NALs per keyframe).

	// Smaller KeyInt → more keyframes.
	short := base()
	short.KeyInt = 15
	long := base()
	long.KeyInt = 1000
	shortIDR := countNALType(encode(short, 120), idr)
	longIDR := countNALType(encode(long, 120), idr)
	if shortIDR <= longIDR {
		t.Errorf("smaller KeyInt should yield more IDR keyframes: KeyInt=15 -> %d, KeyInt=1000 -> %d", shortIDR, longIDR)
	}

	// Intra-refresh leaves only the initial keyframe.
	refresh := base()
	refresh.KeyInt = 15
	refresh.IntraRefresh = true
	refreshIDR := countNALType(encode(refresh, 120), idr)
	if refreshIDR >= shortIDR {
		t.Errorf("IntraRefresh should suppress periodic IDR keyframes: KeyInt=15 -> %d, with IntraRefresh -> %d", shortIDR, refreshIDR)
	}
}
