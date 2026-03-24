package repository

import (
	"context"
	"errors"

	"teacher-platform/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubmissionRepository struct {
	db *pgxpool.Pool
}

type UpdateSubmissionParams struct {
	AnswerText string
	FileURL    string
}

type ReviewSubmissionParams struct {
	Status string
	Score  int
}

type CreateSubmissionParams struct {
	AssignmentID int64
	StudentID    int64
	AnswerText   string
	FileURL      string
}

func NewSubmissionRepository(db *pgxpool.Pool) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func (r *SubmissionRepository) Create(ctx context.Context, params CreateSubmissionParams) (*models.Submission, error) {
	query := `
		INSERT INTO submissions (assignment_id, student_id, answer_text, file_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, assignment_id, student_id, answer_text, file_url, status, score, submitted_at, checked_at, created_at, updated_at
	`

	var submission models.Submission

	err := r.db.QueryRow(
		ctx,
		query,
		params.AssignmentID,
		params.StudentID,
		params.AnswerText,
		params.FileURL,
	).Scan(
		&submission.ID,
		&submission.AssignmentID,
		&submission.StudentID,
		&submission.AnswerText,
		&submission.FileURL,
		&submission.Status,
		&submission.Score,
		&submission.SubmittedAt,
		&submission.CheckedAt,
		&submission.CreatedAt,
		&submission.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &submission, nil
}

func (r *SubmissionRepository) GetByAssignmentID(ctx context.Context, assignmentID int64) ([]models.Submission, error) {
	query := `
		SELECT id, assignment_id, student_id, answer_text, file_url, status, score, submitted_at, checked_at, created_at, updated_at
		FROM submissions
		WHERE assignment_id = $1
		ORDER BY submitted_at DESC, id DESC
	`

	rows, err := r.db.Query(ctx, query, assignmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []models.Submission

	for rows.Next() {
		var submission models.Submission

		err := rows.Scan(
			&submission.ID,
			&submission.AssignmentID,
			&submission.StudentID,
			&submission.AnswerText,
			&submission.FileURL,
			&submission.Status,
			&submission.Score,
			&submission.SubmittedAt,
			&submission.CheckedAt,
			&submission.CreatedAt,
			&submission.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		submissions = append(submissions, submission)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return submissions, nil
}

func (r *SubmissionRepository) GetByID(ctx context.Context, id int64) (*models.Submission, error) {
	query := `
		SELECT id, assignment_id, student_id, answer_text, file_url, status, score, submitted_at, checked_at, created_at, updated_at
		FROM submissions
		WHERE id = $1
	`

	var submission models.Submission

	err := r.db.QueryRow(ctx, query, id).Scan(
		&submission.ID,
		&submission.AssignmentID,
		&submission.StudentID,
		&submission.AnswerText,
		&submission.FileURL,
		&submission.Status,
		&submission.Score,
		&submission.SubmittedAt,
		&submission.CheckedAt,
		&submission.CreatedAt,
		&submission.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &submission, nil
}

func (r *SubmissionRepository) ReviewByID(ctx context.Context, id int64, params ReviewSubmissionParams) (*models.Submission, error) {
	query := `
		UPDATE submissions
		SET status = $2,
		    score = $3,
		    checked_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, assignment_id, student_id, answer_text, file_url, status, score, submitted_at, checked_at, created_at, updated_at
	`

	var submission models.Submission

	err := r.db.QueryRow(ctx, query, id, params.Status, params.Score).Scan(
		&submission.ID,
		&submission.AssignmentID,
		&submission.StudentID,
		&submission.AnswerText,
		&submission.FileURL,
		&submission.Status,
		&submission.Score,
		&submission.SubmittedAt,
		&submission.CheckedAt,
		&submission.CreatedAt,
		&submission.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &submission, nil
}

func (r *SubmissionRepository) UpdateByID(ctx context.Context, id int64, params UpdateSubmissionParams) (*models.Submission, error) {
	query := `
		UPDATE submissions
		SET answer_text = $2,
		    file_url = $3,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, assignment_id, student_id, answer_text, file_url, status, score, submitted_at, checked_at, created_at, updated_at
	`

	var submission models.Submission

	err := r.db.QueryRow(ctx, query, id, params.AnswerText, params.FileURL).Scan(
		&submission.ID,
		&submission.AssignmentID,
		&submission.StudentID,
		&submission.AnswerText,
		&submission.FileURL,
		&submission.Status,
		&submission.Score,
		&submission.SubmittedAt,
		&submission.CheckedAt,
		&submission.CreatedAt,
		&submission.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &submission, nil
}

func (r *SubmissionRepository) DeleteByID(ctx context.Context, id int64) error {
	query := `
		DELETE FROM submissions
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