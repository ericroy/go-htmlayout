package htmlayout
/*
#cgo CFLAGS: -std=gnu89 -I../htmlayout/include
//#cgo LDFLAGS: -l../htmlayout/lib/HTMLayout.lib
#include <htmlayout.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const (
	HLDOM_OK = C.HLDOM_OK
	HLDOM_INVALID_HWND = C.HLDOM_INVALID_HWND
	HLDOM_INVALID_HANDLE = C.HLDOM_INVALID_HANDLE
	HLDOM_PASSIVE_HANDLE = C.HLDOM_PASSIVE_HANDLE
	HLDOM_INVALID_PARAMETER = C.HLDOM_INVALID_PARAMETER
	HLDOM_OPERATION_FAILED = C.HLDOM_OPERATION_FAILED
	HLDOM_OK_NOT_HANDLED = -1
)

func init() {
}


type DomResult int

var errorToString = map[DomResult]string {
	HLDOM_OK: "HLDOM_OK",
	HLDOM_INVALID_HWND: "HLDOM_INVALID_HWND",
	HLDOM_INVALID_HANDLE: "HLDOM_INVALID_HANDLE",
	HLDOM_PASSIVE_HANDLE: "HLDOM_PASSIVE_HANDLE",
	HLDOM_INVALID_PARAMETER: "HLDOM_INVALID_PARAMETER",
	HLDOM_OPERATION_FAILED: "HLDOM_OPERATION_FAILED",
	HLDOM_OK_NOT_HANDLED: "HLDOM_OK_NOT_HANDLED",
}

type DomError struct {
	Result DomResult
	Message string
}

func (self DomError) String() string {
	fmt.Sprintf( "%s: %s", errorToString[self.Result], self.Message )
}

func DomPanic(result DomResult, message string) {
	panic(DomError{result, message})
}


/*
Handle

Simple wrapper around an HELEMENT, provides reference 
counting functionality
*/

type Handle unsafe.Pointer

func (self Handle) use() {
	if dr := C.HTMLayout_UseElement( self ); dr != HLDOM_OK {
		DomPanic( dr, "UseElement" );
	}
}

func (self Handle) unuse() {
	if self {
		if dr := C.HTMLayout_UnuseElement( self ); dr != HLDOM_OK {
			DomPanic( dr, "UnuseElement" );
		}
	}
}


/*
Element

Represents a single DOM element, owns and manages a Handle
*/
type Element struct {
	handle Handle
}


