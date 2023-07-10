package event

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"golang.org/x/exp/slog"
)

type PusherEvent struct {
	log  *slog.Logger
	conn *websocket.Conn
}

type Message struct {
	Channel string                 `json:"channel"`
	Event   string                 `json:"event"`
	Data    map[string]interface{} `json:"data"`
}

func NewPusherEvent(log *slog.Logger, conn *websocket.Conn) *PusherEvent {
	return &PusherEvent{
		log:  log,
		conn: conn,
	}
}

func (p *PusherEvent) TriggerEvent(m Message) error {
	const op = "handlers.event.TriggerEvent"

	var (
		err error
		msg []byte
	)

	msg, err = json.Marshal(m)

	if err != nil {
		p.log.Error("failed to marshal message")

		return fmt.Errorf("%s: %w", op, err)
	}

	if err = p.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		p.log.Error("failed to trigger event")

		return fmt.Errorf("%s: %w", op, err)
	}

	p.log.Info("event triggered")

	return nil
}
