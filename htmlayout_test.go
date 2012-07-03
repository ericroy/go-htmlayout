package gohl

import (
	"log"
	"math"
	"regexp"
	"strconv"
	"syscall"
	"testing"
	"unsafe"
)

const (
	WM_CREATE     = 1
	WM_DESTROY    = 2
	WM_CLOSE      = 16
	WM_QUIT       = 0x0012
	WM_ERASEBKGND = 0x0014
	WM_SHOWWINDOW = 0x0018
	WM_USER       = 0x0400
	ERROR_SUCCESS = 0
)

var (
	moduser32 = syscall.NewLazyDLL("user32.dll")

	procRegisterClassExW = moduser32.NewProc("RegisterClassExW")
	procCreateWindowExW  = moduser32.NewProc("CreateWindowExW")
	procDefWindowProcW   = moduser32.NewProc("DefWindowProcW")
	procDestroyWindow    = moduser32.NewProc("DestroyWindow")
	procPostQuitMessage  = moduser32.NewProc("PostQuitMessage")
	procGetMessageW      = moduser32.NewProc("GetMessageW")
	procTranslateMessage = moduser32.NewProc("TranslateMessage")
	procDispatchMessageW = moduser32.NewProc("DispatchMessageW")
	procSendMessageW     = moduser32.NewProc("SendMessageW")
	procPostMessageW     = moduser32.NewProc("PostMessageW")

	registeredClasses = make(map[string]bool, 4)
	testPages         = make(chan string, 1)
	testFuncs         = make(chan func(uint32), 1)
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

// Utility functions for creating a testing window, etc

func registerWindow(callbacks MsgHandlerMap, windowName string) {

	wproc := syscall.NewCallback(func(hwnd, msg uint32, wparam uintptr, lparam uintptr) uintptr {
		if result, handled := ProcNoDefault(hwnd, msg, wparam, lparam); handled {
			return result
		}

		var rc interface{} = nil
		if cb := callbacks[msg]; cb != nil {
			rc = cb(hwnd)
		}

		// Handler provided a return code
		if rc != nil {
			if code, ok := rc.(int); !ok {
				panic("window msg response should be int")
			} else {
				return uintptr(code)
			}
		}

		// Handler did not provide a return code, use the default window procedure
		code := DefWindowProc(hwnd, msg, wparam, lparam)
		return uintptr(code)
	})

	var wc Wndclassex
	wc.Size = uint32(unsafe.Sizeof(wc))
	wc.WndProc = wproc
	wc.Instance = 0
	wc.Icon = 0
	wc.Cursor = 0
	wc.Background = 0
	wc.MenuName = nil
	wc.ClassName = stringToUtf16Ptr(windowName)
	wc.IconSm = 0

	if _, errno := RegisterClassEx(&wc); errno != ERROR_SUCCESS {
		log.Panic(errno)
	}
}

func createWindow(windowName string) uint32 {
	wcname := stringToUtf16Ptr(windowName)
	hwnd, errno := CreateWindowEx(
		0,
		wcname,
		stringToUtf16Ptr("Gohl Test App"),
		0,
		0, 0, 20, 10,
		0, 0, 0, 0)
	if errno != ERROR_SUCCESS {
		log.Panic(errno)
	}
	return hwnd
}

func pump() {
	var m Msg
	for {
		if r, errno := GetMessage(&m, 0, 0, 0); errno != ERROR_SUCCESS {
			panic(errno)
		} else if r == 0 {
			break
		}
		TranslateMessage(&m)
		DispatchMessage(&m)
	}
}

func looseEqual(a, b string) bool {
	if re, err := regexp.Compile(`\s+`); err != nil {
		log.Panic(err)
	} else {
		a = re.ReplaceAllLiteralString(a, "")
		b = re.ReplaceAllLiteralString(b, "")
	}
	return a == b
}

func expectDomError(code HLDOM_RESULT) {
	if err := recover(); err == nil {
		log.Panic("Expected a DomError but got no error")
	} else if de, ok := err.(*DomError); !ok {
		log.Panic("Expected DomError, instead got: ", err)
	} else if de.Result != code {
		log.Panicf("Expected DomError with code %s, but got code %s instead ", domResultAsString(code), domResultAsString(de.Result))
	}
}

func expectPanic() {
	if err := recover(); err == nil {
		log.Panic("Expected a panic but didn't get one")
	}
}

func testWithHtml(html string, test func(hwnd uint32)) {
	if !registeredClasses["html"] {
		m := make(MsgHandlerMap, 32)
		for k, v := range defaultHandlerMap {
			m[k] = v
		}
		m[WM_CREATE] = func(hwnd uint32) interface{} {
			ret := defaultHandlerMap[WM_CREATE](hwnd)
			if err := LoadHtml(hwnd, []byte(<-testPages), ""); err != nil {
				log.Panic(err)
			}
			(<-testFuncs)(hwnd)
			PostMessage(hwnd, WM_CLOSE, 0, 0)
			return ret
		}
		registerWindow(m, "html")
		registeredClasses["html"] = true
	}
	testPages <- html
	testFuncs <- test
	_ = createWindow("html")
	pump()
}

// Variables and types for testing

type MsgHandler func(uint32) interface{}
type MsgHandlerMap map[uint32]MsgHandler

var defaultHandlerMap = MsgHandlerMap{
	WM_CREATE: func(hwnd uint32) interface{} {
		//log.Print("WM_CREATE, hwnd = ", hwnd)
		AttachNotifyHandler(hwnd, notifyHandler)
		AttachWindowEventHandler(hwnd, windowEventHandler)
		return 0
	},
	WM_SHOWWINDOW: func(hwnd uint32) interface{} {
		return 0
	},
	WM_ERASEBKGND: func(hwnd uint32) interface{} {
		return 0
	},
	WM_CLOSE: func(hwnd uint32) interface{} {
		//log.Print("WM_CLOSE, hwnd = ", hwnd)
		DetachWindowEventHandler(hwnd)
		DetachNotifyHandler(hwnd)
		DestroyWindow(hwnd)
		return nil
	},
	WM_DESTROY: func(hwnd uint32) interface{} {
		//log.Print("WM_DESTROY, hwnd = ", hwnd)
		//DumpObjectCounts()
		PostQuitMessage(0)
		return 0
	},
}

// Page templates used for various tests
var pages = map[string]string{
	"empty":       ``,
	"page":        `<html><body></body></html>`,
	"one-div":     `<div id="a"></div>`,
	"two-divs":    `<div id="a"></div><div id="b"></div>`,
	"three-divs":  `<div id="a"></div><div id="b"></div><div id="c"></div>`,
	"nested-divs": `<div id="a"><div id="b"></div></div>`,
	"attr":        `<div id="a" first="5" second="5.1" third="yes"></div>`,
	"css":         `<div style="left:10; opacity:0.5; text-align:center;"></div>`,
	"classes":     `<div class="one  two three"></div><div></div>`,
}

// Notify handler deals with WM_NOTIFY messages sent by htmlayout
var notifyHandler = &NotifyHandler{}

// Window event handler gets first and last chance to process events
var windowEventHandler = &EventHandler{}

// Tests:

func TestBasicWindow(t *testing.T) {
	// A channel to receive the hwnd once the window is created
	created := make(chan uint32)

	// Wait until the window is created, then post a message to close it
	go func() {
		PostMessage(<-created, WM_CLOSE, 0, 0)
	}()

	if !registeredClasses["basic"] {
		// Setup the window message handlers
		// Customize the WM_CREATE handler so that sends the hwnd to our channel
		handler := make(MsgHandlerMap, 32)
		for k, v := range defaultHandlerMap {
			handler[k] = v
		}
		handler[WM_CREATE] = func(hwnd uint32) interface{} {
			created <- hwnd
			return defaultHandlerMap[WM_CREATE](hwnd)
		}
		registerWindow(handler, "basic")
	}
	_ = createWindow("basic")
	pump()
}

func TestLoadHtml(t *testing.T) {
	testWithHtml(pages["page"], func(hwnd uint32) {})
}

func TestRootElement(t *testing.T) {
	testWithHtml(pages["one-div"], func(hwnd uint32) {
		if e := RootElement(hwnd); e == nil {
			t.Fatal("Could not get root elem")
		}
	})
}

func TestHandle(t *testing.T) {
	testWithHtml(pages["one-div"], func(hwnd uint32) {
		e := RootElement(hwnd)
		if h := e.Handle(); h == nil {
			t.Fatal("Handle was nil")
		}
	})
}

func TestRelease(t *testing.T) {
	testWithHtml(pages["one-div"], func(hwnd uint32) {
		e := RootElement(hwnd)
		e.Release()
		if h := e.Handle(); h != nil {
			t.Fatal("Released but handle is not nil, finalizer not called?")
		}
	})
}

func TestChildCount(t *testing.T) {
	testWithHtml(pages["two-divs"], func(hwnd uint32) {
		root := RootElement(hwnd)
		if count := root.ChildCount(); count != 2 {
			t.Fatal("Expected two divs as children")
		}
	})
}

func TestChildCount2(t *testing.T) {
	testWithHtml(pages["nested-divs"], func(hwnd uint32) {
		root := RootElement(hwnd)
		if count := root.ChildCount(); count != 1 {
			t.Fatal("Expected one divs as child")
		}
	})
}

func TestChild(t *testing.T) {
	testWithHtml(pages["two-divs"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d1 := root.Child(0)
		d2 := root.Child(1)
		if d1 == nil || d2 == nil {
			t.Fatal("A child element could not be retrieved by index")
		}
	})
}

func TestIndex(t *testing.T) {
	testWithHtml(pages["two-divs"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d1 := root.Child(0)
		d2 := root.Child(1)
		if d1.Index() != 0 {
			t.Fatal("Expected index 0")
		}
		if d2.Index() != 1 {
			t.Fatal("Expected index 1")
		}
	})
}

func TestEquals(t *testing.T) {
	testWithHtml(pages["one-div"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d1 := root.Child(0)
		if root.Equals(d1) {
			t.Fatal("Distinct elems should not be equal")
		}
		if !d1.Equals(d1) {
			t.Fatal("Same elements should be equal")
		}
	})
}

func TestParent(t *testing.T) {
	testWithHtml(pages["nested-divs"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d1 := root.Child(0)
		d2 := d1.Child(0)

		if !d2.Parent().Equals(d1) {
			t.Fatal("Parent was not the expected elem")
		}
		if !d1.Parent().Equals(root) {
			t.Fatal("Parent was not the expected elem")
		}
		if root.Parent() != nil {
			t.Fatal("Root's parent should be nil")
		}
	})
}

func TestSelect(t *testing.T) {
	testWithHtml(pages["nested-divs"], func(hwnd uint32) {
		root := RootElement(hwnd)
		results := root.Select("div > div")
		if len(results) != 1 {
			t.Fatal("Expected one result")
		}
		inner := root.Child(0).Child(0)
		if !results[0].Equals(inner) {
			t.Fatal("Expected to match inner div")
		}
	})
}

func TestSelectParent(t *testing.T) {
	testWithHtml(pages["nested-divs"], func(hwnd uint32) {
		root := RootElement(hwnd)
		if result := root.SelectParent("*:has-child-of-type(div):has-child-of-type(div)"); !result.Equals(root) {
			t.Fatal("Expected to match root element")
		}
	})
}

func TestSelectParentLimit(t *testing.T) {
	testWithHtml(pages["nested-divs"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d1 := root.Child(0)
		d2 := d1.Child(0)

		// Depth=0 means search all the way up to root, should match
		if result := d2.SelectParentLimit("html", 0); result == nil || !result.Equals(root) {
			t.Fatal("Expected to match root elem")
		}

		// Depth=1 means only consider receiver element, should not match
		if result := d2.SelectParentLimit("*:has-child-of-type(div)", 1); result != nil {
			t.Fatal("Expected to only check current element and not match it, instead got: ", result.OuterHtml())
		}

		// Depth=2 means consider first parent, should match
		if result := d2.SelectParentLimit("*:has-child-of-type(div)", 2); result == nil || !result.Equals(root.Child(0)) {
			t.Fatal("Expected to match outer div, instead got: ", result.OuterHtml())
		}
	})
}

func TestType(t *testing.T) {
	testWithHtml(pages["nested-divs"], func(hwnd uint32) {
		root := RootElement(hwnd)
		elemType := root.Type()
		if elemType != "html" {
			t.Fatal("Type of root elem should be 'html', instead got: ", elemType)
		}
		elemType = root.Child(0).Type()
		if elemType != "div" {
			t.Fatal("Type of first child elem should be 'div', instead got: ", elemType)
		}
	})
}

func TestOuterHtml(t *testing.T) {
	testWithHtml(pages["nested-divs"], func(hwnd uint32) {
		root := RootElement(hwnd)
		expected := "<html>" + pages["nested-divs"] + "</html>"
		if !looseEqual(expected, root.OuterHtml()) {
			t.Fatal("Outer html of root elem not as expected")
		}
	})
}

func TestHtml(t *testing.T) {
	testWithHtml(pages["nested-divs"], func(hwnd uint32) {
		root := RootElement(hwnd)
		if !looseEqual(root.Html(), pages["nested-divs"]) {
			t.Fatal("Inner html of root elem should match original html")
		}

		d1 := root.Child(0)
		if !looseEqual(d1.Html(), `<div id="b"></div>`) {
			t.Fatal("Inner html of first div should match html of innermost div")
		}

		d2 := d1.Child(0)
		if !looseEqual(d2.Html(), ``) {
			t.Fatal("Inner html of innermost div be the div's contents")
		}
	})
}

func TestInsertChild(t *testing.T) {
	testWithHtml(pages["one-div"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)
		e := NewElement("div")
		d.InsertChild(e, 0)
		if !looseEqual(d.Html(), `<div></div>`) {
			t.Fatal("Inserting element created unexpected html: ", d.Html())
		}
		e = NewElement("span")
		d.InsertChild(e, 0)
		if !looseEqual(d.Html(), `<span></span><div></div>`) {
			t.Fatal("Inserting element created unexpected html: ", d.Html())
		}
	})
}

func TestAppendChild(t *testing.T) {
	testWithHtml(pages["one-div"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)
		e := NewElement("div")
		d.AppendChild(e)
		if !looseEqual(d.Html(), `<div></div>`) {
			t.Fatal("Inserting element created unexpected html: ", d.Html())
		}
		e = NewElement("span")
		d.AppendChild(e)
		if !looseEqual(d.Html(), `<div></div><span></span>`) {
			t.Fatal("Inserting element created unexpected html: ", d.Html())
		}
	})
}

func TestDetach(t *testing.T) {
	testWithHtml(pages["nested-divs"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)
		inner := d.Child(0)

		// Pull the inner div out of the dom
		inner.Detach()
		if !looseEqual(d.Html(), ``) {
			t.Fatal("Element should not have any contents after detaching its only child")
		}

		// Put it back in
		d.AppendChild(inner)
		if !looseEqual(d.Html(), inner.OuterHtml()) {
			t.Fatal("Element should not have any contents after detaching its only child")
		}
	})
}

func TestDelete(t *testing.T) {
	testWithHtml(pages["nested-divs"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)
		inner := d.Child(0)

		inner.Delete()
		if !looseEqual(d.Html(), ``) {
			t.Fatal("Element should not have any contents after detaching its only child")
		}

		// Should not be able to put the deleted element back in, should get invalid handle error
		func() {
			defer expectDomError(HLDOM_INVALID_HANDLE)
			d.AppendChild(inner)
		}()
	})
}

func TestClone(t *testing.T) {
	testWithHtml(`<div>a</div>`, func(hwnd uint32) {
		root := RootElement(hwnd)
		clone := root.Child(0).Clone()
		if clone.Type() != "div" {
			t.Fatal("Clone should have same type as original")
		}
		if !looseEqual(clone.Html(), `a`) {
			t.Fatal("Clone should have same contents as original")
		}
	})
}

func TestSwap(t *testing.T) {
	testWithHtml(pages["two-divs"], func(hwnd uint32) {
		root := RootElement(hwnd)
		a := root.Child(0)
		b := root.Child(1)
		a.Swap(b)
		if !looseEqual(root.Html(), `<div id="b"></div><div id="a"></div>`) {
			t.Fatal("Clone should have same contents as original")
		}
	})
}

func TestSortChildren(t *testing.T) {
	cmp := func(a, b *Element) int {
		first := a.Html()[0]
		second := b.Html()[0]
		if first == second {
			return 0
		} else if first > second {
			return 1
		}
		return -1
	}
	testWithHtml(`<div>c</div><div>b</div><div>a</div>`, func(hwnd uint32) {
		root := RootElement(hwnd)
		root.SortChildren(cmp)
		if !looseEqual(root.Html(), `<div>a</div><div>b</div><div>c</div>`) {
			t.Fatal("Sorted elements should be in alphabetically descending order")
		}
	})
}

func TestSortChildrenRange(t *testing.T) {
	cmp := func(a, b *Element) int {
		first := a.Html()[0]
		second := b.Html()[0]
		if first == second {
			return 0
		} else if first > second {
			return 1
		}
		return -1
	}
	testWithHtml(`<div>c</div><div>b</div><div>a</div>`, func(hwnd uint32) {
		root := RootElement(hwnd)
		root.SortChildrenRange(0, 2, cmp)
		if !looseEqual(root.Html(), `<div>b</div><div>c</div><div>a</div>`) {
			t.Fatal("First two elements should be in alphabetically descending order")
		}
	})
}

func TestHwnd(t *testing.T) {
	testWithHtml(pages["one-div"], func(hwnd uint32) {
		root := RootElement(hwnd)
		if root.Hwnd() != hwnd {
			t.Fatal("Root should report the same hwnd it was created with")
		}
		if root.Child(0).Hwnd() != hwnd {
			t.Fatal("Child should report same hwnd as its root")
		}
	})
}

// TODO: Figure out how this test should differ from the test for Hwnd()
func TestRootHwnd(t *testing.T) {
	testWithHtml(pages["one-div"], func(hwnd uint32) {
		root := RootElement(hwnd)
		if root.RootHwnd() != hwnd {
			t.Fatal("Root should report the same root hwnd it was created with")
		}
		if root.Child(0).RootHwnd() != hwnd {
			t.Fatal("Child should report same root hwnd as its root")
		}
	})
}

func TestSetHtml(t *testing.T) {
	d := NewElement("div")

	// Should not be able to set html on an element that is not part of a dom
	func() {
		defer expectDomError(HLDOM_PASSIVE_HANDLE)
		d.SetHtml("<span></span>")
	}()

	testWithHtml(pages["one-div"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)
		d.SetHtml("<span></span>")
		if !looseEqual(d.Html(), `<span></span>`) {
			t.Fatal("Element should contain new html")
		}
	})
}

func TestPrependHtml(t *testing.T) {
	testWithHtml(pages["one-div"], func(hwnd uint32) {
		root := RootElement(hwnd)
		root.PrependHtml("<span></span>")
		if !looseEqual(root.Html(), `<span></span><div id="a"></div>`) {
			t.Fatal("Expected prepended html in front of existing contents")
		}
	})
}

func TestAppendHtml(t *testing.T) {
	testWithHtml(pages["one-div"], func(hwnd uint32) {
		root := RootElement(hwnd)
		root.AppendHtml("<span></span>")
		if !looseEqual(root.Html(), `<div id="a"></div><span></span>`) {
			t.Fatal("Expected appended html at the end of existing contents")
		}
	})
}

func TestSetText(t *testing.T) {
	testWithHtml(pages["one-div"], func(hwnd uint32) {
		root := RootElement(hwnd)
		root.SetText("Hi")
		if !looseEqual(root.Html(), `Hi`) {
			t.Fatal("Setting the text should have replaced inner html")
		}
	})
}

func TestAttrCount(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)
		if count := d.AttrCount(); count != 4 {
			t.Fatal("Expected four attributes on div")
		}
	})
}

func TestAttrByIndex(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)
		expectedKeys := []string{"id", "first", "second", "third"}
		expectedValues := []string{"a", "5", "5.1", "yes"}
		for i := range expectedKeys {
			key, val := d.AttrByIndex(i)
			if key != expectedKeys[i] || val != expectedValues[i] {
				t.Fatal("Expected (%s,%s), got (%s, %s)", expectedKeys[i], expectedValues[i], key, val)
			}
		}
	})
}

func TestAttr(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)

		if first, exists := d.Attr("first"); !exists {
			t.Fatal("Should exist")
		} else if first != "5" {
			t.Fatal("Unexpected value: ", first)
		}

		if second, exists := d.Attr("second"); !exists {
			t.Fatal("Should exist")
		} else if second != "5.1" {
			t.Fatal("Unexpected value: ", second)
		}

		if third, exists := d.Attr("third"); !exists {
			t.Fatal("Should exist")
		} else if third != "yes" {
			t.Fatal("Unexpected value: ", third)
		}

		if _, exists := d.Attr("invalid"); exists {
			t.Fatal("Should not exist")
		}
	})
}

func TestAttrAsFloatOnInt(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)

		if value, exists, err := d.AttrAsFloat("first"); err != nil {
			t.Fatal(err)
		} else if !exists {
			t.Fatal("Should exist")
		} else if math.Abs(value-float64(5.0)) > float64(0.0001) {
			t.Fatal("Unexpected value: ", value)
		}
	})
}

func TestAttrAsFloatOnFloat(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)

		if value, exists, err := d.AttrAsFloat("second"); err != nil {
			t.Fatal(err)
		} else if !exists {
			t.Fatal("Should exist")
		} else if math.Abs(value-float64(5.1)) > float64(0.0001) {
			t.Fatal("Unexpected value: ", value)
		}
	})
}

func TestAttrAsFloatOnString(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)

		if _, exists, err := d.AttrAsFloat("third"); !exists {
			t.Fatal("Should exist")
		} else if err == nil {
			t.Fatal("Expected error")
		} else if _, ok := err.(*strconv.NumError); !ok {
			t.Fatal("Expected NumError")
		}
	})
}

func TestAttrAsFloatOnInvalid(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)

		if _, exists, err := d.AttrAsFloat("invalid"); err != nil || exists {
			t.Fatal()
		}
	})
}

func TestAttrAsIntOnInt(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)

		if value, exists, err := d.AttrAsInt("first"); err != nil {
			t.Fatal(err)
		} else if !exists {
			t.Fatal("Should Exist")
		} else if value != 5 {
			t.Fatal("Expected integer 5")
		}
	})
}

func TestAttrAsIntOnFloat(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)

		if _, exists, err := d.AttrAsInt("second"); !exists {
			t.Fatal("Should exist")
		} else if err == nil {
			t.Fatal("Expected error while parsing")
		} else if _, ok := err.(*strconv.NumError); !ok {
			t.Fatal("Expected NumError")
		}
	})
}

func TestAttrAsIntOnString(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)

		if _, exists, err := d.AttrAsInt("third"); !exists {
			t.Fatal("Should exist")
		} else if err == nil {
			t.Fatal("Expected error while parsing")
		} else if _, ok := err.(*strconv.NumError); !ok {
			t.Fatal("Expected NumError")
		}
	})
}

func TestAttrAsIntOnInvalid(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)

		if _, exists, err := d.AttrAsFloat("invalid"); err != nil || exists {
			t.Fatal()
		}
	})
}

func TestRemoveAttr(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)
		if _, exists := d.Attr("second"); !exists {
			t.Fatal("Should exist")
		} else {
			d.RemoveAttr("second")
			if _, exists := d.Attr("second"); exists {
				t.Fatal("Should not exist")
			}
		}
		// Removing a nonexistent attr should do nothing
		d.RemoveAttr("invalid")
	})
}

func TestSetAttrFloat32(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)
		pi := 3.14159
		d.SetAttr("myFloat32", float32(pi))

		if s, exists := d.Attr("myFloat32"); !exists {
			t.Fatal("Should exist")
		} else if f, err := strconv.ParseFloat(s, 64); err != nil {
			t.Fatal(err)
		} else if math.Abs(f-float64(pi)) > float64(0.0001) {
			t.Fatal("Unexpected value: ", f)
		}
	})
}

func TestSetAttrFloat64(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)
		pi := 3.14159
		d.SetAttr("myFloat64", float64(pi))

		if s, exists := d.Attr("myFloat64"); !exists {
			t.Fatal("Should exist")
		} else if f, err := strconv.ParseFloat(s, 64); err != nil {
			t.Fatal("Could not parse attr string as float")
		} else if math.Abs(f-float64(pi)) > float64(0.0001) {
			t.Fatal("Unexpected value: ", f)
		}
	})
}

func TestSetAttrString(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)
		d.SetAttr("myString", "hello")

		if s, exists := d.Attr("myString"); !exists {
			t.Fatal("Should exist")
		} else if s != "hello" {
			t.Fatal("Unexpected value: ", s)
		}
	})
}

func TestSetAttrInt(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)
		d.SetAttr("myInt", 9)

		if s, exists := d.Attr("myInt"); !exists {
			t.Fatal("Should exist")
		} else if s != "9" {
			t.Fatal("Unexpected value: ", s)
		}
	})
}

func TestSetAttrOverwrite(t *testing.T) {
	testWithHtml(pages["attr"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)

		// Test overwriting an attr that already exists
		d.SetAttr("id", "hai")
		if s, exists := d.Attr("id"); !exists {
			t.Fatal("Should exist")
		} else if s != "hai" {
			t.Fatal("Expected to overwrite attr")
		}
	})
}

func TestHasClass(t *testing.T) {
	testWithHtml(pages["classes"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)
		d2 := root.Child(1)

		if !d.HasClass("two") {
			t.Fatal("Should have class 'two'")
		}
		if d.HasClass("nope") {
			t.Fatal("Should not have class 'nope'")
		}
		if d2.HasClass("nope") {
			t.Fatal("Should not have class 'nope'")
		}
	})
}

func TestAddClass(t *testing.T) {
	testWithHtml(pages["classes"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)
		d2 := root.Child(1)

		var classes string
		var exists bool

		d.AddClass("four")
		if classes, exists = d.Attr("class"); !exists {
			t.Fatal("Should exist")
		}
		if classes != "one two three four" {
			t.Fatalf("Unexpected class attr value: '%s'", classes)
		}

		d.AddClass("two")
		if classes, exists = d.Attr("class"); !exists {
			t.Fatal("Should exist")
		}
		if classes != "one two three four" {
			t.Fatalf("Unexpected class attr value: '%s'", classes)
		}

		d2.AddClass("one")
		if classes, exists = d2.Attr("class"); !exists {
			t.Fatal("Should exist")
		}
		if classes != "one" {
			t.Fatalf("Unexpected class attr value: '%s'", classes)
		}
	})
}

func TestRemoveClass(t *testing.T) {
	testWithHtml(pages["classes"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)
		d2 := root.Child(1)

		var classes string
		var exists bool

		d.RemoveClass("one")
		if classes, exists = d.Attr("class"); !exists {
			t.Fatal("Should exist")
		}
		if classes != "two three" {
			t.Fatalf("Unexpected class attr value: '%s'", classes)
		}

		d.RemoveClass("three")
		if classes, exists = d.Attr("class"); !exists {
			t.Fatal("Should exist")
		}
		if classes != "two" {
			t.Fatalf("Unexpected class attr value: '%s'", classes)
		}

		d.RemoveClass("nope")
		if classes, exists = d.Attr("class"); !exists {
			t.Fatal("Should exist")
		}
		if classes != "two" {
			t.Fatalf("Unexpected class attr value: '%s'", classes)
		}

		d2.RemoveClass("nope")
		if classes, exists = d2.Attr("class"); exists {
			t.Fatal("Should not exist")
		}
	})
}

func TestStyle(t *testing.T) {
	testWithHtml(pages["css"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)

		if s, exists := d.Style("left"); !exists {
			t.Fatal("Should exist")
		} else if s != "10px" {
			t.Fatal("Unexpected value for style 'left'")
		}

		if s, exists := d.Style("opacity"); !exists {
			t.Fatal("Should exist")
		} else if f, err := strconv.ParseFloat(s, 64); err != nil {
			t.Fatal(err)
		} else if math.Abs(f-float64(0.5)) > float64(0.05) {
			t.Fatal("Unexpected value: ", s)
		}

		if s, exists := d.Style("text-align"); !exists {
			t.Fatal("Should exist")
		} else if s != "center" {
			t.Fatal("Unexpected value: ", s)
		}

		if _, exists := d.Style("invalid"); exists {
			t.Fatal("Should not exist")
		}
	})
}

func TestSetStyleFloat32(t *testing.T) {
	testWithHtml(pages["css"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)

		d.SetStyle("opacity", float32(0.75))
		if s, exists := d.Style("opacity"); !exists {
			t.Fatal("Should exist")
		} else if f, err := strconv.ParseFloat(s, 64); err != nil {
			t.Fatal(err)
		} else if math.Abs(f-float64(0.75)) > float64(0.05) {
			log.Print(s)
			t.Fatal("Unexpected value: ", f)
		}
	})
}

func TestSetStyleFloat64(t *testing.T) {
	testWithHtml(pages["css"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)

		d.SetStyle("opacity", float64(0.75))
		if s, exists := d.Style("opacity"); !exists {
			t.Fatal("Should exist")
		} else if f, err := strconv.ParseFloat(s, 64); err != nil {
			t.Fatal(err)
		} else if math.Abs(f-float64(0.75)) > float64(0.05) {
			t.Fatal("Unexpected value: ", f)
		}
	})
}

func TestSetStyleString(t *testing.T) {
	testWithHtml(pages["css"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)

		d.SetStyle("font-family", "hello")
		if s, exists := d.Style("font-family"); !exists {
			t.Fatal("Should exist")
		} else if s != "hello" {
			t.Fatal("Unexpected value: ", s)
		}
	})
}

func TestSetStyleInt(t *testing.T) {
	testWithHtml(pages["css"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)

		d.SetStyle("height", 9)
		if s, exists := d.Style("height"); !exists {
			t.Fatal("Should exist")
		} else if s != "9px" {
			t.Fatal("Unexpected value: ", s)
		}
	})
}

func TestSetStyleOverwrite(t *testing.T) {
	testWithHtml(pages["css"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d := root.Child(0)

		// Test overwriting a style that already exists
		d.SetStyle("left", 99)
		if s, exists := d.Style("left"); !exists {
			t.Fatal("Should exist")
		} else if s != "99px" {
			t.Fatal("Expected to overwrite style")
		}
	})
}

func TestAttachSameHandlerToTwoElements(t *testing.T) {
	testWithHtml(pages["two-divs"], func(hwnd uint32) {
		root := RootElement(hwnd)
		d1 := root.Child(0)
		d2 := root.Child(1)

		handler := &EventHandler{}
		d1.AttachHandler(handler)
		d2.AttachHandler(handler)
		d2.DetachHandler(handler)
		d1.DetachHandler(handler)
	})
}
