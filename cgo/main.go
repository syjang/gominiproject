package main

// #cgo CXXFLAGS: -std=c++11 -I.. -O2 -fomit-frame-pointer -Wall
// #include "test.h"
// #include <stdlib.h>
import "C"

import (
	"math/rand"
	"unsafe"
)

func init() {
	// names := make([]string, 5)
	// names[0] = "asdfsdf"
	// names[1] = "qwqweqwe1"
	// names[2] = "qwqweqwe2"
	// names[3] = "qwqweqwe3"
	// names[4] = "qwqweqwe4"

	// // var features [2][512]float32
	// // features := make([][]float32, 2, 512)
	// var features [][512]float32
	// var t [512]float32
	// for i := 0; i < 512; i++ {
	// 	t[i] = (float32)((float32)(i) / 512.0)
	// }
	// features = append(features, t)

	// arr := make([]*C.char, len(names))
	// for i, s := range names {
	// 	cs := C.CString(s)
	// 	defer C.free(unsafe.Pointer(cs))
	// 	arr[i] = cs
	// }

	// test := unsafe.Pointer(&features[0])
	// C.test(&arr[0], C.int(len(names)), (**C.float)(test), C.int(1))
}

func ThisIsSecond() []int {
	return []int{1, 2, 3}
}

func main() {
	var fixed [6]float32

	for i := range fixed {
		fixed[i] = rand.Float32()
	}
	invokeC([]string{"anton", "bruni", "ceaser", "dora"}, fixed[:6])

	ttt := C.app()
	tta := C.GoString(ttt)
	print(tta)

}

func invokeC(names []string, features []float32) {
	cnames := makeCNames(names)
	defer func() {
		for _, ptr := range cnames {
			C.free(unsafe.Pointer(ptr))
		}
	}()

	cfeatures := makeCFeatures(features, []uint16{1, 2, 3})
	defer C.free(cfeatures)

	C.test(
		unsafe.Pointer(&cnames[0]), C.int(len(cnames)),
		cfeatures, C.int(len(features)))

}

func makeCNames(names []string) (list []*C.char) {
	for i := range names {
		list = append(list, C.CString(names[i]))
	}
	return
}

func makeCFeatures(features []float32, strides []uint16) (cfeatures unsafe.Pointer) {
	total := uint16(0)
	for i := range strides {
		total += strides[i]
	}

	clen := C.ulong(total * 4) // sizeof(float32) * accumulate(strides)
	ptr := C.malloc(clen)
	src := unsafe.Pointer(&features[0])
	C.memcpy(ptr, src, clen)
	cfeatures = unsafe.Pointer(ptr)
	return
}
