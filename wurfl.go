package wurfl

/*
// http://www.scientiamobile.com/docs/c-api-user-guide.html/
#cgo LDFLAGS: -lwurfl -L/usr/local/lib

#include <stdlib.h>
#include <wurfl/wurfl.h>
*/
import "C"

import (
	"sync"
	"unsafe"
)

type Wurfl struct {
	wurfl C.wurfl_handle
	mu    sync.Mutex
}

func New(wurflxml string) *Wurfl {
	w := &Wurfl{}

	w.wurfl = C.wurfl_create()
	C.wurfl_set_engine_target(w.wurfl, C.WURFL_ENGINE_TARGET_HIGH_PERFORMANCE)
	C.wurfl_set_cache_provider(w.wurfl, C.WURFL_CACHE_PROVIDER_DOUBLE_LRU, C.CString("10000,3000"))
	wxml := C.CString(wurflxml)
	C.wurfl_set_root(w.wurfl, wxml)
	C.free(unsafe.Pointer(wxml))

	if C.wurfl_load(w.wurfl) != C.WURFL_OK {
		return nil
	}

	return w
}

func (w *Wurfl) Lookup(useragent string) map[string]string {

	w.mu.Lock()
	defer w.mu.Unlock()

	device := C.wurfl_lookup_useragent(w.wurfl, C.CString(useragent))

	if device == nil {
		return nil
	}

	m := make(map[string]string)

	enumerator := C.wurfl_device_get_capability_enumerator(device)

	for C.wurfl_device_capability_enumerator_is_valid(enumerator) != 0 {
		name := C.wurfl_device_capability_enumerator_get_name(enumerator)
		val := C.wurfl_device_capability_enumerator_get_value(enumerator)

		if name != nil && val != nil {
			gname := C.GoString(name)
			gval := C.GoString(val)
			m[gname] = gval
		}

		C.wurfl_device_capability_enumerator_move_next(enumerator)
	}
	C.wurfl_device_destroy(device)

	return m
}
