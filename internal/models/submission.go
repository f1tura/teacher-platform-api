package models

import "time"

type Submission struct {
	ID           int64      `json:"id"`
	AssignmentID int64      `json:"assignment_id"`
	StudentID    int64      `json:"student_id"`
	AnswerText   string     `json:"answer_text"`
	FileURL      string     `json:"file_url"`
	Status       string     `json:"status"`
	Score        *int       `json:"score"`
	SubmittedAt  time.Time  `json:"submitted_at"`
	CheckedAt    *time.Time `json:"checked_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}