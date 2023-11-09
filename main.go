package main

import (
	"log"
	"net/http"

	"github.com/doganarif/go-wasm-socket/pkg/socket"
)

func main() {
	setupAPI()

	// Serve on port :8080, fudge yeah hardcoded port
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// setupAPI will start all Routes and their Handlers
func setupAPI() {
	manager := socket.NewManager()

	// Serve the ./frontend directory at Route /
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.Handle("/ws", http.HandlerFunc(manager.ServeWS))
}
