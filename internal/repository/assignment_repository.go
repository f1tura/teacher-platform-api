package repository

import (
	"context"
	"errors"

	"teacher-platform/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AssignmentRepository struct {
	db *pgxpool.Pool
}

type UpdateAssignmentParams struct {
	Title       string
	Description string
	DueDate     string
	MaxScore    int
	IsPublished bool
}


type CreateAssignmentParams struct {
	LessonID    int64
	Title       string
	Description string
	DueDate     string
	MaxScore    int
	IsPublished bool
}

func NewAssignmentRepository(db *pgxpool.Pool) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

func (r *AssignmentRepository) Create(ctx context.Context, params CreateAssignmentParams) (*models.Assignment, error) {
	query := `
		INSERT INTO assignments (lesson_id, title, description, due_date, max_score, is_published)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, lesson_id, title, description, due_date, max_score, is_published, created_at, updated_at
	`

	var assignment models.Assignment

	err := r.db.QueryRow(
		ctx,
		query,
		params.LessonID,
		params.Title,
		params.Description,
		params.DueDate,
		params.MaxScore,
		params.IsPublished,
	).Scan(
		&assignment.ID,
		&assignment.LessonID,
		&assignment.Title,
		&assignment.Description,
		&assignment.DueDate,
		&assignment.MaxScore,
		&assignment.IsPublished,
		&assignment.CreatedAt,
		&assignment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &assignment, nil
}

func (r *AssignmentRepository) GetByLessonID(ctx context.Context, lessonID int64) ([]models.Assignment, error) {
	query := `
		SELECT id, lesson_id, title, description, due_date, max_score, is_published, created_at, updated_at
		FROM assignments
		WHERE lesson_id = $1
		ORDER BY due_date ASC, id ASC
	`

	rows, err := r.db.Query(ctx, query, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []models.Assignment

	for rows.Next() {
		var assignment models.Assignment

		err := rows.Scan(
			&assignment.ID,
			&assignment.LessonID,
			&assignment.Title,
			&assignment.Description,
			&assignment.DueDate,
			&assignment.MaxScore,
			&assignment.IsPublished,
			&assignment.CreatedAt,
			&assignment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return assignments, nil
}

func (r *AssignmentRepository) GetByID(ctx context.Context, id int64) (*models.Assignment, error) {
	query := `
		SELECT id, lesson_id, title, description, due_date, max_score, is_published, created_at, updated_at
		FROM assignments
		WHERE id = $1
	`

	var assignment models.Assignment

	err := r.db.QueryRow(ctx, query, id).Scan(
		&assignment.ID,
		&assignment.LessonID,
		&assignment.Title,
		&assignment.Description,
		&assignment.DueDate,
		&assignment.MaxScore,
		&assignment.IsPublished,
		&assignment.CreatedAt,
		&assignment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &assignment, nil
}

func (r *AssignmentRepository) UpdateByID(ctx context.Context, id int64, params UpdateAssignmentParams) (*models.Assignment, error) {
	query := `
		UPDATE assignments
		SET title = $2,
		    description = $3,
		    due_date = $4,
		    max_score = $5,
		    is_published = $6,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, lesson_id, title, description, due_date, max_score, is_published, created_at, updated_at
	`

	var assignment models.Assignment

	err := r.db.QueryRow(
		ctx,
		query,
		id,
		params.Title,
		params.Description,
		params.DueDate,
		params.MaxScore,
		params.IsPublished,
	).Scan(
		&assignment.ID,
		&assignment.LessonID,
		&assignment.Title,
		&assignment.Description,
		&assignment.DueDate,
		&assignment.MaxScore,
		&assignment.IsPublished,
		&assignment.CreatedAt,
		&assignment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &assignment, nil
}

func (r *AssignmentRepository) DeleteByID(ctx context.Context, id int64) error {
	query := `
		DELETE FROM assignments
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