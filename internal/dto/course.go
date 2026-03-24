package dto

type CreateCourseInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type UpdateCourseInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active" binding:"required"`
}