package main

import (
	"checkerbox/internal/config"
	"checkerbox/internal/device"
	"checkerbox/internal/event"
	"checkerbox/internal/test"
	"fmt"
	"log"
)

type applicationContext struct {
	config        *config.Config
	devices       []device.Device
	resultChannel chan test.Result
	eventBus      *event.EventBus
	deviceErrors  []error
}

func main() {
	var ctx applicationContext
	reloadConfiguration(&ctx)

	ctx.eventBus.Publish(event.Event{
		Type: "SequenceEvent",
		Data: event.SequenceEvent{
			DeviceName: "genericuart",
			Site:       1,
			Function:   "Send-Receive",
			Data:       "Test",
			Threshold:  "Test",
		},
	})

	result := <-ctx.resultChannel
	fmt.Println(result)
}

func reloadConfiguration(ctx *applicationContext) {
	// Load specified config file
	// TODO - Figure out idle state in which application loads first - before conf selection
	loadedConfig, err := config.NewConfig("config/config.yml")
	if err != nil {
		log.Fatal(err.Error())
	}
	ctx.config = loadedConfig

	// Isolate configuration of devices, spawn instances and print errors that occured durning initialization
	ctx.devices, ctx.deviceErrors = config.HardwareConfigResolver(ctx.config.GetHardwareConfig())
	if len(ctx.deviceErrors) > 0 {
		for _, value := range ctx.deviceErrors {
			fmt.Println(value.Error())
		}
	}
	for _, device := range ctx.devices {
		device.Print()
	}

	// Instantiate variables regarding event structure
	// Create event bus and subsribe device modules to events of type "SequenceEvent"
	ctx.resultChannel = make(chan test.Result)
	ctx.eventBus = event.NewEventBus()
	for _, device := range ctx.devices {
		ctx.eventBus.Subscribe("SequenceEvent", device.GetEventChannel())
	}

	// Start goroutines from device modules that handle events sent
	for _, device := range ctx.devices {
		go device.SequenceEventHandler(ctx.resultChannel)
	}
}
