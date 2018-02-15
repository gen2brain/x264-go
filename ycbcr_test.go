package x264

import (
	"image"
	"testing"
)

func TestYCbCr(t *testing.T) {
	var img image.Image

	rgba := image.NewRGBA(image.Rect(0, 0, 640, 480))

	ycbcr := NewYCbCr(rgba.Bounds())
	ycbcr.ToYCbCr(rgba)

	img = ycbcr

	_, ok := img.(*YCbCr)
	if !ok {
		t.Error("ToYCbCr failed")
	}
}
