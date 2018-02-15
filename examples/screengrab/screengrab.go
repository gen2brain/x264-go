package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/gen2brain/x264-go"
	"github.com/kbinani/screenshot"
)

func main() {
	buf := bytes.NewBuffer(make([]byte, 0))

	bounds := screenshot.GetDisplayBounds(0)

	opts := &x264.Options{
		Width:     bounds.Dx(),
		Height:    bounds.Dy(),
		FrameRate: 10,
		Tune:      "zerolatency",
		Preset:    "veryfast",
		Profile:   "baseline",
		LogLevel:  x264.LogDebug,
	}

	enc, err := x264.NewEncoder(buf, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	defer enc.Close()

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-s:
			enc.Flush()

			err = ioutil.WriteFile("screen.264", buf.Bytes(), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			os.Exit(0)
		default:
			img, err := screenshot.CaptureRect(bounds)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				continue
			}

			err = enc.Encode(img)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			}
		}
	}
}
