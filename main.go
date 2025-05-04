package main

import (
	"checkerbox/internal/config"
	"checkerbox/internal/device"
	"checkerbox/internal/event"
	"checkerbox/internal/test"
	"checkerbox/internal/util"
	"fmt"
	"log"
	"strconv"
	"time"
)

type applicationContext struct {
	noError            bool
	appSettings        *config.AppSettings
	config             *config.Config
	devices            []device.Device
	sequenceEventLists map[int]*util.Queue[event.Event]
	resultChannel      chan test.Result
	eventBus           *event.EventBus
	deviceErrors       []error
}

func main() {
	var ctx applicationContext
	ctx.noError = false

	loadAppSettings(&ctx)

	reloadConfiguration(&ctx)
	// ctx.eventBus.Publish(event.Event{
	// 	Type: "SequenceEvent",
	// 	Data: event.SequenceEvent{
	// 		DeviceName: "genericuart",
	// 		Site:       1,
	// 		Function:   "Send-Receive",
	// 		Data:       "Test",
	// 		Threshold:  "Test",
	// 	},
	// })
	//
	// result := <-ctx.resultChannel
	// fmt.Println(result)
	for _, sequenceEventList := range ctx.sequenceEventLists {
		fmt.Println(sequenceEventList)
		go handleSequence(sequenceEventList, &ctx)
	}

	time.Sleep(time.Second * 3)
	for _, sequenceEventList := range ctx.sequenceEventLists {
		fmt.Println(sequenceEventList)
	}
}

func handleSequence(sequenceEventsList *util.Queue[event.Event], ctx *applicationContext) {
	for range sequenceEventsList.Len() {
		singleSequenceEvent := sequenceEventsList.Dequeue()
		var result test.Result
		for range singleSequenceEvent.Data.(event.SequenceEvent).Retry {
			ctx.eventBus.Publish(singleSequenceEvent)
			result = <-ctx.resultChannel
			fmt.Println(result)
			if result.Result == test.Pass || result.Result == test.Error {
				break
			}
		}
		if (result.Result == test.Fail || result.Result == test.Error) && !ctx.noError {
			sequenceEventsList.Flush()
			break
		}
	}
}

func loadAppSettings(ctx *applicationContext) {
	// Load basic app settings on startup
	ctx.appSettings = config.NewAppSettings()
	ctx.sequenceEventLists = make(map[int]*util.Queue[event.Event])
	ctx.eventBus = event.NewEventBus()
}

func reloadConfiguration(ctx *applicationContext) {
	// Load specified config file
	loadedConfig, err := config.NewConfig("config/config.yml")
	if err != nil {
		log.Fatal(err.Error())
	}
	ctx.config = loadedConfig

	// Load sequence events into lists marked with site number
	for i := 0; i <= ctx.appSettings.Sites-1; i++ {
		ctx.sequenceEventLists[i] = util.NewQueue[event.Event]()
		for n, sequenceConfigNode := range ctx.config.GetSequenceConfig() {
			retries, err := strconv.ParseInt(sequenceConfigNode["retry"], 10, 32)
			if err != nil || retries == 0 {
				retries = 1
			}
			ctx.sequenceEventLists[i].Enqueue(event.Event{
				Type: "SequenceEvent",
				Data: event.SequenceEvent{
					Id:         uint(n),
					Label:      sequenceConfigNode["step_label"],
					Site:       i,
					Retry:      int(retries),
					DeviceName: sequenceConfigNode["device"],
					Function:   sequenceConfigNode["function"],
					Data:       sequenceConfigNode["data"],
					Threshold:  sequenceConfigNode["threshold"],
					Timeout:    sequenceConfigNode["timeout"],
				},
			})
		}
	}

	for i := 0; i <= ctx.appSettings.Sites-1; i++ {
		ctx.devices = append(ctx.devices, device.NewSequenceDevice(i))
	}
	// Init individual device based on config
	// TODO - after UI design - send UI events based on succesful or unsuccesful initialization instead of printing
	for _, deviceDeclaration := range ctx.config.GetHardwareConfig() {
		initializedDevice, initDeviceErrorTable := config.DeviceEntryResolver(deviceDeclaration)

		if initializedDevice != nil {
			ctx.devices = append(ctx.devices, initializedDevice)
		}
		for _, errs := range initDeviceErrorTable {
			fmt.Println(errs.Error())
		}
	}
	for _, device := range ctx.devices {
		device.Print()
	}

	// Instantiate variables regarding event structure
	// Create event bus and subsribe device modules to events of type "SequenceEvent"
	ctx.resultChannel = make(chan test.Result)
	for _, device := range ctx.devices {
		ctx.eventBus.Subscribe("SequenceEvent", device.GetEventChannel())
	}

	// Start goroutines from device modules that handle events sent
	for _, device := range ctx.devices {
		go device.SequenceEventHandler(ctx.resultChannel)
	}
}
