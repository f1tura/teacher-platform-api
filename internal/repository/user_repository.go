package repository

import (
	"context"
	"errors"

	"teacher-platform/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

type UpdateUserParams struct {
	Name     string
	Email    string
	Role     string
	IsActive bool
}

type CreateUserParams struct {
	Name         string
	Email        string
	PasswordHash string
	Role         string
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, params CreateUserParams) (*models.User, error) {
	query := `
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, password_hash, role, is_active, created_at, updated_at
	`

	var user models.User

	err := r.db.QueryRow(
		ctx,
		query,
		params.Name,
		params.Email,
		params.PasswordHash,
		params.Role,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetAll(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var user models.User

		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) DeleteByID(ctx context.Context, id int64) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *UserRepository) UpdateByID(ctx context.Context, id int64, params UpdateUserParams) (*models.User, error) {
	query := `
		UPDATE users
		SET name = $2,
		    email = $3,
		    role = $4,
		    is_active = $5,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, email, password_hash, role, is_active, created_at, updated_at
	`

	var user models.User

	err := r.db.QueryRow(ctx, query, id, params.Name, params.Email, params.Role, params.IsActive).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user models.User

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &user, nil
}