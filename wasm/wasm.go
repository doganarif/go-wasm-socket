package main

import (
	"context"
	"fmt"
	"log"
	"syscall/js"

	"nhooyr.io/websocket"
)

// Conn wraps a WebSocket connection.
type Conn struct {
	wsConn *websocket.Conn
}

// NewConn establishes a new WebSocket connection to a specified URL.
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
	// Channel to keep the main function running until it's closed.
	c := make(chan struct{}, 0)

	println("WASM Go Initialized")
	// Establish a new WebSocket connection.
	conn := NewConn()

	// Register the onButtonClick function in the global JavaScript context.
	js.Global().Set("onButtonClick", onButtonClickFunc(conn))

	// Start reading messages in a new goroutine.
	go conn.readMessage()

	// Wait indefinitely.
	<-c
}

// onButtonClickFunc returns a js.Func that sends a "HELLO" message over WebSocket when invoked.
func onButtonClickFunc(conn *Conn) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		println("Button Clicked!")
		// Send a message through the WebSocket connection.
		err := conn.wsConn.Write(context.Background(), websocket.MessageText, []byte("HELLO"))
		if err != nil {
			log.Println("Error writing to WebSocket:", err)
		}
		return nil
	})
}

// readMessage handles incoming WebSocket messages and updates the DOM accordingly.
func (c *Conn) readMessage() {
	defer func() {
		// Close the WebSocket connection when the function returns.
		c.wsConn.Close(websocket.StatusGoingAway, "BYE")
	}()

	for {
		// Read a message from the WebSocket connection.
		messageType, payload, err := c.wsConn.Read(context.Background())

		if err != nil {
			// Log and panic if there is an error reading the message.
			log.Panicf(err.Error())
		}

		// Update the DOM with the received message.
		updateDOMContent(string(payload))

		// Log the message type and payload for debugging.
		log.Println("MessageType: ", messageType)
		log.Println("Payload: ", string(payload))
	}
}

// updateDOMContent updates the text content of the DOM element with the given text.
func updateDOMContent(text string) {
	// Get the document object from the global JavaScript context.
	document := js.Global().Get("document")
	// Get the DOM element by its ID.
	element := document.Call("getElementById", "text")
	// Set the innerText of the element to the provided text.
	element.Set("innerText", text)
}
