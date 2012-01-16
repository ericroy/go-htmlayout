package gohl

import "C"
import (
	"log"
)

type NotifyHandler interface {
	HandleCreateControl(params *NmhlCreateControl) uintptr
	HandleControlCreated(params *NmhlCreateControl) uintptr
	HandleDestroyControl(params *NmhlDestroyControl) uintptr
	HandleLoadData(params *NmhlLoadData) uintptr
	HandleDataLoaded(params *NmhlDataLoaded) uintptr
	HandleDocumentComplete() uintptr
	HandleAttachBehavior(params *NmhlAttachBehavior) uintptr
}

type EventHandlerConstructor func() EventHandler

// Default implementation of the NotifyHandler interface that does nothing
type NotifyHandlerBase struct {
	behaviors map[string]EventHandlerConstructor
}

func NewNotifyHandlerBase() *NotifyHandlerBase {
	return &NotifyHandlerBase{make(map[string]EventHandlerConstructor, 16)}
}

func (n *NotifyHandlerBase) RegisterBehavior(name string, constructor EventHandlerConstructor) {
	n.behaviors[name] = constructor
}

func (n *NotifyHandlerBase) UnregisterBehavior(name string) {
	n.behaviors[name] = nil, false
}

func (n *NotifyHandlerBase) HandleCreateControl(params *NmhlCreateControl) uintptr {
	return 0
}

func (n *NotifyHandlerBase) HandleControlCreated(params *NmhlCreateControl) uintptr {
	return 0
}

func (n *NotifyHandlerBase) HandleDestroyControl(params *NmhlDestroyControl) uintptr {
	return 0
}

func (n *NotifyHandlerBase) HandleLoadData(params *NmhlLoadData) uintptr {
	return 0
}

func (n *NotifyHandlerBase) HandleDataLoaded(params *NmhlDataLoaded) uintptr {
	return 0
}

func (n *NotifyHandlerBase) HandleDocumentComplete() uintptr {
	return 0
}

func (n *NotifyHandlerBase) HandleAttachBehavior(params *NmhlAttachBehavior) uintptr {
	key := C.GoString(params.BehaviorName)
	if constructor, exists := n.behaviors[key]; exists {
		NewElement(params.Element).AttachHandler(constructor())
	} else {
		log.Panic("No such behavior: ", key)
	}
	return 0
}
