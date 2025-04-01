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

type Recording struct {
	ID        uint   `gorm:"primarykey"`
	SessionID string `gorm:"index;not null"`
	Topic     string `gorm:"index;not null"`
	Duration  int    `gorm:"not null"` // Total duration in milliseconds
	Chunks    int    `gorm:"not null"` // Total number of chunks
	Metadata  JSON   `gorm:"type:jsonb;default:'{}'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AudioChunk struct {
	ID            uint        `gorm:"primarykey"`
	SessionID     string      `gorm:"index;not null"`
	S3Key         string      `gorm:"not null"` // Path to the chunk file in S3
	ChunkNumber   int         `gorm:"not null"`
	Duration      int         `gorm:"not null"` // Duration in milliseconds
	Size          int         `gorm:"not null"` // File size in bytes
	Status        ChunkStatus `gorm:"not null;default:'new'"`
	Transcription string      `gorm:"type:text"`
	Metadata      JSON        `gorm:"type:jsonb;default:'{}'"`
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Unique constraint on session_id and chunk_number
	UniqueConstraint struct {
		SessionID   string `gorm:"uniqueIndex:idx_session_chunk"`
		ChunkNumber int    `gorm:"uniqueIndex:idx_session_chunk"`
	} `gorm:"embedded"`
}

func (s ChunkStatus) String() string {
	return string(s)
}

type JSON map[string]interface{}
