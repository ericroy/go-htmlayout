package gohl

/*
#cgo CFLAGS: -I./htmlayout/include
#cgo LDFLAGS: ./htmlayout/lib/HTMLayout.lib

#include <stdlib.h>
#include <htmlayout.h>
*/
import "C"

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"reflect"
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
	HLDOM_OK_NOT_HANDLED    = C.HLDOM_OK_NOT_HANDLED

	STATE_LINK       = 0x00000001 // selector :link,    any element having href attribute
	STATE_HOVER      = 0x00000002 // selector :hover,   element is under the cursor, mouse hover  
	STATE_ACTIVE     = 0x00000004 // selector :active,  element is activated, e.g. pressed  
	STATE_FOCUS      = 0x00000008 // selector :focus,   element is in focus  
	STATE_VISITED    = 0x00000010 // selector :visited, aux flag - not used internally now.
	STATE_CURRENT    = 0x00000020 // selector :current, current item in collection, e.g. current <option> in <select>
	STATE_CHECKED    = 0x00000040 // selector :checked, element is checked (or selected), e.g. check box or itme in multiselect
	STATE_DISABLED   = 0x00000080 // selector :disabled, element is disabled, behavior related flag.
	STATE_READONLY   = 0x00000100 // selector :read-only, element is read-only, behavior related flag.
	STATE_EXPANDED   = 0x00000200 // selector :expanded, element is in expanded state - nodes in tree view e.g. <options> in <select>
	STATE_COLLAPSED  = 0x00000400 // selector :collapsed, mutually exclusive with EXPANDED
	STATE_INCOMPLETE = 0x00000800 // selector :incomplete, element has images (back/fore/bullet) requested but not delivered.
	STATE_ANIMATING  = 0x00001000 // selector :animating, is currently animating 
	STATE_FOCUSABLE  = 0x00002000 // selector :focusable, shall accept focus
	STATE_ANCHOR     = 0x00004000 // selector :anchor, first element in selection (<select miltiple>), STATE_CURRENT is the current.
	STATE_SYNTHETIC  = 0x00008000 // selector :synthetic, synthesized DOM elements - e.g. all missed cells in tables (<td>) are getting this flag
	STATE_OWNS_POPUP = 0x00010000 // selector :owns-popup, anchor(owner) element of visible popup. 
	STATE_TABFOCUS   = 0x00020000 // selector :tab-focus, element got focus by tab traversal. engine set it together with :focus.
	STATE_EMPTY      = 0x00040000 // selector :empty - element is empty. 
	STATE_BUSY       = 0x00080000 // selector :busy, element is busy. HTMLayoutRequestElementData will set this flag if
	// external data was requested for the element. When data will be delivered engine will reset this flag on the element. 

	STATE_DRAG_OVER   = 0x00100000 // drag over the block that can accept it (so is current drop target). Flag is set for the drop target block. At any given moment of time it can be only one such block.
	STATE_DROP_TARGET = 0x00200000 // active drop target. Multiple elements can have this flag when D&D is active. 
	STATE_MOVING      = 0x00400000 // dragging/moving - the flag is set for the moving element (copy of the drag-source).
	STATE_COPYING     = 0x00800000 // dragging/copying - the flag is set for the copying element (copy of the drag-source).
	STATE_DRAG_SOURCE = 0x00C00000 // is set in element that is being dragged.

	STATE_POPUP   = 0x40000000 // this element is in popup state and presented to the user - out of flow now
	STATE_PRESSED = 0x04000000 // pressed - close to active but has wider life span - e.g. in MOUSE_UP it 
	// is still on, so behavior can check it in MOUSE_UP to discover CLICK condition.
	STATE_HAS_CHILDREN = 0x02000000 // has more than one child.    
	STATE_HAS_CHILD    = 0x01000000 // has single child.

	STATE_IS_LTR = 0x20000000 // selector :ltr, the element or one of its nearest container has @dir and that dir has "ltr" value
	STATE_IS_RTL = 0x10000000 // selector :rtl, the element or one of its nearest container has @dir and that dir has "rtl" value    

	RESET_STYLE_THIS = 0x0020 // reset styles - this may require if you have styles dependent from attributes,
	RESET_STYLE_DEEP = 0x0010 // use these flags after SetAttribute then. RESET_STYLE_THIS is faster than RESET_STYLE_DEEP.
	MEASURE_INPLACE  = 0x0001 // use this flag if you do not expect any dimensional changes - this is faster than REMEASURE
	MEASURE_DEEP     = 0x0002 // use this flag if changes of some attributes/content may cause change of dimensions of the element  
	REDRAW_NOW       = 0x8000

	BAD_HELEMENT = HELEMENT(unsafe.Pointer(uintptr(0)))
)

var errorToString = map[HLDOM_RESULT]string{
	HLDOM_OK:                "HLDOM_OK",
	HLDOM_INVALID_HWND:      "HLDOM_INVALID_HWND",
	HLDOM_INVALID_HANDLE:    "HLDOM_INVALID_HANDLE",
	HLDOM_PASSIVE_HANDLE:    "HLDOM_PASSIVE_HANDLE",
	HLDOM_INVALID_PARAMETER: "HLDOM_INVALID_PARAMETER",
	HLDOM_OPERATION_FAILED:  "HLDOM_OPERATION_FAILED",
	HLDOM_OK_NOT_HANDLED:    "HLDOM_OK_NOT_HANDLED",
}

var whitespaceSplitter = regexp.MustCompile(`(\S+)`)

// DomError represents an htmlayout error with an associated
// dom error code
type DomError struct {
	Result  HLDOM_RESULT
	Message string
}

func (e *DomError) Error() string {
	return fmt.Sprintf("%s: %s", errorToString[e.Result], e.Message)
}

func domResultAsString(result HLDOM_RESULT) string {
	return errorToString[result]
}

func domPanic(result C.HLDOM_RESULT, message ...interface{}) {
	panic(DomError{HLDOM_RESULT(result), fmt.Sprint(message...)})
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
		panic("null cstring")
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
func NewElementFromHandle(h HELEMENT) *Element {
	if h == BAD_HELEMENT {
		panic("Nil helement")
	}
	e := &Element{BAD_HELEMENT}
	e.setHandle(h)
	runtime.SetFinalizer(e, (*Element).finalize)
	return e
}

func NewElement(tagName string) *Element {
	var handle HELEMENT = BAD_HELEMENT
	szName := C.CString(tagName)
	defer C.free(unsafe.Pointer(szName))
	if ret := C.HTMLayoutCreateElement((*C.CHAR)(szName), nil, (*C.HELEMENT)(&handle)); ret != HLDOM_OK {
		domPanic(ret, "Failed to create new element")
	}
	return NewElementFromHandle(handle)
}

func RootElement(hwnd uint32) *Element {
	var handle HELEMENT = BAD_HELEMENT
	if ret := C.HTMLayoutGetRootElement(C.HWND(C.HANDLE(uintptr(hwnd))), (*C.HELEMENT)(&handle)); ret != HLDOM_OK {
		domPanic(ret, "Failed to get root element")
	}
	return NewElementFromHandle(handle)
}

func FocusedElement(hwnd uint32) *Element {
	var handle HELEMENT = BAD_HELEMENT
	if ret := C.HTMLayoutGetFocusElement(C.HWND(C.HANDLE(uintptr(hwnd))), (*C.HELEMENT)(&handle)); ret != HLDOM_OK {
		domPanic(ret, "Failed to get focus element")
	}
	if handle != BAD_HELEMENT {
		return NewElementFromHandle(handle)
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
	return other != nil && e.handle == other.handle
}

// This is the same as AttachHandler, except that behaviors are singleton instances stored
// in a master map.  They may be shared among many elements since they have no state.
// The only reason we keep a separate set of the behaviors is so that the event handler
// dispatch method can tell if an event handler is a behavior or a regular handler.
func (e *Element) attachBehavior(handler *EventHandler) {
	tag := uintptr(unsafe.Pointer(handler))
	behaviors[tag] = handler
	if subscription := handler.Subscription(); subscription == HANDLE_ALL {
		if ret := C.HTMLayoutAttachEventHandler(e.handle, (*[0]byte)(unsafe.Pointer(goElementProc)), C.LPVOID(tag)); ret != HLDOM_OK {
			domPanic(ret, "Failed to attach event handler to element")
		}
	} else {
		if ret := C.HTMLayoutAttachEventHandlerEx(e.handle, (*[0]byte)(unsafe.Pointer(goElementProc)), C.LPVOID(tag), C.UINT(subscription)); ret != HLDOM_OK {
			domPanic(ret, "Failed to attach event handler to element")
		}
	}
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
		if ret := C.HTMLayoutAttachEventHandlerEx(e.handle, (*[0]byte)(unsafe.Pointer(goElementProc)), C.LPVOID(tag), C.UINT(subscription)); ret != HLDOM_OK {
			domPanic(ret, "Failed to attach event handler to element")
		}
	}
}

func (e *Element) DetachHandler(handler *EventHandler) {
	tag := uintptr(unsafe.Pointer(handler))
	if _, exists := eventHandlers[tag]; exists {
		if ret := C.HTMLayoutDetachEventHandler(e.handle, (*[0]byte)(unsafe.Pointer(goElementProc)), C.LPVOID(tag)); ret != HLDOM_OK {
			domPanic(ret, "Failed to detach event handler from element")
		}
		delete(eventHandlers, tag)
	} else {
		panic("cannot detach, handler was not registered")
	}
}

func (e *Element) Update(restyle, restyleDeep, remeasure, remeasureDeep, render bool) {
	var flags uint32
	if restyle {
		if restyleDeep {
			flags |= RESET_STYLE_DEEP
		} else {
			flags |= RESET_STYLE_THIS
		}
	}
	if remeasure {
		if remeasureDeep {
			flags |= MEASURE_DEEP
		} else {
			flags |= MEASURE_INPLACE
		}
	}
	if render {
		flags |= REDRAW_NOW
	}
	if ret := C.HTMLayoutUpdateElementEx(e.handle, C.UINT(flags)); ret != HLDOM_OK {
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
		return NewElementFromHandle(HELEMENT(parent))
	}
	return nil
}

func (e *Element) SelectParent(selector string) *Element {
	return e.SelectParentLimit(selector, 0)
}

// For delivering programmatic events to this element.
// Returns true if the event was handled, false otherwise
func (e *Element) SendEvent(eventCode uint, source *Element, reason uint32) bool {
	var handled C.BOOL = 0
	if ret := C.HTMLayoutSendEvent(e.handle, C.UINT(eventCode), source.handle, C.UINT_PTR(reason), &handled); ret != HLDOM_OK {
		domPanic(ret, "Failed to send event")
	}
	return handled != 0
}

// For asynchronously delivering programmatic events to this element.
func (e *Element) PostEvent(eventCode uint, source *Element, reason uint32) {
	if ret := C.HTMLayoutPostEvent(e.handle, C.UINT(eventCode), source.handle, C.UINT(reason)); ret != HLDOM_OK {
		domPanic(ret, "Failed to post event")
	}
}

//
// DOM structure accessors/modifiers:
//

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
	return NewElementFromHandle(HELEMENT(child))
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
		return NewElementFromHandle(HELEMENT(parent))
	}
	return nil
}

func (e *Element) InsertChild(child *Element, index uint) {
	if ret := C.HTMLayoutInsertElement(child.handle, e.handle, C.UINT(index)); ret != HLDOM_OK {
		domPanic(ret, "Failed to insert child element at index: ", index)
	}
}

func (e *Element) AppendChild(child *Element) {
	count := e.ChildCount()
	if ret := C.HTMLayoutInsertElement(child.handle, e.handle, C.UINT(count)); ret != HLDOM_OK {
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
	return NewElementFromHandle(HELEMENT(clone))
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
	if ret := C.HTMLayoutGetElementHtml(e.handle, (*C.LPBYTE)(unsafe.Pointer(&data)), C.BOOL(0)); ret != HLDOM_OK {
		domPanic(ret, "Failed to get inner html")
	}
	return C.GoString(data)
}

func (e *Element) OuterHtml() string {
	var data *C.char
	if ret := C.HTMLayoutGetElementHtml(e.handle, (*C.LPBYTE)(unsafe.Pointer(&data)), C.BOOL(1)); ret != HLDOM_OK {
		domPanic(ret, "Failed to get outer html")
	}
	return C.GoString(data)
}

func (e *Element) Type() string {
	var data *C.char
	if ret := C.HTMLayoutGetElementType(e.handle, (*C.LPCSTR)(unsafe.Pointer(&data))); ret != HLDOM_OK {
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

//
// HTML attribute accessors/modifiers:
//

// Returns the value of attr and a boolean indicating whether or not that attr exists.
// If the boolean is true, then the returned string is valid.
func (e *Element) Attr(key string) (string, bool) {
	szValue := (*C.WCHAR)(nil)
	szKey := C.CString(key)
	defer C.free(unsafe.Pointer(szKey))
	if ret := C.HTMLayoutGetAttributeByName(e.handle, (*C.CHAR)(szKey), (*C.LPCWSTR)(&szValue)); ret != HLDOM_OK {
		domPanic(ret, "Failed to get attribute: ", key)
	}
	if szValue != nil {
		return utf16ToString((*uint16)(szValue)), true
	}
	return "", false
}

func (e *Element) AttrAsFloat(key string) (float64, bool, error) {
	var f float64
	var err error
	if s, exists := e.Attr(key); !exists {
		return 0.0, false, nil
	} else if f, err = strconv.ParseFloat(s, 64); err != nil {
		return 0.0, true, err
	}
	return float64(f), true, nil
}

func (e *Element) AttrAsInt(key string) (int, bool, error) {
	var i int
	var err error
	if s, exists := e.Attr(key); !exists {
		return 0, false, nil
	} else if i, err = strconv.Atoi(s); err != nil {
		return 0, true, err
	}
	return i, true, nil
}

func (e *Element) SetAttr(key string, value interface{}) {
	szKey := C.CString(key)
	defer C.free(unsafe.Pointer(szKey))
	var ret C.HLDOM_RESULT = HLDOM_OK
	switch v := value.(type) {
	case string:
		ret = C.HTMLayoutSetAttributeByName(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(v)))
	case float32:
		ret = C.HTMLayoutSetAttributeByName(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(strconv.FormatFloat(float64(v), 'g', -1, 64))))
	case float64:
		ret = C.HTMLayoutSetAttributeByName(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(strconv.FormatFloat(float64(v), 'g', -1, 64))))
	case int:
		ret = C.HTMLayoutSetAttributeByName(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(strconv.Itoa(v))))
	case int32:
		ret = C.HTMLayoutSetAttributeByName(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(strconv.FormatInt(int64(v), 10))))
	case int64:
		ret = C.HTMLayoutSetAttributeByName(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(stringToUtf16Ptr(strconv.FormatInt(v, 10))))
	case nil:
		ret = C.HTMLayoutSetAttributeByName(e.handle, (*C.CHAR)(szKey), nil)
	default:
		panic(fmt.Sprintf("Don't know how to format this argument type: %s", reflect.TypeOf(v)))
	}
	if ret != HLDOM_OK {
		domPanic(ret, "Failed to set attribute: "+key)
	}
}

func (e *Element) RemoveAttr(key string) {
	e.SetAttr(key, nil)
}

func (e *Element) AttrByIndex(index int) (string, string) {
	szValue := (*C.WCHAR)(nil)
	szName := (*C.CHAR)(nil)
	if ret := C.HTMLayoutGetNthAttribute(e.handle, C.UINT(index), (*C.LPCSTR)(&szName), (*C.LPCWSTR)(&szValue)); ret != HLDOM_OK {
		domPanic(ret, fmt.Sprintf("Failed to get attribute by index: %u", index))
	}
	return C.GoString((*C.char)(szName)), utf16ToString((*uint16)(szValue))
}

func (e *Element) AttrCount() uint {
	var count C.UINT = 0
	if ret := C.HTMLayoutGetAttributeCount(e.handle, &count); ret != HLDOM_OK {
		domPanic(ret, "Failed to get attribute count")
	}
	return uint(count)
}

//
// CSS style attribute accessors/mutators
//

// Returns the value of the style and a boolean indicating whether or not that style exists.
// If the boolean is true, then the returned string is valid.
func (e *Element) Style(key string) (string, bool) {
	szValue := (*C.WCHAR)(nil)
	szKey := C.CString(key)
	defer C.free(unsafe.Pointer(szKey))
	if ret := C.HTMLayoutGetStyleAttribute(e.handle, (*C.CHAR)(szKey), (*C.LPCWSTR)(&szValue)); ret != HLDOM_OK {
		domPanic(ret, "Failed to get style: "+key)
	}
	if szValue != nil {
		return utf16ToString((*uint16)(szValue)), true
	}
	return "", false
}

func (e *Element) SetStyle(key string, value interface{}) {
	szKey := C.CString(key)
	defer C.free(unsafe.Pointer(szKey))
	var valuePtr *uint16 = nil

	switch v := value.(type) {
	case string:
		valuePtr = stringToUtf16Ptr(v)
	case float32:
		valuePtr = stringToUtf16Ptr(strconv.FormatFloat(float64(v), 'g', -1, 64))
	case float64:
		valuePtr = stringToUtf16Ptr(strconv.FormatFloat(float64(v), 'g', -1, 64))
	case int:
		valuePtr = stringToUtf16Ptr(strconv.Itoa(v))
	case int32:
		valuePtr = stringToUtf16Ptr(strconv.FormatInt(int64(v), 10))
	case int64:
		valuePtr = stringToUtf16Ptr(strconv.FormatInt(v, 10))
	case nil:
		valuePtr = nil
	default:
		panic(fmt.Sprintf("Don't know how to format this argument type: %s", reflect.TypeOf(v)))
	}

	if ret := C.HTMLayoutSetStyleAttribute(e.handle, (*C.CHAR)(szKey), (*C.WCHAR)(valuePtr)); ret != HLDOM_OK {
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

//
// Element state manipulation
//

// Gets the whole set of state flags for this element
func (e *Element) GetStateFlags() uint32 {
	var state C.UINT
	if ret := C.HTMLayoutGetElementState(e.handle, &state); ret != HLDOM_OK {
		domPanic(ret, "Failed to get element state flags")
	}
	return uint32(state)
}

// Replaces the whole set of state flags with the specified value
func (e *Element) SetStateFlags(flags uint32) {
	shouldUpdate := C.BOOL(1)
	if ret := C.HTMLayoutSetElementState(e.handle, C.UINT(flags), C.UINT(^flags), shouldUpdate); ret != HLDOM_OK {
		domPanic(ret, "Failed to set element state flags")
	}
}

// Returns true if the specified flag is "on"
func (e *Element) GetState(flag uint32) bool {
	return e.GetStateFlags()&flag != 0
}

// Sets the specified flag to "on" or "off" according to the value of the provided boolean
func (e *Element) SetState(flag uint32, on bool) {
	addBits := uint32(0)
	clearBits := uint32(0)
	if on {
		addBits = flag
	} else {
		clearBits = flag
	}
	shouldUpdate := C.BOOL(1)
	if ret := C.HTMLayoutSetElementState(e.handle, C.UINT(addBits), C.UINT(clearBits), shouldUpdate); ret != HLDOM_OK {
		domPanic(ret, "Failed to set element state flag")
	}
}

//
// Functions for retrieving the various dimensions of an element
//

func (e *Element) getRect(rectTypeFlags uint32) (left, top, right, bottom int) {
	r := Rect{}
	if ret := C.HTMLayoutGetElementLocation(e.handle, (C.LPRECT)(unsafe.Pointer(&r)), C.UINT(rectTypeFlags)); ret != HLDOM_OK {
		domPanic(ret, "Failed to get element rect")
	}
	return int(r.Left), int(r.Top), int(r.Right), int(r.Bottom)
}

func (e *Element) ContentBox() (left, top, right, bottom int) {
	return e.getRect(CONTENT_BOX)
}

func (e *Element) ContentBoxSize() (width, height int) {
	l, t, r, b := e.getRect(CONTENT_BOX)
	return int(r - l), int(b - t)
}

func (e *Element) PaddingBox() (left, top, right, bottom int) {
	return e.getRect(PADDING_BOX)
}

func (e *Element) PaddingBoxSize() (width, height int) {
	l, t, r, b := e.getRect(PADDING_BOX)
	return int(r - l), int(b - t)
}

func (e *Element) BorderBox() (left, top, right, bottom int) {
	return e.getRect(BORDER_BOX)
}

func (e *Element) BorderBoxSize() (width, height int) {
	l, t, r, b := e.getRect(BORDER_BOX)
	return int(r - l), int(b - t)
}

func (e *Element) MarginBox() (left, top, right, bottom int) {
	return e.getRect(MARGIN_BOX)
}

func (e *Element) MarginBoxSize() (width, height int) {
	l, t, r, b := e.getRect(MARGIN_BOX)
	return int(r - l), int(b - t)
}

//
// The following are not strictly wrappers of htmlayout functions, but rather convenience 
// functions that are helpful in common use cases
//

func (e *Element) Describe() string {
	s := e.Type()
	if value, exists := e.Attr("id"); exists {
		s += "#" + value
	}
	if value, exists := e.Attr("class"); exists {
		values := strings.Split(value, " ")
		for _, v := range values {
			s += "." + v
		}
	}
	return s
}

// Returns the first of the child elements matching the selector.  If no elements
// match, the function panics
func (e *Element) SelectFirst(selector string) *Element {
	results := e.Select(selector)
	if len(results) == 0 {
		panic(fmt.Sprintf("No elements match selector '%s'", selector))
	}
	return results[0]
}

// Returns the only child element that matches the selector.  If no elements match
// or more than one element matches, the function panics
func (e *Element) SelectUnique(selector string) *Element {
	results := e.Select(selector)
	if len(results) == 0 {
		panic(fmt.Sprintf("No elements match selector '%s'", selector))
	} else if len(results) > 1 {
		panic(fmt.Sprintf("More than one element match selector '%s'", selector))
	}
	return results[0]
}

// A wrapper of SelectUnique that auto-prepends a hash to the provided id.
// Useful when selecting elements base on a programmatically retrieved id (which does
// not already have the hash on it)
func (e *Element) SelectId(id string) *Element {
	return e.SelectUnique(fmt.Sprintf("#%s", id))
}

//
// Functions for manipulating the set of classes applied to this element:
//

// Returns true if the specified class is among those listed in the "class" attribute.
func (e *Element) HasClass(class string) bool {
	if classList, exists := e.Attr("class"); !exists {
		return false
	} else if classes := whitespaceSplitter.FindAllString(classList, -1); classes == nil {
		return false
	} else {
		for _, item := range classes {
			if class == item {
				return true
			}
		}
	}
	return false
}

// Adds the specified class to the classes listed in the "class" attribute, or does nothing
// if this class is already included in the list.
func (e *Element) AddClass(class string) {
	if classList, exists := e.Attr("class"); !exists {
		e.SetAttr("class", class)
	} else if classes := whitespaceSplitter.FindAllString(classList, -1); classes == nil {
		e.SetAttr("class", class)
	} else {
		for _, item := range classes {
			if class == item {
				return
			}
		}
		classes = append(classes, class)
		e.SetAttr("class", strings.Join(classes, " "))
	}
}

// Removes the specified class from the classes listed in the "class" attribute, or does nothing
// if this class is not included in the list.
func (e *Element) RemoveClass(class string) {
	if classList, exists := e.Attr("class"); exists {
		if classes := whitespaceSplitter.FindAllString(classList, -1); classes != nil {
			for i, item := range classes {
				if class == item {
					// Delete the item from the list
					classes = append(classes[:i], classes[i+1:]...)
					e.SetAttr("class", strings.Join(classes, " "))
					return
				}
			}
		}
	}
}
