package server

import (
	"log"
	"net/http"
	"os"

	"voicethread/internal/storage"

	"github.com/gorilla/websocket"
)

type Server struct {
	store    storage.Storage
	upgrader websocket.Upgrader
}

func New(store storage.Storage) *Server {
	return &Server{
		store: store,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	}
}

func (s *Server) Start() error {
	// WebSocket handler
	http.HandleFunc("/ws", s.handleWebSocket)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	return http.ListenAndServe(":"+port, nil)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		// Echo the message back
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("Error writing message: %v", err)
			break
		}
	}
}
