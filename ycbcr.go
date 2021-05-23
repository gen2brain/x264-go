package x264

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/gen2brain/x264-go/yuv"
)

// YCbCr is an in-memory image of Y'CbCr colors.
type YCbCr struct {
	*image.YCbCr
}

// NewYCbCr returns a new YCbCr image with the given bounds and subsample ratio.
func NewYCbCr(r image.Rectangle) *YCbCr {
	return &YCbCr{image.NewYCbCr(r, image.YCbCrSubsampleRatio420)}
}

// Set sets pixel color.
func (p *YCbCr) Set(x, y int, c color.Color) {
	p.setYCbCr(x, y, p.ColorModel().Convert(c).(color.YCbCr))
}

func (p *YCbCr) setYCbCr(x, y int, c color.YCbCr) {
	if !image.Pt(x, y).In(p.Rect) {
		return
	}

	yi := p.YOffset(x, y)
	ci := p.COffset(x, y)

	p.Y[yi] = c.Y
	p.Cb[ci] = c.Cb
	p.Cr[ci] = c.Cr
}

// ToYCbCrDraw converts image.Image to YCbCr.
func (p *YCbCr) ToYCbCrDraw(src image.Image) {
	bounds := src.Bounds()
	draw.Draw(p, bounds, src, bounds.Min, draw.Src)
}

// ToYCbCrColor converts image.Image to YCbCr.
func (p *YCbCr) ToYCbCrColor(src image.Image) {
	bounds := src.Bounds()

	for row := 0; row < bounds.Max.Y; row++ {
		for col := 0; col < bounds.Max.X; col++ {
			r, g, b, _ := src.At(col, row).RGBA()
			y, cb, cr := color.RGBToYCbCr(uint8(r), uint8(g), uint8(b))

			p.Y[p.YOffset(col, row)] = y
			p.Cb[p.COffset(col, row)] = cb
			p.Cr[p.COffset(col, row)] = cr
		}
	}
}

// ToYCbCr converts image.Image to YCbCr.
func (p *YCbCr) ToYCbCr(src image.Image) {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	lumaSize := int32(width * height)
	chromaSize := int32(width*height) / 4

	yuvProc := yuv.NewYuvImgProcessor(width, height)
	yCbCr := yuvProc.Process(src.(*image.RGBA)).Get()

	p.Y = yCbCr[:lumaSize]
	p.Cb = yCbCr[lumaSize : lumaSize+chromaSize]
	p.Cr = yCbCr[lumaSize+chromaSize:]
}
