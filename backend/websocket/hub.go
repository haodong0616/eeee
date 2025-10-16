package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client connected, total: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("Client disconnected, total: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastTrade(data interface{}) {
	message := Message{
		Type: "trade",
		Data: data,
	}
	h.broadcastMessage(message)
}

func (h *Hub) BroadcastOrderBook(data interface{}) {
	message := Message{
		Type: "orderbook",
		Data: data,
	}
	h.broadcastMessage(message)
}

func (h *Hub) BroadcastTicker(data interface{}) {
	message := Message{
		Type: "ticker",
		Data: data,
	}
	h.broadcastMessage(message)
}

func (h *Hub) broadcastMessage(message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}
	h.broadcast <- data
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
