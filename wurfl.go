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
	"strconv"
	"sync"
	"unsafe"
)

type Wurfl struct {
	wurfl C.wurfl_handle
	mu    sync.Mutex
}

type Device struct {
	Device              string                 `json:"device"`
	Capabilities        map[string]interface{} `json:"capabilities"`
	VirtualCapabilities map[string]interface{} `json:"virtual"`
}

func New(wurflxml string, patches ...string) (*Wurfl, error) {
	w := &Wurfl{}

	w.wurfl = C.wurfl_create()
	C.wurfl_set_engine_target(w.wurfl, C.WURFL_ENGINE_TARGET_HIGH_ACCURACY)
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

func concreteProperty(val string) interface{} {

	if val == "true" || val == "false" {
		return val == "true"
	}

	// check for numbers
	n, err := strconv.Atoi(val)
	if err == nil {
		return n
	}

	return val
}

func (w *Wurfl) LookupProperties(useragent string, proplist []string, vproplist []string) *Device {

	w.mu.Lock()
	defer w.mu.Unlock()

	ua := C.CString(useragent)
	device := C.wurfl_lookup_useragent(w.wurfl, ua)
	C.free(unsafe.Pointer(ua))

	if device == nil {
		return nil
	}

	m := make(map[string]interface{})

	for _, prop := range proplist {
		cprop := C.CString(prop)
		val := C.wurfl_device_get_capability(device, cprop)
		C.free(unsafe.Pointer(cprop))
		m[prop] = concreteProperty(C.GoString(val))
	}

	// get the virtual properties
	mv := make(map[string]interface{})
	for _, prop := range vproplist {
		cprop := C.CString(prop)
		val := C.wurfl_device_get_virtual_capability(device, cprop)
		C.free(unsafe.Pointer(cprop))
		mv[prop] = concreteProperty(C.GoString(val))
	}

	d := &Device{
		Device:              C.GoString(C.wurfl_device_get_id(device)),
		Capabilities:        m,
		VirtualCapabilities: mv,
	}
	C.wurfl_device_destroy(device)

	return d
}

func wurfl_capability_enumerate(enumerator C.wurfl_device_capability_enumerator_handle) map[string]interface{} {

	m := make(map[string]interface{})

	for C.wurfl_device_capability_enumerator_is_valid(enumerator) != 0 {
		name := C.wurfl_device_capability_enumerator_get_name(enumerator)
		val := C.wurfl_device_capability_enumerator_get_value(enumerator)

		if name != nil && val != nil {
			gname := C.GoString(name)
			gval := C.GoString(val)
			m[gname] = concreteProperty(gval)
		}

		C.wurfl_device_capability_enumerator_move_next(enumerator)
	}

	return m
}

func (w *Wurfl) Lookup(useragent string) *Device {

	w.mu.Lock()
	defer w.mu.Unlock()

	ua := C.CString(useragent)
	device := C.wurfl_lookup_useragent(w.wurfl, ua)
	C.free(unsafe.Pointer(ua))

	if device == nil {
		return nil
	}

	enumerator := C.wurfl_device_get_capability_enumerator(device)
	m := wurfl_capability_enumerate(enumerator)
	C.wurfl_device_capability_enumerator_destroy(enumerator)

	enumerator = C.wurfl_device_get_virtual_capability_enumerator(device)
	mv := wurfl_capability_enumerate(enumerator)
	C.wurfl_device_capability_enumerator_destroy(enumerator)

	d := &Device{
		Device:              C.GoString(C.wurfl_device_get_id(device)),
		Capabilities:        m,
		VirtualCapabilities: mv,
	}
	C.wurfl_device_destroy(device)

	return d
}
