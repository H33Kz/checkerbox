package userinterface

import (
	"checkerbox/internal/event"
	"checkerbox/internal/test"
	"fmt"

	"github.com/rivo/tview"
)

type TviewInterface struct {
	eventChannel chan event.Event
	application  tview.Application
	sites        int
	resultLists  map[int][]test.Result
}

func NewTviewInterace(sites int) *TviewInterface {
	return &TviewInterface{
		eventChannel: make(chan event.Event),
		sites:        sites,
		resultLists:  make(map[int][]test.Result),
	}
}

func (t *TviewInterface) GetEventChannel() chan event.Event {
	return t.eventChannel
}

func (t *TviewInterface) GraphicEventHandler() {
	mainBox := tview.NewFlex()
	var siteBoxes []tview.TextView
	for i := range t.sites {
		siteBoxes = append(siteBoxes, *tview.NewTextView())
		siteBoxes[i].SetBorder(true).SetTitle("Site" + fmt.Sprintf("%v", i))
	}
	for i := range t.sites {
		mainBox.AddItem(&siteBoxes[i], 0, 1, false)
	}
	mainBox.SetBorder(true).SetTitle("checkerbox")

	go func() {
		for receivedEvent := range t.eventChannel {
			graphicEvent, ok := receivedEvent.Data.(event.GraphicEvent)
			if !ok {
				continue
			}

			if graphicEvent.Type == "QUIT" {
				fmt.Println("UI quit")
				t.application.Stop()
			}

			t.resultLists[graphicEvent.Result.Site] = append(t.resultLists[graphicEvent.Result.Site], graphicEvent.Result)
			t.refreshSites(siteBoxes)
		}
	}()

	if err := t.application.SetRoot(mainBox, true).Run(); err != nil {
		panic(err)
	}
}

func (t *TviewInterface) refreshSites(siteBoxes []tview.TextView) {
	for idx := range siteBoxes {
		siteBoxes[idx].Clear()
		for _, result := range t.resultLists[idx] {
			fmt.Fprintf(&siteBoxes[idx], "%v: %v \n", result.Label, result.Message)
		}
	}
}
