package models

import (
	"time"
)

type Topic struct {
	ID                string             `json:"id" gorm:"primaryKey"`
	Body              string             `json:"body"`
	InterviewSections []InterviewSection `json:"interview_sections" gorm:"foreignKey:TopicID"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

type InterviewSection struct {
	ID        string     `json:"id" gorm:"primaryKey"`
	TopicID   string     `json:"topic_id"`
	Body      string     `json:"body"`
	Questions []Question `json:"questions" gorm:"foreignKey:InterviewSectionID"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Question struct {
	ID                 string    `json:"id" gorm:"primaryKey"`
	InterviewSectionID string    `json:"interview_section_id"`
	Body               string    `json:"body"`
	URL                string    `json:"url"`
	Duration           float64   `json:"duration"`
	Status             string    `json:"status"`
	IsFollowupQuestion bool      `json:"is_followup_question"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
