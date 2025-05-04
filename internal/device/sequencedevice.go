package device

import (
	"checkerbox/internal/event"
	"checkerbox/internal/test"
	"fmt"
	"strconv"
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
		siteResultChannel <- result
	}
}

func (s *SequenceDevice) GetEventChannel() chan event.Event {
	return s.eventChannel
}

func (s *SequenceDevice) functionResolver(sequenceEvent event.SequenceEvent) test.Result {
	switch sequenceEvent.Function {
	case "Wait":
		timeToWait, err := strconv.ParseInt(sequenceEvent.Data, 10, 32)
		if err != nil {
			return test.Result{Result: test.Error, Message: err.Error(), Site: sequenceEvent.Site, Id: sequenceEvent.Id, Label: sequenceEvent.Label}
		}
		time.Sleep(time.Duration(timeToWait) * time.Millisecond)
		return test.Result{Result: test.Done, Message: "Wait " + sequenceEvent.Data + "mS", Site: sequenceEvent.Site, Id: sequenceEvent.Id, Label: sequenceEvent.Label}
	default:
		return test.Result{Result: test.Error, Message: "Function not found: " + sequenceEvent.Label, Site: sequenceEvent.Site}
	}
}

func (s *SequenceDevice) Print() {
	fmt.Println("Sequence device at site: " + fmt.Sprintf("%v", s.site))
}
