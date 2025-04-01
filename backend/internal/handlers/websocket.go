package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"voicethread/internal/storage"

	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	upgrader websocket.Upgrader
	store    storage.Storage
}

type AudioMessage struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionId"`
	Data      []byte `json:"data"`
}

func NewWebSocketHandler(upgrader websocket.Upgrader, store storage.Storage) *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: upgrader,
		store:    store,
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		if messageType == websocket.BinaryMessage {
			// Handle binary audio data
			var audioMsg AudioMessage
			if err := json.Unmarshal(message, &audioMsg); err != nil {
				log.Printf("Failed to unmarshal audio message: %v", err)
				continue
			}

			// Save the audio chunk
			key, err := h.store.Save(context.Background(), audioMsg.SessionID, audioMsg.Data)
			if err != nil {
				log.Printf("Failed to save audio chunk: %v", err)
				continue
			}

			// Send acknowledgment
			response := struct {
				Type      string `json:"type"`
				Status    string `json:"status"`
				Key       string `json:"key"`
				SessionID string `json:"sessionId"`
			}{
				Type:      "ack",
				Status:    "success",
				Key:       key,
				SessionID: audioMsg.SessionID,
			}

			if err := conn.WriteJSON(response); err != nil {
				log.Printf("Failed to send acknowledgment: %v", err)
			}
		}
	}
} 