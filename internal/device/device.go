package device

import (
	"checkerbox/internal/event"
	"checkerbox/internal/test"
)

type Device interface {
	SequenceEventHandler()
	functionResolver(event.SequenceEvent) test.Result
	GetEventChannel() chan event.Event
	Print()
}
