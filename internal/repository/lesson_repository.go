package repository

import (
	"context"
	"errors"

	"teacher-platform/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LessonRepository struct {
	db *pgxpool.Pool
}

type UpdateLessonParams struct {
	Title       string
	Description string
	LessonDate  string
	Status      string
}

type CreateLessonParams struct {
	CourseID    int64
	Title       string
	Description string
	LessonDate  string
	Status      string
}

func NewLessonRepository(db *pgxpool.Pool) *LessonRepository {
	return &LessonRepository{db: db}
}

func (r *LessonRepository) Create(ctx context.Context, params CreateLessonParams) (*models.Lesson, error) {
	query := `
		INSERT INTO lessons (course_id, title, description, lesson_date, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, course_id, title, description, lesson_date, status, created_at, updated_at
	`

	var lesson models.Lesson

	err := r.db.QueryRow(
		ctx,
		query,
		params.CourseID,
		params.Title,
		params.Description,
		params.LessonDate,
		params.Status,
	).Scan(
		&lesson.ID,
		&lesson.CourseID,
		&lesson.Title,
		&lesson.Description,
		&lesson.LessonDate,
		&lesson.Status,
		&lesson.CreatedAt,
		&lesson.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &lesson, nil
}

func (r *LessonRepository) GetByCourseID(ctx context.Context, courseID int64) ([]models.Lesson, error) {
	query := `
		SELECT id, course_id, title, description, lesson_date, status, created_at, updated_at
		FROM lessons
		WHERE course_id = $1
		ORDER BY lesson_date ASC, id ASC
	`

	rows, err := r.db.Query(ctx, query, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lessons []models.Lesson

	for rows.Next() {
		var lesson models.Lesson

		err := rows.Scan(
			&lesson.ID,
			&lesson.CourseID,
			&lesson.Title,
			&lesson.Description,
			&lesson.LessonDate,
			&lesson.Status,
			&lesson.CreatedAt,
			&lesson.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		lessons = append(lessons, lesson)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lessons, nil
}

func (r *LessonRepository) GetByID(ctx context.Context, id int64) (*models.Lesson, error) {
	query := `
		SELECT id, course_id, title, description, lesson_date, status, created_at, updated_at
		FROM lessons
		WHERE id = $1
	`

	var lesson models.Lesson

	err := r.db.QueryRow(ctx, query, id).Scan(
		&lesson.ID,
		&lesson.CourseID,
		&lesson.Title,
		&lesson.Description,
		&lesson.LessonDate,
		&lesson.Status,
		&lesson.CreatedAt,
		&lesson.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &lesson, nil
}

func (r *LessonRepository) UpdateByID(ctx context.Context, id int64, params UpdateLessonParams) (*models.Lesson, error) {
	query := `
		UPDATE lessons
		SET title = $2,
		    description = $3,
		    lesson_date = $4,
		    status = $5,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, course_id, title, description, lesson_date, status, created_at, updated_at
	`

	var lesson models.Lesson

	err := r.db.QueryRow(
		ctx,
		query,
		id,
		params.Title,
		params.Description,
		params.LessonDate,
		params.Status,
	).Scan(
		&lesson.ID,
		&lesson.CourseID,
		&lesson.Title,
		&lesson.Description,
		&lesson.LessonDate,
		&lesson.Status,
		&lesson.CreatedAt,
		&lesson.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &lesson, nil
}

func (r *LessonRepository) DeleteByID(ctx context.Context, id int64) error {
	query := `
		DELETE FROM lessons
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