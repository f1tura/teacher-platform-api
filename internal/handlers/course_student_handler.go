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

type CourseStudentHandler struct {
	courseRepo        *repository.CourseRepository
	userRepo          *repository.UserRepository
	courseStudentRepo *repository.CourseStudentRepository
}

func NewCourseStudentHandler(
	courseRepo *repository.CourseRepository,
	userRepo *repository.UserRepository,
	courseStudentRepo *repository.CourseStudentRepository,
) *CourseStudentHandler {
	return &CourseStudentHandler{
		courseRepo:        courseRepo,
		userRepo:          userRepo,
		courseStudentRepo: courseStudentRepo,
	}
}

func (h *CourseStudentHandler) AddStudent(c *gin.Context) {
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

	var input dto.AddStudentToCourseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
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
			"error": "you can only manage students in your own courses",
		})
		return
	}

	student, err := h.userRepo.GetByID(ctx, input.StudentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "student not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get student",
		})
		return
	}

	if student.Role != "student" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "only users with role student can be added to a course",
		})
		return
	}

	err = h.courseStudentRepo.AddStudent(ctx, courseID, input.StudentID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "student is already enrolled in this course",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to add student to course",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "student added to course",
	})
}

func (h *CourseStudentHandler) GetStudents(c *gin.Context) {
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
			"error": "you can only view students in your own courses",
		})
		return
	}

	students, err := h.courseStudentRepo.GetStudentsByCourseID(ctx, courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get course students",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": students,
	})
}

func (h *CourseStudentHandler) RemoveStudent(c *gin.Context) {
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

	studentIDParam := c.Param("studentId")
	studentID, err := strconv.ParseInt(studentIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid student id",
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
			"error": "you can only manage students in your own courses",
		})
		return
	}

	err = h.courseStudentRepo.RemoveStudent(ctx, courseID, studentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "student is not enrolled in this course",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to remove student from course",
		})
		return
	}

	c.Status(http.StatusNoContent)
}