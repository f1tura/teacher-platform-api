package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"teacher-platform/internal/dto"
	"teacher-platform/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type SubmissionHandler struct {
	submissionRepo    *repository.SubmissionRepository
	assignmentRepo    *repository.AssignmentRepository
	lessonRepo        *repository.LessonRepository
	courseRepo        *repository.CourseRepository
	courseStudentRepo *repository.CourseStudentRepository
}

func NewSubmissionHandler(
	submissionRepo *repository.SubmissionRepository,
	assignmentRepo *repository.AssignmentRepository,
	lessonRepo *repository.LessonRepository,
	courseRepo *repository.CourseRepository,
	courseStudentRepo *repository.CourseStudentRepository,
) *SubmissionHandler {
	return &SubmissionHandler{
		submissionRepo:    submissionRepo,
		assignmentRepo:    assignmentRepo,
		lessonRepo:        lessonRepo,
		courseRepo:        courseRepo,
		courseStudentRepo: courseStudentRepo,
	}
}

// CreateSubmission godoc
// @Summary Create submission
// @Description Student submits a solution for an assignment
// @Tags submissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Assignment ID"
// @Param input body dto.CreateSubmissionInput true "Create submission input"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /assignments/{id}/submissions [post]
func (h *SubmissionHandler) Create(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	studentID, ok := userIDValue.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id type in context"})
		return
	}

	assignmentIDParam := c.Param("id")
	assignmentID, err := strconv.ParseInt(assignmentIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignment id"})
		return
	}

	var input dto.CreateSubmissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	assignment, err := h.assignmentRepo.GetByID(ctx, assignmentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get assignment"})
		return
	}

	if time.Now().After(assignment.DueDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "assignment deadline has passed"})
		return
	}

	lesson, err := h.lessonRepo.GetByID(ctx, assignment.LessonID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get lesson"})
		return
	}

	enrolled, err := h.courseStudentRepo.IsStudentEnrolled(ctx, lesson.CourseID, studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check course enrollment"})
		return
	}

	if !enrolled {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not enrolled in this course"})
		return
	}

	submission, err := h.submissionRepo.Create(ctx, repository.CreateSubmissionParams{
		AssignmentID: assignmentID,
		StudentID:    studentID,
		AnswerText:   input.AnswerText,
		FileURL:      input.FileURL,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "you have already submitted this assignment",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create submission"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": submission})
}

// GetSubmissionsByAssignmentID godoc
// @Summary Get submissions by assignment
// @Description Teacher/Admin gets all submissions for an assignment
// @Tags submissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Assignment ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /assignments/{id}/submissions [get]
func (h *SubmissionHandler) GetByAssignmentID(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id type in context"})
		return
	}

	roleValue, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "role not found in context"})
		return
	}

	role, ok := roleValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid role type in context"})
		return
	}

	assignmentIDParam := c.Param("id")
	assignmentID, err := strconv.ParseInt(assignmentIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignment id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	assignment, err := h.assignmentRepo.GetByID(ctx, assignmentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get assignment"})
		return
	}

	lesson, err := h.lessonRepo.GetByID(ctx, assignment.LessonID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get lesson"})
		return
	}

	course, err := h.courseRepo.GetByID(ctx, lesson.CourseID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get course"})
		return
	}

	if role != "admin" && course.TeacherID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "you can only view submissions in your own courses",
		})
		return
	}

	submissions, err := h.submissionRepo.GetByAssignmentID(ctx, assignmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get submissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": submissions})
}

// GetSubmissionByID godoc
// @Summary Get submission by id
// @Description Get one submission by id
// @Tags submissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Submission ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /submissions/{id} [get]
func (h *SubmissionHandler) GetByID(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id type in context"})
		return
	}

	roleValue, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "role not found in context"})
		return
	}

	role, ok := roleValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid role type in context"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	submission, err := h.submissionRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get submission"})
		return
	}

	if role == "student" {
		if submission.StudentID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "you can only view your own submission"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": submission})
		return
	}

	if role == "admin" {
		c.JSON(http.StatusOK, gin.H{"data": submission})
		return
	}

	assignment, err := h.assignmentRepo.GetByID(ctx, submission.AssignmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get assignment"})
		return
	}

	lesson, err := h.lessonRepo.GetByID(ctx, assignment.LessonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get lesson"})
		return
	}

	course, err := h.courseRepo.GetByID(ctx, lesson.CourseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get course"})
		return
	}

	if course.TeacherID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only view submissions in your own courses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": submission})
}

// ReviewSubmission godoc
// @Summary Review submission
// @Description Teacher/Admin reviews a submission and sets score/status
// @Tags submissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Submission ID"
// @Param input body dto.ReviewSubmissionInput true "Review submission input"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /submissions/{id}/review [patch]
func (h *SubmissionHandler) Review(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id type in context"})
		return
	}

	roleValue, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "role not found in context"})
		return
	}

	role, ok := roleValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid role type in context"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission id"})
		return
	}

	var input dto.ReviewSubmissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Status != "checked" && input.Status != "returned" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "status must be one of: checked, returned",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	submission, err := h.submissionRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get submission"})
		return
	}

	assignment, err := h.assignmentRepo.GetByID(ctx, submission.AssignmentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get assignment"})
		return
	}

	if input.Score < 0 || input.Score > assignment.MaxScore {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "score must be between 0 and max_score",
		})
		return
	}

	lesson, err := h.lessonRepo.GetByID(ctx, assignment.LessonID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get lesson"})
		return
	}

	course, err := h.courseRepo.GetByID(ctx, lesson.CourseID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get course"})
		return
	}

	if role != "admin" && course.TeacherID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "you can only review submissions in your own courses",
		})
		return
	}

	updatedSubmission, err := h.submissionRepo.ReviewByID(ctx, id, repository.ReviewSubmissionParams{
		Status: input.Status,
		Score:  input.Score,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to review submission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updatedSubmission})
}

// UpdateSubmission godoc
// @Summary Update submission
// @Description Student updates own submission before deadline
// @Tags submissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Submission ID"
// @Param input body dto.UpdateSubmissionInput true "Update submission input"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /submissions/{id} [patch]
func (h *SubmissionHandler) UpdateByID(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	studentID, ok := userIDValue.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id type in context"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission id"})
		return
	}

	var input dto.UpdateSubmissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	submission, err := h.submissionRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get submission"})
		return
	}

	if submission.StudentID != studentID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only update your own submission"})
		return
	}

	if submission.Status == "checked" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "checked submission cannot be updated"})
		return
	}

	assignment, err := h.assignmentRepo.GetByID(ctx, submission.AssignmentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get assignment"})
		return
	}

	if time.Now().After(assignment.DueDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "assignment deadline has passed"})
		return
	}

	updatedSubmission, err := h.submissionRepo.UpdateByID(ctx, id, repository.UpdateSubmissionParams{
		AnswerText: input.AnswerText,
		FileURL:    input.FileURL,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update submission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updatedSubmission})
}

// DeleteSubmission godoc
// @Summary Delete submission
// @Description Student deletes own submission before deadline
// @Tags submissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Submission ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /submissions/{id} [delete]
func (h *SubmissionHandler) DeleteByID(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	studentID, ok := userIDValue.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id type in context"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	submission, err := h.submissionRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get submission"})
		return
	}

	if submission.StudentID != studentID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own submission"})
		return
	}

	if submission.Status == "checked" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "checked submission cannot be deleted"})
		return
	}

	assignment, err := h.assignmentRepo.GetByID(ctx, submission.AssignmentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get assignment"})
		return
	}

	if time.Now().After(assignment.DueDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "assignment deadline has passed"})
		return
	}

	err = h.submissionRepo.DeleteByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete submission"})
		return
	}

	c.Status(http.StatusNoContent)
}