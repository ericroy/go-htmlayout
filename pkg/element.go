package gohl
/*
#cgo CFLAGS: -I../htmlayout/include
#cgo LDFLAGS: ../htmlayout/lib/HTMLayout.lib

#include <stdlib.h>
#include <htmlayout.h>

extern BOOL goMainElementProc(LPVOID, HELEMENT, UINT, LPVOID);

// Main event function that dispatches to the appropriate event handler
BOOL CALLBACK MainElementProc(LPVOID tag, HELEMENT he, UINT evtg, LPVOID prms )
{
	return goMainElementProc(tag, he, evtg, prms);
}
LPELEMENT_EVENT_PROC MainElementProcAddr = &MainElementProc;

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
	HLDOM_OK = C.HLDOM_OK
	HLDOM_INVALID_HWND = C.HLDOM_INVALID_HWND
	HLDOM_INVALID_HANDLE = C.HLDOM_INVALID_HANDLE
	HLDOM_PASSIVE_HANDLE = C.HLDOM_PASSIVE_HANDLE
	HLDOM_INVALID_PARAMETER = C.HLDOM_INVALID_PARAMETER
	HLDOM_OPERATION_FAILED = C.HLDOM_OPERATION_FAILED
	HLDOM_OK_NOT_HANDLED = C.int(-1)

	// EventGroups
	HANDLE_INITIALIZATION = C.HANDLE_INITIALIZATION 	/** attached/detached */
	HANDLE_MOUSE = C.HANDLE_MOUSE						/** mouse events */ 
	HANDLE_KEY = C.HANDLE_KEY							/** key events */  
	HANDLE_FOCUS = C.HANDLE_FOCUS						/** focus events, if this flag is set it also means that element it attached to is focusable */ 
	HANDLE_SCROLL = C.HANDLE_SCROLL						/** scroll events */ 
	HANDLE_TIMER = C.HANDLE_TIMER						/** timer event */ 
	HANDLE_SIZE = C.HANDLE_SIZE							/** size changed event */ 
	HANDLE_DRAW = C.HANDLE_DRAW							/** drawing request (event) */
	HANDLE_DATA_ARRIVED = C.HANDLE_DATA_ARRIVED			/** requested data () has been delivered */
	HANDLE_BEHAVIOR_EVENT = C.HANDLE_BEHAVIOR_EVENT		/** secondary, synthetic events: 
														BUTTON_CLICK, HYPERLINK_CLICK, etc., 
														a.k.a. notifications from intrinsic behaviors */
	HANDLE_METHOD_CALL = C.HANDLE_METHOD_CALL			/** behavior specific methods */
	HANDLE_EXCHANGE = C.HANDLE_EXCHANGE					/** system drag-n-drop */
	HANDLE_GESTURE = C.HANDLE_GESTURE					/** touch input events */
	HANDLE_ALL = C.HANDLE_ALL 							/** all of them */
	DISABLE_INITIALIZATION = C.DISABLE_INITIALIZATION 	/** disable INITIALIZATION events to be sent. */

	// MouseEvents
	MOUSE_ENTER = C.MOUSE_ENTER
	MOUSE_LEAVE = C.MOUSE_LEAVE
	MOUSE_MOVE = C.MOUSE_MOVE
	MOUSE_UP = C.MOUSE_UP
	MOUSE_DOWN = C.MOUSE_DOWN
	MOUSE_DCLICK = C.MOUSE_DCLICK
	MOUSE_WHEEL = C.MOUSE_WHEEL
	MOUSE_TICK = C.MOUSE_TICK		// mouse pressed ticks
	MOUSE_IDLE = C.MOUSE_IDLE		// mouse stay idle for some time

	DROP = C.DROP 					// item dropped, target is that dropped item 
	DRAG_ENTER = C.DRAG_ENTER 		// drag arrived to the target element that is one of current drop targets.  
	DRAG_LEAVE = C.DRAG_LEAVE 		// drag left one of current drop targets. target is the drop target element.  
	DRAG_REQUEST = C.DRAG_REQUEST  	// drag src notification before drag start. To cancel - return true from handler.
	MOUSE_CLICK = C.MOUSE_CLICK		// mouse click event
	DRAGGING = C.DRAGGING			// This flag is 'ORed' with MOUSE_ENTER..MOUSE_DOWN codes if dragging operation is in effect.
									// E.g. event DRAGGING | MOUSE_MOVE is sent to underlying DOM elements while dragging.
)

var errorToString = map[C.HLDOM_RESULT]string {
	C.HLDOM_OK: "HLDOM_OK",
	C.HLDOM_INVALID_HWND: "HLDOM_INVALID_HWND",
	C.HLDOM_INVALID_HANDLE: "HLDOM_INVALID_HANDLE",
	C.HLDOM_PASSIVE_HANDLE: "HLDOM_PASSIVE_HANDLE",
	C.HLDOM_INVALID_PARAMETER: "HLDOM_INVALID_PARAMETER",
	C.HLDOM_OPERATION_FAILED: "HLDOM_OPERATION_FAILED",
	C.HLDOM_OK_NOT_HANDLED: "HLDOM_OK_NOT_HANDLED",
}

// Hang on to any attached event handlers so that they don't
// get garbage collected
var eventHandlers = make(map[uintptr]EventHandler, 128)


// Main event handler that dispatches to the right element handler
//export goMainElementProc 
func goMainElementProc(tag uintptr, he unsafe.Pointer, evtg C.UINT, params unsafe.Pointer) C.BOOL {
	handled := false
	if key := uintptr(tag); key != 0 {
		handler := eventHandlers[key]
		switch evtg {
		case C.HANDLE_INITIALIZATION:
			if p := (*initializationParams)(params); p.Cmd == C.BEHAVIOR_ATTACH {
				handler.Attached(HELEMENT(he))
			} else if p.Cmd == C.BEHAVIOR_DETACH {
				handler.Detached(HELEMENT(he))
			}
			handled = true
		case C.HANDLE_MOUSE:
			p := (*MouseParams)(params);
			handled = handler.HandleMouse(HELEMENT(he), p)
		}
	}
	if handled {
		return C.TRUE
	}
	return C.FALSE
}



type HELEMENT C.HELEMENT

type JsonValue struct {
	T 	uint32
	U 	uint32
	D 	uint64
}

type initializationParams struct {
	Cmd		uint32
}

type BehaviorEventParams struct {
	Cmd		uint32
	Target 	HELEMENT
	Source	HELEMENT
	Reason 	uint32
	Data 	JsonValue
}

type Point struct {
	X 		int32
	Y		int32
}

type MouseParams struct {
	Cmd 			uint32
	Target 			HELEMENT
	Pos 			Point
	DocumentPos 	Point
	ButtonState 	uint32
	AltState 		uint32
	CursorType 		uint32
	IsOnIcon 		int32

	Dragging 		HELEMENT
	DraggingMode 	uint32
}


type EventHandler interface {
	Attached(he HELEMENT)
	Detached(he HELEMENT)

	HandleMouse(he HELEMENT, params *MouseParams) bool
}



// DomError represents an htmlayout error with an associated
// dom error code
type DomError struct {
	Result C.HLDOM_RESULT
	Message string
}

func (self DomError) String() string {
	return fmt.Sprintf( "%s: %s", errorToString[self.Result], self.Message )
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
		domPanic(dr, "UseElement");
	}
}

func unuse(handle HELEMENT) {
	if handle != nil {
		if dr := C.HTMLayout_UnuseElement(handle); dr != HLDOM_OK {
			domPanic(dr, "UnuseElement");
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
	tag := uintptr(unsafe.Pointer(&handler))
	if _, exists := eventHandlers[tag]; !exists {
		eventHandlers[tag] = handler
		if ret := C.HTMLayoutAttachEventHandlerEx(e.handle, C.MainElementProcAddr, C.LPVOID(tag), C.UINT(subscription)); ret != HLDOM_OK {
			domPanic(ret, "Failed to attach event handler to element")
		}
	}
}

func (e *Element) AttachHandlerAll(handler EventHandler) {
	tag := uintptr(unsafe.Pointer(&handler))
	if _, exists := eventHandlers[tag]; !exists {
		eventHandlers[tag] = handler
		if ret := C.HTMLayoutAttachEventHandler(e.handle, C.MainElementProcAddr, C.LPVOID(tag)); ret != HLDOM_OK {
			domPanic(ret, "Failed to attach event handler to element")
		}
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
	return nil;
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
	return nil;
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








