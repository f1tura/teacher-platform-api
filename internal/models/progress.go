package models

type StudentProgress struct {
	StudentID            int64   `json:"student_id"`
	TotalAssignments     int     `json:"total_assignments"`
	SubmittedAssignments int     `json:"submitted_assignments"`
	CheckedAssignments   int     `json:"checked_assignments"`
	AverageScore         float64 `json:"average_score"`
	CompletionRate       float64 `json:"completion_rate"`
}