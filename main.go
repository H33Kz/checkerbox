package main

import (
	"checkerbox/internal/config"
	"fmt"
	"log"
)

func main() {
	loadedConfig, error := config.NewConfig("config/config.yml")
	if error != nil {
		log.Fatal(error.Error())
	}
	hardware := loadedConfig.GetHardwareConfig()
	devices, errors := config.HardwareConfigResolver(hardware)
	fmt.Println(hardware)
	if len(errors) > 0 {
		for _, value := range errors {
			fmt.Println(value.Error())
		}
	}
	for _, device := range devices {
		device.Print()
	}
}
