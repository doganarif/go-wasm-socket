.PHONY: build buildwasm run

run: buildwasm
	go run main.go

buildwasm:
	GOOS=js GOARCH=wasm go build -o ./public/main.wasm ./wasm/wasm.go
