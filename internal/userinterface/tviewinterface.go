package userinterface

import (
	"checkerbox/internal/event"
	"checkerbox/internal/test"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TviewInterface struct {
	eventChannel    chan event.Event
	returnChannel   chan event.ControlEvent
	sites           int
	sitesFinished   int
	sequenceRunning bool
}

func NewTviewInterace(sites int, returnChannel chan event.ControlEvent) *TviewInterface {
	return &TviewInterface{
		eventChannel:    make(chan event.Event),
		returnChannel:   returnChannel,
		sites:           sites,
		sitesFinished:   0,
		sequenceRunning: false,
	}
}

func (t *TviewInterface) GetEventChannel() chan event.Event {
	return t.eventChannel
}

func (t *TviewInterface) GraphicEventHandler() {
	// Map used for displaying results sent by main routine
	resultLists := make(map[int][]test.Result)

	// Instatiate tview app struct and pages struct which is main container for all widgets
	app := tview.NewApplication()
	pages := tview.NewPages()
	masterLayout := tview.NewFlex()

	// Create layout for test page - test results and site status
	testBox := tview.NewFlex()
	// Site boxes - containers for results of tests
	siteBoxes := make(map[int]*tview.TextView)
	for i := range t.sites {
		siteBoxes[i] = tview.NewTextView().SetDynamicColors(true).SetWordWrap(true)
		siteBoxes[i].SetBorder(true).SetTitle("Site" + fmt.Sprintf("%v", i))
	}
	for i := range siteBoxes {
		testBox.AddItem(siteBoxes[i], 0, 1, false)
	}
	testBox.SetBorder(true).SetTitle("Sequence")

	info := tview.NewTextView().
		SetRegions(true).
		SetDynamicColors(true).
		SetWrap(false)
	fmt.Fprintf(info, `F1 [darkcyan]Sequence [white][""]`)
	fmt.Fprintf(info, `F2 [darkcyan]DebugInfo [white][""]`)

	debugBox := tview.NewFlex()
	debugTextField := tview.NewTextView()
	debugBox.AddItem(debugTextField, 0, 1, false)
	debugBox.SetBorder(true).SetTitle("Debug Info")

	// Place created pages into main container and set keyboard shortcuts
	pages.AddPage("Sequence", testBox, true, true)
	pages.AddPage("DebugInfo", debugBox, true, false)
	masterLayout.SetInputCapture(func(tcellEvent *tcell.EventKey) *tcell.EventKey {
		if tcellEvent.Key() == tcell.KeyF1 {
			pages.SwitchToPage("Sequence")
		} else if tcellEvent.Key() == tcell.KeyF2 {
			pages.SwitchToPage("DebugInfo")
		} else if tcellEvent.Key() == tcell.KeyEnter {
			if !t.sequenceRunning {
				t.sequenceRunning = true
				for k := range resultLists {
					delete(resultLists, k)
				}
				for _, siteBox := range siteBoxes {
					siteBox.Clear()
					siteBox.SetTextColor(tcell.ColorWhite)
				}
				t.returnChannel <- event.ControlEvent{
					Type: "START",
				}
			}
		} else if tcellEvent.Key() == tcell.KeyEsc {
			app.Stop()
			t.returnChannel <- event.ControlEvent{
				Type: "QUIT",
			}
		}
		return tcellEvent
	})

	masterLayout.
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true).
		AddItem(info, 1, 1, false)

	// Main event loop - started in separate goroutine
	// Sets UI elements based on events received from main goroutine
	go func() {
		// Wait for event in a loop
		for receivedEvent := range t.eventChannel {
			// If event is of type "graphicEvent" proceed
			graphicEvent, ok := receivedEvent.Data.(event.GraphicEvent)
			if !ok {
				continue
			}

			// Switch on type of graphic event and modify related fields
			switch graphicEvent.Type {
			case "deviceInit":
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(siteBoxes[graphicEvent.Result.Site], "%s %s\n", graphicEvent.Result.Result, graphicEvent.Result.Label)
				})
			// Event on start of the test. Sets new line in textview in referenced site unless there is already test referenced with the same ID
			case "testStarted":
				app.QueueUpdateDraw(func() {
					if len(resultLists[graphicEvent.Result.Site]) > 0 {
						if resultLists[graphicEvent.Result.Site][len(resultLists[graphicEvent.Result.Site])-1].Id == graphicEvent.Result.Id {
							resultLists[graphicEvent.Result.Site][len(resultLists[graphicEvent.Result.Site])-1] = graphicEvent.Result
						} else {
							resultLists[graphicEvent.Result.Site] = append(resultLists[graphicEvent.Result.Site], graphicEvent.Result)
							siteBoxes[graphicEvent.Result.Site].Clear()
							for _, result := range resultLists[graphicEvent.Result.Site] {
								fmt.Fprintf(siteBoxes[graphicEvent.Result.Site], "%v %s %v: %v \n", result.Id, result.Result, result.Label, result.Message)
							}
						}
					} else {
						resultLists[graphicEvent.Result.Site] = append(resultLists[graphicEvent.Result.Site], graphicEvent.Result)
						siteBoxes[graphicEvent.Result.Site].Clear()
						for _, result := range resultLists[graphicEvent.Result.Site] {
							fmt.Fprintf(siteBoxes[graphicEvent.Result.Site], "%v %s %v: %v \n", result.Id, result.Result, result.Label, result.Message)
						}
					}
				})
				// Event indicating end of a test. Changes line previously set by testStarted event.
			case "testResult":
				app.QueueUpdateDraw(func() {
					resultLists[graphicEvent.Result.Site][len(resultLists[graphicEvent.Result.Site])-1] = graphicEvent.Result
					siteBoxes[graphicEvent.Result.Site].Clear()
					for _, result := range resultLists[graphicEvent.Result.Site] {
						if result.Retried > 0 {
							fmt.Fprintf(siteBoxes[graphicEvent.Result.Site], "%v %s %v: %v (%v) \n", result.Id, result.Result, result.Label, result.Message, result.Retried+1)
						} else {
							fmt.Fprintf(siteBoxes[graphicEvent.Result.Site], "%v %s %v: %v \n", result.Id, result.Result, result.Label, result.Message)
						}
					}
				})
				// Event for end of the test sequence. Changes color of text based of and outcome
			case "sequenceEnd":
				t.sitesFinished++
				if t.sitesFinished == t.sites {
					t.sequenceRunning = false
					t.sitesFinished = 0
				}
				app.QueueUpdateDraw(func() {
					if graphicEvent.Result.Result == test.Pass {
						siteBoxes[graphicEvent.Result.Site].SetTextColor(tcell.ColorGreen)
					} else {
						siteBoxes[graphicEvent.Result.Site].SetTextColor(tcell.ColorRed)
					}
				})
			case "debugInfo":
				app.QueueUpdateDraw(func() {
					result := graphicEvent.Result
					fmt.Fprintf(debugTextField, "Site:%d %d %s %s: %s \n", result.Site, result.Id, result.Result, result.Label, result.Message)
				})
			default:
				continue
			}
		}
	}()

	// Start tview application
	if err := app.SetRoot(masterLayout, true).Run(); err != nil {
		panic(err)
	}
}
