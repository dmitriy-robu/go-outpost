package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
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
}

func newHub() *Hub {
	return &Hub{
		Channels:  make(map[string]map[*websocket.Conn]bool),
		Broadcast: make(chan Message),
		Subscribe: make(chan Subscription),
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (h *Hub) run() {
	for {
		select {
		case sub := <-h.Subscribe:
			if h.Channels[sub.Channel] == nil {
				h.Channels[sub.Channel] = make(map[*websocket.Conn]bool)
			}
			h.Channels[sub.Channel][sub.Conn] = true
		case message := <-h.Broadcast:
			if receivers, ok := h.Channels[message.Channel]; ok {
				data, err := json.Marshal(message)
				if err != nil {
					log.Printf("error: %v", err)
					continue
				}

				log.Printf("Outgoing message to channel '%s': %s\n", message.Channel, string(data))

				for conn := range receivers {
					if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
						log.Println(err)
					}
				}
			}
		}
	}
}

var hub = newHub()

func handleConnection(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("Received data: %s\n", p)
		message := &Message{}
		err = json.Unmarshal(p, message)
		if err != nil {
			log.Printf("error: %v", err)
			continue
		}

		// Log incoming message
		log.Printf("Incoming message from channel '%s': %s\n", message.Channel, string(p))

		if hub.Channels[message.Channel] == nil {
			hub.Subscribe <- Subscription{Conn: ws, Channel: message.Channel}
		}

		hub.Broadcast <- *message
	}
}

func main() {
	go hub.run()

	http.HandleFunc("/ws", handleConnection)

	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}
