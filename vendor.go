//go:build required

package x264

import (
	_ "github.com/gen2brain/x264-go/x264c/external"
	_ "github.com/gen2brain/x264-go/x264c/external/x264"
	_ "github.com/gen2brain/x264-go/x264c/external/x264/common"
	_ "github.com/gen2brain/x264-go/x264c/external/x264/encoder"
)
