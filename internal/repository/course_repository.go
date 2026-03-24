package repository

import (
	"context"
	"errors"

	"teacher-platform/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CourseRepository struct {
	db *pgxpool.Pool
}

type UpdateCourseParams struct {
	Title       string
	Description string
	IsActive    bool
}

type CreateCourseParams struct {
	TeacherID   int64
	Title       string
	Description string
}

func NewCourseRepository(db *pgxpool.Pool) *CourseRepository {
	return &CourseRepository{db: db}
}

func (r *CourseRepository) Create(ctx context.Context, params CreateCourseParams) (*models.Course, error) {
	query := `
		INSERT INTO courses (teacher_id, title, description)
		VALUES ($1, $2, $3)
		RETURNING id, teacher_id, title, description, is_active, created_at, updated_at
	`

	var course models.Course

	err := r.db.QueryRow(ctx, query, params.TeacherID, params.Title, params.Description).Scan(
		&course.ID,
		&course.TeacherID,
		&course.Title,
		&course.Description,
		&course.IsActive,
		&course.CreatedAt,
		&course.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &course, nil
}

func (r *CourseRepository) GetAll(ctx context.Context) ([]models.Course, error) {
	query := `
		SELECT id, teacher_id, title, description, is_active, created_at, updated_at
		FROM courses
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []models.Course

	for rows.Next() {
		var course models.Course

		err := rows.Scan(
			&course.ID,
			&course.TeacherID,
			&course.Title,
			&course.Description,
			&course.IsActive,
			&course.CreatedAt,
			&course.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		courses = append(courses, course)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return courses, nil
}

func (r *CourseRepository) GetByID(ctx context.Context, id int64) (*models.Course, error) {
	query := `
		SELECT id, teacher_id, title, description, is_active, created_at, updated_at
		FROM courses
		WHERE id = $1
	`

	var course models.Course

	err := r.db.QueryRow(ctx, query, id).Scan(
		&course.ID,
		&course.TeacherID,
		&course.Title,
		&course.Description,
		&course.IsActive,
		&course.CreatedAt,
		&course.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &course, nil
}

func (r *CourseRepository) UpdateByID(ctx context.Context, id int64, params UpdateCourseParams) (*models.Course, error) {
	query := `
		UPDATE courses
		SET title = $2,
		    description = $3,
		    is_active = $4,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, teacher_id, title, description, is_active, created_at, updated_at
	`

	var course models.Course

	err := r.db.QueryRow(ctx, query, id, params.Title, params.Description, params.IsActive).Scan(
		&course.ID,
		&course.TeacherID,
		&course.Title,
		&course.Description,
		&course.IsActive,
		&course.CreatedAt,
		&course.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &course, nil
}

func (r *CourseRepository) DeleteByID(ctx context.Context, id int64) error {
	query := `
		DELETE FROM courses
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