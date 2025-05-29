package device

import (
	"checkerbox/internal/event"
	"checkerbox/internal/test"
	"fmt"
	"time"

	"go.bug.st/serial"
)

type GenericUart struct {
	eventChannel chan event.Event
	site         int
	port         serial.Port
}

func NewGenericUart(site int, address string, baudrate int) (*GenericUart, error) {
	port, portError := initPort(address, baudrate)
	if portError != nil {
		return nil, portError
	}

	return &GenericUart{
		site:         site,
		port:         port,
		eventChannel: make(chan event.Event, 100),
	}, nil
}

func (u *GenericUart) GetEventChannel() chan event.Event {
	return u.eventChannel
}

func initPort(addres string, baudrate int) (serial.Port, error) {
	port, error := serial.Open(addres, &serial.Mode{
		BaudRate: baudrate,
	})
	if error == nil {
		port.SetReadTimeout(1000)
	}
	return port, error
}

func (u *GenericUart) SequenceEventHandler() {
	for receivedEvent := range u.eventChannel {
		sequenceEvent, ok := receivedEvent.Data.(event.SequenceEvent)
		if !ok || sequenceEvent.DeviceName != "genericuart" || sequenceEvent.Site != u.site {
			continue
		}
		siteResultChannel := receivedEvent.ReturnChannel
		result := u.functionResolver(sequenceEvent)
		result.Site = sequenceEvent.Site
		result.Id = sequenceEvent.Id
		result.Label = sequenceEvent.Label
		siteResultChannel <- result
	}
}

func (u *GenericUart) functionResolver(sequenceEvent event.SequenceEvent) test.Result {
	function, ok := sequenceEvent.StepSettings["function"].(string)
	if !ok {
		return test.Result{Result: test.Error, Message: "Error parsing function name"}
	}

	switch function {
	case "Read":
		return u.read(sequenceEvent.StepSettings["threshold"].(string))
	case "Write":
		return u.write(sequenceEvent.StepSettings["data"].(string))
	case "Send-Receive":
		return u.sendReceive(sequenceEvent.StepSettings["data"].(string), sequenceEvent.StepSettings["threshold"].(string))
	default:
		return test.Result{Result: test.Error, Message: "Function not found: " + sequenceEvent.Label, Site: sequenceEvent.Site}
	}
}

func (u *GenericUart) sendReceive(data, threshold string) test.Result {
	writeResult := u.write(data)
	if writeResult.Result == test.Error {
		return writeResult
	} else {
		time.Sleep(time.Millisecond * 20)
		return u.read(threshold)
	}
}

func (u *GenericUart) read(threshold string) test.Result {
	buff := make([]byte, 128)
	n, err := u.port.Read(buff)
	if err != nil {
		return test.Result{Result: test.Error, Message: err.Error()}
	}
	readBuff := "Rx: " + string(buff[:n])
	if threshold == "" {
		return test.Result{Result: test.Done, Message: readBuff}
	}
	if threshold == string(buff[:n]) {
		return test.Result{Result: test.Pass, Message: readBuff}
	} else {
		return test.Result{Result: test.Fail, Message: readBuff}
	}
}

func (u *GenericUart) write(data string) test.Result {
	_, err := u.port.Write([]byte(data))
	if err != nil {
		return test.Result{Result: test.Error, Message: err.Error()}
	} else {
		return test.Result{Result: test.Done, Message: "Tx: " + data}
	}
}

func (u *GenericUart) Print() {
	fmt.Println("GenericUart device created for site: " + fmt.Sprintf("%v", u.site))
}
