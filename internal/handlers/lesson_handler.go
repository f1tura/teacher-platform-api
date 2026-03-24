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

type LessonHandler struct {
	lessonRepo *repository.LessonRepository
	courseRepo *repository.CourseRepository
}

func NewLessonHandler(lessonRepo *repository.LessonRepository, courseRepo *repository.CourseRepository) *LessonHandler {
	return &LessonHandler{
		lessonRepo: lessonRepo,
		courseRepo: courseRepo,
	}
}

// CreateLesson godoc
// @Summary Create lesson
// @Description Create a new lesson in a course
// @Tags lessons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Course ID"
// @Param input body dto.CreateLessonInput true "Create lesson input"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /courses/{id}/lessons [post]
func (h *LessonHandler) Create(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user_id not found in context",
		})
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	roleValue, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "role not found in context",
		})
		return
	}

	role, ok := roleValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid role type in context",
		})
		return
	}

	courseIDParam := c.Param("id")
	courseID, err := strconv.ParseInt(courseIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid course id",
		})
		return
	}

	var input dto.CreateLessonInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	status := input.Status
	if status == "" {
		status = "planned"
	}

	if status != "planned" && status != "completed" && status != "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "status must be one of: planned, completed, cancelled",
		})
		return
	}

	lessonDate, err := time.Parse(time.RFC3339, input.LessonDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "lesson_date must be in RFC3339 format, e.g. 2026-03-25T18:00:00Z",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	course, err := h.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "course not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get course",
		})
		return
	}

	if role != "admin" && course.TeacherID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "you can only create lessons in your own courses",
		})
		return
	}

	lesson, err := h.lessonRepo.Create(ctx, repository.CreateLessonParams{
		CourseID:    courseID,
		Title:       input.Title,
		Description: input.Description,
		LessonDate:  lessonDate.Format(time.RFC3339),
		Status:      status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create lesson",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": lesson,
	})
}

// GetLessonsByCourseID godoc
// @Summary Get lessons by course
// @Description Get all lessons for a course
// @Tags lessons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Course ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /courses/{id}/lessons [get]
func (h *LessonHandler) GetByCourseID(c *gin.Context) {
	courseIDParam := c.Param("id")
	courseID, err := strconv.ParseInt(courseIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid course id",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	_, err = h.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "course not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get course",
		})
		return
	}

	lessons, err := h.lessonRepo.GetByCourseID(ctx, courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get lessons",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": lessons,
	})
}

// GetLessonByID godoc
// @Summary Get lesson by id
// @Description Get one lesson by id
// @Tags lessons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Lesson ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /lessons/{id} [get]
func (h *LessonHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid lesson id",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	lesson, err := h.lessonRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "lesson not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get lesson",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": lesson,
	})
}

// UpdateLesson godoc
// @Summary Update lesson
// @Description Update lesson by id
// @Tags lessons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Lesson ID"
// @Param input body dto.UpdateLessonInput true "Update lesson input"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /lessons/{id} [patch]
func (h *LessonHandler) UpdateByID(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	var input dto.UpdateLessonInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Status != "planned" && input.Status != "completed" && input.Status != "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "status must be one of: planned, completed, cancelled",
		})
		return
	}

	lessonDate, err := time.Parse(time.RFC3339, input.LessonDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "lesson_date must be in RFC3339 format, e.g. 2026-03-25T18:00:00Z",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	existingLesson, err := h.lessonRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get lesson"})
		return
	}

	course, err := h.courseRepo.GetByID(ctx, existingLesson.CourseID)
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
			"error": "you can only update lessons in your own courses",
		})
		return
	}

	updatedLesson, err := h.lessonRepo.UpdateByID(ctx, id, repository.UpdateLessonParams{
		Title:       input.Title,
		Description: input.Description,
		LessonDate:  lessonDate.Format(time.RFC3339),
		Status:      input.Status,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update lesson"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updatedLesson})
}

// DeleteLesson godoc
// @Summary Delete lesson
// @Description Delete lesson by id
// @Tags lessons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Lesson ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /lessons/{id} [delete]
func (h *LessonHandler) DeleteByID(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	existingLesson, err := h.lessonRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get lesson"})
		return
	}

	course, err := h.courseRepo.GetByID(ctx, existingLesson.CourseID)
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
			"error": "you can only delete lessons in your own courses",
		})
		return
	}

	err = h.lessonRepo.DeleteByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete lesson"})
		return
	}

	c.Status(http.StatusNoContent)
}