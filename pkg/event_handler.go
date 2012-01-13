package gohl

import (
	"log"
	"unsafe"
)

type EventHandler interface {
	// Should return the pointer to the actual event handler instance.  This
	// will be used as a key to look up the event handler later when being
	// unregistered
	GetAddress() uintptr

	// Called when the event handler is attached to and detached from the element.
	// These function calls are bumpers to all the others below.
	Attached(he HELEMENT)
	Detached(he HELEMENT)

	HandleMouse(he HELEMENT, params *MouseParams) bool
	HandleKey(he HELEMENT, params *KeyParams) bool
	HandleFocus(he HELEMENT, params *FocusParams) bool
	HandleDraw(he HELEMENT, params *DrawParams) bool
	HandleTimer(he HELEMENT, params *TimerParams) bool
	HandleBehaviorEvent(he HELEMENT, params *BehaviorEventParams) bool
	HandleMethodCall(he HELEMENT, params *MethodParams) bool
	HandleDataArrived(he HELEMENT, params *DataArrivedParams) bool
	HandleSize(he HELEMENT)
	HandleScroll(he HELEMENT, params *ScrollParams) bool
	HandleExchange(he HELEMENT, params *ExchangeParams) bool
	HandleGesture(he HELEMENT, params *GestureParams) bool
}

// Default implementation of the EventHandler interface that does nothing
type EventHandlerBase struct{}

func (e *EventHandlerBase) GetAddress() uintptr {
	return uintptr(unsafe.Pointer(e))
}

func (e *EventHandlerBase) Attached(he HELEMENT) {
	log.Print("Event handler attached")
}

func (e *EventHandlerBase) Detached(he HELEMENT) {
	log.Print("Event handler detached")
}

func (e *EventHandlerBase) HandleMouse(he HELEMENT, params *MouseParams) bool {
	return false
}

func (e *EventHandlerBase) HandleKey(he HELEMENT, params *KeyParams) bool {
	return false
}

func (e *EventHandlerBase) HandleFocus(he HELEMENT, params *FocusParams) bool {
	return false
}

func (e *EventHandlerBase) HandleDraw(he HELEMENT, params *DrawParams) bool {
	return false
}

func (e *EventHandlerBase) HandleTimer(he HELEMENT, params *TimerParams) bool {
	return false
}

func (e *EventHandlerBase) HandleBehaviorEvent(he HELEMENT, params *BehaviorEventParams) bool {
	return false
}

func (e *EventHandlerBase) HandleMethodCall(he HELEMENT, params *MethodParams) bool {
	return false
}

func (e *EventHandlerBase) HandleDataArrived(he HELEMENT, params *DataArrivedParams) bool {
	return false
}

func (e *EventHandlerBase) HandleSize(he HELEMENT) {
}

func (e *EventHandlerBase) HandleScroll(he HELEMENT, params *ScrollParams) bool {
	return false
}

func (e *EventHandlerBase) HandleExchange(he HELEMENT, params *ExchangeParams) bool {
	return false
}

func (e *EventHandlerBase) HandleGesture(he HELEMENT, params *GestureParams) bool {
	return false
}