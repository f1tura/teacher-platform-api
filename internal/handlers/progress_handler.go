package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"teacher-platform/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type ProgressHandler struct {
	progressRepo      *repository.ProgressRepository
	userRepo          *repository.UserRepository
	courseStudentRepo *repository.CourseStudentRepository
}

func NewProgressHandler(
	progressRepo *repository.ProgressRepository,
	userRepo *repository.UserRepository,
	courseStudentRepo *repository.CourseStudentRepository,
) *ProgressHandler {
	return &ProgressHandler{
		progressRepo:      progressRepo,
		userRepo:          userRepo,
		courseStudentRepo: courseStudentRepo,
	}
}

// GetStudentProgress godoc
// @Summary Get student progress
// @Description Get aggregated progress for a student
// @Tags progress
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Student ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /students/{id}/progress [get]
func (h *ProgressHandler) GetStudentProgress(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	currentUserID, ok := userIDValue.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id type in context"})
		return
	}

	roleValue, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "role not found in context"})
		return
	}

	currentRole, ok := roleValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid role type in context"})
		return
	}

	studentIDParam := c.Param("id")
	studentID, err := strconv.ParseInt(studentIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	student, err := h.userRepo.GetByID(ctx, studentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get student"})
		return
	}

	if student.Role != "student" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "progress is available only for users with role student"})
		return
	}

	switch currentRole {
	case "student":
		if currentUserID != studentID {
			c.JSON(http.StatusForbidden, gin.H{"error": "you can only view your own progress"})
			return
		}
	case "teacher":
		allowed, err := h.courseStudentRepo.IsStudentInTeacherCourses(ctx, studentID, currentUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check teacher-student relation"})
			return
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "you can only view progress of your own students"})
			return
		}
	case "admin":
		// allowed
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	progress, err := h.progressRepo.GetStudentProgress(ctx, studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get student progress",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": progress,
	})
}