package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gen2brain/x264-go"
	"github.com/kbinani/screenshot"
)

func main() {
	file, err := os.Create("screen.264")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	bounds := screenshot.GetDisplayBounds(0)

	opts := &x264.Options{
		Width:     bounds.Dx(),
		Height:    bounds.Dy(),
		FrameRate: 25,
		Tune:      "zerolatency",
		Preset:    "ultrafast",
		Profile:   "baseline",
		LogLevel:  x264.LogError,
	}

	enc, err := x264.NewEncoder(file, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	defer enc.Close()

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(time.Second / time.Duration(25))

	start := time.Now()
	frame := 0

	for range ticker.C {
		select {
		case <-s:
			enc.Flush()

			err = file.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			os.Exit(0)
		default:
			frame++
			log.Printf("frame: %v", frame)
			img, err := screenshot.CaptureRect(bounds)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				continue
			}

			err = enc.Encode(img)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			}
			log.Printf("t: %v", time.Since(start))
			start = time.Now()
		}
	}
}
