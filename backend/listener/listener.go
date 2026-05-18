package listener

import (
	"fmt"

	"whatsapp_multi_session/commandhandler"
	"whatsapp_multi_session/config"
	"whatsapp_multi_session/primitive"

	"github.com/gookit/event"
)

type Listener struct {
	CommandHandler commandhandler.CommandHandler
}

func NewListener(commandHandler commandhandler.CommandHandler) Listener {
	return Listener{
		CommandHandler: commandHandler,
	}
}

// TriggerStartUp sends a signal to the repository and performs start up actions.
// this call should be not initiated on event because we can just call it on the main.go
func (l Listener) TriggerStartUp() {
	if config.Conf.StartUp.EnableAutoLogin {
		fmt.Println("trigger TriggerStartUp for EnableAutoLogin is enabled")
		l.CommandHandler.HandleLoginAllDevices()
	}
}

// ListenForShutdownEvent listen on the shutdown event
// look utils/ShutDownEvent constant.
func (l Listener) ListenForShutdownEvent() {
	event.On(primitive.ShutDownEvent, event.ListenerFunc(func(e event.Event) error {
		// TriggerShutdown wrapping action for the shutdown event
		l.TriggerShutDown()
		return nil
	}))
}

// TriggerShutDown sends a signal to the code handler and performs shutdown actions.
// this call should be not initiated on event because we can just call it on the main.go
func (l Listener) TriggerShutDown() {
	if config.Conf.ShutDown.EnableAutoShutDown {
		fmt.Println("trigger TriggerShutDown for EnableAutoShutDown is enabled")
		l.CommandHandler.HandleDisconnectAllDevices()
	}
}
