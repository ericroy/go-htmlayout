package gohl

import (
	"testing"
	"log"
	"syscall"
	"unsafe"
)

const (
	WM_CREATE      = 1
	WM_DESTROY     = 2
	WM_CLOSE       = 16
	ERROR_SUCCESS = 0
)

var (
	moduser32   = syscall.NewLazyDLL("user32.dll")

	procRegisterClassExW        = moduser32.NewProc("RegisterClassExW")
	procCreateWindowExW         = moduser32.NewProc("CreateWindowExW")
	procDefWindowProcW          = moduser32.NewProc("DefWindowProcW")
	procDestroyWindow           = moduser32.NewProc("DestroyWindow")
	procPostQuitMessage         = moduser32.NewProc("PostQuitMessage")
	procGetMessageW             = moduser32.NewProc("GetMessageW")
	procTranslateMessage        = moduser32.NewProc("TranslateMessage")
	procDispatchMessageW        = moduser32.NewProc("DispatchMessageW")
	procSendMessageW            = moduser32.NewProc("SendMessageW")
	procPostMessageW            = moduser32.NewProc("PostMessageW")
)

type Wndclassex struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uint32
	Icon       uint32
	Cursor     uint32
	Background uint32
	MenuName   *uint16
	ClassName  *uint16
	IconSm     uint32
}

type Msg struct {
	Hwnd    uint32
	Message uint32
	Wparam  int32
	Lparam  int32
	Time    uint32
	Pt      Point
}

func RegisterClassEx(wndclass *Wndclassex) (atom uint16, err syscall.Errno) {
	r0, _, e1 := syscall.Syscall(procRegisterClassExW.Addr(), 1, uintptr(unsafe.Pointer(wndclass)), 0, 0)
	atom = uint16(r0)
	if atom == 0 {
		if e1 != 0 {
			err = syscall.Errno(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func CreateWindowEx(exstyle uint32, classname *uint16, windowname *uint16, style uint32, x int32, y int32, width int32, height int32, wndparent uint32, menu uint32, instance uint32, param uintptr) (hwnd uint32, err syscall.Errno) {
	r0, _, e1 := syscall.Syscall12(procCreateWindowExW.Addr(), 12, uintptr(exstyle), uintptr(unsafe.Pointer(classname)), uintptr(unsafe.Pointer(windowname)), uintptr(style), uintptr(x), uintptr(y), uintptr(width), uintptr(height), uintptr(wndparent), uintptr(menu), uintptr(instance), uintptr(param))
	hwnd = uint32(r0)
	if hwnd == 0 {
		if e1 != 0 {
			err = syscall.Errno(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func DefWindowProc(hwnd uint32, msg uint32, wparam uintptr, lparam uintptr) (lresult int32) {
	r0, _, _ := syscall.Syscall6(procDefWindowProcW.Addr(), 4, uintptr(hwnd), uintptr(msg), uintptr(wparam), uintptr(lparam), 0, 0)
	lresult = int32(r0)
	return
}

func DestroyWindow(hwnd uint32) (err syscall.Errno) {
	r1, _, e1 := syscall.Syscall(procDestroyWindow.Addr(), 1, uintptr(hwnd), 0, 0)
	if int(r1) == 0 {
		if e1 != 0 {
			err = syscall.Errno(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func PostQuitMessage(exitcode int32) {
	syscall.Syscall(procPostQuitMessage.Addr(), 1, uintptr(exitcode), 0, 0)
	return
}

func GetMessage(msg *Msg, hwnd uint32, MsgFilterMin uint32, MsgFilterMax uint32) (ret int32, err syscall.Errno) {
	r0, _, e1 := syscall.Syscall6(procGetMessageW.Addr(), 4, uintptr(unsafe.Pointer(msg)), uintptr(hwnd), uintptr(MsgFilterMin), uintptr(MsgFilterMax), 0, 0)
	ret = int32(r0)
	if ret == -1 {
		if e1 != 0 {
			err = syscall.Errno(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func TranslateMessage(msg *Msg) (done bool) {
	r0, _, _ := syscall.Syscall(procTranslateMessage.Addr(), 1, uintptr(unsafe.Pointer(msg)), 0, 0)
	done = bool(r0 != 0)
	return
}

func DispatchMessage(msg *Msg) (ret int32) {
	r0, _, _ := syscall.Syscall(procDispatchMessageW.Addr(), 1, uintptr(unsafe.Pointer(msg)), 0, 0)
	ret = int32(r0)
	return
}

func SendMessage(hwnd uint32, msg uint32, wparam uintptr, lparam uintptr) (lresult int32) {
	r0, _, _ := syscall.Syscall6(procSendMessageW.Addr(), 4, uintptr(hwnd), uintptr(msg), uintptr(wparam), uintptr(lparam), 0, 0)
	lresult = int32(r0)
	return
}

func PostMessage(hwnd uint32, msg uint32, wparam uintptr, lparam uintptr) (err syscall.Errno) {
	r1, _, e1 := syscall.Syscall6(procPostMessageW.Addr(), 4, uintptr(hwnd), uintptr(msg), uintptr(wparam), uintptr(lparam), 0, 0)
	if int(r1) == 0 {
		if e1 != 0 {
			err = syscall.Errno(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}


// Notify handler deals with WM_NOTIFY messages sent by htmlayout
var notifyHandler = &NotifyHandler{}

// Window event handler gets first and last chance to process events
var windowEventHandler = &EventHandler{}


func makeWndProc(html string, onCreate func(), onDestroy func()) func(uint32, uint32, uintptr, uintptr) uintptr {
	return func(hwnd, msg uint32, wparam uintptr, lparam uintptr) uintptr {
		if result, handled := ProcNoDefault(hwnd, msg, wparam, lparam); handled {
			return result
		}

		var rc int32
		
		switch msg {
		case WM_CREATE:
			log.Print("WM_CREATE")
			AttachNotifyHandler(hwnd, notifyHandler)
			AttachWindowEventHandler(hwnd, windowEventHandler)
			if err := LoadHtml(hwnd, []byte(html), ""); err != nil {
				log.Panic(err)
			}
			onCreate()
			rc = 0
		case WM_CLOSE:
			log.Print("WM_CLOSE")
			DetachWindowEventHandler(hwnd)
			DetachNotifyHandler(hwnd)
			DestroyWindow(hwnd)
		case WM_DESTROY:
			log.Print("WM_DESTROY")
			DumpObjectCounts()		
			PostQuitMessage(0)
			onDestroy()
		default:
			rc = DefWindowProc(hwnd, msg, wparam, lparam)
		}

		return uintptr(rc)
	}
}



func makeWindow(html string, onCreate func(), onDestroy func()) uint32 {

	wproc := syscall.NewCallback(makeWndProc(html, onCreate, onDestroy))

	// RegisterClassEx
	wcname := stringToUtf16Ptr("gohlWindowClass")
	var wc Wndclassex
	wc.Size = uint32(unsafe.Sizeof(wc))
	wc.WndProc = wproc
	wc.Instance = 0
	wc.Icon = 0
	wc.Cursor = 0
	wc.Background = 0
	wc.MenuName = nil
	wc.ClassName = wcname
	wc.IconSm = 0

	if _, errno := RegisterClassEx(&wc); errno != ERROR_SUCCESS {
		panic(errno)
	}

	hwnd, errno := CreateWindowEx(
		0,
		wcname,
		stringToUtf16Ptr("Gohl Test App"),
		0,
		0, 0, 20, 10,
		0, 0, 0, 0)
	if errno != ERROR_SUCCESS {
		panic(errno)
	}

	return hwnd
}

func pump() {
	var m Msg
	for {
		if r, errno := GetMessage(&m, 0, 0, 0); errno != ERROR_SUCCESS {
			panic(errno)
		} else if r == 0 {
			// WM_QUIT received -> get out
			break
		}
		TranslateMessage(&m)
		DispatchMessage(&m)
	}
}


func TestBasicWindow(t *testing.T) {
	hwnd := makeWindow(
		"<html></html>",
		func() {
			log.Print("Created")
		},
		func() {
			log.Print("Destroyed")
		})
	go pump()
	SendMessage(hwnd, WM_CLOSE, 0, 0)
	log.Print("Done")
}
