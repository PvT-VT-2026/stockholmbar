package handlers

import (
	"bytes"
	"db-client/internal/models"
	"db-client/internal/services"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
)


type SubmissionHandler struct {
	service *services.SubmissionService
}

func NewSubmissionHandler(s *services.SubmissionService) *SubmissionHandler {
	return &SubmissionHandler{service: s}
}

// ------ CLIENT FACING HANDLERS ------

// Creates a submission, which an admin will later have to approve or reject
// in the admin panel. 
// This handler is the only way for clientside data to enter the database.
// Expects a CreateSubmissionRequest struct as json in the request body, and a user id to be passed
// in the context.
func (h *SubmissionHandler) CreateSubmission(w http.ResponseWriter, r *http.Request) {
	// The user UUID should come from some auth middleware and should be sent in the request context
	// and be handled like this:
		// userID, ok := r.Context().Value("userID").(uuid.UUID)
		// if !ok {
		// 	    http.Error(w, "unauthorized", http.StatusUnauthorized)
		// 	    return
		// }
	// For now, the .env file needs a user uuid which points to a user in the supabase project. 
	stringID := os.Getenv("TEST_USER_UUID")
	if stringID == "" {
        panic("Failed to load a test user id")
    }
	userID, err := uuid.Parse(stringID)
	if err != nil {
		panic("Failed to parse test user id")
	}

	// If the request body can not be parsed into a CreateSubmissionRequest struct,
	// it indicates a frontend issue.
	var input models.CreateSubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		fmt.Printf("SubmissionHandler.CreateSubmission: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if input.Category != "venue" && input.Category != "unit" {
		http.Error(w, "unimplemented method", http.StatusNotImplemented)
		return
	}

	if err := h.service.CreateSubmission(r.Context(), userID, input); err != nil {
		fmt.Printf("SubmissionHandler.CreateSubmission: %v", err)
        http.Error(w, "failed to create submission", http.StatusInternalServerError)
        return
    }

	w.WriteHeader(http.StatusCreated)
}



// ------ ADMIN FACING HANDLERS ------

// Retrieves an entire submission, including payload, by id
func (h *SubmissionHandler) GetByID(w http.ResponseWriter, r *http.Request) {

	submissionID, err := uuid.Parse(r.PathValue("id"))
    if err != nil {
		fmt.Printf("SubmissionHandler.GetByID: %v", err)
        http.Error(w, "invalid id parameter", http.StatusBadRequest)
        return
    }

	submission, err := h.service.GetByID(r.Context(), submissionID)
	if err != nil {
		fmt.Printf("SubmissionHandler.GetByID: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if submission == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submission)
}	

// Retreives the image associated with a submission, if there is one
func (h *SubmissionHandler) GetImageByID(w http.ResponseWriter, r *http.Request) {
    submissionID, err := uuid.Parse(r.PathValue("id"))
    if err != nil {
		fmt.Printf("SubmissionHandler.GetImageByID: %v", err)
        http.Error(w, "invalid id parameter", http.StatusBadRequest)
        return
    }

	imageBytes, err := h.service.GetImageByID(r.Context(), submissionID)
	if err != nil {
		fmt.Printf("SubmissionHandler.GetImageByID: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if imageBytes == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}


	// Set content length according to image size and use io.Copy to write
	// the response instead of w.Write, to avoid partial read/write errors.
	w.Header().Set("Content-Type", http.DetectContentType(imageBytes))
	w.Header().Set("Content-Length", strconv.Itoa(len(imageBytes)))
    
    io.Copy(w, bytes.NewReader(imageBytes))
}	


// Retrieves the oldest pending submission. 
// Mainly used by the admin panel to easily get the next in line submission.
func (h *SubmissionHandler) GetOldestPending(w http.ResponseWriter, r *http.Request) {
	submission, err := h.service.GetOldestPending(r.Context())
	if err != nil {
		fmt.Printf("SubmissionHandler.GetOldestPending: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if submission == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submission)
}	

// Retrieves a list of submissions. Setting ?status=pending in the url
// filters by only pending submissions. The same can be done with accepted and rejected submissions.
// If there is no status parameter, all submissions are fetched. 
// This list does not include the full payload of each submission, only the metadata. 
// Submissions are ordered by created_at, listing the oldest first.
func (h *SubmissionHandler) ListSubmissions(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status != "pending" && status != "accepted" && status != "rejected" && status != "" {
		http.Error(w, "invalid parameter", http.StatusBadRequest)
		return
	}

	submissions, err := h.service.ListSubmissions(r.Context(), status)
	if err != nil {
		fmt.Printf("SubmissionHandler.ListSubmissions: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submissions)
}

func (h *SubmissionHandler) Accept(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")
	submissionID, err := uuid.Parse(idParam)
	if err != nil {
		fmt.Printf("SubmissionHandler.Accept: %v", err)
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	err = h.service.Accept(r.Context(), submissionID)	
	if err != nil {
		fmt.Printf("SubmissionHandler.Accept: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *SubmissionHandler) Reject(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")
	submissionID, err := uuid.Parse(idParam)
	if err != nil {
		fmt.Printf("SubmissionHandler.Reject: %v", err)
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}
	
	err = h.service.Reject(r.Context(), submissionID)	
	if err != nil {
		fmt.Printf("SubmissionHandler.Reject: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}





// User facing
// POST   /submissions          // create a submission
// GET    /submissions          // get own submissions (filtered by auth.uid())
// GET    /submissions/:id      // get a specific submission

// Admin facing  
// GET    /admin/submissions           // list all submissions, filterable by status/category
// GET    /admin/submissions/:id       // get a specific submission with full payload
// POST   /admin/submissions/:id/accept  // accept and apply
// POST   /admin/submissions/:id/reject  // reject with optional reason