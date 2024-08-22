package main

import (
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

const serverURL = "ws://localhost:8080/ws"

func connectAndRequest(wg *sync.WaitGroup, index int) {
	defer wg.Done()

	// Establish a connection to the WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		slog.Error("ws dial failed", "err", err)
		return
	}
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte("request")); err != nil {
		slog.Error("ws write message failed", "err", err)
		return
	}

	// Wait for the server's response
	_, message, err := conn.ReadMessage()
	if err != nil {
		slog.Error("ws read message failed", "err", err)
		return
	}

	// Print the unique random big integer received
	slog.Info("client request successed", "connection", index, "id", string(message))
}

func main() {
	var wg sync.WaitGroup

	wg.Add(10)
	for i := 1; i <= 10; i++ {
		go connectAndRequest(&wg, i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	slog.Info("All connections completed")
}
