package services

import (
	"context"
	"db-client/internal/models"
	"db-client/internal/stores"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type SubmissionService struct {
	submissionStore *stores.SubmissionStore
	unitStore       *stores.UnitStore
	venueStore      *stores.VenueStore
}

func NewSubmissionService(sub *stores.SubmissionStore, unit *stores.UnitStore, venue *stores.VenueStore) *SubmissionService {
	return &SubmissionService{submissionStore: sub, unitStore: unit, venueStore: venue}
}

// Before a submission is passed to the submissionStore in order to create a submission entry
// in the database, the submission service tries to parse the json payload into a 
// request struct depending on the category. 
// Ex. If the submission category is "unit", that means the json payload should be able to 
// be unmarshalled into a CreateUnitsPayload.
// This means all submissions that reach the admin panel for verification are valid types,
// and are ready for insertion.
func (s *SubmissionService) CreateSubmission(ctx context.Context, userID uuid.UUID, input models.CreateSubmissionRequest) error {
	
	switch (input.Category) {
		case "unit": 
			var unitPayload *models.CreateUnitsPayload
			if err := json.Unmarshal(input.Payload, &unitPayload); err != nil {
				return fmt.Errorf("bad request: %w", err)
			}
		case "venue":
			var venuePayload *models.CreateVenuePayload
			if err := json.Unmarshal(input.Payload, &venuePayload); err != nil {
				return fmt.Errorf("bad request: %w", err)
			}
		default:
			return fmt.Errorf("unknown category: %s", input.Category)
	}
	
	return s.submissionStore.Create(ctx, userID, input)
}

func (s *SubmissionService) ListSubmissions(ctx context.Context, status string) (*models.ListSubmissionsResponse, error) {
	return s.submissionStore.List(ctx, status)
}

func (s *SubmissionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Submission, error) {
	return s.submissionStore.GetByID(ctx, id)
}

func (s *SubmissionService) GetImageByID(ctx context.Context, id uuid.UUID) ([]byte, error) {
	return s.submissionStore.GetImageByID(ctx, id)
}

func (s *SubmissionService) GetOldestPending(ctx context.Context) (*models.Submission, error) {
	return s.submissionStore.GetOldestPending(ctx)
}

func (s *SubmissionService) Accept(ctx context.Context, submissionID uuid.UUID) error {
    submission, err := s.submissionStore.GetByID(ctx, submissionID)
	if err != nil {
		return err
	}

	// Accepting an already accepted submission has no effect
	if submission.Status == "accepted" {
		return nil
	}

    switch submission.Category {
    case "unit":
		var unitPayload models.CreateUnitsPayload
		if err := json.Unmarshal(submission.Payload, &unitPayload); err != nil {
			return fmt.Errorf("unable to parse unit payload: %w", err)
		}
        if err := s.unitStore.Create(ctx, &unitPayload); err != nil {
			return err
		}
	case "venue":
		var venuePayload models.CreateVenuePayload
		if err := json.Unmarshal(submission.Payload, &venuePayload); err != nil {
			return fmt.Errorf("unable to parse venue payload: %w", err)
		}
        if err := s.venueStore.Create(ctx, &venuePayload); err != nil {
			return err
		}
	
	default:
		return fmt.Errorf("unknown submission category: %s", submission.Category)
    }

	return s.submissionStore.UpdateStatus(ctx, submissionID, "accepted")
}

func (s *SubmissionService) Reject(ctx context.Context, submissionID uuid.UUID) error {
	submission, err := s.submissionStore.GetByID(ctx, submissionID)
	if err != nil {
		return err
	}

	// Rejecting an already rejected submission has no effect
	if submission.Status == "reject" {
		return nil
	}

	// TODO: 
	// Rejecting a previously accepted submission is a bit tricky,
	// as it requires removal from tables and essentially backtracking the inserts. 
	if submission.Status == "accepted" {
		return fmt.Errorf("Logic for rejecting an accepted submission is not yet implemented")
	}

	return s.submissionStore.UpdateStatus(ctx, submissionID, "rejected")
}

// func (s *SubmissionService) Reject(ctx context.Context, submissionID uuid.UUID) error {
//     // submission, err := s.submissions.GetByID(ctx, submissionID)

//     // switch submission.Category {
//     // case "unit":
//     //     return s.units.CreateUnits(ctx, payloadToCreateUnitsRequest(submission.Payload))
//     // case "venue":
//     //     return s.venues.Create(ctx, ...)
//     // }
// }