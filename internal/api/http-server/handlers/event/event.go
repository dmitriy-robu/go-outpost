package event

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"go-outpost/internal/lib/logger/sl"
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
		p.log.Error("failed to marshal message", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	p.log.Info("triggering event")

	if err = p.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		p.log.Error("failed to trigger event", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	p.log.Info("event triggered")

	return nil
}
