package models

import (
	"time"
)

// ChunkStatus represents the status of an audio chunk
type ChunkStatus string

const (
	ChunkStatusNew         ChunkStatus = "new"
	ChunkStatusTranscribed ChunkStatus = "transcribed"
)

type InterviewSession struct {
	ID                    string       `json:"id" gorm:"primaryKey"`
	UserID                string       `json:"user_id"`
	TopicID               string       `json:"topic_id"`
	Status                string       `json:"status"`
	Metadata              string       `json:"metadata" gorm:"type:jsonb"`
	FinalAudioURL         string       `json:"final_audio_url"`
	TotalDuration         float64      `json:"total_duration"`
	PolishedTranscription string       `json:"polished_transcription"`
	RawTranscript         string       `json:"raw_transcript"`
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`
	AudioChunks           []AudioChunk `json:"audio_chunks" gorm:"foreignKey:InterviewSessionID"`
}

type AudioChunk struct {
	ID                 string    `json:"id" gorm:"primaryKey"`
	InterviewSessionID string    `json:"interview_session_id"`
	S3Key              string    `json:"s3_key"`
	ChunkNumber        int       `json:"chunk_number"`
	Duration           float64   `json:"duration"`
	Size               int64     `json:"size"`
	Status             string    `json:"status"`
	Transcription      string    `json:"transcription"`
	Metadata           string    `json:"metadata" gorm:"type:jsonb"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (s ChunkStatus) String() string {
	return string(s)
}

type JSON map[string]interface{}
