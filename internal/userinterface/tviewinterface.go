package userinterface

import (
	"checkerbox/internal/event"
	"checkerbox/internal/test"
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TviewInterface struct {
	eventChannel    chan event.Event
	returnChannel   chan event.ControlEvent
	sites           int
	sitesFinished   int
	sequenceRunning bool
	noError         bool
}

func NewTviewInterace(sites int, returnChannel chan event.ControlEvent) *TviewInterface {
	return &TviewInterface{
		eventChannel:    make(chan event.Event),
		returnChannel:   returnChannel,
		sites:           sites,
		sitesFinished:   0,
		sequenceRunning: false,
		noError:         false,
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
	sequenceBox := tview.NewFlex()
	testBox := tview.NewFlex()
	resultBox := tview.NewFlex()
	// Site boxes - containers for results of tests
	siteBoxes := make(map[int]*tview.TextView)
	resultBoxes := make(map[int]*tview.TextView)
	for i := range t.sites {
		siteBoxes[i] = tview.NewTextView().SetDynamicColors(true).SetWordWrap(true)
		siteBoxes[i].SetBorder(true).SetTitle("Site" + fmt.Sprintf("%v", i))
		resultBoxes[i] = tview.NewTextView().SetTextAlign(tview.AlignCenter)
		resultBoxes[i].SetBorder(true)
	}
	for i := range siteBoxes {
		testBox.AddItem(siteBoxes[i], 0, 1, false)
		resultBox.AddItem(resultBoxes[i], 0, 1, false)
	}
	sequenceBox.SetDirection(tview.FlexRow)
	sequenceBox.AddItem(testBox, 0, 1, false)
	sequenceBox.AddItem(resultBox, 6, 1, false)
	sequenceBox.SetBorder(true).SetTitle(" Sequence ")

	// Create layout for navigation section at the bottom of the screen
	navBar := tview.NewFlex()
	info := tview.NewTextView().
		SetText("F1 [darkcyan]Sequence [white] F2 [darkcyan]DebugInfo [white] F3 [darkcyan]ConfigPicker [white]").
		SetRegions(true).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	info2 := tview.NewTextView().
		SetText("F10 [darkcyan]noError [white] F12 [darkcyan]SeqStart [white] CTRL+Q [darkcyan]Exit [white]").
		SetRegions(true).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight)
	navBar.
		SetDirection(tview.FlexColumn).
		AddItem(info, 0, 1, false).
		AddItem(info2, 0, 1, false)

	// Create page for debug information
	debugBox := tview.NewFlex()
	debugTextField := tview.NewTextView()
	debugTextField.SetDynamicColors(true)
	debugBox.AddItem(debugTextField, 0, 1, false)
	debugBox.SetBorder(true).SetTitle(" Debug Info ")

	// Create page for choosing config file
	configBox := tview.NewFlex()
	configList := tview.NewList()
	configFiles, err := os.ReadDir("./config/")
	if err != nil {
		fmt.Fprintf(debugTextField, "%s \n", err.Error())
	}
	for i, file := range configFiles {
		configList.AddItem(file.Name(), "", rune(i+1), func() {
			t.returnChannel <- event.ControlEvent{
				Type: "CONFIGPICK",
				Data: file.Name(),
			}
			app.Stop()
		})
	}
	modalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("Choose config file").SetTextAlign(tview.AlignCenter), 2, 1, false).
		AddItem(configList, 0, 2, true)
	modalFlex.SetBorder(true)
	configBoxWidthLayout := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(modalFlex, 40, 0, true).
		AddItem(nil, 0, 1, false)
	configboxHeightLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(configBoxWidthLayout, 10, 0, true).
		AddItem(nil, 0, 1, false)
	configBox.AddItem(configboxHeightLayout, 0, 1, true)
	configBox.SetBorder(true).SetTitle(" Config Picker ")

	// Place created pages into main container and set keyboard shortcuts
	pages.AddPage("Sequence", sequenceBox, true, true)
	pages.AddPage("DebugInfo", debugBox, true, false)
	pages.AddPage("ConfigPicker", configBox, true, false)
	masterLayout.SetInputCapture(func(tcellEvent *tcell.EventKey) *tcell.EventKey {
		if tcellEvent.Key() == tcell.KeyF1 {
			pages.SwitchToPage("Sequence")
		} else if tcellEvent.Key() == tcell.KeyF2 {
			pages.SwitchToPage("DebugInfo")
		} else if tcellEvent.Key() == tcell.KeyF10 {
			t.noError = !t.noError
			if t.noError {
				info2.SetText("F10 [red]noError [white] F12 [darkcyan]SeqStart [white] CTRL+Q [darkcyan]Exit [white]")
			} else {
				info2.SetText("F10 [darkcyan]noError [white] F12 [darkcyan]SeqStart [white] CTRL+Q [darkcyan]Exit [white]")
			}
			t.returnChannel <- event.ControlEvent{
				Type: "NOERROR",
			}
		} else if tcellEvent.Key() == tcell.KeyF12 {
			if !t.sequenceRunning {
				t.sequenceRunning = true
				for k := range resultLists {
					delete(resultLists, k)
				}
				for _, siteBox := range siteBoxes {
					siteBox.Clear()
					siteBox.SetTextColor(tcell.ColorWhite)
				}
				for _, resultBox := range resultBoxes {
					resultBox.Clear()
					resultBox.SetBackgroundColor(tcell.ColorDarkBlue)
					fmt.Fprintf(resultBox, "Test in progress")
				}
				t.returnChannel <- event.ControlEvent{
					Type: "START",
				}
			}
		} else if tcellEvent.Key() == tcell.KeyF3 {
			pages.SwitchToPage("ConfigPicker")
		} else if tcellEvent.Key() == tcell.KeyCtrlQ {
			app.Stop()
			t.returnChannel <- event.ControlEvent{
				Type: "QUIT",
			}
		}
		return tcellEvent
	})

	// Set Master layout that will be root of UI
	masterLayout.
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true).
		AddItem(navBar, 1, 1, false)

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
					if graphicEvent.Result.Result == test.Pass {
						fmt.Fprintf(siteBoxes[graphicEvent.Result.Site], "[green]%s [white]%s\n", graphicEvent.Result.Result, graphicEvent.Result.Label)
					} else {
						fmt.Fprintf(siteBoxes[graphicEvent.Result.Site], "[red]%s [white]%s\n", graphicEvent.Result.Result, graphicEvent.Result.Label)
					}
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
					siteBoxes[graphicEvent.Result.Site].ScrollToEnd()
				})
				// Event for end of the test sequence. Changes flags to unlock starting test control
			case "sequenceEnd":
				t.sitesFinished++
				if t.sitesFinished == t.sites {
					t.sequenceRunning = false
					t.sitesFinished = 0
				}
				app.QueueUpdateDraw(func() {
					resultBoxes[graphicEvent.Result.Site].Clear()
					if graphicEvent.Result.Result == test.Pass {
						// siteBoxes[graphicEvent.Result.Site].SetBackgroundColor(tcell.ColorDarkGreen)
						resultBoxes[graphicEvent.Result.Site].SetBackgroundColor(tcell.ColorDarkGreen)
						fmt.Fprintf(resultBoxes[graphicEvent.Result.Site], "%s", graphicEvent.Result.Result)
					} else {
						// siteBoxes[graphicEvent.Result.Site].SetBackgroundColor(tcell.ColorDarkRed)
						resultBoxes[graphicEvent.Result.Site].SetBackgroundColor(tcell.ColorDarkRed)
						fmt.Fprintf(resultBoxes[graphicEvent.Result.Site], "%s", graphicEvent.Result.Result)
					}
				})
			// Event adding debug information to debug page
			case "debugInfo":
				app.QueueUpdateDraw(func() {
					// result := graphicEvent.Result
					// fmt.Fprintf(debugTextField, "[gray]%s: [white]Site:%d %d %s %s: %s \n", time.Now().Format("15:4:5"), result.Site, result.Id, result.Result, result.Label, result.Message)
					fmt.Fprintf(debugTextField, "[gray]%s [white]%s", time.Now().Format("15:4:5"), graphicEvent.Log.Print())
					debugTextField.ScrollToEnd()
				})
			default:
				continue
			}
		}
	}()

	// Start tview application
	if err := app.SetRoot(masterLayout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
