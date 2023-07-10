package handler

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"go-outpost/internal/lib/logger/sl"
	"golang.org/x/exp/slog"
	"net/http"
	"sync"
)

type Message struct {
	Channel string                 `json:"channel"`
	Event   string                 `json:"event"`
	Data    map[string]interface{} `json:"data"`
}

type Subscription struct {
	Conn    *websocket.Conn
	Channel string
}

type Hub struct {
	Channels  map[string]map[*websocket.Conn]bool
	Broadcast chan Message
	Subscribe chan Subscription
	mutex     sync.RWMutex
	log       *slog.Logger
}

func NewHub(
	log *slog.Logger,
) *Hub {
	return &Hub{
		Channels:  make(map[string]map[*websocket.Conn]bool),
		Broadcast: make(chan Message),
		Subscribe: make(chan Subscription),
		log:       log,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (hub *Hub) run() {
	var (
		sub       Subscription
		err       error
		data      []byte
		conn      *websocket.Conn
		receivers map[*websocket.Conn]bool
		ok        bool
	)

	for {
		select {
		case sub = <-hub.Subscribe:
			if hub.Channels[sub.Channel] == nil {
				hub.Channels[sub.Channel] = make(map[*websocket.Conn]bool)
			}
			hub.Channels[sub.Channel][sub.Conn] = true
		case message := <-hub.Broadcast:
			if receivers, ok = hub.Channels[message.Channel]; ok {
				data, err = json.Marshal(message)
				if err != nil {
					hub.log.Error("failed to marshal message", sl.Err(err))

					continue
				}

				hub.log.Info("broadcasting message", sl.String("channel", message.Channel),
					sl.String("event", message.Event),
					sl.Any("data", message.Data))

				for conn = range receivers {
					if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
						hub.log.Error("failed to write message", sl.Err(err))
					}
				}
			}
		}
	}
}

func (hub *Hub) HandleConnection(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		ws      *websocket.Conn
		p       []byte
		message *Message
	)

	ws, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		hub.log.Error("failed to upgrade connection", sl.Err(err))

		return
	}
	defer func(ws *websocket.Conn) {
		err = ws.Close()
		if err != nil {
			hub.log.Error("failed to close connection", sl.Err(err))
		}
	}(ws)

	for {
		_, p, err = ws.ReadMessage()
		if err != nil {
			hub.log.Error("failed to read message", sl.Err(err))
			return
		}

		err = json.Unmarshal(p, message)
		if err != nil {
			hub.log.Error("failed to unmarshal message", sl.Err(err))

			continue
		}

		// Log incoming message
		hub.log.Info("incoming message", sl.String("channel", message.Channel),
			sl.String("event", message.Event),
			sl.Any("data", message.Data))

		if hub.Channels[message.Channel] == nil {
			hub.Subscribe <- Subscription{Conn: ws, Channel: message.Channel}
		}

		hub.Broadcast <- *message
	}
}

func (hub *Hub) RunServer() {
	go hub.run()
}
