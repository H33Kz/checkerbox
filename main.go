package main

import (
	"checkerbox/internal/config"
	"checkerbox/internal/device"
	"checkerbox/internal/event"
	"checkerbox/internal/test"
	"checkerbox/internal/userinterface"
	"checkerbox/internal/util"
	"fmt"
	"log"
	"sync"
	"time"
)

type applicationContext struct {
	ctxMutex           sync.Mutex
	noError            bool
	appSettings        *config.AppSettings
	config             *config.Config
	devices            []device.Device
	sequenceEventLists map[int]*util.Queue[event.Event]
	eventBus           *event.EventBus
	deviceErrors       []error
	graphicInterface   userinterface.GraphicInterface
	uiReturnChannel    chan event.ControlEvent
}

func main() {
	var ctx applicationContext
	loadAppSettings(&ctx)

	if ctx.graphicInterface != nil {
	out:
		for receivedEvent := range ctx.uiReturnChannel {
			switch receivedEvent.Type {
			case "START":
				for i, sequenceEventList := range ctx.sequenceEventLists {
					go handleSequence(*sequenceEventList, &ctx, i)
				}
			case "QUIT":
				break out
			case "CONFIGPICK":
				newCtx := applicationContext{}
				newCtx.graphicInterface = ctx.graphicInterface
				newCtx.uiReturnChannel = ctx.uiReturnChannel
				ctx = newCtx
				loadAppSettings(&ctx)
				reloadConfiguration(&ctx, "./config/"+receivedEvent.Data.(string))
			}
		}
	} else {
		reloadConfiguration(&ctx, "./config/config.yml")
		for i, sequenceEventList := range ctx.sequenceEventLists {
			go handleSequence(*sequenceEventList, &ctx, i)
		}
		time.Sleep(time.Second * 10)
	}
}

func handleSequence(sequenceEventsList util.Queue[event.Event], ctx *applicationContext, siteId int) {
	siteResultChannel := make(chan test.Result)
	sequenceFailed := false
	for range sequenceEventsList.Len() {
		singleSequenceEvent := sequenceEventsList.Dequeue()
		singleSequenceEvent.ReturnChannel = siteResultChannel
		var result test.Result
		for retried := range singleSequenceEvent.Data.(event.SequenceEvent).Retry {

			ctx.ctxMutex.Lock()
			ctx.eventBus.Publish(singleSequenceEvent)
			sequenceEventForUI := singleSequenceEvent.Data.(event.SequenceEvent)
			SendTestStartedEvent(ctx, sequenceEventForUI.Id, sequenceEventForUI.Site, sequenceEventForUI.Label)
			SendDebugInfoEvent(ctx, test.InProgress, sequenceEventForUI.Site, sequenceEventForUI.Label, "Test started")
			ctx.ctxMutex.Unlock()

			select {
			case result = <-siteResultChannel:
				result.Retried = retried
			case <-time.After(time.Millisecond * time.Duration(sequenceEventForUI.Timeout)):
				result = test.Result{
					Result:  test.Error,
					Site:    sequenceEventForUI.Site,
					Id:      sequenceEventForUI.Id,
					Label:   sequenceEventForUI.Label,
					Message: "Timeout",
				}
			}
			ctx.ctxMutex.Lock()
			SendTestResultEvent(ctx, result)
			SendDebugInfoEvent(ctx, result.Result, result.Site, result.Label, "Test finished with result: "+result.Message+" On retry: "+fmt.Sprintf("%v", result.Retried))
			ctx.ctxMutex.Unlock()
			if ctx.graphicInterface == nil {
				fmt.Println(result)
			}

			if result.Result == test.Pass || result.Result == test.Error || result.Result == test.Done {
				break
			}
		}
		if (result.Result == test.Fail || result.Result == test.Error) && !ctx.noError {
			sequenceFailed = true
			sequenceEventsList.Flush()
			SendSequenceEndEvent(ctx, test.Fail, siteId)
			break
		} else if (result.Result == test.Fail || result.Result == test.Error) && ctx.noError {
			sequenceFailed = true
		}
	}
	if !sequenceFailed {
		SendSequenceEndEvent(ctx, test.Pass, siteId)
	}
}

func loadAppSettings(ctx *applicationContext) {
	// Load basic app settings on startup
	ctx.appSettings = config.NewAppSettings()
	ctx.sequenceEventLists = make(map[int]*util.Queue[event.Event])
	ctx.eventBus = event.NewEventBus()
	if ctx.graphicInterface == nil {
		ctx.uiReturnChannel = make(chan event.ControlEvent)
		ctx.graphicInterface = config.GraphicalInterfaceResolver(*ctx.appSettings, ctx.uiReturnChannel)
	}
	ctx.noError = false

	if ctx.graphicInterface != nil {
		ctx.eventBus.Subscribe("graphicEvent", ctx.graphicInterface.GetEventChannel())
		go ctx.graphicInterface.GraphicEventHandler()
	}
}

func reloadConfiguration(ctx *applicationContext, path string) {
	// Load specified config file
	// loadedConfig, err := config.NewConfig("config/config.yml")
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }
	loadedConfig, err := config.NewConfig(path)
	if err != nil {
		log.Fatal(err.Error())
	}
	ctx.config = loadedConfig

	// Load sequence events into lists marked with site number
	for i := 0; i <= ctx.appSettings.Sites-1; i++ {
		ctx.sequenceEventLists[i] = util.NewQueue[event.Event]()
		for n, sequenceConfigNode := range ctx.config.GetSequenceConfig() {
			ctx.sequenceEventLists[i].Enqueue(event.Event{
				Type: "SequenceEvent",
				Data: event.SequenceEvent{
					Id:           uint(n),
					Label:        sequenceConfigNode.StepLabel,
					Site:         i,
					Retry:        sequenceConfigNode.Retry,
					DeviceName:   sequenceConfigNode.Device,
					StepSettings: sequenceConfigNode.StepSettings,
					Timeout:      sequenceConfigNode.Timeout,
				},
			})
		}
	}

	// I dont know why but first event sent is never received by UI routine and changing timing doesn't help
	// Doing it twice dodges the issue
	SendDebugInfoEvent(ctx, test.Pass, 0, "", "Configuration Loading started\n")
	SendDebugInfoEvent(ctx, test.Pass, 0, "", "Configuration Loading started\n")

	for i := 0; i <= ctx.appSettings.Sites-1; i++ {
		ctx.devices = append(ctx.devices, device.NewSequenceDevice(i))
	}
	// Init individual device based on config
	// TODO - after UI design - send UI events based on succesful or unsuccesful initialization instead of printing
	// TODO - add check if device initialized are out of site number spec
	for _, deviceDeclaration := range ctx.config.GetHardwareConfig() {
		initializedDevice, initDeviceErrorTable := config.DeviceEntryResolver(deviceDeclaration)

		deviceInitErrorString := ""
		for _, err = range initDeviceErrorTable {
			deviceInitErrorString += err.Error() + "\n"
		}

		if initializedDevice != nil {
			ctx.devices = append(ctx.devices, initializedDevice)
			SendDeviceInitEvent(ctx, test.Pass, deviceDeclaration.Site, deviceDeclaration.DeviceName)
			SendDebugInfoEvent(ctx, test.Pass, deviceDeclaration.Site, deviceDeclaration.DeviceName, "Device initiated\n"+deviceInitErrorString)
		} else {
			SendDeviceInitEvent(ctx, test.Error, deviceDeclaration.Site, deviceDeclaration.DeviceName)
			SendDebugInfoEvent(ctx, test.Error, deviceDeclaration.Site, deviceDeclaration.DeviceName, "Error while initializing device:\n"+deviceInitErrorString)
		}
	}

	// Instantiate variables regarding event structure
	// Subsribe device modules to events of type "SequenceEvent"
	for _, device := range ctx.devices {
		ctx.eventBus.Subscribe("SequenceEvent", device.GetEventChannel())
	}

	// Start goroutines from device modules that handle events sent
	for _, device := range ctx.devices {
		go device.SequenceEventHandler()
	}
}

func SendDebugInfoEvent(ctx *applicationContext, result test.ResultType, site int, label, message string) {
	ctx.eventBus.Publish(event.Event{
		Type: "graphicEvent",
		Data: event.GraphicEvent{
			Type: "debugInfo",
			Result: test.Result{
				Result:  result,
				Label:   label,
				Site:    site,
				Message: message,
			},
		},
	})
}

func SendDeviceInitEvent(ctx *applicationContext, result test.ResultType, site int, label string) {
	ctx.eventBus.Publish(event.Event{
		Type: "graphicEvent",
		Data: event.GraphicEvent{
			Type: "deviceInit",
			Result: test.Result{
				Result: result,
				Label:  label,
				Site:   site,
			},
		},
	})
}

func SendSequenceEndEvent(ctx *applicationContext, result test.ResultType, site int) {
	ctx.eventBus.Publish(event.Event{
		Type: "graphicEvent",
		Data: event.GraphicEvent{
			Type: "sequenceEnd",
			Result: test.Result{
				Result: result,
				Site:   site,
			},
		},
	})
}

func SendTestResultEvent(ctx *applicationContext, result test.Result) {
	ctx.eventBus.Publish(event.Event{
		Type: "graphicEvent",
		Data: event.GraphicEvent{
			Type:   "testResult",
			Result: result,
		},
	})
}

func SendTestStartedEvent(ctx *applicationContext, id uint, site int, label string) {
	ctx.eventBus.Publish(event.Event{
		Type: "graphicEvent",
		Data: event.GraphicEvent{
			Type: "testStarted",
			Result: test.Result{
				Site:    site,
				Id:      id,
				Label:   label,
				Message: "...",
				Result:  test.InProgress,
			},
		},
	})
}
