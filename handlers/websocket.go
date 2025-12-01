package handlers

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	gameID string
}

type Hub struct {
	clients    map[string]map[*Client]bool
	broadcast  chan *BroadcastMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type BroadcastMessage struct {
	GameID  string
	Message []byte
}

var GlobalHub = NewHub()

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan *BroadcastMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.gameID] == nil {
				h.clients[client.gameID] = make(map[*Client]bool)
			}
			h.clients[client.gameID][client] = true
			h.mu.Unlock()
			log.Printf("Client registered for game: %s. Total: %d", client.gameID, len(h.clients[client.gameID]))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.gameID][client]; ok {
				delete(h.clients[client.gameID], client)
				close(client.send)
				log.Printf("Client unregistered from game: %s", client.gameID)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[message.GameID]
			h.mu.RUnlock()

			for client := range clients {
				select {
				case client.send <- message.Message:
				default:
					h.mu.Lock()
					close(client.send)
					delete(h.clients[message.GameID], client)
					h.mu.Unlock()
				}
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
}

func ServeWs(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	if gameID == "" {
		gameID = "global"
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		hub:    GlobalHub,
		conn:   conn,
		send:   make(chan []byte, 256),
		gameID: gameID,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func BroadcastLeaderboardUpdate(gameID string, data interface{}) {
	message, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshaling leaderboard update:", err)
		return
	}

	GlobalHub.broadcast <- &BroadcastMessage{
		GameID:  gameID,
		Message: message,
	}
}
