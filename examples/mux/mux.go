// Command mux muxes x264-go's raw H.264 stream into a playable .mp4 using
// github.com/Eyevinn/mp4ff: SPS/PPS go into the avcC descriptor, each frame's
// Annex-B NALs become a length-prefixed AVCC sample, written as a single-file
// fragmented MP4. Tune "zerolatency" gives no B-frames (DTS == PTS).
package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"

	"github.com/Eyevinn/mp4ff/avc"
	"github.com/Eyevinn/mp4ff/mp4"
	"github.com/gen2brain/x264-go"
)

// nalWriter records each Write separately: unit[0] is the SPS/PPS header, then one per coded frame.
type nalWriter struct {
	units [][]byte
}

func (w *nalWriter) Write(p []byte) (int, error) {
	u := make([]byte, len(p))
	copy(u, p)
	w.units = append(w.units, u)
	return len(p), nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "mux:", err)
		os.Exit(1)
	}
}

func run() error {
	const fps = 25

	opts := &x264.Options{
		Width:     640,
		Height:    480,
		FrameRate: fps,
		Tune:      "zerolatency",
		Preset:    "veryfast",
		Profile:   "baseline",
		LogLevel:  x264.LogWarning,
	}

	// Encode a short generated clip into per-frame Annex-B units.
	nw := &nalWriter{}
	enc, err := x264.NewEncoder(nw, opts)
	if err != nil {
		return fmt.Errorf("new encoder: %w", err)
	}

	img := x264.NewYCbCr(image.Rect(0, 0, opts.Width, opts.Height))
	for i := 0; i < opts.Width/2; i++ {
		draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)
		for y := 0; y < opts.Height; y++ {
			img.Set(i, y, color.RGBA{R: 255, A: 255})
		}
		if err := enc.Encode(img); err != nil {
			return fmt.Errorf("encode: %w", err)
		}
	}
	if err := enc.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}

	if len(nw.units) < 2 {
		return fmt.Errorf("expected headers + at least one frame, got %d units", len(nw.units))
	}
	headers, frames := nw.units[0], nw.units[1:]

	// Build the init segment: one video track described by the SPS/PPS avcC.
	spss, ppss := avc.GetParameterSetsFromByteStream(headers)
	if len(spss) == 0 || len(ppss) == 0 {
		return fmt.Errorf("no SPS/PPS found in stream headers")
	}

	init := mp4.CreateEmptyInit()
	trak := init.AddEmptyTrack(fps, "video", "und")
	if err := trak.SetAVCDescriptor("avc1", spss, ppss, true); err != nil {
		return fmt.Errorf("set avc descriptor: %w", err)
	}

	// One fragment holding every frame as an AVCC sample, 1 tick (1/fps s) each.
	frag, err := mp4.CreateFragment(1, trak.Tkhd.TrackID)
	if err != nil {
		return fmt.Errorf("create fragment: %w", err)
	}
	for i, unit := range frames {
		// Build the AVCC sample, dropping SPS/PPS/AUD (parameter sets live in avcC).
		var sample []byte
		isSync := i == 0
		for _, nalu := range avc.ExtractNalusFromByteStream(unit) {
			switch avc.GetNaluType(nalu[0]) {
			case avc.NALU_SPS, avc.NALU_PPS, avc.NALU_AUD:
				continue
			case avc.NALU_IDR:
				isSync = true
			}
			var length [4]byte
			binary.BigEndian.PutUint32(length[:], uint32(len(nalu)))
			sample = append(sample, length[:]...)
			sample = append(sample, nalu...)
		}

		flags := mp4.SetNonSyncSampleFlags(0)
		if isSync {
			flags = mp4.SetSyncSampleFlags(0)
		}

		frag.AddFullSample(mp4.FullSample{
			Sample: mp4.Sample{
				Flags: flags,
				Dur:   1,
				Size:  uint32(len(sample)),
			},
			DecodeTime: uint64(i),
			Data:       sample,
		})
	}

	out, err := os.Create("out.mp4")
	if err != nil {
		return fmt.Errorf("create out.mp4: %w", err)
	}
	defer out.Close()

	if err := init.Encode(out); err != nil {
		return fmt.Errorf("encode init: %w", err)
	}
	if err := frag.Encode(out); err != nil {
		return fmt.Errorf("encode fragment: %w", err)
	}

	fmt.Printf("wrote out.mp4: %d frames, %dx%d @ %d fps\n", len(frames), opts.Width, opts.Height, fps)
	return nil
}
