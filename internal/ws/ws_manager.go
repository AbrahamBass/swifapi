package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Room string

const (
	RoomAll   Room = "*" // Broadcast global
	RoomLobby Room = "lobby"
)

type Event string

const (
	EventMessage    Event = "message"
	EventClose      Event = "close"
	EventError      Event = "error"
	EventDisconnect Event = "disconnect"
)

type WebsocketManager struct {
	upgrader *websocket.Upgrader
	conn     *websocket.Conn
	logger   *zap.Logger
	events   map[string]func(data interface{})
	rooms    map[Room]struct{}
	done     chan struct{}
	clientID string
	mu       sync.RWMutex
	wg       sync.WaitGroup
}

type EventPayload struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type WebsocketHub struct {
	clients map[string]*WebsocketManager // Todos los clientes
	rooms   map[Room]map[string]struct{} // Mapa de salas
	mu      sync.RWMutex
}

var hub = &WebsocketHub{
	clients: make(map[string]*WebsocketManager),
	rooms:   make(map[Room]map[string]struct{}),
}

type WebsocketEmitter struct {
	source *WebsocketManager
	room   Room
}

func (e *WebsocketEmitter) Emit(event string, data interface{}) error {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	clients := hub.rooms[e.room]
	var errs []error

	for clientID := range clients {
		if client := hub.clients[clientID]; client != nil && client != e.source {
			if err := client.Replit(event, data); err != nil {
				errs = append(errs, err)
				e.source.logger.Error("Error en emisión",
					zap.String("room", string(e.room)),
					zap.Error(err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errores durante la emisión: %v", errs)
	}
	return nil
}

func (h *WebsocketHub) register(client *WebsocketManager) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.clientID] = client

	room := RoomAll
	if _, exists := h.rooms[room]; !exists {
		h.rooms[room] = make(map[string]struct{})
	}
	h.rooms[room][client.clientID] = struct{}{}

	client.mu.Lock()
	client.rooms[room] = struct{}{}
	client.mu.Unlock()
}

func (h *WebsocketHub) unregister(client *WebsocketManager) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients, client.clientID)
	for room := range client.rooms {
		delete(h.rooms[room], client.clientID)

		if len(h.rooms[room]) == 0 {
			delete(h.rooms, room)
		}
	}
}

func (m *WebsocketManager) CleanupRoom(room Room) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	if clients, exists := hub.rooms[room]; exists && len(clients) == 0 {
		delete(hub.rooms, room)
	}
}

func (c *WebsocketManager) Broadcast(event string, data interface{}) {
	c.To(RoomAll).Emit(event, data)
}

func (c *WebsocketManager) Join(room Room) {
	c.mu.Lock()
	c.rooms[room] = struct{}{}
	c.mu.Unlock()

	hub.mu.Lock()
	if _, exists := hub.rooms[room]; !exists {
		hub.rooms[room] = make(map[string]struct{})
	}
	hub.rooms[room][c.clientID] = struct{}{}
	hub.mu.Unlock()
}

func (c *WebsocketManager) Leave(room Room) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.rooms, room)
	hub.mu.Lock()
	defer hub.mu.Unlock()

	delete(hub.rooms[room], c.clientID)
}

func (c *WebsocketManager) To(room Room) *WebsocketEmitter {
	return &WebsocketEmitter{
		source: c,
		room:   room,
	}
}

func (h *WebsocketManager) On(event string, handler func(data interface{})) {
	h.events[event] = handler
}

func (h *WebsocketManager) Replit(event string, data interface{}) error {
	payload := EventPayload{
		Event: event,
		Data:  data,
	}

	msg, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return h.conn.WriteMessage(websocket.TextMessage, msg)
}

func (h *WebsocketManager) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.conn != nil {
		h.conn.Close()
		h.conn = nil
		hub.unregister(h)
	}

	h.wg.Wait()
	return nil
}

func (h *WebsocketManager) Start() {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.readLoop()
	}()
}

func (h *WebsocketManager) Connect(w http.ResponseWriter, r *http.Request) error {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	h.conn = conn
	hub.register(h)
	return nil
}

func (h *WebsocketManager) readLoop() {
	defer h.Close()

	for {
		_, message, err := h.conn.ReadMessage()
		if err != nil {
			var eventData string

			switch {
			case websocket.IsCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived):
				eventData = "Normal closure by the customer"

			case websocket.IsUnexpectedCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived):
				h.logger.Error("Unexpected closure", zap.Error(err))
				eventData = fmt.Sprintf("Error: %v", err)

			default:
				h.logger.Error("Read error", zap.Error(err))
				eventData = fmt.Sprintf("Critical Failure: %v", err)
			}

			h.triggerEvent(EventDisconnect, map[string]interface{}{
				"reason":   eventData,
				"clientID": h.clientID,
			})
			return
		}

		var payload EventPayload
		if err := json.Unmarshal(message, &payload); err != nil {
			h.triggerEvent(EventError, err)
			continue
		}

		h.triggerEvent(Event(payload.Event), payload.Data)
	}
}

func (h *WebsocketManager) triggerEvent(event Event, data interface{}) {
	if handler, exists := h.events[string(event)]; exists {
		go handler(data)
	}
}

func NewWebsocketManager(upgrader *websocket.Upgrader, logger *zap.Logger) *WebsocketManager {
	client := &WebsocketManager{
		upgrader: upgrader,
		logger:   logger,
		events:   make(map[string]func(data interface{})),
		done:     make(chan struct{}),
		rooms:    make(map[Room]struct{}),
		clientID: uuid.New().String(),
	}

	return client
}
