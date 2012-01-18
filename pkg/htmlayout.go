package gohl
/*
#cgo CFLAGS: -I../htmlayout/include
#cgo LDFLAGS: ../htmlayout/lib/HTMLayout.lib

#include <stdlib.h>
#include <htmlayout.h>

// Main event function that dispatches to the appropriate event handler
BOOL CALLBACK ElementProc(LPVOID tag, HELEMENT he, UINT evtg, LPVOID prms)
{
	extern BOOL goElementProc(LPVOID, HELEMENT, UINT, LPVOID);
	return goElementProc(tag, he, evtg, prms);
}
LPELEMENT_EVENT_PROC ElementProcAddr = &ElementProc;

// Main event function that dispatches notify messages
LRESULT CALLBACK NotifyProc(UINT uMsg, WPARAM wParam, LPARAM lParam, LPVOID vParam)
{
	extern BOOL goNotifyProc(UINT, WPARAM, LPARAM, LPVOID);
	return goNotifyProc(uMsg, wParam, lParam, vParam);
}
LPHTMLAYOUT_NOTIFY NotifyProcAddr = &NotifyProc;

// Callback for results found during a select operation
BOOL CALLBACK SelectCallback(HELEMENT he, LPVOID param)
{
	extern BOOL goSelectCallback(HELEMENT, LPVOID);
	return goSelectCallback(he, param);
}
HTMLayoutElementCallback *SelectCallbackAddr = &SelectCallback;


INT ElementComparator(HELEMENT he1, HELEMENT he2, LPVOID pArg)
{
	extern INT goElementComparator(HELEMENT, HELEMENT, LPVOID);
	return goElementComparator(he1, he2, pArg);
}
ELEMENT_COMPARATOR *ElementComparatorAddr = (ELEMENT_COMPARATOR *)&ElementComparator;

*/
import "C"

import (
	"os"
	"log"
	"unsafe"
)

const (

	// HTMLayout Notify Events
	HLN_CREATE_CONTROL    = C.HLN_CREATE_CONTROL
	HLN_LOAD_DATA         = C.HLN_LOAD_DATA
	HLN_CONTROL_CREATED   = C.HLN_CONTROL_CREATED
	HLN_DATA_LOADED       = C.HLN_DATA_LOADED
	HLN_DOCUMENT_COMPLETE = C.HLN_DOCUMENT_COMPLETE
	HLN_UPDATE_UI         = C.HLN_UPDATE_UI
	HLN_DESTROY_CONTROL   = C.HLN_DESTROY_CONTROL
	HLN_ATTACH_BEHAVIOR   = C.HLN_ATTACH_BEHAVIOR
	HLN_BEHAVIOR_CHANGED  = C.HLN_BEHAVIOR_CHANGED
	HLN_DIALOG_CREATED    = C.HLN_DIALOG_CREATED
	HLN_DIALOG_CLOSE_RQ   = C.HLN_DIALOG_CLOSE_RQ
	HLN_DOCUMENT_LOADED   = C.HLN_DOCUMENT_LOADED

	// PhaseMask
	BUBBLING = uint32(C.BUBBLING) // bubbling (emersion) phase
	SINKING  = uint32(C.SINKING)  // capture (immersion) phase, this flag is or'ed with EVENTS codes below
	HANDLED  = uint32(C.HANDLED)  // event already processed.

	// EventGroups
	HANDLE_INITIALIZATION = C.HANDLE_INITIALIZATION /** attached/detached */
	HANDLE_MOUSE          = C.HANDLE_MOUSE          /** mouse events */
	HANDLE_KEY            = C.HANDLE_KEY            /** key events */
	HANDLE_FOCUS          = C.HANDLE_FOCUS          /** focus events, if this flag is set it also means that element it attached to is focusable */
	HANDLE_SCROLL         = C.HANDLE_SCROLL         /** scroll events */
	HANDLE_TIMER          = C.HANDLE_TIMER          /** timer event */
	HANDLE_SIZE           = C.HANDLE_SIZE           /** size changed event */
	HANDLE_DRAW           = C.HANDLE_DRAW           /** drawing request (event) */
	HANDLE_DATA_ARRIVED   = C.HANDLE_DATA_ARRIVED   /** requested data () has been delivered */
	HANDLE_BEHAVIOR_EVENT = C.HANDLE_BEHAVIOR_EVENT /** secondary, synthetic events: 
	BUTTON_CLICK, HYPERLINK_CLICK, etc., 
	a.k.a. notifications from intrinsic behaviors */
	HANDLE_METHOD_CALL     = C.HANDLE_METHOD_CALL     /** behavior specific methods */
	HANDLE_EXCHANGE        = C.HANDLE_EXCHANGE        /** system drag-n-drop */
	HANDLE_GESTURE         = C.HANDLE_GESTURE         /** touch input events */
	HANDLE_ALL             = C.HANDLE_ALL             /** all of them */
	DISABLE_INITIALIZATION = C.DISABLE_INITIALIZATION /** disable INITIALIZATION events to be sent. */

	// KeyboardStates
	CONTROL_KEY_PRESSED = C.CONTROL_KEY_PRESSED
	SHIFT_KEY_PRESSED   = C.SHIFT_KEY_PRESSED
	ALT_KEY_PRESSED     = C.ALT_KEY_PRESSED

	// InitializationEvents
	BEHAVIOR_DETACH = C.BEHAVIOR_DETACH
	BEHAVIOR_ATTACH = C.BEHAVIOR_ATTACH

	// DraggingType
	NO_DRAGGING   = C.NO_DRAGGING
	DRAGGING_MOVE = C.DRAGGING_MOVE
	DRAGGING_COPY = C.DRAGGING_COPY

	// MouseButtons
	MAIN_MOUSE_BUTTON   = C.MAIN_MOUSE_BUTTON //aka left button
	PROP_MOUSE_BUTTON   = C.PROP_MOUSE_BUTTON //aka right button
	MIDDLE_MOUSE_BUTTON = C.MIDDLE_MOUSE_BUTTON
	X1_MOUSE_BUTTON     = C.X1_MOUSE_BUTTON
	X2_MOUSE_BUTTON     = C.X2_MOUSE_BUTTON

	// MouseEvents
	MOUSE_ENTER  = C.MOUSE_ENTER
	MOUSE_LEAVE  = C.MOUSE_LEAVE
	MOUSE_MOVE   = C.MOUSE_MOVE
	MOUSE_UP     = C.MOUSE_UP
	MOUSE_DOWN   = C.MOUSE_DOWN
	MOUSE_DCLICK = C.MOUSE_DCLICK
	MOUSE_WHEEL  = C.MOUSE_WHEEL
	MOUSE_TICK   = C.MOUSE_TICK   // mouse pressed ticks
	MOUSE_IDLE   = C.MOUSE_IDLE   // mouse stay idle for some time
	DROP         = C.DROP         // item dropped, target is that dropped item 
	DRAG_ENTER   = C.DRAG_ENTER   // drag arrived to the target element that is one of current drop targets.  
	DRAG_LEAVE   = C.DRAG_LEAVE   // drag left one of current drop targets. target is the drop target element.  
	DRAG_REQUEST = C.DRAG_REQUEST // drag src notification before drag start. To cancel - return true from handler.
	MOUSE_CLICK  = C.MOUSE_CLICK  // mouse click event
	DRAGGING     = C.DRAGGING     // This flag is 'ORed' with MOUSE_ENTER..MOUSE_DOWN codes if dragging operation is in effect.
	// E.g. event DRAGGING | MOUSE_MOVE is sent to underlying DOM elements while dragging.

	// CursorType
	CURSOR_ARROW       = C.CURSOR_ARROW       //0
	CURSOR_IBEAM       = C.CURSOR_IBEAM       //1
	CURSOR_WAIT        = C.CURSOR_WAIT        //2
	CURSOR_CROSS       = C.CURSOR_CROSS       //3
	CURSOR_UPARROW     = C.CURSOR_UPARROW     //4
	CURSOR_SIZENWSE    = C.CURSOR_SIZENWSE    //5
	CURSOR_SIZENESW    = C.CURSOR_SIZENESW    //6
	CURSOR_SIZEWE      = C.CURSOR_SIZEWE      //7
	CURSOR_SIZENS      = C.CURSOR_SIZENS      //8
	CURSOR_SIZEALL     = C.CURSOR_SIZEALL     //9 
	CURSOR_NO          = C.CURSOR_NO          //10
	CURSOR_APPSTARTING = C.CURSOR_APPSTARTING //11
	CURSOR_HELP        = C.CURSOR_HELP        //12
	CURSOR_HAND        = C.CURSOR_HAND        //13
	CURSOR_DRAG_MOVE   = C.CURSOR_DRAG_MOVE   //14 
	CURSOR_DRAG_COPY   = C.CURSOR_DRAG_COPY   //15

	// KeyEvents
	KEY_DOWN = C.KEY_DOWN
	KEY_UP   = C.KEY_UP
	KEY_CHAR = C.KEY_CHAR

	// FocusEvents
	FOCUS_LOST = C.FOCUS_LOST
	FOCUS_GOT  = C.FOCUS_GOT

	// FocusCause
	BY_CODE     = C.BY_CODE
	BY_MOUSE    = C.BY_MOUSE
	BY_KEY_NEXT = C.BY_KEY_NEXT
	BY_KEY_PREV = C.BY_KEY_PREV

	// ScrollEvents
	SCROLL_HOME            = C.SCROLL_HOME
	SCROLL_END             = C.SCROLL_END
	SCROLL_STEP_PLUS       = C.SCROLL_STEP_PLUS
	SCROLL_STEP_MINUS      = C.SCROLL_STEP_MINUS
	SCROLL_PAGE_PLUS       = C.SCROLL_PAGE_PLUS
	SCROLL_PAGE_MINUS      = C.SCROLL_PAGE_MINUS
	SCROLL_POS             = C.SCROLL_POS
	SCROLL_SLIDER_RELEASED = C.SCROLL_SLIDER_RELEASED

	// GestureCmd
	GESTURE_REQUEST = C.GESTURE_REQUEST // return true and fill flags if it will handle gestures.
	GESTURE_ZOOM    = C.GESTURE_ZOOM    // The zoom gesture.
	GESTURE_PAN     = C.GESTURE_PAN     // The pan gesture.
	GESTURE_ROTATE  = C.GESTURE_ROTATE  // The rotation gesture.
	GESTURE_TAP1    = C.GESTURE_TAP1    // The tap gesture.
	GESTURE_TAP2    = C.GESTURE_TAP2    // The two-finger tap gesture.

	// GestureState
	GESTURE_STATE_BEGIN   = C.GESTURE_STATE_BEGIN   // starts
	GESTURE_STATE_INERTIA = C.GESTURE_STATE_INERTIA // events generated by inertia processor
	GESTURE_STATE_END     = C.GESTURE_STATE_END     // end, last event of the gesture sequence

	// GestureTypeFlags
	GESTURE_FLAG_ZOOM             = C.GESTURE_FLAG_ZOOM
	GESTURE_FLAG_ROTATE           = C.GESTURE_FLAG_ROTATE
	GESTURE_FLAG_PAN_VERTICAL     = C.GESTURE_FLAG_PAN_VERTICAL
	GESTURE_FLAG_PAN_HORIZONTAL   = C.GESTURE_FLAG_PAN_HORIZONTAL
	GESTURE_FLAG_TAP1             = C.GESTURE_FLAG_TAP1             // press & tap
	GESTURE_FLAG_TAP2             = C.GESTURE_FLAG_TAP2             // two fingers tap
	GESTURE_FLAG_PAN_WITH_GUTTER  = C.GESTURE_FLAG_PAN_WITH_GUTTER  // PAN_VERTICAL and PAN_HORIZONTAL modifiers
	GESTURE_FLAG_PAN_WITH_INERTIA = C.GESTURE_FLAG_PAN_WITH_INERTIA //
	GESTURE_FLAGS_ALL             = C.GESTURE_FLAGS_ALL             //

	// DrawEvents
	DRAW_BACKGROUND = C.DRAW_BACKGROUND
	DRAW_CONTENT    = C.DRAW_CONTENT
	DRAW_FOREGROUND = C.DRAW_FOREGROUND

	// ExchangeEvents
	X_DRAG_ENTER = C.X_DRAG_ENTER
	X_DRAG_LEAVE = C.X_DRAG_LEAVE
	X_DRAG       = C.X_DRAG
	X_DROP       = C.X_DROP

	// ExchangeDataType
	EXF_UNDEFINED = C.EXF_UNDEFINED
	EXF_TEXT      = C.EXF_TEXT      // FETCH_EXCHANGE_DATA will receive UTF8 encoded string - plain text
	EXF_HTML      = C.EXF_HTML      // FETCH_EXCHANGE_DATA will receive UTF8 encoded string - html
	EXF_HYPERLINK = C.EXF_HYPERLINK // FETCH_EXCHANGE_DATA will receive UTF8 encoded string with pair url\0caption (null separated)
	EXF_JSON      = C.EXF_JSON      // FETCH_EXCHANGE_DATA will receive UTF8 encoded string with JSON literal
	EXF_FILE      = C.EXF_FILE      // FETCH_EXCHANGE_DATA will receive UTF8 encoded list of file names separated by nulls

	// ExchangeCommands
	EXC_NONE = C.EXC_NONE
	EXC_COPY = C.EXC_COPY
	EXC_MOVE = C.EXC_MOVE
	EXC_LINK = C.EXC_LINK

	// BehaviorEvents
	BUTTON_CLICK             = C.BUTTON_CLICK             // click on button
	BUTTON_PRESS             = C.BUTTON_PRESS             // mouse down or key down in button
	BUTTON_STATE_CHANGED     = C.BUTTON_STATE_CHANGED     // checkbox/radio/slider changed its state/value 
	EDIT_VALUE_CHANGING      = C.EDIT_VALUE_CHANGING      // before text change
	EDIT_VALUE_CHANGED       = C.EDIT_VALUE_CHANGED       // after text change
	SELECT_SELECTION_CHANGED = C.SELECT_SELECTION_CHANGED // selection in <select> changed
	SELECT_STATE_CHANGED     = C.SELECT_STATE_CHANGED     // node in select expanded/collapsed, heTarget is the node
	POPUP_REQUEST            = C.POPUP_REQUEST            // request to show popup just received, 
	//     here DOM of popup element can be modifed.
	POPUP_READY = C.POPUP_READY // popup element has been measured and ready to be shown on screen,
	//     here you can use functions like ScrollToView.
	POPUP_DISMISSED = C.POPUP_DISMISSED // popup element is closed,
	//     here DOM of popup element can be modifed again - e.g. some items can be removed
	//     to free memory.
	MENU_ITEM_ACTIVE = C.MENU_ITEM_ACTIVE // menu item activated by mouse hover or by keyboard,
	MENU_ITEM_CLICK  = C.MENU_ITEM_CLICK  // menu item click, 
	//   BEHAVIOR_EVENT_PARAMS structure layout
	//   BEHAVIOR_EVENT_PARAMS.cmd - MENU_ITEM_CLICK/MENU_ITEM_ACTIVE   
	//   BEHAVIOR_EVENT_PARAMS.heTarget - the menu item, presumably <li> element
	//   BEHAVIOR_EVENT_PARAMS.reason - BY_MOUSE_CLICK | BY_KEY_CLICK
	CONTEXT_MENU_SETUP   = C.CONTEXT_MENU_SETUP   // evt.he is a menu dom element that is about to be shown. You can disable/enable items in it.      
	CONTEXT_MENU_REQUEST = C.CONTEXT_MENU_REQUEST // "right-click", BEHAVIOR_EVENT_PARAMS::he is current popup menu HELEMENT being processed or NULL.
	// application can provide its own HELEMENT here (if it is NULL) or modify current menu element.
	VISIUAL_STATUS_CHANGED  = C.VISIUAL_STATUS_CHANGED  // broadcast notification, sent to all elements of some container being shown or hidden   
	DISABLED_STATUS_CHANGED = C.DISABLED_STATUS_CHANGED // broadcast notification, sent to all elements of some container that got new value of :disabled state
	POPUP_DISMISSING        = C.POPUP_DISMISSING        // popup is about to be closed

	// "grey" event codes  - notfications from behaviors from this SDK 
	HYPERLINK_CLICK    = C.HYPERLINK_CLICK    // hyperlink click
	TABLE_HEADER_CLICK = C.TABLE_HEADER_CLICK // click on some cell in table header, 
	//     target = C.the cell, 
	//     reason = C.index of the cell (column number, 0..n)
	TABLE_ROW_CLICK = C.TABLE_ROW_CLICK // click on data row in the table, target is the row
	//     target = C.the row, 
	//     reason = C.index of the row (fixed_rows..n)
	TABLE_ROW_DBL_CLICK = C.TABLE_ROW_DBL_CLICK // mouse dbl click on data row in the table, target is the row
	//     target = C.the row, 
	//     reason = C.index of the row (fixed_rows..n)
	ELEMENT_COLLAPSED = C.ELEMENT_COLLAPSED // element was collapsed, so far only behavior:tabs is sending these two to the panels
	ELEMENT_EXPANDED  = C.ELEMENT_EXPANDED  // element was expanded,
	ACTIVATE_CHILD    = C.ACTIVATE_CHILD    // activate (select) child, 
	// used for example by accesskeys behaviors to send activation request, e.g. tab on behavior:tabs. 
	DO_SWITCH_TAB = C.DO_SWITCH_TAB // command to switch tab programmatically, handled by behavior:tabs 
	// use it as HTMLayoutPostEvent(tabsElementOrItsChild, DO_SWITCH_TAB, tabElementToShow, 0);
	INIT_DATA_VIEW    = C.INIT_DATA_VIEW    // request to virtual grid to initialize its view
	ROWS_DATA_REQUEST = C.ROWS_DATA_REQUEST // request from virtual grid to data source behavior to fill data in the table
	// parameters passed throug DATA_ROWS_PARAMS structure.
	UI_STATE_CHANGED = C.UI_STATE_CHANGED // ui state changed, observers shall update their visual states.
	// is sent for example by behavior:richtext when caret position/selection has changed.
	FORM_SUBMIT = C.FORM_SUBMIT // behavior:form detected submission event. BEHAVIOR_EVENT_PARAMS::data field contains data to be posted.
	// BEHAVIOR_EVENT_PARAMS::data is of type T_MAP in this case key/value pairs of data that is about 
	// to be submitted. You can modify the data or discard submission by returning TRUE from the handler.
	FORM_RESET = C.FORM_RESET // behavior:form detected reset event (from button type = C.reset). BEHAVIOR_EVENT_PARAMS::data field contains data to be reset.
	// BEHAVIOR_EVENT_PARAMS::data is of type T_MAP in this case key/value pairs of data that is about 
	// to be rest. You can modify the data or discard reset by returning TRUE from the handler.		 
	DOCUMENT_COMPLETE            = C.DOCUMENT_COMPLETE // behavior:frame have complete document.
	HISTORY_PUSH                 = C.HISTORY_PUSH      // behavior:history stuff
	HISTORY_DROP                 = C.HISTORY_DROP
	HISTORY_PRIOR                = C.HISTORY_PRIOR
	HISTORY_NEXT                 = C.HISTORY_NEXT
	HISTORY_STATE_CHANGED        = C.HISTORY_STATE_CHANGED // behavior:history notification - history stack has changed
	CLOSE_POPUP                  = C.CLOSE_POPUP           // close popup request,
	REQUEST_TOOLTIP              = C.REQUEST_TOOLTIP       // request tooltip, BEHAVIOR_EVENT_PARAMS.he <- is the tooltip element.
	ANIMATION                    = C.ANIMATION             // animation started (reason = C.1) or ended(reason = C.0) on the element.
	FIRST_APPLICATION_EVENT_CODE = C.FIRST_APPLICATION_EVENT_CODE
	// all custom event codes shall be greater
	// than this number. All codes below this will be used
	// solely by application - HTMLayout will not intrepret it 
	// and will do just dispatching.
	// To send event notifications with  these codes use
	// HTMLayoutSend/PostEvent API.

	// EventReason
	BY_MOUSE_CLICK = C.BY_MOUSE_CLICK
	BY_KEY_CLICK   = C.BY_KEY_CLICK
	SYNTHESIZED    = C.SYNTHESIZED

	// EventChangedReason
	BY_INS_CHAR  = C.BY_INS_CHAR  // single char insertion
	BY_INS_CHARS = C.BY_INS_CHARS // character range insertion, clipboard
	BY_DEL_CHAR  = C.BY_DEL_CHAR  // single char deletion
	BY_DEL_CHARS = C.BY_DEL_CHARS // character range deletion (selection)

	// BehaviorMethodIdentifiers
	DO_CLICK                = C.DO_CLICK
	GET_TEXT_VALUE          = C.GET_TEXT_VALUE
	SET_TEXT_VALUE          = C.SET_TEXT_VALUE          // p - TEXT_VALUE_PARAMS
	TEXT_EDIT_GET_SELECTION = C.TEXT_EDIT_GET_SELECTION // p - TEXT_EDIT_SELECTION_PARAMS
	TEXT_EDIT_SET_SELECTION = C.TEXT_EDIT_SET_SELECTION // p - TEXT_EDIT_SELECTION_PARAMS
	// Replace selection content or insert text at current caret position.
	// Replaced text will be selected. 
	TEXT_EDIT_REPLACE_SELECTION = C.TEXT_EDIT_REPLACE_SELECTION // p - TEXT_EDIT_REPLACE_SELECTION_PARAMS
	// Set value of type = C."vscrollbar"/"hscrollbar"
	SCROLL_BAR_GET_VALUE = C.SCROLL_BAR_GET_VALUE
	SCROLL_BAR_SET_VALUE = C.SCROLL_BAR_SET_VALUE
	// get current caret position, it returns rectangle that is relative to origin of the editing element.
	TEXT_EDIT_GET_CARET_POSITION = C.TEXT_EDIT_GET_CARET_POSITION // p - TEXT_CARET_POSITION_PARAMS
	TEXT_EDIT_GET_SELECTION_TEXT = C.TEXT_EDIT_GET_SELECTION_TEXT // p - TEXT_SELECTION_PARAMS, OutputStreamProc will receive stream of WCHARs
	TEXT_EDIT_GET_SELECTION_HTML = C.TEXT_EDIT_GET_SELECTION_HTML // p - TEXT_SELECTION_PARAMS, OutputStreamProc will receive stream of BYTEs - utf8 encoded html fragment.
	TEXT_EDIT_CHAR_POS_AT_XY     = C.TEXT_EDIT_CHAR_POS_AT_XY     // p - TEXT_EDIT_CHAR_POS_AT_XY_PARAMS
	IS_EMPTY                     = C.IS_EMPTY                     // p - IS_EMPTY_PARAMS // set VALUE_PARAMS::is_empty (false/true) reflects :empty state of the element.
	GET_VALUE                    = C.GET_VALUE                    // p - VALUE_PARAMS 
	SET_VALUE                    = C.SET_VALUE                    // p - VALUE_PARAMS 
	XCALL                        = C.XCALL                        // p - XCALL_PARAMS
	FIRST_APPLICATION_METHOD_ID  = C.FIRST_APPLICATION_METHOD_ID


	// Content insertion locations
	SIH_REPLACE_CONTENT = C.SIH_REPLACE_CONTENT
	SIH_INSERT_AT_START = C.SIH_INSERT_AT_START
	SIH_APPEND_AFTER_LAST = C.SIH_APPEND_AFTER_LAST
	SOH_REPLACE = C.SOH_REPLACE
	SOH_INSERT_BEFORE = C.SOH_INSERT_BEFORE
	SOH_INSERT_AFTER = C.SOH_INSERT_AFTER

)

var (
	// Hold a reference to handlers that are in-use so that they don't
	// get garbage collected.
	notifyHandlers = make(map[uintptr]*NotifyHandler, 8)
	eventHandlers = make(map[uintptr]*EventHandler, 128)
)



type HELEMENT C.HELEMENT

type Point struct {
	X int32
	Y int32
}

type Size struct {
	Cx int32
	Cy int32
}

type Rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type JsonValue struct {
	T uint32
	U uint32
	D uint64
}

type InitializationParams struct {
	Cmd uint32
}

type MouseParams struct {
	Cmd         uint32 // MouseEvents
	Target      HELEMENT
	Pos         Point
	DocumentPos Point
	ButtonState uint32
	AltState    uint32
	CursorType  uint32
	IsOnIcon    int32

	Dragging     HELEMENT
	DraggingMode uint32
}

type KeyParams struct {
	Cmd      uint32 // KeyEvents
	Target   HELEMENT
	KeyCode  uint32
	AltState uint32
}

type FocusParams struct {
	Cmd          uint32 // FocusEvents
	Target       HELEMENT
	ByMouseClick int32 // boolean
	Cancel       int32 // boolean
}

type DrawParams struct {
	Cmd      uint32 // DrawEvents
	Hdc      C.HDC
	Area     Rect
	reserved uint32
}

type TimerParams struct {
	TimerId uintptr
}

type BehaviorEventParams struct {
	Cmd    uint32 // Behavior events
	Target HELEMENT
	Source HELEMENT
	Reason uint32
	Data   JsonValue
}

type MethodParams struct {
	MethodId uint32
}

// TODO: Add all the structures derived from MethodParams here...

type DataArrivedParams struct {
	Initiator HELEMENT
	Data      *byte
	DataSize  uint32
	DataType  uint32
	Status    uint32
	Uri       *uint16 // Wide character string
}

type ScrollParams struct {
	Cmd      uint32
	Target   HELEMENT
	Pos      int32
	Vertical int32 // bool
}

type ExchangeParams struct {
	Cmd       uint32
	Target    HELEMENT
	Pos       Point
	PosView   Point
	DataTypes uint32
	DragCmd   uint32
	FetchData uintptr // func pointer: typedef BOOL CALLBACK FETCH_EXCHANGE_DATA(EXCHANGE_PARAMS* params, UINT data_type, LPCBYTE* ppDataStart, UINT* pDataLength );
}

type GestureParams struct {
	Cmd       uint32
	Target    HELEMENT
	Pos       Point
	PosView   Point
	Flags     uint32
	DeltaTime uint32
	DeltaXY   Size
	DeltaV    float64
}

// Notify structures

type NMHDR struct {
	HwndFrom		uint32
	IdFrom			uintptr
	Code 			uint32
}


type NmhlCreateControl struct {
	Header         NMHDR
	Element        HELEMENT
	InHwndParent   uint32
	OutHwndControl uint32
	reserved1      int32
	reserved2      int32
}

type NmhlDestroyControl struct {
	Header           NMHDR
	Element          HELEMENT
	InOutHwndControl uint32
	reserved1        int32
}

type NmhlLoadData struct {
	Header      NMHDR
	Uri         *uint16
	OutData     uintptr
	OutDataSize int32
	DataType    uint32
	Principal   HELEMENT
	Initiator   HELEMENT
}

type NmhlDataLoaded struct {
	Header   NMHDR
	Uri      *uint16
	Data     uintptr
	DataSize int32
	DataType uint32
	Status   uint32
}

type NmhlAttachBehavior struct {
	Header        NMHDR
	Element       HELEMENT
	BehaviorName  *C.char
	ElementProc   uintptr
	ElementTag    uintptr
	ElementEvents uint32
}


// Main event handler that dispatches to the right element handler
//export goElementProc 
func goElementProc(tag uintptr, he unsafe.Pointer, evtg uint32, params unsafe.Pointer) C.BOOL {
	key := uintptr(tag)

	var handler *EventHandler
	var exists bool
	if handler, exists = eventHandlers[key]; !exists {
		log.Print("Warning: No handler for tag ", tag)
		return C.FALSE
	}

	handled := false
	switch evtg {
	case C.HANDLE_INITIALIZATION:
		if p := (*InitializationParams)(params); p.Cmd == BEHAVIOR_ATTACH {
			if handler.OnAttached != nil {
				handler.OnAttached(HELEMENT(he))
			}
		} else if p.Cmd == BEHAVIOR_DETACH {
			if handler.OnDetached != nil {
				handler.OnDetached(HELEMENT(he))
			}
			eventHandlers[key] = handler, false
		}
		handled = true
	case C.HANDLE_MOUSE:
		if handler.OnMouse != nil {
			p := (*MouseParams)(params)
			handled = handler.OnMouse(HELEMENT(he), p)
		}
	case C.HANDLE_KEY:
		if handler.OnKey != nil {
			p := (*KeyParams)(params)
			handled = handler.OnKey(HELEMENT(he), p)
		}
	case C.HANDLE_FOCUS:
		if handler.OnFocus != nil {
			p := (*FocusParams)(params)
			handled = handler.OnFocus(HELEMENT(he), p)
		}
	case C.HANDLE_DRAW:
		if handler.OnDraw != nil {
			p := (*DrawParams)(params)
			handled = handler.OnDraw(HELEMENT(he), p)
		}
	case C.HANDLE_TIMER:
		if handler.OnTimer != nil {
			p := (*TimerParams)(params)
			handled = handler.OnTimer(HELEMENT(he), p)
		}
	case C.HANDLE_BEHAVIOR_EVENT:
		if handler.OnBehaviorEvent != nil {
			p := (*BehaviorEventParams)(params)
			handled = handler.OnBehaviorEvent(HELEMENT(he), p)
		}
	case C.HANDLE_METHOD_CALL:
		if handler.OnMethodCall != nil {
			p := (*MethodParams)(params)
			handled = handler.OnMethodCall(HELEMENT(he), p)
		}
	case C.HANDLE_DATA_ARRIVED:
		if handler.OnDataArrived != nil {
			p := (*DataArrivedParams)(params)
			handled = handler.OnDataArrived(HELEMENT(he), p)
		}
	case C.HANDLE_SIZE:
		if handler.OnSize != nil {
			handler.OnSize(HELEMENT(he))
		}
	case C.HANDLE_SCROLL:
		if handler.OnScroll != nil {
			p := (*ScrollParams)(params)
			handled = handler.OnScroll(HELEMENT(he), p)
		}
	case C.HANDLE_EXCHANGE:
		if handler.OnExchange != nil {
			p := (*ExchangeParams)(params)
			handled = handler.OnExchange(HELEMENT(he), p)
		}
	case C.HANDLE_GESTURE:
		if handler.OnGesture != nil {
			p := (*GestureParams)(params)
			handled = handler.OnGesture(HELEMENT(he), p)
		}
	default:
		log.Panic("unhandled htmlayout event case: ", evtg)
	}

	if handled {
		return C.TRUE
	}
	return C.FALSE
}

//export goNotifyProc
func goNotifyProc(msg uint32, wparam uintptr, lparam uintptr, vparam uintptr) uintptr {
	if handler, exists := notifyHandlers[vparam]; exists {
		phdr := (*C.NMHDR)(unsafe.Pointer(lparam))

		switch phdr.code {
		case HLN_CREATE_CONTROL:
			if handler.OnCreateControl != nil {
				return handler.OnCreateControl((*NmhlCreateControl)(unsafe.Pointer(lparam)))
			}
		case HLN_CONTROL_CREATED:
			if handler.OnControlCreated != nil {
				return handler.OnControlCreated((*NmhlCreateControl)(unsafe.Pointer(lparam)))
			}
		case HLN_DESTROY_CONTROL:
			if handler.OnDestroyControl != nil {
				return handler.OnDestroyControl((*NmhlDestroyControl)(unsafe.Pointer(lparam)))
			}
		case HLN_LOAD_DATA:
			if handler.OnLoadData != nil {
				return handler.OnLoadData((*NmhlLoadData)(unsafe.Pointer(lparam)))
			}
		case HLN_DATA_LOADED:
			if handler.OnDataLoaded != nil {
				return handler.OnDataLoaded((*NmhlDataLoaded)(unsafe.Pointer(lparam)))
			}
		case HLN_DOCUMENT_COMPLETE:
			if handler.OnDocumentComplete != nil {
				return handler.OnDocumentComplete()
			}
		case HLN_ATTACH_BEHAVIOR:
			params := (*NmhlAttachBehavior)(unsafe.Pointer(lparam))
			key := C.GoString(params.BehaviorName)
			if constructor, exists := handler.Behaviors[key]; exists {
				NewElement(params.Element).AttachHandler(constructor())
			} else {
				log.Print("No such behavior: ", key)
			}
		}
	}
	return 0
}

//export goSelectCallback
func goSelectCallback(he unsafe.Pointer, param uintptr) uintptr {
	slice := (*[]*Element)(unsafe.Pointer(param))
	*slice = append(*slice, NewElement(HELEMENT(he)))
	return 0
}

//export goElementComparator
func goElementComparator(he1 unsafe.Pointer, he2 unsafe.Pointer, arg uintptr) int {
	cmp := *(*func(*Element, *Element) int)(unsafe.Pointer(arg))
	return cmp(NewElement(HELEMENT(he1)), NewElement(HELEMENT(he2)))
}



// Main htmlayout wndproc
func ProcNoDefault(hwnd, msg uint32, wparam, lparam uintptr) (uintptr, bool) {
	var handled C.BOOL = 0
	var result C.LRESULT = C.HTMLayoutProcND(C.HWND(C.HANDLE(uintptr(hwnd))), C.UINT(msg),
		C.WPARAM(wparam), C.LPARAM(lparam), &handled)
	return uintptr(result), handled != 0
}

// Load html contents into window
func LoadHtml(hwnd uint32, data []byte, baseUrl string) os.Error {
	if ok := C.HTMLayoutLoadHtmlEx(C.HWND(C.HANDLE(uintptr(hwnd))), (*C.BYTE)(&data[0]),
		C.UINT(len(data)), (*C.WCHAR)(stringToUtf16Ptr(baseUrl))); ok == 0 {
		return os.NewError("HTMLayoutLoadHtmlEx failed")
	}
	return nil
}

// Call this from your NotifyHandler.HandleLoadData method if you want htmlayout to
// process the data right away so you don't have to provide a buffer in the NmhlLoadData structure.
func DataReady(hwnd uint32, uri *uint16, data *byte, dataLength int32) bool {
	return C.HTMLayoutDataReady(C.HWND(C.HANDLE(uintptr(hwnd))), (*C.WCHAR)(uri), (*C.BYTE)(data), C.DWORD(dataLength)) != 0
}

func AttachWindowEventHandler(hwnd uint32, handler *EventHandler) {
	key := uintptr(hwnd)
	if _, exists := eventHandlers[key]; exists {
		if ret := C.HTMLayoutWindowDetachEventHandler(C.HWND(C.HANDLE(key)), C.ElementProcAddr, C.LPVOID(key)); ret != HLDOM_OK {
			domPanic(ret, "Failed to detach event handler from window before adding the new one")
		}
		eventHandlers[key] = nil, false
	}
	eventHandlers[key] = handler

	// Don't let the caller disable ATTACH/DETACH events, otherwise we
	// won't know when to throw out our event handler object
	subscription := handler.GetSubscription()
	subscription &= ^DISABLE_INITIALIZATION

	if ret := C.HTMLayoutWindowAttachEventHandler(C.HWND(C.HANDLE(key)), C.ElementProcAddr, C.LPVOID(key), C.UINT(subscription)); ret != HLDOM_OK {
		domPanic(ret, "Failed to attach event handler to window")
	}
}

func DetachWindowEventHandler(hwnd uint32) {
	key := uintptr(hwnd)
	if handler, exists := eventHandlers[key]; exists {
		if ret := C.HTMLayoutWindowDetachEventHandler(C.HWND(C.HANDLE(key)), C.ElementProcAddr, C.LPVOID(key)); ret != HLDOM_OK {
			domPanic(ret, "Failed to detach event handler from window")
		}
		eventHandlers[key] = handler, false
	}
}

func AttachNotifyHandler(hwnd uint32, handler *NotifyHandler) {
	key := uintptr(hwnd)
	if _, exists := notifyHandlers[key]; exists {
		notifyHandlers[key] = nil, false
	}
	notifyHandlers[key] = handler
	C.HTMLayoutSetCallback(C.HWND(C.HANDLE(key)), C.NotifyProcAddr, C.LPVOID(key))
}

func DetachNotifyHandler(hwnd uint32) {
	key := uintptr(hwnd)
	if handler, exists := notifyHandlers[key]; exists {
		C.HTMLayoutSetCallback(C.HWND(C.HANDLE(key)), nil, nil)
		notifyHandlers[key] = handler, false
	}
}

func DumpObjectCounts() {
	log.Print("Window/element event handlers (", len(eventHandlers), "): ", eventHandlers)
	log.Print("Window notify handlers (", len(notifyHandlers), "): ", notifyHandlers)
}