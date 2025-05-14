package device

import (
	"checkerbox/internal/event"
	"checkerbox/internal/test"
	"fmt"
)

type TestDevice struct {
	eventChannel chan event.Event
	site         int
}

func NewTestDevice(site int) (*TestDevice, error) {
	return &TestDevice{
		site:         int(site),
		eventChannel: make(chan event.Event),
	}, nil
}

func (t *TestDevice) SequenceEventHandler() {
	for receivedEvent := range t.eventChannel {
		sequenceEvent, ok := receivedEvent.Data.(event.SequenceEvent)
		if !ok || sequenceEvent.DeviceName != "testdevice" || sequenceEvent.Site != t.site {
			continue
		}

		siteResultChannel := receivedEvent.ReturnChannel
		result := t.functionResolver(sequenceEvent)
		result.Site = sequenceEvent.Site
		result.Id = sequenceEvent.Id
		result.Label = sequenceEvent.Label
		siteResultChannel <- result
	}
}

func (t *TestDevice) GetEventChannel() chan event.Event {
	return t.eventChannel
}

func (t *TestDevice) functionResolver(sequenceEvent event.SequenceEvent) test.Result {
	function, ok := sequenceEvent.StepSettings["function"].(string)
	if !ok {
		return test.Result{Result: test.Error, Message: "Error parsing function name"}
	}

	switch function {
	case "TestAction1":
		return test.Result{Result: test.Done, Message: "TestAction1"}
	case "TestAction2":
		return test.Result{Result: test.Done, Message: "TestAction2"}
	case "TestAction3":
		return test.Result{Result: test.Done, Message: "TestAction3"}
	default:
		return test.Result{Result: test.Error, Message: "Function not found: "}
	}
}

func (t *TestDevice) Print() {
	fmt.Println("Test device at site: " + fmt.Sprintf("%v", t.site))
}
