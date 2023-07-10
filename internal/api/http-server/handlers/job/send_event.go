package job

import "go-outpost/internal/api/http-server/handlers/event"

type SendEventJob struct {
	EventMessage event.Message
	Event        *event.PusherEvent
}

func (job *SendEventJob) Execute() {
	message := event.Message{
		Channel: job.EventMessage.Channel,
		Event:   job.EventMessage.Event,
		Data:    job.EventMessage.Data,
	}

	err := job.Event.TriggerEvent(message)
	if err != nil {
		return
	}
}
