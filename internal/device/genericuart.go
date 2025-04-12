package device

import (
	"errors"
	"fmt"
	"strconv"

	"go.bug.st/serial"
)

type GenericUart struct {
	site int

	port serial.Port
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
	}

	return &GenericUart{
		site: int(site),
		port: port,
	}, errorTable
}

func initPort(addres string, baudrate int) (serial.Port, error) {
	port, error := serial.Open(addres, &serial.Mode{
		BaudRate: baudrate,
	})
	return port, error
}

func (u *GenericUart) FunctionResolver() {
}

func (u *GenericUart) Print() {
	fmt.Println("GenericUart device created for site: " + fmt.Sprintf("%v", u.site))
}
