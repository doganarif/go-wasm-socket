package socket

import (
	"log"
	"math/rand"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type NotifyEvent struct {
	client  *Client
	message string
}

type Client struct {
	id         uint32
	connection *websocket.Conn
	manager    *Manager

	writeChan chan string
}

type Manager struct {
	clients ClientList

	sync.RWMutex

	notifyChan chan NotifyEvent
}

type ClientList map[*Client]bool

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		id:         rand.Uint32(),
		connection: conn,
		manager:    manager,
		writeChan:  make(chan string),
	}
}

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

func NewManager() *Manager {
	m := &Manager{
		clients:    make(ClientList),
		notifyChan: make(chan NotifyEvent),
	}

	go m.notifyOtherClients()

	return m
}

func (m *Manager) otherClients(client *Client) []*Client {
	clientList := make([]*Client, 0)

	for c := range m.clients {
		if c.id != client.id {
			clientList = append(clientList, c)
		}
	}

	return clientList
}

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

func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[client] = true
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[client]; ok {
		client.connection.Close()
		delete(m.clients, client)
	}
}

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
