package repository

import (
	"context"

	"teacher-platform/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProgressRepository struct {
	db *pgxpool.Pool
}

func NewProgressRepository(db *pgxpool.Pool) *ProgressRepository {
	return &ProgressRepository{db: db}
}

func (r *ProgressRepository) GetStudentProgress(ctx context.Context, studentID int64) (*models.StudentProgress, error) {
	query := `
		WITH student_assignments AS (
			SELECT DISTINCT
				a.id AS assignment_id
			FROM course_students cs
			JOIN lessons l ON l.course_id = cs.course_id
			JOIN assignments a ON a.lesson_id = l.id
			WHERE cs.student_id = $1::BIGINT
			  AND a.is_published = TRUE
		),
		student_submissions AS (
			SELECT
				s.id,
				s.assignment_id,
				s.score,
				s.status
			FROM submissions s
			WHERE s.student_id = $1::BIGINT
		)
		SELECT
			$1::BIGINT AS student_id,
			(SELECT COUNT(*) FROM student_assignments)::INT AS total_assignments,
			(SELECT COUNT(*) FROM student_submissions)::INT AS submitted_assignments,
			(SELECT COUNT(*) FROM student_submissions WHERE status = 'checked')::INT AS checked_assignments,
			COALESCE((SELECT ROUND(AVG(score)::numeric, 2)::float8 FROM student_submissions), 0) AS average_score,
			COALESCE(
				(
					SELECT ROUND(
						CASE
							WHEN COUNT(*) = 0 THEN 0
							ELSE (
								(SELECT COUNT(*) FROM student_submissions)::numeric * 100 / COUNT(*)
							)
						END,
						2
					)::float8
					FROM student_assignments
				),
				0
			) AS completion_rate
	`

	var progress models.StudentProgress

	err := r.db.QueryRow(ctx, query, studentID).Scan(
		&progress.StudentID,
		&progress.TotalAssignments,
		&progress.SubmittedAssignments,
		&progress.CheckedAssignments,
		&progress.AverageScore,
		&progress.CompletionRate,
	)
	if err != nil {
		return nil, err
	}

	return &progress, nil
}