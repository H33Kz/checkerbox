package device

import (
	"checkerbox/internal/event"
	"checkerbox/internal/test"
)

type Device interface {
	SequenceEventHandler(chan test.Result)
	FunctionResolver()
	GetEventChannel() chan event.Event
	Print()
}
