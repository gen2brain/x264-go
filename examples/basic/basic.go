package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"os"

	"github.com/gen2brain/x264-go"
)

func main() {
	opts := &x264.Options{
		Width:     640,
		Height:    480,
		FrameRate: 25,
		Preset:    "veryfast",
		Tune:      "stillimage",
		Profile:   "baseline",
		LogLevel:  x264.LogDebug,
	}

	w, err := os.Create("example.264")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer w.Close()

	enc, err := x264.NewEncoder(w, opts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer enc.Close()

	r, err := os.Open("example.jpg")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	img, _, err := image.Decode(r)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = enc.Encode(img)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = enc.Flush()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
