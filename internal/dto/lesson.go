package dto

type CreateLessonInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	LessonDate  string `json:"lesson_date" binding:"required"`
	Status      string `json:"status"`
}

type UpdateLessonInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	LessonDate  string `json:"lesson_date" binding:"required"`
	Status      string `json:"status" binding:"required"`
}