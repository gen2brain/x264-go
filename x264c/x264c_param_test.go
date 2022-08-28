package x264c

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestRawParamSize(t *testing.T) {
	if rawParamSize == unsafe.Sizeof(Param{}) {
		return
	}

	t.Fatal(fmt.Sprintf("%v != %v", rawParamSize, unsafe.Sizeof(Param{})))
}
