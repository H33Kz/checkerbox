package event

import (
	"checkerbox/internal/data"
	"checkerbox/internal/test"
	"slices"
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
	Timeout      int
	StepSettings map[string]any
}

type GraphicEvent struct {
	Type   string
	Result test.Result
	Log    data.Log
}

type ControlEvent struct {
	Type string
	Data any
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

func (eBus *EventBus) PublishAndDelete(event Event) {
	eBus.mutex.Lock()
	defer eBus.mutex.Unlock()
	for i, subscriber := range eBus.subscribers[event.Type] {
		subscriber <- event
		eBus.subscribers[event.Type] = slices.Delete(eBus.subscribers[event.Type], i, i)
	}
}
