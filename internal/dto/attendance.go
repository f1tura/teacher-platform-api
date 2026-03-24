package dto

type CreateAttendanceInput struct {
	StudentID int64  `json:"student_id" binding:"required"`
	Status    string `json:"status" binding:"required"`
	Comment   string `json:"comment"`
}

type UpdateAttendanceInput struct {
	Status  string `json:"status" binding:"required"`
	Comment string `json:"comment"`
}