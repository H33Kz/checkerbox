package event

type Event struct {
	Type string
	Data interface{}
}

type SequenceEvent struct {
	Label      string
	DeviceName string
	Function   string
	Site       int
	Timeout    int
	Threshold  string
	Data       string
}

type EventBus struct {
	subscribers map[string][]chan<- Event
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan<- Event),
	}
}

func (eBus *EventBus) Subscribe(eventType string, eventChan chan<- Event) {
	eBus.subscribers[eventType] = append(eBus.subscribers[eventType], eventChan)
}

func (eBus *EventBus) Publish(event Event) {
	for _, subscriber := range eBus.subscribers[event.Type] {
		subscriber <- event
	}
}
