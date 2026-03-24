package dto

type CreateUserInput struct {
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	PasswordHash string `json:"password_hash" binding:"required"`
	Role         string `json:"role" binding:"required"`
}

type UpdateUserInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role" binding:"required"`
	IsActive *bool  `json:"is_active" binding:"required"`
}