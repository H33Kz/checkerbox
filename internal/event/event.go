package event

import (
	"checkerbox/internal/test"
	"sync"
)

type Event struct {
	Type          string
	ReturnChannel chan test.Result
	Data          any
}

type SequenceEvent struct {
	Id           uint
	Label        string
	DeviceName   string
	Retry        int
	Site         int
	Timeout      string
	StepSettings map[string]any
}

type GraphicEvent struct {
	Type   string
	Result test.Result
}

type EventBus struct {
	mutex       sync.Mutex
	subscribers map[string][]chan<- Event
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan<- Event),
	}
}

func (eBus *EventBus) Subscribe(eventType string, eventChan chan<- Event) {
	eBus.mutex.Lock()
	defer eBus.mutex.Unlock()
	eBus.subscribers[eventType] = append(eBus.subscribers[eventType], eventChan)
}

func (eBus *EventBus) Publish(event Event) {
	eBus.mutex.Lock()
	defer eBus.mutex.Unlock()
	for _, subscriber := range eBus.subscribers[event.Type] {
		subscriber <- event
	}
}
