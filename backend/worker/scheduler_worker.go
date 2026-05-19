package worker

import (
	"context"
	"encoding/json"
	"time"

	"whatsapp_multi_session/bulksender"
	"whatsapp_multi_session/commandhandler"
	"whatsapp_multi_session/config"
	"whatsapp_multi_session/scheduler"
	"whatsapp_multi_session/utils"
	"whatsapp_multi_session/warmup"

	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types"
)

type SchedulerWorker struct {
	schedulerService *scheduler.Service
	warmupService    *warmup.Service
	ticker           *time.Ticker
	stopChan         chan bool
}

func NewSchedulerWorker(schedulerService *scheduler.Service, warmupService *warmup.Service) *SchedulerWorker {
	return &SchedulerWorker{
		schedulerService: schedulerService,
		warmupService:    warmupService,
		stopChan:         make(chan bool),
	}
}

func (w *SchedulerWorker) Start() {
	w.ticker = time.NewTicker(1 * time.Minute)
	
	log.Info("[SchedulerWorker] Started - checking for pending jobs every minute")

	go func() {
		for {
			select {
			case <-w.ticker.C:
				w.processPendingJobs()
			case <-w.stopChan:
				log.Info("[SchedulerWorker] Stopped")
				return
			}
		}
	}()
}

func (w *SchedulerWorker) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
	}
	w.stopChan <- true
}

func (w *SchedulerWorker) processPendingJobs() {
	jobs, err := w.schedulerService.GetPendingJobs()
	if err != nil {
		log.Errorf("[SchedulerWorker] Failed to get pending jobs: %v", err)
		return
	}

	if len(jobs) == 0 {
		return
	}

	log.Infof("[SchedulerWorker] Found %d pending jobs to process", len(jobs))

	for _, job := range jobs {
		w.processJob(job)
	}
}

func (w *SchedulerWorker) processJob(job scheduler.ScheduledJob) {
	log.Infof("[SchedulerWorker] Processing job %d for sender %s", job.ID, job.SenderJID)

	var recipients []string
	if err := json.Unmarshal([]byte(job.Recipients), &recipients); err != nil {
		log.Errorf("[SchedulerWorker] Failed to unmarshal recipients for job %d: %v", job.ID, err)
		w.schedulerService.UpdateJobStatus(job.ID, "failed", job.SentMessages)
		return
	}

	var messageVariants []string
	if job.MessageVariants != "" {
		if err := json.Unmarshal([]byte(job.MessageVariants), &messageVariants); err != nil {
			log.Warnf("[SchedulerWorker] Failed to unmarshal message variants for job %d: %v", job.ID, err)
		}
	}

	senderJID, err := types.ParseJID(job.SenderJID)
	if err != nil {
		log.Errorf("[SchedulerWorker] Invalid sender JID for job %d: %v", job.ID, err)
		w.schedulerService.UpdateJobStatus(job.ID, "failed", job.SentMessages)
		return
	}

	client, ok := commandhandler.LoadClientConcurrent(senderJID.User)
	if !ok || client == nil {
		log.Errorf("[SchedulerWorker] Client not found for sender %s", job.SenderJID)
		w.schedulerService.UpdateJobStatus(job.ID, "failed", job.SentMessages)
		return
	}

	if !client.IsLoggedIn() {
		log.Errorf("[SchedulerWorker] Client not logged in for sender %s", job.SenderJID)
		w.schedulerService.UpdateJobStatus(job.ID, "failed", job.SentMessages)
		return
	}

	dailyLimit := config.Conf.BulkSend.DailyLimit
	if w.warmupService != nil {
		warmupLimit, err := w.warmupService.GetCurrentDailyLimit(job.SenderJID)
		if err != nil {
			log.Warnf("[SchedulerWorker] Failed to get warmup limit for job %d: %v", job.ID, err)
		} else if warmupLimit > 0 {
			dailyLimit = warmupLimit
			log.Infof("[SchedulerWorker] Using warmup daily limit: %d for job %d", dailyLimit, job.ID)
		}
	}

	dailyCount := bulksender.GetDailyCount(senderJID.User)
	remainingToday := dailyLimit - dailyCount
	
	if remainingToday <= 0 {
		log.Infof("[SchedulerWorker] Daily limit reached for sender %s, will retry tomorrow", job.SenderJID)
		return
	}

	recipientsToSend := recipients[job.SentMessages:]
	if len(recipientsToSend) > remainingToday {
		recipientsToSend = recipientsToSend[:remainingToday]
	}

	log.Infof("[SchedulerWorker] Sending %d messages for job %d (daily limit: %d, sent today: %d)", 
		len(recipientsToSend), job.ID, dailyLimit, dailyCount)

	variables := make(map[string]string)
	
	for i, recipient := range recipientsToSend {
		message := ""
		if len(messageVariants) > 0 {
			variantIndex := (job.SentMessages + i) % len(messageVariants)
			message = messageVariants[variantIndex]
			log.Debugf("[SchedulerWorker] Using message variant %d for recipient %s", variantIndex, recipient)
		}

		results := bulksender.SendBulkSequential(
			context.Background(),
			client,
			senderJID,
			[]string{recipient},
			message,
			variables,
			utils.ParseJID,
			w.warmupService,
		)

		if len(results) > 0 && results[0].Success {
			job.SentMessages++
		}
	}

	if job.SentMessages >= job.TotalMessages {
		w.schedulerService.UpdateJobStatus(job.ID, "completed", job.SentMessages)
		log.Infof("[SchedulerWorker] Job %d completed - sent %d/%d messages", job.ID, job.SentMessages, job.TotalMessages)
	} else {
		w.schedulerService.UpdateJobStatus(job.ID, "pending", job.SentMessages)
		log.Infof("[SchedulerWorker] Job %d in progress - sent %d/%d messages", job.ID, job.SentMessages, job.TotalMessages)
	}
}
