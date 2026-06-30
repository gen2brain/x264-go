// Package x264 provides H.264/MPEG-4 AVC codec encoder based on [x264](https://www.videolan.org/developers/x264.html) library.
package x264

/*
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"image"
	"io"
	"unsafe"

	"github.com/gen2brain/x264-go/x264c"
	"github.com/gen2brain/x264-go/yuv"
)

// Logging constants.
const (
	LogNone int32 = iota - 1
	LogError
	LogWarning
	LogInfo
	LogDebug
)

// Options represent encoding options.
type Options struct {
	// Frame width.
	Width int
	// Frame height.
	Height int
	// Frame rate.
	FrameRate int
	// Tunings: film, animation, grain, stillimage, psnr, ssim, fastdecode, zerolatency.
	Tune string
	// Presets: ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow, placebo.
	Preset string
	// Profiles: baseline, main, high, high10, high422, high444.
	Profile string
	// RateControl: cqp, crf, abr.
	RateControl string
	// Quality target: quantizer for cqp, CRF value for crf.
	RateConstant float32
	// Upper bound: max quantizer for cqp, max CRF for crf, VBV max bitrate (kbit/s) for abr.
	RateMax float32
	// Target average bitrate (kbit/s). Required for abr.
	Bitrate int
	// VBV buffer size (kbit). Optional; bounds rate fluctuation when RateMax is set for abr.
	VbvBufferSize int
	// Log level.
	LogLevel int32
}

// Encoder type.
type Encoder struct {
	e *x264c.T
	w io.Writer

	img  *YCbCr
	opts *Options

	proc   yuv.ImgProcessor
	planes [3]unsafe.Pointer
	ySize  int
	cSize  int

	csp int32
	pts int64
	dts int64

	nnals int32
	nals  []*x264c.Nal

	picIn x264c.Picture

	tpf int64
}

// NewEncoder returns new x264 encoder.
func NewEncoder(w io.Writer, opts *Options) (e *Encoder, err error) {
	e = &Encoder{}

	e.w = w
	e.pts = 0
	e.opts = opts

	e.csp = x264c.CspI420

	e.nals = make([]*x264c.Nal, 3)
	e.img = NewYCbCr(image.Rect(0, 0, e.opts.Width, e.opts.Height))

	e.proc = yuv.NewYuvImgProcessor(e.opts.Width, e.opts.Height)
	e.ySize = e.opts.Width * e.opts.Height
	e.cSize = e.ySize / 4

	param := x264c.Param{}

	if e.opts.Preset != "" && e.opts.Tune != "" {
		ret := x264c.ParamDefaultPreset(&param, e.opts.Preset, e.opts.Tune)
		if ret < 0 {
			err = fmt.Errorf("x264: invalid preset/tune name")
			return
		}
	} else {
		x264c.ParamDefault(&param)
	}

	param.IWidth = int32(e.opts.Width)
	param.IHeight = int32(e.opts.Height)
	param.ICsp = e.csp
	param.ILogLevel = e.opts.LogLevel
	param.IBitdepth = 8

	param.BVfrInput = 0
	param.BRepeatHeaders = 1
	param.BAnnexb = 1

	param.BIntraRefresh = 1
	param.IKeyintMax = int32(e.opts.FrameRate)
	param.IFpsNum = uint32(e.opts.FrameRate)
	param.IFpsDen = 1

	if e.opts.Profile != "" {
		ret := x264c.ParamApplyProfile(&param, e.opts.Profile)
		if ret < 0 {
			err = fmt.Errorf("x264: invalid profile name")
			return
		}
	}

	if e.opts.RateControl != "" {
		switch e.opts.RateControl {
		case "cqp":
			param.Rc.IRcMethod = x264c.RcCqp
			if e.opts.RateConstant != 0 {
				param.Rc.IQpConstant = int32(e.opts.RateConstant)
			}
			if e.opts.RateMax != 0 {
				param.Rc.IQpMax = int32(e.opts.RateMax)
			}
		case "crf":
			param.Rc.IRcMethod = x264c.RcCrf
			if e.opts.RateConstant != 0 {
				param.Rc.FRfConstant = e.opts.RateConstant
			}
			if e.opts.RateMax != 0 {
				param.Rc.FRfConstantMax = e.opts.RateMax
			}
		case "abr":
			if e.opts.Bitrate <= 0 {
				err = fmt.Errorf("x264: abr rate control requires Options.Bitrate > 0")
				return
			}
			param.Rc.IRcMethod = x264c.RcAbr
			param.Rc.IBitrate = int32(e.opts.Bitrate)
			if e.opts.RateMax != 0 {
				param.Rc.IVbvMaxBitrate = int32(e.opts.RateMax)
				if e.opts.VbvBufferSize != 0 {
					param.Rc.IVbvBufferSize = int32(e.opts.VbvBufferSize)
				} else {
					param.Rc.IVbvBufferSize = int32(e.opts.RateMax)
				}
			}
		}
	}

	var picIn x264c.Picture
	x264c.PictureInit(&picIn)
	e.picIn = picIn

	e.e = x264c.EncoderOpen(&param)
	if e.e == nil {
		err = fmt.Errorf("x264: cannot open the encoder")
		return
	}

	ret := x264c.EncoderHeaders(e.e, e.nals, &e.nnals)
	if ret < 0 {
		err = fmt.Errorf("x264: cannot encode headers")
		return
	}

	if ret > 0 {
		b := C.GoBytes(e.nals[0].PPayload, C.int(ret))
		n, er := e.w.Write(b)
		if er != nil {
			err = er
			return
		}

		if int(ret) != n {
			err = fmt.Errorf("x264: error writing headers, size=%d, n=%d", ret, n)
			return
		}
	}

	// Persistent C-side plane buffers, reused by every Encode and freed in Close.
	e.planes[0] = C.malloc(C.size_t(e.ySize))
	e.planes[1] = C.malloc(C.size_t(e.cSize))
	e.planes[2] = C.malloc(C.size_t(e.cSize))

	return
}

// Encode encodes image.
func (e *Encoder) Encode(im image.Image) (err error) {
	var picOut x264c.Picture

	// Planes to encode: caller's image for *YCbCr (read-only), else e.img scratch.
	var src *YCbCr

	switch m := im.(type) {
	case *YCbCr:
		src = m
	case *image.RGBA:
		buf := e.proc.Process(m).Get()
		e.img.Y = buf[:e.ySize]
		e.img.Cb = buf[e.ySize : e.ySize+e.cSize]
		e.img.Cr = buf[e.ySize+e.cSize:]
		src = e.img
	default:
		e.img.ToYCbCrDraw(im)
		src = e.img
	}

	copy(unsafe.Slice((*byte)(e.planes[0]), e.ySize), src.Y)
	copy(unsafe.Slice((*byte)(e.planes[1]), e.cSize), src.Cb)
	copy(unsafe.Slice((*byte)(e.planes[2]), e.cSize), src.Cr)

	picIn := e.picIn

	picIn.Img.ICsp = e.csp

	picIn.Img.IPlane = 3
	picIn.Img.IStride[0] = int32(e.opts.Width)
	picIn.Img.IStride[1] = int32(e.opts.Width) / 2
	picIn.Img.IStride[2] = int32(e.opts.Width) / 2

	picIn.Img.Plane[0] = e.planes[0]
	picIn.Img.Plane[1] = e.planes[1]
	picIn.Img.Plane[2] = e.planes[2]

	picIn.IPts = e.pts
	e.pts++

	ret := x264c.EncoderEncode(e.e, e.nals, &e.nnals, &picIn, &picOut)
	if ret < 0 {
		err = fmt.Errorf("x264: cannot encode picture")
		return
	}

	if ret > 0 {
		b := C.GoBytes(e.nals[0].PPayload, C.int(ret))

		n, er := e.w.Write(b)
		if er != nil {
			err = er
			return
		}

		if int(ret) != n {
			err = fmt.Errorf("x264: error writing payload, size=%d, n=%d", ret, n)
		}
	}

	e.dts = picOut.IDts

	return
}

// Timestamp returns the current PTS and DTS.
func (e *Encoder) Timestamp() (int64, int64) {
	return e.pts, e.dts
}

// Flush flushes encoder.
func (e *Encoder) Flush() (err error) {
	var picOut x264c.Picture

	for x264c.EncoderDelayedFrames(e.e) > 0 {
		ret := x264c.EncoderEncode(e.e, e.nals, &e.nnals, nil, &picOut)
		if ret < 0 {
			err = fmt.Errorf("x264: cannot encode picture")
			return
		}

		if ret > 0 {
			b := C.GoBytes(e.nals[0].PPayload, C.int(ret))

			n, er := e.w.Write(b)
			if er != nil {
				err = er
				return
			}

			if int(ret) != n {
				err = fmt.Errorf("x264: error writing payload, size=%d, n=%d", ret, n)
			}
		}
	}

	return
}

// Close closes encoder.
func (e *Encoder) Close() error {
	picIn := e.picIn
	x264c.PictureClean(&picIn)
	x264c.EncoderClose(e.e)

	C.free(e.planes[0])
	C.free(e.planes[1])
	C.free(e.planes[2])

	return nil
}
