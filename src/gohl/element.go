package gohl

/*
#cgo CFLAGS: -I../../htmlayout/include
#cgo LDFLAGS: ../../htmlayout/lib/HTMLayout.lib

#include <stdlib.h>
#include <htmlayout.h>
*/
import "C"

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
	"unicode/utf16"
	"unsafe"
)

const (
	HLDOM_OK                = C.HLDOM_OK
	HLDOM_INVALID_HWND      = C.HLDOM_INVALID_HWND
	HLDOM_INVALID_HANDLE    = C.HLDOM_INVALID_HANDLE
	HLDOM_PASSIVE_HANDLE    = C.HLDOM_PASSIVE_HANDLE
	HLDOM_INVALID_PARAMETER = C.HLDOM_INVALID_PARAMETER
	HLDOM_OPERATION_FAILED  = C.HLDOM_OPERATION_FAILED
	HLDOM_OK_NOT_HANDLED    = C.int(-1)

	BAD_HELEMENT = HELEMENT(unsafe.Pointer(uintptr(0)))
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

func domPanic(result C.HLDOM_RESULT, message ...interface{}) {
	log.Panic(DomError{result, fmt.Sprint(message...)})
}

// Returns the utf-16 encoding of the utf-8 string s,
// with a terminating NUL added.
func stringToUtf16(s string) []uint16 {
	return utf16.Encode([]rune(s + "\x00"))
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
	if h == BAD_HELEMENT {
		panic("Nil helement")
	}
	e := &Element{BAD_HELEMENT}
	e.setHandle(h)
	runtime.SetFinalizer(e, (*Element).finalize)
	return e
}

func RootElement(hwnd uint32) *Element {
	var handle HELEMENT = BAD_HELEMENT
	if ret := C.HTMLayoutGetRootElement(C.HWND(C.HANDLE(uintptr(hwnd))), (*C.HELEMENT)(&handle)); ret != HLDOM_OK {
		domPanic(ret, "Failed to get root element")
	}
	return NewElement(handle)
}

func FocusedElement(hwnd uint32) *Element {
	var handle HELEMENT = BAD_HELEMENT
	if ret := C.HTMLayoutGetFocusElement(C.HWND(C.HANDLE(uintptr(hwnd))), (*C.HELEMENT)(&handle)); ret != HLDOM_OK {
		domPanic(ret, "Failed to get focus element")
	}
	if handle != BAD_HELEMENT {
		return NewElement(handle)
	}
	return nil
}

// Finalizer method, only to be called from Release or by
// the Go runtime
func (e *Element) finalize() {
	// Release the underlying htmlayout handle
	unuse(e.handle)
	e.handle = BAD_HELEMENT
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

func (e *Element) Handle() HELEMENT {
	return e.handle
}

func (e *Element) Equals(other *Element) bool {
	return e.handle == other.handle
}

func (e *Element) AttachHandler(handler *EventHandler) {
	tag := uintptr(unsafe.Pointer(handler))
	if _, exists := eventHandlers[tag]; exists {
		if ret := C.HTMLayoutDetachEventHandler(e.handle, (*[0]byte)(unsafe.Pointer(goElementProc)), C.LPVOID(tag)); ret != HLDOM_OK {
			domPanic(ret, "Failed to detach event handler from element before attaching it again")
		}
	}
	eventHandlers[tag] = handler

	// Don't let the caller disable ATTACH/DETACH events, otherwise we
	// won't know when to throw out our event handler object
	subscription := handler.Subscription()
	subscription &= ^DISABLE_INITIALIZATION

	if subscription == HANDLE_ALL {
		if ret := C.HTMLayoutAttachEventHandler(e.handle, (*[0]byte)(unsafe.Pointer(goElementProc)), C.LPVOID(tag)); ret != HLDOM_OK {
			domPanic(ret, "Failed to attach event handler to element")
		}
	} else {
		if ret := C.HTMLayoutAttachEventHandlerEx(e.handle,  (*[0]byte)(unsafe.Pointer(goElementProc)), C.LPVOID(tag), C.UINT(subscription)); ret != HLDOM_OK {
			domPanic(ret, "Failed to attach event handler to element")
		}
	}
}

func (e *Element) DetachHandler(handler *EventHandler) {
	tag := uintptr(unsafe.Pointer(handler))
	if _, exists := eventHandlers[tag]; exists {
		if ret := C.HTMLayoutDetachEventHandler(e.handle,  (*[0]byte)(unsafe.Pointer(goElementProc)), C.LPVOID(tag)); ret != HLDOM_OK {
			domPanic(ret, "Failed to detach event handler from element")
		}
		delete(eventHandlers, tag)
	} else {
		panic("cannot detach, handler was not registered")
	}
}

func (e *Element) Update(render bool) {
	var shouldRender C.BOOL
	if render {
		shouldRender = C.BOOL(1)
	}
	if ret := C.HTMLayoutUpdateElement(e.handle, shouldRender); ret != HLDOM_OK {
		domPanic(ret, "Failed to update element")
	}
}

func (e *Element) Capture() {
	if ret := C.HTMLayoutSetCapture(e.handle); ret != HLDOM_OK {
		domPanic(ret, "Failed to set capture for element")
	}
}

func (e *Element) ReleaseCapture() {
	if ok := C.ReleaseCapture(); ok == 0 {
		panic("Failed to release capture for element")
	}
}

// Functions for querying elements

func (e *Element) Select(selector string) []*Element {
	szSelector := C.CString(selector)
	defer C.free(unsafe.Pointer(szSelector))
	results := make([]*Element, 0, 32)
	if ret := C.HTMLayoutSelectElements(e.handle, (*C.CHAR)(szSelector), (*[0]byte)(unsafe.Pointer(goSelectCallback)), C.LPVOID(unsafe.Pointer(&results))); ret != HLDOM_OK {
		domPanic(ret, "Failed to select dom elements, selector: '", selector, "'")
	}
	return results
}

// Searches up the parent chain to find the first element that matches the given selector.
// Includes the element in the search.  Depth indicates how far the search should progress.
// Depth = 1 means only consider this element.  Depth = 0 means search all the way up to the
// root.  Any other positive value of depth limits the length of the search.
func (e *Element) SelectParentLimit(selector string, depth int) *Element {
	szSelector := C.CString(selector)
	defer C.free(unsafe.Pointer(szSelector))
	var parent C.HELEMENT
	if ret := C.HTMLayoutSelectParent(e.handle, (*C.CHAR)(szSelector), C.UINT(depth), &parent); ret != HLDOM_OK {
		domPanic(ret, "Failed to select parent dom elements, selector: '", selector, "'")
	}
	if parent != nil {
		return NewElement(HELEMENT(parent))
	}
	return nil
}

func (e *Element) SelectParent(selector string) *Element {
	return e.SelectParentLimit(selector, 0)
}

// For delivering programmatic events to the elements
// Returns true if the event was handled, false otherwise
func (e *Element) SendEvent(destination *Element, eventCode uint, source *Element, reason uintptr) bool {
	var handled C.BOOL = 0
	if ret := C.HTMLayoutSendEvent(destination.handle, C.UINT(eventCode), source.handle, C.UINT_PTR(reason), &handled); ret != HLDOM_OK {
		domPanic(ret, "Failed to send event")
	}
	return handled != 0
}

// DOM structure accessors/modifiers:

func (e *Element) ChildCount() uint {
	var count C.UINT
	if ret := C.HTMLayoutGetChildrenCount(e.handle, &count); ret != HLDOM_OK {
		domPanic(ret, "Failed to get child count")
	}
	return uint(count)
}

func (e *Element) Child(index uint) *Element {
	var child C.HELEMENT
	if ret := C.HTMLayoutGetNthChild(e.handle, C.UINT(index), &child); ret != HLDOM_OK {
		domPanic(ret, "Failed to get child at index: ", index)
	}
	return NewElement(HELEMENT(child))
}

func (e *Element) Children() []*Element {
	slice := make([]*Element, 0, 32)
	for i := uint(0); i < e.ChildCount(); i++ {
		slice = append(slice, e.Child(i))
	}
	return slice
}

func (e *Element) Index() uint {
	var index C.UINT
	if ret := C.HTMLayoutGetElementIndex(e.handle, &index); ret != HLDOM_OK {
		domPanic(ret, "Failed to get element's index")
	}
	return uint(index)
}

func (e *Element) Parent() *Element {
	var parent C.HELEMENT
	if ret := C.HTMLayoutGetParentElement(e.handle, &parent); ret != HLDOM_OK {
		domPanic(ret, "Failed to get parent")
	}
	if parent != nil {
		return NewElement(HELEMENT(parent))
	}
	return nil
}

func (e *Element) InsertChild(child *Element, index uint) {
	if ret := C.HTMLayoutInsertElement(e.handle, child.handle, C.UINT(index)); ret != HLDOM_OK {
		domPanic(ret, "Failed to insert child element at index: ", index)
	}
}

func (e *Element) AppendChild(child *Element) {
	count := e.ChildCount()
	if ret := C.HTMLayoutInsertElement(e.handle, child.handle, C.UINT(count)); ret != HLDOM_OK {
		domPanic(ret, "Failed to append child element")
	}
}

func (e *Element) Detach() {
	if ret := C.HTMLayoutDetachElement(e.handle); ret != HLDOM_OK {
		domPanic(ret, "Failed to detach element from dom")
	}
}

func (e *Element) Delete() {
	if ret := C.HTMLayoutDeleteElement(e.handle); ret != HLDOM_OK {
		domPanic(ret, "Failed to delete element from dom")
	}
	e.finalize()
}

// Makes a deep clone of the receiver, the resulting subtree is not attached to the dom.
func (e *Element) Clone() *Element {
	var clone C.HELEMENT
	if ret := C.HTMLayoutCloneElement(e.handle, &clone); ret != HLDOM_OK {
		domPanic(ret, "Failed to clone element")
	}
	return NewElement(HELEMENT(clone))
}

func (e *Element) Swap(other *Element) {
	if ret := C.HTMLayoutSwapElements(e.handle, other.handle); ret != HLDOM_OK {
		domPanic(ret, "Failed to swap elements")
	}
}

// Sorts 'count' child elements starting at index 'start'.  Uses comparator to define the
// order.  Comparator should return -1, or 0, or 1 to indicate less, equal or greater
func (e *Element) SortChildrenRange(start, count uint, comparator func(*Element, *Element) int) {
	end := start + count
	arg := uintptr(unsafe.Pointer(&comparator))
	if ret := C.HTMLayoutSortElements(e.handle, C.UINT(start), C.UINT(end), (*[0]byte)(unsafe.Pointer(goElementComparator)), C.LPVOID(arg)); ret != HLDOM_OK {
		domPanic(ret, "Failed to sort elements")
	}
}

func (e *Element) SortChildren(comparator func(*Element, *Element) int) {
	e.SortChildrenRange(0, e.ChildCount(), comparator)
}

func (e *Element) SetTimer(ms int) {
	if ret := C.HTMLayoutSetTimer(e.handle, C.UINT(ms)); ret != HLDOM_OK {
		domPanic(ret, "Failed to set timer")
	}
}

func (e *Element) CancelTimer() {
	e.SetTimer(0)
}

func (e *Element) Hwnd() uint32 {
	var hwnd uint32
	if ret := C.HTMLayoutGetElementHwnd(e.handle, (*C.HWND)(unsafe.Pointer(&hwnd)), 0); ret != HLDOM_OK {
		domPanic(ret, "Failed to get element's hwnd")
	}
	return hwnd
}

func (e *Element) RootHwnd() uint32 {
	var hwnd uint32
	if ret := C.HTMLayoutGetElementHwnd(e.handle, (*C.HWND)(unsafe.Pointer(&hwnd)), 1); ret != HLDOM_OK {
		domPanic(ret, "Failed to get element's root hwnd")
	}
	return hwnd
}

func (e *Element) Html() string {
	var data *C.char
	if ret := C.HTMLayoutGetElementHtml(e.handle, (*C.LPBYTE)(unsafe.Pointer(data)), C.BOOL(0)); ret != HLDOM_OK {
		domPanic(ret, "Failed to get inner html")
	}
	return C.GoString(data)
}

func (e *Element) OuterHtml() string {
	var data *C.char
	if ret := C.HTMLayoutGetElementHtml(e.handle, (*C.LPBYTE)(unsafe.Pointer(data)), C.BOOL(1)); ret != HLDOM_OK {
		domPanic(ret, "Failed to get inner html")
	}
	return C.GoString(data)
}

func (e *Element) Type() string {
	var data *C.char
	if ret := C.HTMLayoutGetElementType(e.handle, (*C.LPCSTR)(unsafe.Pointer(data))); ret != HLDOM_OK {
		domPanic(ret, "Failed to get element type")
	}
	return C.GoString(data)
}

func (e *Element) SetHtml(html string) {
	szHtml := C.CString(html)
	defer C.free(unsafe.Pointer(szHtml))
	if ret := C.HTMLayoutSetElementHtml(e.handle, (*C.BYTE)(unsafe.Pointer(szHtml)), C.DWORD(len(html)), SIH_REPLACE_CONTENT); ret != HLDOM_OK {
		domPanic(ret, "Failed to replace element's html")
	}
}

func (e *Element) PrependHtml(prefix string) {
	szHtml := C.CString(prefix)
	defer C.free(unsafe.Pointer(szHtml))
	if ret := C.HTMLayoutSetElementHtml(e.handle, (*C.BYTE)(unsafe.Pointer(szHtml)), C.DWORD(len(prefix)), SIH_INSERT_AT_START); ret != HLDOM_OK {
		domPanic(ret, "Failed to prepend to element's html")
	}
}

func (e *Element) AppendHtml(suffix string) {
	szHtml := C.CString(suffix)
	defer C.free(unsafe.Pointer(szHtml))
	if ret := C.HTMLayoutSetElementHtml(e.handle, (*C.BYTE)(unsafe.Pointer(szHtml)), C.DWORD(len(suffix)), SIH_APPEND_AFTER_LAST); ret != HLDOM_OK {
		domPanic(ret, "Failed to append to element's html")
	}
}

func (e *Element) SetText(text string) {
	szText := C.CString(text)
	defer C.free(unsafe.Pointer(szText))
	if ret := C.HTMLayoutSetElementInnerText(e.handle, (*C.BYTE)(unsafe.Pointer(szText)), C.UINT(len(text))); ret != HLDOM_OK {
		domPanic(ret, "Failed to replace element's text")
	}
}

// HTML attribute accessors/modifiers:

func (e *Element) Attr(key string) *string {
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

func (e *Element) AttrAsFloat(key string) *float32 {
	if s := e.Attr(key); s != nil {
		if f, err := strconv.ParseFloat(*s, 32); err != nil {
			panic(err)
		} else {
			f32 := float32(f)
			return &f32
		}
	}
	return nil
}

func (e *Element) AttrAsInt(key string) *int {
	if s := e.Attr(key); s != nil {
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
	switch v := value.(type) {
	case string:
		ret = C.HTMLayoutSetAttributeByName(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(v)))
	case float32:
		ret = C.HTMLayoutSetAttributeByName(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(strconv.FormatFloat(float64(v), 'e', -1, 32))))
	case int:
		ret = C.HTMLayoutSetAttributeByName(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(strconv.Itoa(v))))
	case nil:
		ret = C.HTMLayoutSetAttributeByName(e.handle, (*C.CHAR)(szKey), nil)
	default:
		log.Panic("Don't know how to format this argument type")
	}
	if ret != HLDOM_OK {
		domPanic(ret, "Failed to set attribute: "+key)
	}
}

func (e *Element) RemoveAttr(key string) {
	e.SetAttr(key, nil)
}

func (e *Element) AttrValueByIndex(index uint) string {
	szValue := (*C.WCHAR)(nil)
	if ret := C.HTMLayoutGetNthAttribute(e.handle, C.UINT(index), nil, (*C.LPCWSTR)(&szValue)); ret != HLDOM_OK {
		domPanic(ret, fmt.Sprintf("Failed to get attribute name by index: %u", index))
	}
	return utf16ToString((*uint16)(szValue))
}

func (e *Element) AttrNameByIndex(index uint) string {
	szName := (*C.CHAR)(nil)
	if ret := C.HTMLayoutGetNthAttribute(e.handle, C.UINT(index), (*C.LPCSTR)(&szName), nil); ret != HLDOM_OK {
		domPanic(ret, fmt.Sprintf("Failed to get attribute name by index: %u", index))
	}
	return C.GoString((*C.char)(szName))
}

func (e *Element) AttrCount(index uint) uint {
	var count C.UINT = 0
	if ret := C.HTMLayoutGetAttributeCount(e.handle, &count); ret != HLDOM_OK {
		domPanic(ret, "Failed to get attribute count")
	}
	return uint(count)
}

// CSS style attribute accessors/mutators

func (e *Element) Style(key string) *string {
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

func (e *Element) StyleAsFloat(key string) *float32 {
	if s := e.Style(key); s != nil {
		if f, err := strconv.ParseFloat(*s, 32); err != nil {
			panic(err)
		} else {
			f32 := float32(f)
			return &f32
		}
	}
	return nil
}

func (e *Element) StyleAsInt(key string) *int {
	if s := e.Style(key); s != nil {
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
	switch v := value.(type) {
	case string:
		ret = C.HTMLayoutSetStyleAttribute(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(v)))
	case float32:
		ret = C.HTMLayoutSetStyleAttribute(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(strconv.FormatFloat(float64(v), 'e', -1, 32))))
	case int:
		ret = C.HTMLayoutSetStyleAttribute(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(strconv.Itoa(v))))
	case nil:
		ret = C.HTMLayoutSetStyleAttribute(e.handle, (*C.CHAR)(szKey), nil)
	default:
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
