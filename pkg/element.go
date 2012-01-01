package htmlayout
/*
#cgo CFLAGS: -I../htmlayout/include
#cgo LDFLAGS: ../htmlayout/lib/HTMLayout.lib
#include <htmlayout.h>
*/
import "C"

import (
	"fmt"
)

const (
	HLDOM_OK = C.HLDOM_OK
	HLDOM_INVALID_HWND = C.HLDOM_INVALID_HWND
	HLDOM_INVALID_HANDLE = C.HLDOM_INVALID_HANDLE
	HLDOM_PASSIVE_HANDLE = C.HLDOM_PASSIVE_HANDLE
	HLDOM_INVALID_PARAMETER = C.HLDOM_INVALID_PARAMETER
	HLDOM_OPERATION_FAILED = C.HLDOM_OPERATION_FAILED
	HLDOM_OK_NOT_HANDLED = C.int(-1)
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

type DomError struct {
	Result C.HLDOM_RESULT
	Message string
}

func (self DomError) String() string {
	return fmt.Sprintf( "%s: %s", errorToString[self.Result], self.Message )
}

func DomPanic(result C.HLDOM_RESULT, message string) {
	panic(DomError{result, message})
}



type Handle C.HELEMENT

func use(handle Handle) {
	if dr := C.HTMLayout_UseElement( handle ); dr != HLDOM_OK {
		DomPanic( dr, "UseElement" );
	}
}

func unuse(handle Handle) {
	if handle != nil {
		if dr := C.HTMLayout_UnuseElement( handle ); dr != HLDOM_OK {
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


