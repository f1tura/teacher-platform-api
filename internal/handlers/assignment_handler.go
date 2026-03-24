package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"teacher-platform/internal/dto"
	"teacher-platform/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type AssignmentHandler struct {
	assignmentRepo *repository.AssignmentRepository
	lessonRepo     *repository.LessonRepository
	courseRepo     *repository.CourseRepository
}

func NewAssignmentHandler(
	assignmentRepo *repository.AssignmentRepository,
	lessonRepo *repository.LessonRepository,
	courseRepo *repository.CourseRepository,
) *AssignmentHandler {
	return &AssignmentHandler{
		assignmentRepo: assignmentRepo,
		lessonRepo:     lessonRepo,
		courseRepo:     courseRepo,
	}
}


// CreateAssignment godoc
// @Summary Create assignment
// @Description Create a new assignment in a lesson
// @Tags assignments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Lesson ID"
// @Param input body dto.CreateAssignmentInput true "Create assignment input"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /lessons/{id}/assignments [post]
func (h *AssignmentHandler) Create(c *gin.Context) {
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

	lessonIDParam := c.Param("id")
	lessonID, err := strconv.ParseInt(lessonIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	var input dto.CreateAssignmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.MaxScore <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "max_score must be greater than 0"})
		return
	}

	dueDate, err := time.Parse(time.RFC3339, input.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "due_date must be in RFC3339 format, e.g. 2026-03-30T18:00:00Z",
		})
		return
	}

	isPublished := true
	if input.IsPublished != nil {
		isPublished = *input.IsPublished
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	lesson, err := h.lessonRepo.GetByID(ctx, lessonID)
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
			"error": "you can only create assignments in your own courses",
		})
		return
	}

	assignment, err := h.assignmentRepo.Create(ctx, repository.CreateAssignmentParams{
		LessonID:    lessonID,
		Title:       input.Title,
		Description: input.Description,
		DueDate:     dueDate.Format(time.RFC3339),
		MaxScore:    input.MaxScore,
		IsPublished: isPublished,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create assignment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": assignment})
}


// GetAssignmentsByLessonID godoc
// @Summary Get assignments by lesson
// @Description Get all assignments for a lesson
// @Tags assignments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Lesson ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /lessons/{id}/assignments [get]
func (h *AssignmentHandler) GetByLessonID(c *gin.Context) {
	lessonIDParam := c.Param("id")
	lessonID, err := strconv.ParseInt(lessonIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	_, err = h.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get lesson"})
		return
	}

	assignments, err := h.assignmentRepo.GetByLessonID(ctx, lessonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get assignments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": assignments})
}

// GetAssignmentByID godoc
// @Summary Get assignment by id
// @Description Get one assignment by id
// @Tags assignments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Assignment ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /assignments/{id} [get]
func (h *AssignmentHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignment id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	assignment, err := h.assignmentRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get assignment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": assignment})
}

// UpdateAssignment godoc
// @Summary Update assignment
// @Description Update assignment by id
// @Tags assignments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Assignment ID"
// @Param input body dto.UpdateAssignmentInput true "Update assignment input"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /assignments/{id} [patch]
func (h *AssignmentHandler) UpdateByID(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignment id"})
		return
	}

	var input dto.UpdateAssignmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.MaxScore <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "max_score must be greater than 0"})
		return
	}

	dueDate, err := time.Parse(time.RFC3339, input.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "due_date must be in RFC3339 format, e.g. 2026-03-30T18:00:00Z",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	existingAssignment, err := h.assignmentRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get assignment"})
		return
	}

	lesson, err := h.lessonRepo.GetByID(ctx, existingAssignment.LessonID)
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
			"error": "you can only update assignments in your own courses",
		})
		return
	}

	updatedAssignment, err := h.assignmentRepo.UpdateByID(ctx, id, repository.UpdateAssignmentParams{
		Title:       input.Title,
		Description: input.Description,
		DueDate:     dueDate.Format(time.RFC3339),
		MaxScore:    input.MaxScore,
		IsPublished: *input.IsPublished,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update assignment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updatedAssignment})
}

func (h *AssignmentHandler) DeleteByID(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignment id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	existingAssignment, err := h.assignmentRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get assignment"})
		return
	}

	lesson, err := h.lessonRepo.GetByID(ctx, existingAssignment.LessonID)
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
			"error": "you can only delete assignments in your own courses",
		})
		return
	}

	err = h.assignmentRepo.DeleteByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete assignment"})
		return
	}

	c.Status(http.StatusNoContent)
}