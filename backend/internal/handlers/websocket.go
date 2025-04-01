package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
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
	Data      []byte `json:"data"`
	SessionID string `json:"session_id"`
}

type SilenceMessage struct {
	Type      string `json:"type"`
	SessionID string `json:"session_id"`
}

type QuestionRequestMessage struct {
	Type    string `json:"type"`
	TopicID string `json:"topicId"`
	UserID  string `json:"userId"`
}

type StatusUpdateMessage struct {
	Type      string  `json:"type"`
	SessionID string  `json:"session_id"`
	Status    string  `json:"status"`
	Progress  float64 `json:"progress"`
	Message   string  `json:"message"`
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
				if err := h.handleSilenceDetected(context.Background(), SilenceMessage{SessionID: sessionID}); err != nil {
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
			if err := h.handleAudioChunk(context.Background(), audioMsg); err != nil {
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
				Key:       audioMsg.SessionID,
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

func (h *WebSocketHandler) handleAudioChunk(ctx context.Context, msg AudioMessage) error {
	// Generate a unique key for this chunk
	key := fmt.Sprintf("chunks/%s/%d.wav", msg.SessionID, time.Now().UnixNano())

	// Save the audio chunk to S3
	if err := h.store.SaveAudio(ctx, key, msg.Data); err != nil {
		return fmt.Errorf("failed to save audio chunk: %v", err)
	}

	// Create a database record for this chunk
	chunk := models.AudioChunk{
		InterviewSessionID: msg.SessionID,
		S3Key:              key,
		ChunkNumber:        1, // TODO: Implement proper chunk numbering
		Duration:           0, // TODO: Calculate actual duration
		Size:               int64(len(msg.Data)),
		Status:             "processing",
	}

	if err := database.DB.Create(&chunk).Error; err != nil {
		// TODO: Implement S3 cleanup on database failure
		return fmt.Errorf("failed to create audio chunk record: %v", err)
	}

	// Send acknowledgment back to the client
	return nil
}

func (h *WebSocketHandler) handleSilenceDetected(ctx context.Context, msg SilenceMessage) error {
	// Get the current recording state
	state, ok := h.activeRecordings.Load(msg.SessionID)
	if !ok {
		return fmt.Errorf("no active recording found for session %s", msg.SessionID)
	}

	recordingState := state.(RecordingState)

	// Close the current chunk
	if err := h.closeCurrentChunk(ctx, recordingState); err != nil {
		return err
	}

	// Update the recording state
	recordingState.CurrentKey = ""
	recordingState.ChunkCount++
	h.activeRecordings.Store(msg.SessionID, recordingState)

	// Send acknowledgment back to the client
	return nil
}

func (h *WebSocketHandler) closeCurrentChunk(ctx context.Context, state RecordingState) error {
	if state.CurrentKey == "" {
		return nil
	}

	// Update the chunk status in the database
	if err := database.DB.Model(&models.AudioChunk{}).
		Where("s3_key = ?", state.CurrentKey).
		Update("status", "complete").Error; err != nil {
		return fmt.Errorf("failed to update chunk status: %v", err)
	}

	// Check if all chunks are complete
	var totalChunks int64
	var completedChunks int64

	if err := database.DB.Model(&models.AudioChunk{}).
		Where("interview_session_id = ?", state.CurrentKey).
		Count(&totalChunks).Error; err != nil {
		return fmt.Errorf("failed to count total chunks: %v", err)
	}

	if err := database.DB.Model(&models.AudioChunk{}).
		Where("interview_session_id = ? AND status = ?", state.CurrentKey, "complete").
		Count(&completedChunks).Error; err != nil {
		return fmt.Errorf("failed to count completed chunks: %v", err)
	}

	// If all chunks are complete, notify Rails
	if totalChunks > 0 && totalChunks == completedChunks {
		if err := h.notifySessionCompletion(ctx, state.CurrentKey); err != nil {
			return fmt.Errorf("failed to notify session completion: %v", err)
		}
	}

	return nil
}

func (h *WebSocketHandler) notifySessionCompletion(ctx context.Context, sessionID string) error {
	// Get the session details
	var session models.InterviewSession
	if err := database.DB.First(&session, sessionID).Error; err != nil {
		return fmt.Errorf("failed to fetch session: %v", err)
	}

	// Calculate total duration
	var totalDuration float64
	if err := database.DB.Model(&models.AudioChunk{}).
		Where("interview_session_id = ?", sessionID).
		Select("COALESCE(SUM(duration), 0)").
		Scan(&totalDuration).Error; err != nil {
		return fmt.Errorf("failed to calculate total duration: %v", err)
	}

	// Update session status
	session.Status = "completed"
	session.TotalDuration = totalDuration
	if err := database.DB.Save(&session).Error; err != nil {
		return fmt.Errorf("failed to update session status: %v", err)
	}

	// Send webhook to Rails
	webhookURL := fmt.Sprintf("%s/webhooks/golang", os.Getenv("RAILS_APP_URL"))
	payload := map[string]interface{}{
		"session_id": sessionID,
		"status":     "completed",
		"duration":   totalDuration,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %v", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}
