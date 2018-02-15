package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/gen2brain/cam2ip/camera"
	"github.com/gen2brain/x264-go"
)

func main() {
	buf := bytes.NewBuffer(make([]byte, 0))

	opts := &x264.Options{
		Width:     640,
		Height:    480,
		FrameRate: 15,
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

	cam, err := camera.New(1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	cam.SetProperty(camera.PropFrameWidth, float64(opts.Width))
	cam.SetProperty(camera.PropFrameHeight, float64(opts.Height))

	defer cam.Close()

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-s:
			enc.Flush()

			err = ioutil.WriteFile("camera.264", buf.Bytes(), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			os.Exit(0)
		default:
			img, err := cam.Read()
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
