package wurfl

/*
// http://www.scientiamobile.com/docs/c-api-user-guide.html/
#cgo LDFLAGS: -lwurfl -L/usr/local/lib

#include <stdlib.h>
#include <wurfl/wurfl.h>
*/
import "C"

import (
	"errors"
	"sync"
	"unsafe"
)

type Wurfl struct {
	wurfl C.wurfl_handle
	mu    sync.Mutex
}

type Device struct {
	Device       string            `json:"device"`
	Capabilities map[string]string `json:"capabilities"`
}

func New(wurflxml string, patches ...string) (*Wurfl, error) {
	w := &Wurfl{}

	w.wurfl = C.wurfl_create()
	C.wurfl_set_engine_target(w.wurfl, C.WURFL_ENGINE_TARGET_HIGH_ACCURACY)
	C.wurfl_set_cache_provider(w.wurfl, C.WURFL_CACHE_PROVIDER_DOUBLE_LRU, C.CString("10000,3000"))
	wxml := C.CString(wurflxml)
	C.wurfl_set_root(w.wurfl, wxml)

	var freeme []*C.char
	defer func() {
		C.free(unsafe.Pointer(wxml))
		for _, f := range freeme {
			C.free(unsafe.Pointer(f))
		}
	}()

	for _, pxml := range patches {
		p := C.CString(pxml)
		freeme = append(freeme, p)
		C.wurfl_add_patch(w.wurfl, p)
	}

	if C.wurfl_load(w.wurfl) != C.WURFL_OK {
		err := C.wurfl_get_error_message(w.wurfl)
		return nil, errors.New(C.GoString(err))
	}

	return w, nil
}

func (w *Wurfl) Lookup(useragent string) *Device {

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

	d := &Device{
		Device:       C.GoString(C.wurfl_device_get_id(device)),
		Capabilities: m,
	}
	C.wurfl_device_destroy(device)

	return d
}
