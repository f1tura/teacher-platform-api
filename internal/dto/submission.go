package dto

type CreateSubmissionInput struct {
	AnswerText string `json:"answer_text" binding:"required"`
	FileURL    string `json:"file_url"`
}

type ReviewSubmissionInput struct {
	Status string `json:"status" binding:"required"`
	Score  int    `json:"score" binding:"required"`
}

type UpdateSubmissionInput struct {
	AnswerText string `json:"answer_text" binding:"required"`
	FileURL    string `json:"file_url"`
}