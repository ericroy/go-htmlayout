package gohl

type EventHandler struct {
	OnAttached func(he HELEMENT)
	OnDetached func(he HELEMENT)
	OnMouse func(he HELEMENT, params *MouseParams) bool
	OnKey func(he HELEMENT, params *KeyParams) bool
	OnFocus func(he HELEMENT, params *FocusParams) bool
	OnDraw func(he HELEMENT, params *DrawParams) bool
	OnTimer func(he HELEMENT, params *TimerParams) bool
	OnBehaviorEvent func(he HELEMENT, params *BehaviorEventParams) bool
	OnMethodCall func(he HELEMENT, params *MethodParams) bool
	OnDataArrived func(he HELEMENT, params *DataArrivedParams) bool
	OnSize func(he HELEMENT)
	OnScroll func(he HELEMENT, params *ScrollParams) bool
	OnExchange func(he HELEMENT, params *ExchangeParams) bool
	OnGesture func(he HELEMENT, params *GestureParams) bool
}

func (e *EventHandler) GetSubscription() uint32 {
	var subscription uint32 = 0
	add := func (f interface{}, flag uint32) {
		if f != nil {
			subscription |= flag
		}
	}

	// OnAttached and OnDetached purposely omitted, since we must receive these events
	add(e.OnMouse, HANDLE_MOUSE)
	add(e.OnKey, HANDLE_KEY)
	add(e.OnFocus, HANDLE_FOCUS)
	add(e.OnDraw, HANDLE_DRAW)
	add(e.OnTimer, HANDLE_TIMER)
	add(e.OnBehaviorEvent, HANDLE_BEHAVIOR_EVENT)
	add(e.OnMethodCall, HANDLE_METHOD_CALL)
	add(e.OnDataArrived, HANDLE_DATA_ARRIVED)
	add(e.OnSize, HANDLE_SIZE)
	add(e.OnScroll, HANDLE_SCROLL)
	add(e.OnExchange, HANDLE_EXCHANGE)
	add(e.OnGesture, HANDLE_GESTURE)

	return subscription
}