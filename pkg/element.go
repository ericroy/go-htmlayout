package gohl
/*
#cgo CFLAGS: -I../htmlayout/include
#cgo LDFLAGS: ../htmlayout/lib/HTMLayout.lib

#include <stdlib.h>
#include <htmlayout.h>

extern BOOL CALLBACK ElementProc(LPVOID tag, HELEMENT he, UINT evtg, LPVOID prms );
extern LPELEMENT_EVENT_PROC ElementProcAddr;

*/
import "C"

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
	"unsafe"
	"utf16"
)

const (
	HLDOM_OK                = C.HLDOM_OK
	HLDOM_INVALID_HWND      = C.HLDOM_INVALID_HWND
	HLDOM_INVALID_HANDLE    = C.HLDOM_INVALID_HANDLE
	HLDOM_PASSIVE_HANDLE    = C.HLDOM_PASSIVE_HANDLE
	HLDOM_INVALID_PARAMETER = C.HLDOM_INVALID_PARAMETER
	HLDOM_OPERATION_FAILED  = C.HLDOM_OPERATION_FAILED
	HLDOM_OK_NOT_HANDLED    = C.int(-1)
)

var errorToString = map[C.HLDOM_RESULT]string{
	C.HLDOM_OK:                "HLDOM_OK",
	C.HLDOM_INVALID_HWND:      "HLDOM_INVALID_HWND",
	C.HLDOM_INVALID_HANDLE:    "HLDOM_INVALID_HANDLE",
	C.HLDOM_PASSIVE_HANDLE:    "HLDOM_PASSIVE_HANDLE",
	C.HLDOM_INVALID_PARAMETER: "HLDOM_INVALID_PARAMETER",
	C.HLDOM_OPERATION_FAILED:  "HLDOM_OPERATION_FAILED",
	C.HLDOM_OK_NOT_HANDLED:    "HLDOM_OK_NOT_HANDLED",
}


// DomError represents an htmlayout error with an associated
// dom error code
type DomError struct {
	Result  C.HLDOM_RESULT
	Message string
}

func (self DomError) String() string {
	return fmt.Sprintf("%s: %s", errorToString[self.Result], self.Message)
}

func domPanic(result C.HLDOM_RESULT, message string) {
	log.Panic(DomError{result, message})
}

// Returns the utf-16 encoding of the utf-8 string s,
// with a terminating NUL added.
func stringToUtf16(s string) []uint16 {
	return utf16.Encode([]int(s + "\x00"))
}

// Returns the utf-8 encoding of the utf-16 sequence s,
// with a terminating NUL removed.
func utf16ToString(s *uint16) string {
	if s == nil {
		log.Panic("null cstring")
	}
	us := make([]uint16, 0, 256)
	for p := uintptr(unsafe.Pointer(s)); ; p += 2 {
		u := *(*uint16)(unsafe.Pointer(p))
		if u == 0 {
			return string(utf16.Decode(us))
		}
		us = append(us, u)
	}
	return ""
}

// Returns pointer to the utf-16 encoding of
// the utf-8 string s, with a terminating NUL added.
func stringToUtf16Ptr(s string) *uint16 {
	return &stringToUtf16(s)[0]
}

func use(handle HELEMENT) {
	if dr := C.HTMLayout_UseElement(handle); dr != HLDOM_OK {
		domPanic(dr, "UseElement")
	}
}

func unuse(handle HELEMENT) {
	if handle != nil {
		if dr := C.HTMLayout_UnuseElement(handle); dr != HLDOM_OK {
			domPanic(dr, "UnuseElement")
		}
	}
}

/*
Element

Represents a single DOM element, owns and manages a Handle
*/
type Element struct {
	handle HELEMENT
}

// Constructors
func NewElement(h HELEMENT) *Element {
	e := &Element{nil}
	e.setHandle(h)
	runtime.SetFinalizer(e, (*Element).finalize)
	return e
}

func GetRootElement(hwnd uint32) *Element {
	var handle HELEMENT = nil
	if ret := C.HTMLayoutGetRootElement(C.HWND(C.HANDLE(uintptr(hwnd))), (*C.HELEMENT)(&handle)); ret != HLDOM_OK {
		domPanic(ret, "Failed to get root element")
	}
	return NewElement(handle)
}

// Finalizer method, only to be called from Release or by
// the Go runtime
func (e *Element) finalize() {
	// Release the underlying htmlayout handle
	unuse(e.handle)
	e.handle = nil
}

func (e *Element) Release() {
	// Unregister the finalizer so that it does not get called by Go
	// and then explicitly finalize this element
	runtime.SetFinalizer(e, nil)
	e.finalize()
}

func (e *Element) setHandle(h HELEMENT) {
	use(h)
	unuse(e.handle)
	e.handle = h
}

func (e *Element) GetHandle() HELEMENT {
	return e.handle
}

func (e *Element) AttachHandler(handler EventHandler, subscription uint32) {
	tag := handler.GetAddress()
	if _, exists := eventHandlers[tag]; !exists {
		eventHandlers[tag] = handler
		if ret := C.HTMLayoutAttachEventHandlerEx(e.handle, C.ElementProcAddr, C.LPVOID(tag), C.UINT(subscription)); ret != HLDOM_OK {
			domPanic(ret, "Failed to attach event handler to element")
		}
	}
}

func (e *Element) AttachHandlerAll(handler EventHandler) {
	tag := handler.GetAddress()
	if _, exists := eventHandlers[tag]; !exists {
		eventHandlers[tag] = handler
		if ret := C.HTMLayoutAttachEventHandler(e.handle, C.ElementProcAddr, C.LPVOID(tag)); ret != HLDOM_OK {
			domPanic(ret, "Failed to attach event handler to element")
		}
	}
}

func (e *Element) DetachHandler(handler EventHandler) {
	tag := handler.GetAddress()
	if handler, exists := eventHandlers[tag]; exists {
		if ret := C.HTMLayoutDetachEventHandler(e.handle, C.ElementProcAddr, C.LPVOID(tag)); ret != HLDOM_OK {
			domPanic(ret, "Failed to detach event handler from element")
		}
		eventHandlers[tag] = handler, false
	} else {
		panic("cannot detach, handler was not registered")
	}
}

func (e *Element) SetCapture() {
	if ret := C.HTMLayoutSetCapture(e.handle); ret != HLDOM_OK {
		domPanic(ret, "Failed to SetCapture for element")
	}
}

func (e *Element) ReleaseCapture() {
	if ok := C.ReleaseCapture(); ok == 0 {
		panic("Failed to ReleaseCapture for element");
	}
}

// HTML attribute accessors/modifiers:

func (e *Element) GetAttr(key string) *string {
	szValue := (*C.WCHAR)(nil)
	szKey := C.CString(key)
	defer C.free(unsafe.Pointer(szKey))
	if ret := C.HTMLayoutGetAttributeByName(e.handle, (*C.CHAR)(szKey), (*C.LPCWSTR)(&szValue)); ret != HLDOM_OK {
		domPanic(ret, "Failed to get attribute: "+key)
	}
	if szValue != nil {
		s := utf16ToString((*uint16)(szValue))
		return &s
	}
	return nil
}

func (e *Element) GetAttrAsFloat(key string) *float32 {
	if s := e.GetAttr(key); s != nil {
		if f, err := strconv.Atof32(*s); err != nil {
			panic(err)
		} else {
			return &f
		}
	}
	return nil
}

func (e *Element) GetAttrAsInt(key string) *int {
	if s := e.GetAttr(key); s != nil {
		if i, err := strconv.Atoi(*s); err != nil {
			panic(err)
		} else {
			return &i
		}
	}
	return nil
}

func (e *Element) SetAttr(key string, value interface{}) {
	szKey := C.CString(key)
	defer C.free(unsafe.Pointer(szKey))
	var ret C.HLDOM_RESULT = HLDOM_OK
	if v, ok := value.(string); ok {
		ret = C.HTMLayoutSetAttributeByName(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(v)))
	} else if v, ok := value.(float32); ok {
		ret = C.HTMLayoutSetAttributeByName(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(strconv.Ftoa32(v, 'e', 6))))
	} else if v, ok := value.(int); ok {
		ret = C.HTMLayoutSetAttributeByName(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(strconv.Itoa(v))))
	} else if value == nil {
		ret = C.HTMLayoutSetAttributeByName(e.handle, (*C.CHAR)(szKey), nil)
	} else {
		log.Panic("Don't know how to format this argument type")
	}
	if ret != HLDOM_OK {
		domPanic(ret, "Failed to set attribute: "+key)
	}
}

func (e *Element) RemoveAttr(key string) {
	e.SetAttr(key, nil)
}

func (e *Element) GetAttrValueByIndex(index int) string {
	szValue := (*C.WCHAR)(nil)
	if ret := C.HTMLayoutGetNthAttribute(e.handle, (C.UINT)(index), nil, (*C.LPCWSTR)(&szValue)); ret != HLDOM_OK {
		domPanic(ret, fmt.Sprintf("Failed to get attribute name by index: %d", index))
	}
	return utf16ToString((*uint16)(szValue))
}

func (e *Element) GetAttrNameByIndex(index int) string {
	szName := (*C.CHAR)(nil)
	if ret := C.HTMLayoutGetNthAttribute(e.handle, (C.UINT)(index), (*C.LPCSTR)(&szName), nil); ret != HLDOM_OK {
		domPanic(ret, fmt.Sprintf("Failed to get attribute name by index: %d", index))
	}
	return C.GoString((*C.char)(szName))
}

func (e *Element) GetAttrCount(index int) int {
	var count C.UINT = 0
	if ret := C.HTMLayoutGetAttributeCount(e.handle, &count); ret != HLDOM_OK {
		domPanic(ret, "Failed to get attribute count")
	}
	return int(count)
}

// CSS style attribute accessors/mutators

func (e *Element) GetStyle(key string) *string {
	szValue := (*C.WCHAR)(nil)
	szKey := C.CString(key)
	defer C.free(unsafe.Pointer(szKey))
	if ret := C.HTMLayoutGetStyleAttribute(e.handle, (*C.CHAR)(szKey), (*C.LPCWSTR)(&szValue)); ret != HLDOM_OK {
		domPanic(ret, "Failed to get style: "+key)
	}
	if szValue != nil {
		s := utf16ToString((*uint16)(szValue))
		return &s
	}
	return nil
}

func (e *Element) GetStyleAsFloat(key string) *float32 {
	if s := e.GetStyle(key); s != nil {
		if f, err := strconv.Atof32(*s); err != nil {
			panic(err)
		} else {
			return &f
		}
	}
	return nil
}

func (e *Element) GetStyleAsInt(key string) *int {
	if s := e.GetStyle(key); s != nil {
		if i, err := strconv.Atoi(*s); err != nil {
			panic(err)
		} else {
			return &i
		}
	}
	return nil
}

func (e *Element) SetStyle(key string, value interface{}) {
	szKey := C.CString(key)
	defer C.free(unsafe.Pointer(szKey))
	var ret C.HLDOM_RESULT = HLDOM_OK
	if v, ok := value.(string); ok {
		ret = C.HTMLayoutSetStyleAttribute(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(v)))
	} else if v, ok := value.(float32); ok {
		ret = C.HTMLayoutSetStyleAttribute(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(strconv.Ftoa32(v, 'e', 6))))
	} else if v, ok := value.(int); ok {
		ret = C.HTMLayoutSetStyleAttribute(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(strconv.Itoa(v))))
	} else if value == nil {
		ret = C.HTMLayoutSetStyleAttribute(e.handle, (*C.CHAR)(szKey), nil)
	} else {
		log.Panic("Don't know how to format this argument type")
	}
	if ret != HLDOM_OK {
		domPanic(ret, "Failed to set style: "+key)
	}
}

func (e *Element) RemoveStyle(key string) {
	e.SetStyle(key, nil)
}

func (e *Element) ClearStyles(key string) {
	if ret := C.HTMLayoutSetStyleAttribute(e.handle, nil, nil); ret != HLDOM_OK {
		domPanic(ret, "Failed to clear all styles")
	}
}
