package userinterface

import (
	"checkerbox/internal/event"
)

type GraphicInterface interface {
	GraphicEventHandler()
	GetEventChannel() chan event.Event
}
