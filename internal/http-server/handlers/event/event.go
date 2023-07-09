package event

import (
	"github.com/pusher/pusher-http-go/v5"
	"golang.org/x/exp/slog"
)

type PusherEvent struct {
	log    *slog.Logger
	pusher *pusher.Client
}

func NewPusherEvent(log *slog.Logger, pusherClient *pusher.Client) *PusherEvent {
	return &PusherEvent{
		log:    log,
		pusher: pusherClient,
	}
}

func (p *PusherEvent) TriggerEvent(channel string, eventName string, data map[string]interface{}) error {
	if err := p.pusher.Trigger(channel, eventName, data); err != nil {
		p.log.Error("failed to trigger pusher event")
		return err
	}
	return nil
}
