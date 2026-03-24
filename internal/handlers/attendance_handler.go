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

type AttendanceHandler struct {
	attendanceRepo    *repository.AttendanceRepository
	lessonRepo        *repository.LessonRepository
	courseRepo        *repository.CourseRepository
	userRepo          *repository.UserRepository
	courseStudentRepo *repository.CourseStudentRepository
}

func NewAttendanceHandler(
	attendanceRepo *repository.AttendanceRepository,
	lessonRepo *repository.LessonRepository,
	courseRepo *repository.CourseRepository,
	userRepo *repository.UserRepository,
	courseStudentRepo *repository.CourseStudentRepository,
) *AttendanceHandler {
	return &AttendanceHandler{
		attendanceRepo:    attendanceRepo,
		lessonRepo:        lessonRepo,
		courseRepo:        courseRepo,
		userRepo:          userRepo,
		courseStudentRepo: courseStudentRepo,
	}
}

func isValidAttendanceStatus(status string) bool {
	switch status {
	case "present", "absent", "late":
		return true
	default:
		return false
	}
}

// CreateAttendance godoc
// @Summary Create attendance
// @Description Teacher/Admin sets attendance for a student in a lesson
// @Tags attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Lesson ID"
// @Param input body dto.CreateAttendanceInput true "Create attendance input"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /lessons/{id}/attendance [post]
func (h *AttendanceHandler) Create(c *gin.Context) {
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

	var input dto.CreateAttendanceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isValidAttendanceStatus(input.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "status must be one of: present, absent, late",
		})
		return
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
			"error": "you can only manage attendance in your own courses",
		})
		return
	}

	student, err := h.userRepo.GetByID(ctx, input.StudentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get student"})
		return
	}

	if student.Role != "student" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "attendance can only be set for users with role student",
		})
		return
	}

	enrolled, err := h.courseStudentRepo.IsStudentEnrolled(ctx, lesson.CourseID, input.StudentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check course enrollment"})
		return
	}

	if !enrolled {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "student is not enrolled in this course",
		})
		return
	}

	attendance, err := h.attendanceRepo.Create(ctx, repository.CreateAttendanceParams{
		LessonID:  lessonID,
		StudentID: input.StudentID,
		Status:    input.Status,
		Comment:   input.Comment,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "attendance for this student and lesson already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create attendance"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": attendance})
}

// GetAttendanceByLessonID godoc
// @Summary Get attendance by lesson
// @Description Teacher/Admin gets attendance list for a lesson
// @Tags attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Lesson ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /lessons/{id}/attendance [get]
func (h *AttendanceHandler) GetByLessonID(c *gin.Context) {
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
			"error": "you can only view attendance in your own courses",
		})
		return
	}

	items, err := h.attendanceRepo.GetByLessonID(ctx, lessonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get attendance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// UpdateAttendance godoc
// @Summary Update attendance
// @Description Teacher/Admin updates attendance by id
// @Tags attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Attendance ID"
// @Param input body dto.UpdateAttendanceInput true "Update attendance input"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /attendance/{id} [patch]
func (h *AttendanceHandler) UpdateByID(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attendance id"})
		return
	}

	var input dto.UpdateAttendanceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isValidAttendanceStatus(input.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "status must be one of: present, absent, late",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	existingAttendance, err := h.attendanceRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "attendance not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get attendance"})
		return
	}

	lesson, err := h.lessonRepo.GetByID(ctx, existingAttendance.LessonID)
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
			"error": "you can only update attendance in your own courses",
		})
		return
	}

	updatedAttendance, err := h.attendanceRepo.UpdateByID(ctx, id, repository.UpdateAttendanceParams{
		Status:  input.Status,
		Comment: input.Comment,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "attendance not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update attendance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updatedAttendance})
}