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

type UserHandler struct {
	repo *repository.UserRepository
}

func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) Create(c *gin.Context) {
	var input dto.CreateUserInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !isValidRole(input.Role) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "role must be one of: admin, teacher, student",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	user, err := h.repo.Create(ctx, repository.CreateUserParams{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: input.PasswordHash,
		Role:         input.Role,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "user with this email already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": user,
	})
}

func (h *UserHandler) GetAll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	users, err := h.repo.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
	})
}

func (h *UserHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	user, err := h.repo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

func isValidRole(role string) bool {
	switch role {
	case "admin", "teacher", "student":
		return true
	default:
		return false
	}
}

func (h *UserHandler) DeleteByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	err = h.repo.DeleteByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete user",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *UserHandler) UpdateByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id",
		})
		return
	}

	var input dto.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !isValidRole(input.Role) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "role must be one of: admin, teacher, student",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	user, err := h.repo.UpdateByID(ctx, id, repository.UpdateUserParams{
		Name:     input.Name,
		Email:    input.Email,
		Role:     input.Role,
		IsActive: *input.IsActive,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
			return
		}

		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "user with this email already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}