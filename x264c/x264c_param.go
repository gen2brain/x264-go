// +build !compat

package x264c

/*
#include "stdint.h"
#include "x264.h"
*/
import "C"

import "unsafe"

// Param type.
type Param struct {
	// CPU flags.
	Cpu uint32
	// Encode multiple frames in parallel.
	IThreads int32
	// Multiple threads for lookahead analysis.
	ILookaheadThreads int32
	// Whether to use slice-based threading.
	BSlicedThreads int32
	// Whether to allow non-deterministic optimizations when threaded.
	BDeterministic int32
	// Force canonical behavior rather than cpu-dependent optimal algorithms.
	BCpuIndependent int32
	// Threaded lookahead buffer.
	ISyncLookahead int32

	// Video Properties.
	IWidth  int32
	IHeight int32
	// CSP of encoded bitstream.
	ICsp      int32
	IBitdepth int32
	ILevelIdc int32
	// Number of frames to encode if known, else 0.
	IFrameTotal int32

	// NAL HRD.
	// Uses Buffering and Picture Timing SEIs to signal HRD. The HRD in H.264 was not designed with VFR in mind.
	// It is therefore not recommendeded to use NAL HRD with VFR.
	// Furthermore, reconfiguring the VBV (via x264_encoder_reconfig) will currently generate invalid HRD.
	INalHrd int32

	Vui Vui

	// Bitstream parameters.
	// Maximum number of reference frames.
	IFrameReference int32
	// Force a DPB size larger than that implied by B-frames and reference frames.
	// Useful in combination with interactive error resilience.
	IDpbSize int32
	// Force an IDR keyframe at this interval.
	IKeyintMax int32
	// Scenecuts closer together than this are coded as I, not IDR.
	IKeyintMin int32
	// How aggressively to insert extra I frames.
	IScenecutThreshold int32
	// Whether or not to use periodic intra refresh instead of IDR frames.
	BIntraRefresh int32

	// How many b-frame between 2 references pictures.
	IBframe         int32
	IBframeAdaptive int32
	IBframeBias     int32
	// Keep some B-frames as references: 0=off, 1=strict hierarchical, 2=normal.
	IBframePyramid int32
	BOpenGop       int32
	BBlurayCompat  int32
	IAvcintraClass int32

	BDeblockingFilter int32
	// [-6, 6] -6 light filter, 6 strong.
	IDeblockingFilterAlphac0 int32
	// [-6, 6]  idem.
	IDeblockingFilterBeta int32

	BCabac        int32
	ICabacInitIdc int32

	BInterlaced       int32
	BConstrainedIntra int32

	ICqmPreset int32
	_          [4]byte
	// Filename (in UTF-8) of CQM file, JM format.
	PszCqmFile *int8

	// Used only if i_cqm_preset == X264_CQM_CUSTOM.
	Cqm4iy [16]byte
	Cqm4py [16]byte
	Cqm4ic [16]byte
	Cqm4pc [16]byte
	Cqm8iy [64]byte
	Cqm8py [64]byte
	Cqm8ic [64]byte
	Cqm8pc [64]byte

	// Log.
	PfLog       *[0]byte
	PLogPrivate unsafe.Pointer
	ILogLevel   int32
	// Fully reconstruct frames, even when not necessary for encoding. Implied by psz_dump_yuv.
	BFullRecon int32
	// Filename (in UTF-8) for reconstructed frames.
	PszDumpYuv *int8

	// Encoder analyser parameters.
	Analyse Analyse

	_ [4]byte

	// Rate control parameters.
	Rc Rc

	// Cropping Rectangle parameters: added to those implicitly defined by non-mod16 video resolutions.
	CropRect CropRect

	// Frame packing arrangement flag.
	IFramePacking int32

	// Muxing parameters.
	// Generate access unit delimiters.
	BAud int32
	// Put SPS/PPS before each keyframe.
	BRepeatHeaders int32
	// If set, place start codes (4 bytes) before NAL units, otherwise place size (4 bytes) before NAL units.
	BAnnexb int32
	// SPS and PPS id number.
	ISpsId int32
	// VFR input. If 1, use timebase and timestamps for ratecontrol purposes. If 0, use fps only.
	BVfrInput int32
	// Use explicitly set timebase for CFR.
	BPulldown int32
	IFpsNum   uint32
	IFpsDen   uint32
	// Timebase numerator.
	ITimebaseNum uint32
	// Timebase denominator.
	ITimebaseDen uint32

	BTff int32

	// The correct pic_struct must be passed with each input frame.
	// The input timebase should be the timebase corresponding to the output framerate. This should be constant.
	// e.g. for 3:2 pulldown timebase should be 1001/30000.
	// The PTS passed with each frame must be the PTS of the frame after pulldown is applied.
	// Frame doubling and tripling require BVfrInput set to zero (see H.264 Table D-1)
	//
	// Pulldown changes are not clearly defined in H.264. Therefore, it is the calling app's responsibility to manage this.
	BPicStruct int32

	// Used only when b_interlaced=0. Setting this flag makes it possible to flag the stream as PAFF interlaced yet
	// encode all frames progessively. It is useful for encoding 25p and 30p Blu-Ray streams.
	BFakeInterlaced int32

	// Don't optimize header parameters based on video content, e.g. ensure that splitting an input video, compressing
	// each part, and stitching them back together will result in identical SPS/PPS. This is necessary for stitching
	// with container formats that don't allow multiple SPS/PPS.
	BStitchable int32

	// Use OpenCL when available.
	BOpencl int32
	// Specify count of GPU devices to skip, for CLI users.
	IOpenclDevice int32
	_             [4]byte
	// Pass explicit cl_device_id as void*, for API users.
	OpenclDeviceId unsafe.Pointer
	// Filename (in UTF-8) of the compiled OpenCL kernel cache file.
	PszClbinFile *int8

	// Slicing parameters
	// Max size per slice in bytes; includes estimated NAL overhead.
	ISliceMaxSize int32
	// Max number of MBs per slice; overrides i_slice_count.
	ISliceMaxMbs int32
	// Min number of MBs per slice.
	ISliceMinMbs int32
	// Number of slices per frame: forces rectangular slices.
	ISliceCount int32
	// Absolute cap on slices per frame; stops applying slice-max-size and slice-max-mbs if this is reached.
	ISliceCountMax int32

	_           [4]byte
	ParamFree   *[0]byte
	NaluProcess *[0]byte
}

// cptr return C pointer.
func (p *Param) cptr() *C.x264_param_t {
	return (*C.x264_param_t)(unsafe.Pointer(p))
}
