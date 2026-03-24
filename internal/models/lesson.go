package models

import "time"

type Lesson struct {
	ID          int64     `json:"id"`
	CourseID    int64     `json:"course_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	LessonDate  time.Time `json:"lesson_date"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}