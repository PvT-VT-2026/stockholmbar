package handlers

import (
	"bytes"
	"context"
	"db-client/internal/middleware"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestSubmissionHandler_CreateSubmission(t *testing.T) {
	t.Parallel()

	h := &SubmissionHandler{}

	// Setup mock auth for all CreateSubmission tests
	secret := []byte("test-secret")
	keyfunc := func(token *jwt.Token) (any, error) { return secret, nil }
	auth := middleware.Auth(keyfunc)

	signToken := func(sub string) string {
		t.Helper()
		claims := jwt.MapClaims{
			"sub": sub,
			"aud": "authenticated",
			"exp": time.Now().Add(time.Hour).Unix(),
		}
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := tok.SignedString(secret)
		if err != nil {
			t.Fatalf("sign token: %v", err)
		}
		return signed
	}

	withAuth := func(req *http.Request, sub string) *http.Request {
		req.Header.Set("Authorization", "Bearer "+signToken(sub))
		return req
	}

	// Trying to create a submission without a user in the context should result in StatusUnauthorized
	t.Run("missing user in context", func(t *testing.T) {
		t.Parallel()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/submission/create", nil)

		h.CreateSubmission(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	// Trying to create a submission without avalid user id in context should result in StatusUnauthorized
	t.Run("invalid user id in context", func(t *testing.T) {
		t.Parallel()
		rec := httptest.NewRecorder()
		req := withAuth(httptest.NewRequest(http.MethodPost, "/submission/create", nil), "not-a-uuid")

		auth(http.HandlerFunc(h.CreateSubmission)).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	// Trying to create a submission without a valid, parseable json body, should result in StatusBadRequests
	t.Run("invalid json body", func(t *testing.T) {
		t.Parallel()
		rec := httptest.NewRecorder()
		req := withAuth(httptest.NewRequest(http.MethodPost, "/submission/create", bytes.NewBufferString("{")), uuid.New().String())

		auth(http.HandlerFunc(h.CreateSubmission)).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	// Trying to create a submission that is not of category unit or venue should result in StatusNotImplemented. 
	// Its questionable if this is a good design decision since we dont plan on implementing more submission types, 
	// but this is what the submission handler currently does. 
	t.Run("unsupported category", func(t *testing.T) {
		t.Parallel()
		body, _ := json.Marshal(map[string]any{
			"category": "beverage",
			"payload":  map[string]any{},
		})
		rec := httptest.NewRecorder()
		req := withAuth(httptest.NewRequest(http.MethodPost, "/submission/create", bytes.NewReader(body)), uuid.New().String())

		auth(http.HandlerFunc(h.CreateSubmission)).ServeHTTP(rec, req)

		if rec.Code != http.StatusNotImplemented {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotImplemented)
		}
	})
}

// Faulty request to submission/list
// Querying status=bogus should result in StatusBadRequest.
func TestSubmissionHandler_ListSubmissions_invalidStatus(t *testing.T) {
	t.Parallel()

	h := &SubmissionHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/submission/list?status=bogus", nil)

	h.ListSubmissions(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// Faulty request to submission/{id}
// Querying with an invalid uuid should result in StatusBadRequest.
func TestSubmissionHandler_GetByID_invalidUUID(t *testing.T) {
	t.Parallel()

	h := &SubmissionHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/submission/bad-id", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "bad-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
