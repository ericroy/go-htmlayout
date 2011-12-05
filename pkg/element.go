package htmlayout
/*
#cgo CFLAGS: -I../htmlayout/include
#include <htmlayout.h>
#include <htmlayout_dom.h>
*/
import "C"

import (
	"fmt"
)

const (

	// HLDOM_RESULT values
	HLDOM_OK = 0
	HLDOM_INVALID_HWND = 1 
	HLDOM_INVALID_HANDLE = 2
	HLDOM_PASSIVE_HANDLE = 3
	HLDOM_INVALID_PARAMETER = 4
	HLDOM_OPERATION_FAILED = 5
	HLDOM_OK_NOT_HANDLED = -1
)

// Essentially just typedefs
type DomResult C.HLDOM_RESULT

var errorToString = make(map[DomResult]string)

func init() {
	errorToString[HLDOM_OK] = "HLDOM_OK"
	errorToString[HLDOM_INVALID_HWND] = "HLDOM_INVALID_HWND"
	errorToString[HLDOM_INVALID_HANDLE] = "HLDOM_INVALID_HANDLE"
	errorToString[HLDOM_PASSIVE_HANDLE] = "HLDOM_PASSIVE_HANDLE"
	errorToString[HLDOM_INVALID_PARAMETER] = "HLDOM_INVALID_PARAMETER"
	errorToString[HLDOM_OPERATION_FAILED] = "HLDOM_OPERATION_FAILED"
	errorToString[HLDOM_OK_NOT_HANDLED] = "HLDOM_OK_NOT_HANDLED"
}

func hlpanic( dr DomResult, msg string ) {
	panic( fmt.Sprintf( "%s: %s", errorToString[dr], msg ) )
}


/*
Handle

Simple wrapper around an HELEMENT, provides reference 
counting functionality
*/
type Handle C.HELEMENT

func (h Handle) use() {
	if dr := HTMLayout_UseElement( h ); dr != HLDOM_OK {
		hlpanic( dr, "UseElement" );
	}
}

func (h Handle) unuse() {
	if h {
		if dr := HTMLayout_UnuseElement( h ); dr != HLDOM_OK {
			hlpanic( dr, "UnuseElement" );
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


