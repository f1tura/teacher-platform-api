package models

import "time"

type Course struct {
	ID          int64     `json:"id"`
	TeacherID   int64     `json:"teacher_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}