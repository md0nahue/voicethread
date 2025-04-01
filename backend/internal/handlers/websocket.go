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
	// Track asked questions per session
	askedQuestions sync.Map
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

type QuestionRequestMessage struct {
	Type    string `json:"type"`
	TopicID string `json:"topicId"`
	UserID  string `json:"userId"`
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
			// Handle JSON messages (silence detection, question requests)
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

			switch msgType {
			case "silence":
				sessionID, ok := msg["sessionId"].(string)
				if !ok {
					log.Printf("Invalid session ID")
					continue
				}
				if err := h.handleSilenceDetected(context.Background(), sessionID); err != nil {
					log.Printf("Failed to handle silence detection: %v", err)
					continue
				}

			case "request_questions":
				var questionMsg QuestionRequestMessage
				if err := json.Unmarshal(message, &questionMsg); err != nil {
					log.Printf("Failed to unmarshal question request: %v", err)
					continue
				}
				if err := h.handleQuestionRequest(context.Background(), conn, questionMsg); err != nil {
					log.Printf("Failed to handle question request: %v", err)
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

func (h *WebSocketHandler) handleQuestionRequest(ctx context.Context, conn *websocket.Conn, msg QuestionRequestMessage) error {
	// Query the database for the topic and its questions
	var topic models.Topic
	if err := database.DB.Preload("InterviewSections").First(&topic, msg.TopicID).Error; err != nil {
		return fmt.Errorf("failed to fetch topic: %v", err)
	}

	// Get all section IDs for this topic
	var sectionIDs []string
	for _, section := range topic.InterviewSections {
		sectionIDs = append(sectionIDs, section.ID)
	}

	// Get asked questions for this session
	askedQuestionsMap, _ := h.askedQuestions.LoadOrStore(msg.UserID, make(map[string]bool))
	askedQuestions := askedQuestionsMap.(map[string]bool)

	// Query questions that haven't been asked yet
	var questions []models.Question
	query := database.DB.Where("interview_section_id IN ?", sectionIDs)

	// Order by is_followup_question DESC (prioritize follow-ups) and then by created_at
	query = query.Order("is_followup_question DESC, created_at ASC")

	if err := query.Find(&questions).Error; err != nil {
		return fmt.Errorf("failed to fetch questions: %v", err)
	}

	// Find the first unasked question
	var nextQuestion *models.Question
	for _, question := range questions {
		if !askedQuestions[question.ID] {
			nextQuestion = &question
			break
		}
	}

	// If no unasked questions found, return empty response
	if nextQuestion == nil {
		response := struct {
			Type      string `json:"type"`
			TopicID   string `json:"topic_id"`
			TopicBody string `json:"topic_body"`
			Questions []struct {
				ID          string  `json:"id"`
				Body        string  `json:"body"`
				AudioURL    string  `json:"audio_url"`
				Duration    float64 `json:"duration"`
				IsFollowup  bool    `json:"is_followup"`
				SectionID   string  `json:"section_id"`
				SectionBody string  `json:"section_body"`
			} `json:"questions"`
		}{
			Type:      "interview_questions",
			TopicID:   topic.ID,
			TopicBody: topic.Body,
		}
		return conn.WriteJSON(response)
	}

	// Mark the question as asked
	askedQuestions[nextQuestion.ID] = true
	h.askedQuestions.Store(msg.UserID, askedQuestions)

	// Find the section for this question
	var sectionBody string
	for _, section := range topic.InterviewSections {
		if section.ID == nextQuestion.InterviewSectionID {
			sectionBody = section.Body
			break
		}
	}

	// Format the response with just the next question
	response := struct {
		Type      string `json:"type"`
		TopicID   string `json:"topic_id"`
		TopicBody string `json:"topic_body"`
		Questions []struct {
			ID          string  `json:"id"`
			Body        string  `json:"body"`
			AudioURL    string  `json:"audio_url"`
			Duration    float64 `json:"duration"`
			IsFollowup  bool    `json:"is_followup"`
			SectionID   string  `json:"section_id"`
			SectionBody string  `json:"section_body"`
		} `json:"questions"`
	}{
		Type:      "interview_questions",
		TopicID:   topic.ID,
		TopicBody: topic.Body,
		Questions: []struct {
			ID          string  `json:"id"`
			Body        string  `json:"body"`
			AudioURL    string  `json:"audio_url"`
			Duration    float64 `json:"duration"`
			IsFollowup  bool    `json:"is_followup"`
			SectionID   string  `json:"section_id"`
			SectionBody string  `json:"section_body"`
		}{
			{
				ID:          nextQuestion.ID,
				Body:        nextQuestion.Body,
				AudioURL:    nextQuestion.URL,
				Duration:    nextQuestion.Duration,
				IsFollowup:  nextQuestion.IsFollowupQuestion,
				SectionID:   nextQuestion.InterviewSectionID,
				SectionBody: sectionBody,
			},
		},
	}

	return conn.WriteJSON(response)
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
