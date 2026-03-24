package repository

import (
	"context"
	"errors"

	"teacher-platform/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CourseStudentRepository struct {
	db *pgxpool.Pool
}

func NewCourseStudentRepository(db *pgxpool.Pool) *CourseStudentRepository {
	return &CourseStudentRepository{db: db}
}

func (r *CourseStudentRepository) AddStudent(ctx context.Context, courseID, studentID int64) error {
	query := `
		INSERT INTO course_students (course_id, student_id)
		VALUES ($1, $2)
	`

	_, err := r.db.Exec(ctx, query, courseID, studentID)
	return err
}

func (r *CourseStudentRepository) GetStudentsByCourseID(ctx context.Context, courseID int64) ([]models.User, error) {
	query := `
		SELECT u.id, u.name, u.email, u.password_hash, u.role, u.is_active, u.created_at, u.updated_at
		FROM course_students cs
		JOIN users u ON u.id = cs.student_id
		WHERE cs.course_id = $1
		ORDER BY u.id ASC
	`

	rows, err := r.db.Query(ctx, query, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []models.User

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

		students = append(students, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return students, nil
}

func (r *CourseStudentRepository) RemoveStudent(ctx context.Context, courseID, studentID int64) error {
	query := `
		DELETE FROM course_students
		WHERE course_id = $1 AND student_id = $2
	`

	cmdTag, err := r.db.Exec(ctx, query, courseID, studentID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func IsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func (r *CourseStudentRepository) IsStudentEnrolled(ctx context.Context, courseID, studentID int64) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM course_students
			WHERE course_id = $1 AND student_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, courseID, studentID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *CourseStudentRepository) IsStudentInTeacherCourses(ctx context.Context, studentID, teacherID int64) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM course_students cs
			JOIN courses c ON c.id = cs.course_id
			WHERE cs.student_id = $1
			  AND c.teacher_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, studentID, teacherID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}