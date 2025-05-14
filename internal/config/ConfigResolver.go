package config

import (
	"checkerbox/internal/device"
	"checkerbox/internal/userinterface"
	"errors"
)

func DeviceEntryResolver(deviceEntry DeviceSettings) (device.Device, []error) {
	var errorTable []error
	switch deviceEntry.DeviceName {
	case "genericuart":
		baudrate, ok := deviceEntry.Settings["baudrate"].(int)
		if !ok {
			errorTable = append(errorTable, errors.New("Unable to parse baudrate for: "+deviceEntry.DeviceName+"\nSetting default of: 115200"))
			baudrate = 115200
		}

		address, ok := deviceEntry.Settings["address"].(string)
		if !ok {
			errorTable = append(errorTable, errors.New("Unable to parse address for: "+deviceEntry.DeviceName))
			return nil, errorTable
		}
		genericUartDevice, err := device.NewGenericUart(deviceEntry.Site, address, baudrate)
		if err != nil {
			errorTable = append(errorTable, err)
			return nil, errorTable
		}
		return genericUartDevice, errorTable
	case "testdevice":
		testDevice, err := device.NewTestDevice(deviceEntry.Site)
		if err != nil {
			errorTable = append(errorTable, err)
			return nil, errorTable
		}
		return testDevice, errorTable
	default:
		errorTable = append(errorTable, errors.New("Specified device not supported: "+deviceEntry.DeviceName))
		return nil, errorTable

	}
}

func GraphicalInterfaceResolver(settingsNode AppSettings) userinterface.GraphicInterface {
	switch settingsNode.Uiengine {
	case "tview":
		return userinterface.NewTviewInterace(settingsNode.Sites)
	default:
		return nil
	}
}
