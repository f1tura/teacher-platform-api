package repository

import (
	"context"
	"errors"

	"teacher-platform/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AttendanceRepository struct {
	db *pgxpool.Pool
}

type CreateAttendanceParams struct {
	LessonID  int64
	StudentID int64
	Status    string
	Comment   string
}

type UpdateAttendanceParams struct {
	Status  string
	Comment string
}

func NewAttendanceRepository(db *pgxpool.Pool) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

func (r *AttendanceRepository) Create(ctx context.Context, params CreateAttendanceParams) (*models.Attendance, error) {
	query := `
		INSERT INTO attendance (lesson_id, student_id, status, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING id, lesson_id, student_id, status, comment, created_at, updated_at
	`

	var attendance models.Attendance

	err := r.db.QueryRow(
		ctx,
		query,
		params.LessonID,
		params.StudentID,
		params.Status,
		params.Comment,
	).Scan(
		&attendance.ID,
		&attendance.LessonID,
		&attendance.StudentID,
		&attendance.Status,
		&attendance.Comment,
		&attendance.CreatedAt,
		&attendance.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &attendance, nil
}

func (r *AttendanceRepository) GetByLessonID(ctx context.Context, lessonID int64) ([]models.Attendance, error) {
	query := `
		SELECT id, lesson_id, student_id, status, comment, created_at, updated_at
		FROM attendance
		WHERE lesson_id = $1
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, query, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Attendance

	for rows.Next() {
		var attendance models.Attendance

		err := rows.Scan(
			&attendance.ID,
			&attendance.LessonID,
			&attendance.StudentID,
			&attendance.Status,
			&attendance.Comment,
			&attendance.CreatedAt,
			&attendance.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, attendance)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *AttendanceRepository) GetByID(ctx context.Context, id int64) (*models.Attendance, error) {
	query := `
		SELECT id, lesson_id, student_id, status, comment, created_at, updated_at
		FROM attendance
		WHERE id = $1
	`

	var attendance models.Attendance

	err := r.db.QueryRow(ctx, query, id).Scan(
		&attendance.ID,
		&attendance.LessonID,
		&attendance.StudentID,
		&attendance.Status,
		&attendance.Comment,
		&attendance.CreatedAt,
		&attendance.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &attendance, nil
}

func (r *AttendanceRepository) UpdateByID(ctx context.Context, id int64, params UpdateAttendanceParams) (*models.Attendance, error) {
	query := `
		UPDATE attendance
		SET status = $2,
		    comment = $3,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, lesson_id, student_id, status, comment, created_at, updated_at
	`

	var attendance models.Attendance

	err := r.db.QueryRow(ctx, query, id, params.Status, params.Comment).Scan(
		&attendance.ID,
		&attendance.LessonID,
		&attendance.StudentID,
		&attendance.Status,
		&attendance.Comment,
		&attendance.CreatedAt,
		&attendance.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &attendance, nil
}