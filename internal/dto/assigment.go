package dto

type CreateAssignmentInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	DueDate     string `json:"due_date" binding:"required"`
	MaxScore    int    `json:"max_score" binding:"required"`
	IsPublished *bool  `json:"is_published"`
}

type UpdateAssignmentInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	DueDate     string `json:"due_date" binding:"required"`
	MaxScore    int    `json:"max_score" binding:"required"`
	IsPublished *bool  `json:"is_published" binding:"required"`
}