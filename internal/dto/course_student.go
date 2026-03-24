package dto

type AddStudentToCourseInput struct {
	StudentID int64 `json:"student_id" binding:"required"`
}