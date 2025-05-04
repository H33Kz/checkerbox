package config

import (
	"checkerbox/internal/device"
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
