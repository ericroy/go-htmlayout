package gohl
/*
#cgo CFLAGS: -I../htmlayout/include
#cgo LDFLAGS: ../htmlayout/lib/HTMLayout.lib

#include <stdlib.h>
#include <htmlayout.h>
*/
import "C"

import (
	"os"
	"unsafe"
)

const (
	HLN_CREATE_CONTROL = C.HLN_CREATE_CONTROL 
	HLN_LOAD_DATA = C.HLN_LOAD_DATA
	HLN_CONTROL_CREATED = C.HLN_CONTROL_CREATED 
	HLN_DATA_LOADED = C.HLN_DATA_LOADED
	HLN_DOCUMENT_COMPLETE = C.HLN_DOCUMENT_COMPLETE 
	HLN_UPDATE_UI = C.HLN_UPDATE_UI
	HLN_DESTROY_CONTROL = C.HLN_DESTROY_CONTROL
	HLN_ATTACH_BEHAVIOR = C.HLN_ATTACH_BEHAVIOR
	HLN_BEHAVIOR_CHANGED = C.HLN_BEHAVIOR_CHANGED
	HLN_DIALOG_CREATED = C.HLN_DIALOG_CREATED
	HLN_DIALOG_CLOSE_RQ = C.HLN_DIALOG_CLOSE_RQ
	HLN_DOCUMENT_LOADED = C.HLN_DOCUMENT_LOADED
)

/*
// HTMLayout Window Proc without call of DefWindowProc.
EXTERN_C LRESULT HLAPI HTMLayoutProcND(HWND hwnd, UINT msg, WPARAM wParam, LPARAM lParam, BOOL* pbHandled);
*/
func HTMLayoutProcND(hwnd, msg uint32, wparam, lparam int32) (uintptr, bool) {
	var handled C.BOOL = 0
	var result C.LRESULT = C.HTMLayoutProcND(C.HWND(C.HANDLE(uintptr(hwnd))), C.UINT(msg),
		C.WPARAM(wparam), C.LPARAM(lparam), &handled)
	return uintptr(result), handled != 0
}

/*
Set \link #HTMLAYOUT_NOTIFY() notification callback function \endlink.
 
 \param[in] hWndHTMLayout \b HWND, HTMLayout window handle.
 \param[in] cb \b HTMLAYOUT_NOTIFY*, \link #HTMLAYOUT_NOTIFY() callback function \endlink.
 \param[in] cbParam \b LPVOID, parameter that will be passed to \link #HTMLAYOUT_NOTIFY() callback function \endlink as vParam paramter.
 
EXTERN_C VOID HLAPI     HTMLayoutSetCallback(HWND hWndHTMLayout, LPHTMLAYOUT_NOTIFY cb, LPVOID cbParam);
*/
func HTMLayoutSetCallback(hwnd uint32, callback uintptr, callbackParam uintptr) {
	C.HTMLayoutSetCallback(C.HWND(C.HANDLE(uintptr(hwnd))), (*C.HTMLAYOUT_NOTIFY)(unsafe.Pointer(callback)), C.LPVOID(callbackParam))
}

/*
Load HTML from in memory buffer with base.

 \param[in] hWndHTMLayout \b HWND, HTMLayout window handle.
 \param[in] html \b LPCBYTE, Address of HTML to load.
 \param[in] htmlSize \b UINT, Length of the array pointed by html parameter.
 \param[in] baseUrl \b LPCWSTR, base URL. All relative links will be resolved against this URL.
 \return \b BOOL, \c TRUE if the text was parsed and loaded successfully, FALSE otherwise.
 
EXTERN_C BOOL HLAPI     HTMLayoutLoadHtmlEx(HWND hWndHTMLayout, LPCBYTE html, UINT htmlSize, LPCWSTR baseUrl);
*/
func HTMLayoutLoadHtmlEx(hwnd uint32, data []byte, baseUrl string) os.Error {
	if ok := C.HTMLayoutLoadHtmlEx(C.HWND(C.HANDLE(uintptr(hwnd))), (*C.BYTE)(&data[0]),
		C.UINT(len(data)), (*C.WCHAR)(stringToUtf16Ptr(baseUrl))); ok == 0 {
		return os.NewError("HTMLayoutLoadHtmlEx failed")
	}
	return nil
}