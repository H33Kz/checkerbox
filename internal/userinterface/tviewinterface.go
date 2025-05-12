package userinterface

import (
	"checkerbox/internal/event"
	"checkerbox/internal/test"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TviewInterface struct {
	eventChannel chan event.Event
	sites        int
}

func NewTviewInterace(sites int) *TviewInterface {
	return &TviewInterface{
		eventChannel: make(chan event.Event),
		sites:        sites,
	}
}

func (t *TviewInterface) GetEventChannel() chan event.Event {
	return t.eventChannel
}

func (t *TviewInterface) GraphicEventHandler() {
	resultLists := make(map[int][]test.Result)

	app := tview.NewApplication()
	mainBox := tview.NewFlex()
	siteBoxes := make(map[int]*tview.TextView)
	for i := range t.sites {
		siteBoxes[i] = tview.NewTextView().SetDynamicColors(true).SetWordWrap(true)
		siteBoxes[i].SetBorder(true).SetTitle("Site" + fmt.Sprintf("%v", i))
	}
	for i := range siteBoxes {
		mainBox.AddItem(siteBoxes[i], 0, 1, false)
	}
	mainBox.SetBorder(true).SetTitle("checkerbox")

	go func() {
		for receivedEvent := range t.eventChannel {
			graphicEvent, ok := receivedEvent.Data.(event.GraphicEvent)
			if !ok {
				continue
			}

			switch graphicEvent.Type {
			case "QUIT":
				app.Stop()
			case "testStarted":
				app.QueueUpdateDraw(func() {
					resultLists[graphicEvent.Result.Site] = append(resultLists[graphicEvent.Result.Site], graphicEvent.Result)
					siteBoxes[graphicEvent.Result.Site].Clear()
					for _, result := range resultLists[graphicEvent.Result.Site] {
						fmt.Fprintf(siteBoxes[graphicEvent.Result.Site], "%v %s %v: %v \n", result.Id, result.Result, result.Label, result.Message)
					}
				})
			case "testResult":
				app.QueueUpdateDraw(func() {
					if graphicEvent.Result.Result == test.Fail {
						siteBoxes[graphicEvent.Result.Site].SetTextColor(tcell.ColorRed)
					}
					resultLists[graphicEvent.Result.Site][len(resultLists[graphicEvent.Result.Site])-1] = graphicEvent.Result
					siteBoxes[graphicEvent.Result.Site].Clear()
					for _, result := range resultLists[graphicEvent.Result.Site] {
						fmt.Fprintf(siteBoxes[graphicEvent.Result.Site], "%v %s %v: %v \n", result.Id, result.Result, result.Label, result.Message)
					}
				})
			default:
				continue
			}
		}
	}()

	if err := app.SetRoot(mainBox, true).Run(); err != nil {
		panic(err)
	}
}
