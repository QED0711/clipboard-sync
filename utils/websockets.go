package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/savsgio/gotils/uuid"
)

/* ============================================= CLIENT ============================================= */

type Client struct {
	ID   string
	Conn *websocket.Conn
}

/* ============================================= WEBSOCKET HUB ============================================= */

type WebSocketHub struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

var wsHubOnce sync.Once
var wsHub *WebSocketHub

func GetWebSocketHub() *WebSocketHub {
	wsHubOnce.Do(func() {
		wsHub = &WebSocketHub{
			clients: make(map[string]*Client),
		}
	})
	return wsHub
}

func (hub *WebSocketHub) Register(client *Client) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	hub.clients[client.ID] = client
}

func (hub *WebSocketHub) Unregister(clientID string) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	delete(hub.clients, clientID)
}

func (hub *WebSocketHub) BroadcastFrom(senderID string, message []byte) {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	for id, client := range hub.clients {
		if id == senderID {
			continue
		}
		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error broadcasting to %s: %v\n", id, err)
		}
	}

}

func (hub *WebSocketHub) GetClientByID(clientID string) (*Client, bool) {
	for id, client := range hub.clients {
		if id == clientID {
			return client, true
		}
	}
	return &Client{}, false
}

func (hub *WebSocketHub) RespondTo(clientID string, message []byte) {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	client, ok := hub.GetClientByID(clientID)
	if ok {
		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error Responding to %s: %v\n", clientID, err)
		}
	}
}

/* ============================================= HANDLER ============================================= */

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(req *http.Request) bool {
		return true // allow all connections
	},
}

func WebsocketReceiverHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	client := &Client{
		ID:   uuid.V4(),
		Conn: conn,
	}

	hub := GetWebSocketHub()
	hub.Register(client)
	defer hub.Unregister(client.ID)

	for {
		// Receive messages
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Websocket Read Error: ", err)
			return
		}

		// Convert message into request struct
		var req Request
		err = json.Unmarshal([]byte(p), &req) // convert message to struct
		if err != nil {
			log.Println("JSON unmarshal error: ", err)
			return
		}

		req.HandleRequest(client, hub)

	}
}
