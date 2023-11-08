package main

import (
	"context"
	"fmt"
	"log"
	"syscall/js"

	"nhooyr.io/websocket"
)

type Conn struct {
	wsConn *websocket.Conn
}

func NewConn() *Conn {
	c, _, err := websocket.Dial(context.Background(), "ws://localhost:8080/ws", nil)
	if err != nil {
		fmt.Println(err, "ERROR")
	}

	return &Conn{
		wsConn: c,
	}
}
func main() {
	c := make(chan struct{}, 0)

	println("WASM Go Initialized")
	// register functions
	conn := NewConn()

	js.Global().Set("logMessage", logMessageFunc(conn))

	go conn.readMessage()
	<-c
}

func logMessageFunc(conn *Conn) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		println("Button Clicked!")
		// Use the Conn instance that was passed in when creating this function
		err := conn.wsConn.Write(context.Background(), websocket.MessageText, []byte("HELLO"))
		if err != nil {
			log.Println("Error writing to WebSocket:", err)
		}
		return nil
	})
}

func (c *Conn) readMessage() {
	defer func() {
		c.wsConn.Close(websocket.StatusGoingAway, "BYE")
	}()

	for {
		messageType, payload, err := c.wsConn.Read(context.Background())

		if err != nil {
			log.Panicf(err.Error())
		}

		updateDOMContent(string(payload))

		log.Println("MessageType: ", messageType)
		log.Println("Payload: ", string(payload))
	}

}

func updateDOMContent(text string) {
	document := js.Global().Get("document")
	element := document.Call("getElementById", "count")
	element.Set("innerText", text)
}
