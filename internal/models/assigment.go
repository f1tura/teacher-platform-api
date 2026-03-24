package models

import "time"

type Assignment struct {
	ID          int64     `json:"id"`
	LessonID    int64     `json:"lesson_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	MaxScore    int       `json:"max_score"`
	IsPublished bool      `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}