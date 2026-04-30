package stores

import (
	"context"
	"database/sql"
	"db-client/internal/db"
	"db-client/internal/models"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type SubmissionStore struct {
	db *db.DBClient
}

func NewSubmissionStore(db *db.DBClient) *SubmissionStore {
	return &SubmissionStore{db: db}
}

func (s *SubmissionStore) Create(ctx context.Context, userID uuid.UUID, input models.CreateSubmissionRequest) error {
	_, err := s.db.DB().ExecContext(ctx, `
		INSERT INTO submission (submitted_by, category, status, payload)
		VALUES ($1, $2, 'pending', $3)
	`, 
	userID, input.Category, input.Payload)
	if err != nil {
		return fmt.Errorf("SubmissionStore.Create: %w", err)
	}

	return nil
}

func (s *SubmissionStore) List(ctx context.Context, status string) (*models.ListSubmissionsResponse, error) {
	query := `
	SELECT id, submitted_by, category, status, reviewed_at, created_at
	FROM submission
	WHERE deleted_at IS NULL	
	`

   	args := []any{}
    if status != "" {
        query += " AND status = $1"
        args = append(args, status)
    }	

	rows, err := s.db.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("SubmissionStore.List: %w", err)
	}
	defer rows.Close()

	var submissions models.ListSubmissionsResponse
    for rows.Next() {
        var s models.SubmissionListItem
        if err := rows.Scan(&s.ID, &s.SubmittedBy, &s.Category, &s.Status, &s.ReviewedAt, &s.CreatedAt); err != nil {
            return nil, fmt.Errorf("SubmissionStore.List: scan: %w", err)
        }
        submissions.Submissions = append(submissions.Submissions, &s)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("SubmissionStore.List: rows: %w", err)
    }

    return &submissions, nil
}

func (s *SubmissionStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Submission, error) {
    var submission models.Submission
    
	err := s.db.DB().QueryRowContext(ctx, `
        SELECT id, submitted_by, category, status, payload, reviewed_at, created_at
        FROM submission
        WHERE id = $1 AND deleted_at IS NULL
    `, id).Scan(
        &submission.ID,
        &submission.SubmittedBy,
        &submission.Category,
        &submission.Status,
        &submission.Payload,
        &submission.ReviewedAt,
        &submission.CreatedAt,
    )
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("SubmissionStore.GetByID: %w", err)
    }

    return &submission, nil
}

func (s *SubmissionStore) GetOldestPending(ctx context.Context) (*models.Submission, error) {
    var submission models.Submission
    
	err := s.db.DB().QueryRowContext(ctx, `
        SELECT id, submitted_by, category, status, payload, reviewed_at, created_at
        FROM submission
        WHERE status = 'pending' AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT 1
    `).Scan(
        &submission.ID,
        &submission.SubmittedBy,
        &submission.Category,
        &submission.Status,
        &submission.Payload,
        &submission.ReviewedAt,
        &submission.CreatedAt,
    )
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("SubmissionStore.GetOldestPending: %w", err)
    }

    return &submission, nil
}


func (s *SubmissionStore) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := s.db.DB().ExecContext(ctx, `
        UPDATE submission SET status = $1, reviewed_at = NOW() WHERE id = $2
    `, status, id)
    if err != nil {
        return fmt.Errorf("SubmissionStore.UpdateStatus: %w", err)
    }

    return nil
}