package socket

import (
	"log"
	"math/rand"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Pre-configure the upgrader, which is responsible for upgrading
// an HTTP connection to a WebSocket connection.
var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// NotifyEvent represents an event that contains a reference
// to the client who initiated the event and the message to be notified.
type NotifyEvent struct {
	client  *Client
	message string
}

// Client represents a single WebSocket connection.
// It holds the client's ID, the WebSocket connection itself, and
// the manager that controls all clients.
type Client struct {
	id         uint32
	connection *websocket.Conn
	manager    *Manager

	writeChan chan string
}

// Manager keeps track of all active clients and broadcasts messages.
type Manager struct {
	clients ClientList

	sync.RWMutex

	notifyChan chan NotifyEvent
}

// ClientList is a map of clients to keep track of their presence.
type ClientList map[*Client]bool

// NewClient creates a new Client instance with a unique ID, its connection,
// and a reference to the Manager.
func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		id:         rand.Uint32(),
		connection: conn,
		manager:    manager,
		writeChan:  make(chan string),
	}
}

// readMessages continuously reads messages from the WebSocket connection.
// It will send any received messages to the manager's notification channel.
func (c *Client) readMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	for {
		messageType, payload, err := c.connection.ReadMessage()

		c.manager.notifyChan <- NotifyEvent{client: c, message: string(payload)}

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}
		log.Println("MessageType: ", messageType)
		log.Println("Payload: ", string(payload))
	}
}

// writeMessages listens on the client's write channel for messages
// and writes any received messages to the WebSocket connection.
func (c *Client) writeMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	for {
		select {
		case data := <-c.writeChan:
			c.connection.WriteMessage(websocket.TextMessage, []byte(data))
		}
	}
}

// NewManager creates a new Manager instance, initializes the client list,
// and starts the goroutine responsible for notifying other clients.
func NewManager() *Manager {
	m := &Manager{
		clients:    make(ClientList),
		notifyChan: make(chan NotifyEvent),
	}

	go m.notifyOtherClients()

	return m
}

// otherClients returns a slice of clients excluding the provided client.
func (m *Manager) otherClients(client *Client) []*Client {
	clientList := make([]*Client, 0)

	for c := range m.clients {
		if c.id != client.id {
			clientList = append(clientList, c)
		}
	}

	return clientList
}

// notifyOtherClients waits for notify events and broadcasts the message
// to all clients except the one who sent the message.
func (m *Manager) notifyOtherClients() {
	for {
		select {
		case e := <-m.notifyChan:
			otherClients := m.otherClients(e.client)

			for _, c := range otherClients {
				c.writeChan <- e.message
			}
		}
	}
}

// addClient adds a new client to the manager's client list.
func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[client] = true
}

// removeClient removes a client from the manager's client list and
// closes the WebSocket connection.
func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[client]; ok {
		client.connection.Close()
		delete(m.clients, client)
	}
}

// ServeWS is an HTTP handler that upgrades the HTTP connection to a
// WebSocket connection and registers the new client with the manager.
func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
	log.Println("New Connection")

	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := NewClient(conn, m)
	m.addClient(client)

	go client.readMessages()
	go client.writeMessages()
}
