package models

import "time"

type Attendance struct {
	ID        int64     `json:"id"`
	LessonID  int64     `json:"lesson_id"`
	StudentID int64     `json:"student_id"`
	Status    string    `json:"status"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}