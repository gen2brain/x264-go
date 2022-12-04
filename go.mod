module github.com/gen2brain/x264-go

go 1.19

replace github.com/gen2brain/x264-go/x264c => ./x264c

replace github.com/gen2brain/x264-go/yuv => ./yuv

require (
	github.com/gen2brain/x264-go/x264c v0.0.0-00010101000000-000000000000
	github.com/gen2brain/x264-go/yuv v0.0.0-00010101000000-000000000000
)
