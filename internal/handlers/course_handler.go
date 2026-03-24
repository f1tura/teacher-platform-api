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

type CourseHandler struct {
	repo *repository.CourseRepository
}

func NewCourseHandler(repo *repository.CourseRepository) *CourseHandler {
	return &CourseHandler{repo: repo}
}

// CreateCourse godoc
// @Summary Create course
// @Description Create a new course for teacher/admin
// @Tags courses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body dto.CreateCourseInput true "Create course input"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /courses [post]
func (h *CourseHandler) Create(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user_id not found in context",
		})
		return
	}

	teacherID, ok := userIDValue.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	var input dto.CreateCourseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	course, err := h.repo.Create(ctx, repository.CreateCourseParams{
		TeacherID:   teacherID,
		Title:       input.Title,
		Description: input.Description,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create course",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": course,
	})
}

// GetAllCourses godoc
// @Summary Get all courses
// @Description Get list of courses
// @Tags courses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /courses [get]
func (h *CourseHandler) GetAll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	courses, err := h.repo.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get courses",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": courses,
	})
}

// GetCourseByID godoc
// @Summary Get course by id
// @Description Get one course by id
// @Tags courses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Course ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /courses/{id} [get]
func (h *CourseHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid course id",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	course, err := h.repo.GetByID(ctx, id)
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

	c.JSON(http.StatusOK, gin.H{
		"data": course,
	})
}

// UpdateCourse godoc
// @Summary Update course
// @Description Update course by id
// @Tags courses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Course ID"
// @Param input body dto.UpdateCourseInput true "Update course input"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /courses/{id} [patch]
func (h *CourseHandler) UpdateByID(c *gin.Context) {
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

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid course id",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	existingCourse, err := h.repo.GetByID(ctx, id)
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

	if role != "admin" && existingCourse.TeacherID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "you can only update your own courses",
		})
		return
	}

	var input dto.UpdateCourseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	updatedCourse, err := h.repo.UpdateByID(ctx, id, repository.UpdateCourseParams{
		Title:       input.Title,
		Description: input.Description,
		IsActive:    *input.IsActive,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "course not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update course",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": updatedCourse,
	})
}

// DeleteCourse godoc
// @Summary Delete course
// @Description Delete course by id
// @Tags courses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Course ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /courses/{id} [delete]
func (h *CourseHandler) DeleteByID(c *gin.Context) {
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

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid course id",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	existingCourse, err := h.repo.GetByID(ctx, id)
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

	if role != "admin" && existingCourse.TeacherID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "you can only delete your own courses",
		})
		return
	}

	err = h.repo.DeleteByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "course not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete course",
		})
		return
	}

	c.Status(http.StatusNoContent)
}