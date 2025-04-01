package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"voicethread/internal/database"
	"voicethread/internal/models"
	"voicethread/internal/storage"

	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	upgrader websocket.Upgrader
	store    storage.Storage
	// Track active recordings per session
	activeRecordings sync.Map
}

type AudioMessage struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionId"`
	Data      []byte `json:"data"`
}

type SilenceMessage struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionId"`
}

type RecordingState struct {
	CurrentKey string
	ChunkCount int
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

		if messageType == websocket.TextMessage {
			// Handle JSON messages (silence detection)
			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			msgType, ok := msg["type"].(string)
			if !ok {
				log.Printf("Invalid message type")
				continue
			}

			sessionID, ok := msg["sessionId"].(string)
			if !ok {
				log.Printf("Invalid session ID")
				continue
			}

			switch msgType {
			case "silence":
				// Handle silence detection
				if err := h.handleSilenceDetected(context.Background(), sessionID); err != nil {
					log.Printf("Failed to handle silence detection: %v", err)
					continue
				}
			}
		} else if messageType == websocket.BinaryMessage {
			// Handle binary audio data
			var audioMsg AudioMessage
			if err := json.Unmarshal(message, &audioMsg); err != nil {
				log.Printf("Failed to unmarshal audio message: %v", err)
				continue
			}

			// Save the audio chunk
			key, err := h.handleAudioChunk(context.Background(), audioMsg)
			if err != nil {
				log.Printf("Failed to handle audio chunk: %v", err)
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

func (h *WebSocketHandler) handleAudioChunk(ctx context.Context, msg AudioMessage) (string, error) {
	// Get or create recording state for this session
	state, _ := h.activeRecordings.LoadOrStore(msg.SessionID, &RecordingState{
		ChunkCount: 0,
	})
	recordingState := state.(*RecordingState)
	recordingState.ChunkCount++

	// Generate a new key for this chunk
	key := h.generateChunkKey(msg.SessionID, recordingState.ChunkCount)

	// Save the chunk to S3
	_, err := h.store.Save(ctx, key, msg.Data)
	if err != nil {
		return "", fmt.Errorf("failed to save chunk to storage: %v", err)
	}

	// Create database record for the chunk
	chunk := &models.AudioChunk{
		SessionID:   msg.SessionID,
		S3Key:       key,
		ChunkNumber: recordingState.ChunkCount,
		Duration:    0, // TODO: Calculate actual duration
		Size:        len(msg.Data),
		Status:      models.ChunkStatusNew,
		Metadata:    models.JSON{"source": "websocket"},
	}

	if err := database.DB.Create(chunk).Error; err != nil {
		// If database creation fails, we should clean up the S3 object
		// TODO: Implement cleanup of S3 object
		return "", fmt.Errorf("failed to create chunk record: %v", err)
	}

	return key, nil
}

func (h *WebSocketHandler) handleSilenceDetected(ctx context.Context, sessionID string) error {
	// Get current recording state
	state, exists := h.activeRecordings.Load(sessionID)
	if !exists {
		return nil
	}

	recordingState := state.(*RecordingState)

	// Close current chunk if it exists
	if recordingState.CurrentKey != "" {
		if err := h.store.CloseChunk(ctx, recordingState.CurrentKey); err != nil {
			log.Printf("Failed to close chunk %s: %v", recordingState.CurrentKey, err)
			return err
		}
	}

	// Clear current key to indicate new chunk should be created
	recordingState.CurrentKey = ""
	recordingState.ChunkCount++ // Increment chunk counter for next chunk

	return nil
}

func (h *WebSocketHandler) generateChunkKey(sessionID string, chunkNumber int) string {
	return fmt.Sprintf("%s/chunk_%d.webm", sessionID, chunkNumber)
}
