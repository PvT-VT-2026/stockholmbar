package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// Tests the venuehandler list validation.
// This handler parses all filter options from the url, and should throw BadRequest if any
// filters are malformed, or time is missing.
func TestVenueHandler_List_validation(t *testing.T) {
	t.Parallel()

	h := &VenueHandler{}

	tests := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{name: "missing time", query: "", wantStatus: http.StatusBadRequest},
		{name: "invalid time format", query: "time=25:99", wantStatus: http.StatusBadRequest},
		{name: "invalid max_price", query: "time=12:00&max_price=abc", wantStatus: http.StatusBadRequest},
		{name: "invalid happy_hour", query: "time=12:00&happy_hour=maybe", wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec := httptest.NewRecorder()
			url := "/database/venues/list"
			if tt.query != "" {
				url += "?" + tt.query
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			h.List(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}


// Attempting to fetch a specific venue with an invalid uuid should result in StatusBadRequest.
func TestVenueHandler_GetByID_invalidUUID(t *testing.T) {
	t.Parallel()

	h := &VenueHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/database/venues/not-a-uuid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "not-a-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
