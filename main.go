package main

import (
	"checkerbox/internal/config"
	"checkerbox/internal/data"
	"checkerbox/internal/device"
	"checkerbox/internal/event"
	"checkerbox/internal/test"
	"checkerbox/internal/userinterface"
	"checkerbox/internal/util"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type applicationContext struct {
	ctxMutex           sync.Mutex
	configSource       string
	noError            bool
	appSettings        *config.AppSettings
	config             *config.Config
	devices            []device.Device
	sequenceEventLists map[int]*util.Queue[event.Event]
	eventBus           *event.EventBus
	deviceErrors       []error
	graphicInterface   userinterface.GraphicInterface
	uiReturnChannel    chan event.ControlEvent
	reportDatabase     *gorm.DB
	logDatabase        *gorm.DB
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
				ctx.configSource = receivedEvent.Data.(string)
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

	report := data.NewReport()
	ctx.ctxMutex.Lock()
	report.SetSource(ctx.configSource)
	ctx.ctxMutex.Unlock()
	report.SetSite(siteId)
	report.AppendReportString("Sequence Started \n")

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
			ctx.logDatabase.Create(data.NewResultLog(sequenceEventForUI.DeviceName, result))
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
		report.AppendReportString(fmt.Sprintf("%v %s %v: %v (%v) \n", result.Id, result.Result, result.Label, result.Message, result.Retried+1))
		if (result.Result == test.Fail || result.Result == test.Error) && !ctx.noError {
			sequenceFailed = true
			sequenceEventsList.Flush()
			SendSequenceEndEvent(ctx, test.Fail, siteId)
			report.SetOverallResult(test.Fail)
			break
		} else if (result.Result == test.Fail || result.Result == test.Error) && ctx.noError {
			sequenceFailed = true
			report.SetOverallResult(test.Fail)
		}
	}
	if !sequenceFailed {
		SendSequenceEndEvent(ctx, test.Pass, siteId)
		report.SetOverallResult(test.Pass)
	}
	ctx.ctxMutex.Lock()
	SendDBData(ctx, report)
	ctx.ctxMutex.Unlock()
}

func loadAppSettings(ctx *applicationContext) {
	// Load basic app settings on startup
	var err error
	ctx.appSettings = config.NewAppSettings()
	ctx.sequenceEventLists = make(map[int]*util.Queue[event.Event])
	ctx.eventBus = event.NewEventBus()
	ctx.reportDatabase, err = gorm.Open(sqlite.Open("reports.db"), &gorm.Config{})
	if err != nil {
		ctx.reportDatabase = nil
	} else {
		ctx.reportDatabase.AutoMigrate(&data.Report{})
	}
	ctx.logDatabase, err = gorm.Open(sqlite.Open("log.db"), &gorm.Config{})
	if err != nil {
		ctx.logDatabase = nil
	} else {
		ctx.logDatabase.AutoMigrate(&data.Log{})
	}
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
	ctx.logDatabase.Create(data.NewCustomLog("mainloop", "Configuration loading started", 99, data.INFO))

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
			ctx.logDatabase.Create(data.NewCustomLog(deviceDeclaration.DeviceName, "Device initiated", deviceDeclaration.Site, data.INFO))
		} else {
			SendDeviceInitEvent(ctx, test.Error, deviceDeclaration.Site, deviceDeclaration.DeviceName)
			SendDebugInfoEvent(ctx, test.Error, deviceDeclaration.Site, deviceDeclaration.DeviceName, "Error while initializing device:\n"+deviceInitErrorString)
			ctx.logDatabase.Create(data.NewCustomLog(deviceDeclaration.DeviceName, "Error while initializing device:"+deviceInitErrorString, deviceDeclaration.Site, data.ERROR))
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

func SendDBData(ctx *applicationContext, value any) {
	if ctx.reportDatabase != nil {
		ctx.reportDatabase.Create(value)
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
