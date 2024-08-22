package main

import (
	"crypto/rand"
	"log/slog"
	"math/big"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type idgenState struct {
	idgenFunc func() (*big.Int, error)

	mu   sync.Mutex
	used map[*big.Int]struct{}
}

func (s *idgenState) nextId() (*big.Int, error) {
	for {
		id, err := s.idgenFunc()
		if err != nil {
			return nil, err
		}
		s.mu.Lock()
		if _, ok := s.used[id]; !ok {
			s.used[id] = struct{}{}
			s.mu.Unlock()
			return id, nil
		}
		s.mu.Unlock()
	}
}

type Server struct {
	addr  string
	idgen idgenState
}

var idgenFuncDefault = func() (*big.Int, error) {
	id, err := rand.Int(rand.Reader, big.NewInt(0).Exp(big.NewInt(2), big.NewInt(128), nil))
	if err != nil {
		return nil, err
	}
	return id, nil
}

func NewServer(addr string) *Server {
	return &Server{
		addr: addr,
		idgen: idgenState{
			used:      make(map[*big.Int]struct{}),
			idgenFunc: idgenFuncDefault,
		}}
}

func (s *Server) serveWs(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("ws upgrade failed", "err", err)
		return
	}
	defer conn.Close()

	for {
		// Ignore message
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("ws read message failed", "err", err)
			}
			break
		}

		id, err := s.idgen.nextId()
		if err != nil {
			slog.Error("random id request failed", "err", err)
			break
		}
		if err = conn.WriteMessage(websocket.TextMessage, []byte(id.String())); err != nil {
			slog.Error("ws write message failed", "err", err)
			break
		}
	}

}

func main() {
	s := NewServer("localhost:8080")
	http.HandleFunc("/ws", s.serveWs)
	slog.Info("server starting", "addr", s.addr)
	http.ListenAndServe(s.addr, nil)
}
