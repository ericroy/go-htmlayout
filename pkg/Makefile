include $(GOROOT)/src/Make.inc

TARG=miniondev/gohl
CGOFILES=htmlayout.go element.go notify_handler.go
GOFILES=event_handler.go behavior.go

include $(GOROOT)/src/Make.pkg

.PHONY: format

format:
	gofmt -w .