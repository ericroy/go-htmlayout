package htmlayout
/*
#cgo CFLAGS: -I../htmlayout/include
#cgo LDFLAGS: ../htmlayout/lib/HTMLayout.lib

#include <windows.h>

*/
import "C"

import (
	"os"
)

type Window struct {
	Hwnd C.HWND
}

func (w Window) initInstance() os.Error {
	w.Hwnd = C.CreateWindowEx(0, (*C.CHAR)(C.CString("gohl-window")),
		(*C.CHAR)(C.CString("gohl window title")), C.WS_OVERLAPPEDWINDOW,
		C.CW_USEDEFAULT, C.CW_USEDEFAULT, C.CW_USEDEFAULT, C.CW_USEDEFAULT,
		nil, nil, nil, nil)
	if w.Hwnd == nil {
		return os.NewError("Failed to create window")
	}
	return nil;
}

func (w Window) RunLoop() {
	msg := C.MSG{}
	for ; C.GetMessage(&msg, nil, 0, 0); {
		C.TranslateMessage(&msg)
		C.DispatchMessage(&msg)
	}
}


