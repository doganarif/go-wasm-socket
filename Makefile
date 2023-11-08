.PHONY: build buildwasm run clean serve

# Variables
WASM_DIR := ./wasm
PUBLIC_DIR := ./public
SERVER_FILE := main.go
WASM_SOURCE := $(WASM_DIR)/wasm.go
WASM_TARGET := $(PUBLIC_DIR)/main.wasm
GOOS := js
GOARCH := wasm

# Default rule
all: run

# Run the server
run: serve

# Serve the project
serve: buildwasm
	go run $(SERVER_FILE)

# Build the WebAssembly module
buildwasm:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(WASM_TARGET) $(WASM_SOURCE)

# Clean the built artifacts
clean:
	rm -f $(WASM_TARGET)
