package device

import (
	"checkerbox/internal/event"
	"checkerbox/internal/test"
	"fmt"
	"time"
)

type SequenceDevice struct {
	eventChannel chan event.Event
	site         int
}

func NewSequenceDevice(site int) *SequenceDevice {
	return &SequenceDevice{
		site:         site,
		eventChannel: make(chan event.Event),
	}
}

func (s *SequenceDevice) SequenceEventHandler() {
	for receivedEvent := range s.eventChannel {
		sequenceEvent, ok := receivedEvent.Data.(event.SequenceEvent)
		if !ok || sequenceEvent.DeviceName != "sequence" || sequenceEvent.Site != s.site {
			continue
		}

		siteResultChannel := receivedEvent.ReturnChannel
		result := s.functionResolver(sequenceEvent)
		result.Site = sequenceEvent.Site
		result.Id = sequenceEvent.Id
		result.Label = sequenceEvent.Label
		siteResultChannel <- result
	}
}

func (s *SequenceDevice) GetEventChannel() chan event.Event {
	return s.eventChannel
}

func (s *SequenceDevice) functionResolver(sequenceEvent event.SequenceEvent) test.Result {
	function, ok := sequenceEvent.StepSettings["function"].(string)
	if !ok {
		return test.Result{Result: test.Error, Message: "Error parsing function name"}
	}

	switch function {
	case "Wait":
		data, ok := sequenceEvent.StepSettings["data"].(int)
		if !ok {
			return test.Result{Result: test.Error, Message: "Error parsing time to wait"}
		}
		time.Sleep(time.Duration(data) * time.Millisecond)
		return test.Result{Result: test.Done, Message: "Wait " + fmt.Sprintf("%v", data) + "mS"}
	default:
		return test.Result{Result: test.Error, Message: "Function not found: "}
	}
}

func (s *SequenceDevice) Print() {
	fmt.Println("Sequence device at site: " + fmt.Sprintf("%v", s.site))
}
