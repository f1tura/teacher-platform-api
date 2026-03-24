
package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"teacher-platform/internal/auth"
	"teacher-platform/internal/config"
	"teacher-platform/internal/dto"
	"teacher-platform/internal/repository"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	repo      *repository.UserRepository
	jwtSecret string
}

func NewAuthHandler(repo *repository.UserRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		repo:      repo,
		jwtSecret: cfg.JWTSecret,
	}
}

// Register godoc
// @Summary Register user
// @Description Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param input body dto.RegisterInput true "Register input"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input dto.RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !isValidAuthRole(input.Role) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "role must be one of: admin, teacher, student",
		})
		return
	}

	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to hash password",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	user, err := h.repo.Create(ctx, repository.CreateUserParams{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: passwordHash,
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
			"error": "failed to register user",
		})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Role, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"user":  user,
			"token": token,
		},
	})
}

func isValidAuthRole(role string) bool {
	switch role {
	case "admin", "teacher", "student":
		return true
	default:
		return false
	}
}

// Login godoc
// @Summary Login user
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param input body dto.LoginInput true "Login input"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input dto.LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	user, err := h.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid email or password",
		})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user is inactive",
		})
		return
	}

	if !auth.CheckPasswordHash(input.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid email or password",
		})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Role, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"user":  user,
			"token": token,
		},
	})
}

// Me godoc
// @Summary Get current user
// @Description Get current authorized user by token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	user, err := h.repo.GetByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}