package models

import (
	"time"
)

// ChunkStatus represents the transcription status of an audio chunk
type ChunkStatus int

const (
	ChunkStatusNew ChunkStatus = iota
	ChunkStatusTranscribing
	ChunkStatusTranscribed
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
	ID          uint        `gorm:"primarykey"`
	RecordingID uint        `gorm:"not null"`
	Recording   Recording   `gorm:"foreignKey:RecordingID"`
	ChunkNumber int         `gorm:"not null"`
	FilePath    string      `gorm:"not null"` // Path to the chunk file in storage
	Duration    int         `gorm:"not null"` // Duration in milliseconds
	Size        int         `gorm:"not null"` // File size in bytes
	Status      ChunkStatus `gorm:"not null;default:0"`
	Transcript  string      `gorm:"type:text"`
	Metadata    JSON        `gorm:"type:jsonb;default:'{}'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s ChunkStatus) String() string {
	switch s {
	case ChunkStatusNew:
		return "new"
	case ChunkStatusTranscribing:
		return "transcribing"
	case ChunkStatusTranscribed:
		return "transcribed"
	default:
		return "unknown"
	}
}

type JSON map[string]interface{}
