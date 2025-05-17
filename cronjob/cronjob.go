package cronjob

import (
	"fmt"
	"net/http"

	"whatsapp_multi_session/commandhandler"
	"whatsapp_multi_session/config"
	"whatsapp_multi_session/cronjob/crontab"

	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

type CronJobs struct {
	CommandHandler commandhandler.CommandHandler
}

func NewCronJobs(commandhandler commandhandler.CommandHandler) *CronJobs {
	return &CronJobs{
		CommandHandler: commandhandler,
	}
}

func (c CronJobs) Run() {
	//initiate crontab
	crontabInit := crontab.New()

	//add jobs here based on the configuration
	if config.Conf.Cronjob.AutoPresence.Enable {
		// Add job and print the errors
		schedule := config.Conf.Cronjob.AutoPresence.CronJobSchedule
		if schedule == "" {
			schedule = "* * * * *"
		}
		err := crontabInit.AddJob(schedule, func() {
			err := c.AutoPresence()
			if err != nil {
				log.Errorf("err on job AutoPresence : %v ", err)
				return
			}
		})
		if err != nil {
			log.Errorf("err on job AutoPresence : %v ", err)
			crontabInit.Shutdown()
		}
	}
}

func (c CronJobs) AutoPresence() error {
	var presenceErr error

	// Use sync.Map to iterate over clients
	commandhandler.ClientConcurrent.Range(func(key, value interface{}) bool {
		clientID := key.(string)
		client, ok := value.(*whatsmeow.Client)
		if !ok {
			presenceErr = fmt.Errorf("type assertion failed for client ID %s", clientID)
			return true // Continue iterating even if there's an error
		}

		if client.Store.ID.User != "" {
			if config.Conf.Proxy.Enable {
				proxy := c.CommandHandler.EnabledProxy(client.Store.ID.User)
				if proxy != nil {
					client.SetProxy(http.ProxyURL(proxy))
				}
			}

			// Send presence
			if client.Store.ID != nil {
				err := client.SendPresence(types.PresenceAvailable)
				if err != nil {
					presenceErr = fmt.Errorf("err client.SendPresence for user %s: %v", client.Store.ID.User, err)
					return true // Continue iterating even if there's an error
				}
				log.Printf("Send presence is successful for user %v", client.Store.ID.User)
			}
		}
		return true // Continue iterating over the map
	})

	return presenceErr
}
