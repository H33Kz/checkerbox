package config

import (
	"checkerbox/internal/device"
	"errors"
)

func HardwareConfigResolver(hardwareMap []map[string]string) ([]device.Device, []error) {
	var InitializedDevices []device.Device
	var errorTable []error

	for _, value := range hardwareMap {
		device, error := devicePicker(value)
		InitializedDevices = append(InitializedDevices, device)
		errorTable = append(errorTable, error...)
	}

	return InitializedDevices, errorTable
}

func devicePicker(deviceEntry map[string]string) (device.Device, []error) {
	switch deviceEntry["device"] {
	case "genericuart":
		return device.NewGenericUart(deviceEntry)
	default:
		var errorTable []error
		errorTable = append(errorTable, errors.New("Specified device not supported: "+deviceEntry["device"]))
		return nil, errorTable

	}
}
