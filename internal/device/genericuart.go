package device

import (
	"checkerbox/internal/event"
	"checkerbox/internal/test"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.bug.st/serial"
)

type GenericUart struct {
	eventChannel chan event.Event
	site         int
	port         serial.Port
}

func NewGenericUart(deviceMap map[string]string) (*GenericUart, []error) {
	var errorTable []error
	site, siteError := strconv.ParseInt(deviceMap["site"], 10, 8)
	if siteError != nil {
		errorTable = append(errorTable, errors.New("Unable to parse site name for: "+deviceMap["device"]+"\nSetting of site 1"))
		site = 1
	}
	addres := deviceMap["address"]
	baudrate, baudError := strconv.ParseInt(deviceMap["baudrate"], 10, 32)
	if baudError != nil {
		errorTable = append(errorTable, errors.New("Unable to parse baudrate for: "+deviceMap["device"]+"\nSetting default of: 115200"))
		baudrate = 115200
	}
	port, portError := initPort(addres, int(baudrate))
	if portError != nil {
		errorTable = append(errorTable, portError)
		return nil, errorTable
	}

	return &GenericUart{
		site:         int(site),
		port:         port,
		eventChannel: make(chan event.Event),
	}, errorTable
}

func (u *GenericUart) GetEventChannel() chan event.Event {
	return u.eventChannel
}

func initPort(addres string, baudrate int) (serial.Port, error) {
	port, error := serial.Open(addres, &serial.Mode{
		BaudRate: baudrate,
	})
	port.SetReadTimeout(1000)
	return port, error
}

func (u *GenericUart) SequenceEventHandler(resultChannel chan test.Result) {
	for receivedEvent := range u.eventChannel {
		sequenceEvent, ok := receivedEvent.Data.(event.SequenceEvent)
		if !ok || sequenceEvent.DeviceName != "genericuart" || sequenceEvent.Site != u.site {
			continue
		}

		resultChannel <- u.functionResolver(sequenceEvent)
	}
}

func (u *GenericUart) functionResolver(sequenceEvent event.SequenceEvent) test.Result {
	switch sequenceEvent.Function {
	case "Read":
		return u.read(sequenceEvent)
	case "Write":
		return u.write(sequenceEvent)
	case "Send-Receive":
		return u.sendReceive(sequenceEvent)
	default:
		return test.Result{Result: test.Error, Message: "Function not found: " + sequenceEvent.Label, Site: sequenceEvent.Site}
	}
}

func (u *GenericUart) sendReceive(sequenceEvent event.SequenceEvent) test.Result {
	writeResult := u.write(sequenceEvent)
	if writeResult.Result == test.Error {
		return writeResult
	} else {
		time.Sleep(time.Millisecond * 20)
		return u.read(sequenceEvent)
	}
}

func (u *GenericUart) read(sequenceEvent event.SequenceEvent) test.Result {
	buff := make([]byte, 128)
	n, err := u.port.Read(buff)
	if err != nil {
		return test.Result{Result: test.Error, Message: err.Error(), Site: sequenceEvent.Site, Id: sequenceEvent.Id, Label: sequenceEvent.Label}
	}
	readBuff := "Rx: " + string(buff[:n])
	if sequenceEvent.Threshold == "" {
		return test.Result{Result: test.Done, Message: readBuff, Site: sequenceEvent.Site, Id: sequenceEvent.Id, Label: sequenceEvent.Label}
	}
	if sequenceEvent.Threshold == string(buff[:n]) {
		return test.Result{Result: test.Pass, Message: readBuff, Site: sequenceEvent.Site, Id: sequenceEvent.Id, Label: sequenceEvent.Label}
	} else {
		return test.Result{Result: test.Fail, Message: readBuff, Site: sequenceEvent.Site, Id: sequenceEvent.Id, Label: sequenceEvent.Label}
	}
}

func (u *GenericUart) write(sequenceEvent event.SequenceEvent) test.Result {
	_, err := u.port.Write([]byte(sequenceEvent.Data))
	if err != nil {
		return test.Result{Result: test.Error, Message: err.Error(), Site: sequenceEvent.Site, Id: sequenceEvent.Id, Label: sequenceEvent.Label}
	} else {
		return test.Result{Result: test.Done, Message: "Tx: " + sequenceEvent.Data, Site: sequenceEvent.Site, Id: sequenceEvent.Id, Label: sequenceEvent.Label}
	}
}

func (u *GenericUart) Print() {
	fmt.Println("GenericUart device created for site: " + fmt.Sprintf("%v", u.site))
}
