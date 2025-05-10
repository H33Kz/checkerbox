package config

import (
	"checkerbox/internal/device"
	"checkerbox/internal/userinterface"
	"errors"
)

func DeviceEntryResolver(deviceEntry map[string]string) (device.Device, []error) {
	switch deviceEntry["device"] {
	case "genericuart":
		return device.NewGenericUart(deviceEntry)
	case "testdevice":
		return device.NewTestDevice(deviceEntry)
	default:
		var errorTable []error
		errorTable = append(errorTable, errors.New("Specified device not supported: "+deviceEntry["device"]))
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
