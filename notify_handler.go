package gohl

type NotifyHandler struct {
	Behaviors          map[string]*EventHandler
	OnCreateControl    func(params *NmhlCreateControl) uintptr
	OnControlCreated   func(params *NmhlCreateControl) uintptr
	OnDestroyControl   func(params *NmhlDestroyControl) uintptr
	OnLoadData         func(params *NmhlLoadData) uintptr
	OnDataLoaded       func(params *NmhlDataLoaded) uintptr
	OnDocumentComplete func() uintptr
}
