package stores

import (
	"context"
	"database/sql"
	"db-client/internal/db"
	"db-client/internal/models"
	"encoding/json"
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

type ImageResult struct {
	Data []byte
	URL  string
}

func (s *SubmissionStore) Create(ctx context.Context, userID uuid.UUID, input models.CreateSubmissionRequest) error {
    
    tx, err := s.db.DB().BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    var imageBytes []byte

    if input.Category == "unit" {
        var unitPayload models.CreateUnitsPayload
        if err := json.Unmarshal(input.Payload, &unitPayload); err != nil {
            return fmt.Errorf("invalid unit payload %w", err)
        }
        
        if unitPayload.Image != nil {
            imageBytes, err = decodeBase64Image(*unitPayload.Image)
            if err != nil {
                return fmt.Errorf("invalid image: %w", err)
            }

            // Set the image field to nil so that it is not stored in the submission table
            unitPayload.Image = nil

            // Re-marshal the payload without the image
            stripped, err := json.Marshal(unitPayload)
            if err != nil {
                return err
            }
            input.Payload = stripped
        }
    }

    hash, err := hashPayload(input.Payload)
    if err != nil {
        return err
    }

    // Insert submission and fetch the generated id
    var submissionID uuid.UUID
    err = tx.QueryRowContext(ctx,
        `INSERT INTO submission (submitted_by, category, status, payload, payload_hash)
		VALUES ($1, $2, 'pending', $3, $4)
        RETURNING id`,
    userID, input.Category, input.Payload, hash).Scan(&submissionID)
    if err != nil {
        return fmt.Errorf("Submissionstore.Create: %w", err)
    }

    // Insert the image if one was provided
    if imageBytes != nil {
        _, err := tx.ExecContext(ctx, 
            `INSERT INTO submission_image (submission_id, data)
            VALUES ($1, $2)`,
        submissionID, imageBytes)
        if err != nil {
            return fmt.Errorf("Submissionstore.Create: %w", err)
        }
    }
    
	return tx.Commit()
}

func (s *SubmissionStore) CreateWithImageURL(ctx context.Context, userID uuid.UUID, input models.CreateSubmissionRequest) error {
	tx, err := s.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var unitPayload models.CreateUnitsPayload
	if err := json.Unmarshal(input.Payload, &unitPayload); err != nil {
		return fmt.Errorf("invalid unit payload: %w", err)
	}

	if unitPayload.ImageURL == nil {
		return fmt.Errorf("imageUrl is required")
	}

	if err := validateStorageURL(*unitPayload.ImageURL); err != nil {
		return fmt.Errorf("invalid imageUrl: %w", err)
	}

	imageURL := *unitPayload.ImageURL

	unitPayload.ImageURL = nil
	stripped, err := json.Marshal(unitPayload)
	if err != nil {
		return err
	}
	input.Payload = stripped

	hash, err := hashPayload(input.Payload)
	if err != nil {
		return err
	}

	var submissionID uuid.UUID
	err = tx.QueryRowContext(ctx,
		`INSERT INTO submission (submitted_by, category, status, payload, payload_hash)
		VALUES ($1, $2, 'pending', $3, $4)
		RETURNING id`,
		userID, input.Category, input.Payload, hash).Scan(&submissionID)
	if err != nil {
		return fmt.Errorf("SubmissionStore.CreateWithImageURL: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO submission_image (submission_id, url) VALUES ($1, $2)`,
		submissionID, imageURL)
	if err != nil {
		return fmt.Errorf("SubmissionStore.CreateWithImageURL: %w", err)
	}

	return tx.Commit()
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

func (s *SubmissionStore) GetImageByID(ctx context.Context, id uuid.UUID) (*ImageResult, error) {
	var data []byte
	var url sql.NullString

	err := s.db.DB().QueryRowContext(ctx, `
        SELECT data, url FROM submission_image WHERE submission_id = $1
    `, id).Scan(&data, &url)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("SubmissionStore.GetImageByID: %w", err)
	}

	return &ImageResult{Data: data, URL: url.String}, nil
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