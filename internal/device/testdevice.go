package device

import (
	"checkerbox/internal/event"
	"checkerbox/internal/test"
	"errors"
	"fmt"
	"strconv"
)

type TestDevice struct {
	eventChannel chan event.Event
	site         int
}

func NewTestDevice(deviceMap map[string]string) (*TestDevice, []error) {
	var errorTable []error
	site, siteError := strconv.ParseInt(deviceMap["site"], 10, 8)
	if siteError != nil {
		errorTable = append(errorTable, errors.New("Unable to parse site name for: "+deviceMap["device"]+"\nSetting of site 1"))
		site = 1
	}
	return &TestDevice{
		site:         int(site),
		eventChannel: make(chan event.Event),
	}, errorTable
}

func (t *TestDevice) SequenceEventHandler() {
	for receivedEvent := range t.eventChannel {
		sequenceEvent, ok := receivedEvent.Data.(event.SequenceEvent)
		if !ok || sequenceEvent.DeviceName != "testdevice" || sequenceEvent.Site != t.site {
			continue
		}

		siteResultChannel := receivedEvent.ReturnChannel
		result := t.functionResolver(sequenceEvent)
		siteResultChannel <- result
	}
}

func (t *TestDevice) GetEventChannel() chan event.Event {
	return t.eventChannel
}

func (t *TestDevice) functionResolver(sequenceEvent event.SequenceEvent) test.Result {
	switch sequenceEvent.Function {
	case "TestAction1":
		return test.Result{Result: test.Done, Message: "TestAction1", Site: sequenceEvent.Site, Id: sequenceEvent.Id, Label: sequenceEvent.Label}
	case "TestAction2":
		return test.Result{Result: test.Done, Message: "TestAction2", Site: sequenceEvent.Site, Id: sequenceEvent.Id, Label: sequenceEvent.Label}
	case "TestAction3":
		return test.Result{Result: test.Done, Message: "TestAction3", Site: sequenceEvent.Site, Id: sequenceEvent.Id, Label: sequenceEvent.Label}
	default:
		return test.Result{Result: test.Error, Message: "Function not found: " + sequenceEvent.Label, Site: sequenceEvent.Site}
	}
}

func (t *TestDevice) Print() {
	fmt.Println("Test device at site: " + fmt.Sprintf("%v", t.site))
}
