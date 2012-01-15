package gohl

type NotifyHandler interface {
	HandleCreateControl(params *NmhlCreateControl) uintptr
	HandleControlCreated(params *NmhlCreateControl) uintptr
	HandleDestroyControl(params *NmhlDestroyControl) uintptr
	HandleLoadData(params *NmhlLoadData) uintptr
	HandleDataLoaded(params *NmhlDataLoaded) uintptr
	HandleDocumentComplete() uintptr
	HandleAttachBehavior(params *NmhlAttachBehavior) uintptr
}

// Default implementation of the NotifyHandler interface that does nothing
type NotifyHandlerBase struct {
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
	return 0
}
