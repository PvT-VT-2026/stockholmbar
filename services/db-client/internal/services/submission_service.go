package services

import (
	"context"
	"db-client/internal/clients"
	"db-client/internal/models"
	"db-client/internal/stores"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SubmissionService struct {
	submissionStore *stores.SubmissionStore
	unitStore       *stores.UnitStore
	venueStore      *stores.VenueStore
	placesClient    *clients.PlacesClient
}

func NewSubmissionService(sub *stores.SubmissionStore, unit *stores.UnitStore, venue *stores.VenueStore, places *clients.PlacesClient) *SubmissionService {
	return &SubmissionService{submissionStore: sub, unitStore: unit, venueStore: venue, placesClient: places}
}

// Before a submission is passed to the submissionStore in order to create a submission entry
// in the database, the submission service tries to parse the json payload into a 
// request struct depending on the category. 
// Ex. If the submission category is "unit", that means the json payload should be able to 
// be unmarshalled into a CreateUnitsPayload.
// This means all submissions that reach the admin panel for verification are valid types,
// and are ready for insertion.
func (s *SubmissionService) CreateSubmission(ctx context.Context, userID uuid.UUID, input models.CreateSubmissionRequest) error {
	
	switch input.Category {
	case "unit":
		var unitPayload models.CreateUnitsPayload
		if err := json.Unmarshal(input.Payload, &unitPayload); err != nil {
			return fmt.Errorf("bad request: %w", err)
		}
		if unitPayload.ImageURL != nil {
			return s.submissionStore.CreateWithImageURL(ctx, userID, input)
		}
		return s.submissionStore.Create(ctx, userID, input)
	case "venue":
		var venuePayload *models.CreateVenuePayload
		if err := json.Unmarshal(input.Payload, &venuePayload); err != nil {
			return fmt.Errorf("bad request: %w", err)
		}
		return s.submissionStore.Create(ctx, userID, input)
	default:
		return fmt.Errorf("unknown category: %s", input.Category)
	}
}

func (s *SubmissionService) ListSubmissions(ctx context.Context, status string) (*models.ListSubmissionsResponse, error) {
	return s.submissionStore.List(ctx, status)
}

func (s *SubmissionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Submission, error) {
	return s.submissionStore.GetByID(ctx, id)
}

func (s *SubmissionService) GetImageByID(ctx context.Context, id uuid.UUID) (*stores.ImageResult, error) {
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
		venueID, err := s.venueStore.Create(ctx, &venuePayload)
		if err != nil {
			return err
		}
		if s.placesClient != nil {
			if err := s.enrichWithBusinessHours(ctx, venueID, &venuePayload); err != nil {
				log.Printf("SubmissionService.Accept: could not fetch business hours for venue %s: %v", venueID, err)
			}
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

func (s *SubmissionService) enrichWithBusinessHours(ctx context.Context, venueID uuid.UUID, venue *models.CreateVenuePayload) error {
	results, err := s.placesClient.FindPlace(venue.Name + ", " + venue.City)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		return fmt.Errorf("no Google Places results for %q", venue.Name)
	}

	placeInfo, err := s.placesClient.GetPlaceInfo(results[0].ID)
	if err != nil {
		return err
	}
	if len(placeInfo.OpeningHours) == 0 {
		return fmt.Errorf("no opening hours returned for place %s", results[0].ID)
	}

	hours := parseOpeningHours(placeInfo.OpeningHours)
	if len(hours) == 0 {
		return fmt.Errorf("failed to parse opening hours for place %s, raw strings: %v", results[0].ID, placeInfo.OpeningHours)
	}

	return s.venueStore.CreateBusinessHours(ctx, venueID, hours)
}

var dayNameToWeekday = map[string]int16{
	"Monday":    1,
	"Tuesday":   2,
	"Wednesday": 3,
	"Thursday":  4,
	"Friday":    5,
	"Saturday":  6,
	"Sunday":    0,
	"Måndag":    1,
	"Tisdag":    2,
	"Onsdag":    3,
	"Torsdag":   4,
	"Fredag":    5,
	"Lördag":    6,
	"Söndag":    0,
}

func parseOpeningHours(descriptions []string) []models.BusinessHours {
	var hours []models.BusinessHours
	for _, desc := range descriptions {
		parts := strings.SplitN(desc, ": ", 2)
		if len(parts) != 2 {
			log.Printf("parseOpeningHours: no ': ' separator in %q", desc)
			continue
		}
		dayOfWeek, ok := dayNameToWeekday[parts[0]]
		if !ok {
			log.Printf("parseOpeningHours: unrecognized day name %q in %q", parts[0], desc)
			continue
		}

		h := models.BusinessHours{DayOfWeek: dayOfWeek}
		timeRange := parts[1]

		switch timeRange {
		case "Closed":
			h.IsClosed = true
		case "Open 24 hours":
			open, close := "00:00", "23:59"
			h.OpenTime = &open
			h.CloseTime = &close
		default:
			// en dash (U+2013) is the separator Google Places uses
			timeParts := strings.SplitN(timeRange, "–", 2)
			if len(timeParts) != 2 {
				log.Printf("parseOpeningHours: no en-dash separator in time range %q (bytes: %x)", timeRange, []byte(timeRange))
				continue
			}
			openTime, err := parseTime(strings.TrimSpace(timeParts[0]))
			if err != nil {
				log.Printf("parseOpeningHours: open time parse error in %q: %v", desc, err)
				continue
			}
			closeTime, err := parseTime(strings.TrimSpace(timeParts[1]))
			if err != nil {
				log.Printf("parseOpeningHours: close time parse error in %q: %v", desc, err)
				continue
			}
			h.OpenTime = &openTime
			h.CloseTime = &closeTime
		}

		hours = append(hours, h)
	}
	return hours
}

func parseTime(s string) (string, error) {
	if t, err := time.Parse("3:04 PM", s); err == nil {
		return t.Format("15:04"), nil
	}
	if t, err := time.Parse("15:04", s); err == nil {
		return t.Format("15:04"), nil
	}
	return "", fmt.Errorf("unrecognized time format: %q", s)
}