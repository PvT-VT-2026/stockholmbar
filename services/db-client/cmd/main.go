package main

import (
	"db-client/internal/db"
	"db-client/internal/handlers"
	"db-client/internal/services"
	"db-client/internal/stores"
	"log"
	"net/http"
	"os"
)

func main() {
	connString := os.Getenv("SUPABASE_CONN_STRING")
	if connString == "" {
        panic("Failed to load supabase connection string")
    }

	dbClient, err := db.New(connString)
	if err != nil {
		panic("Failed to open connection to database: " + err.Error())
	}

    // Initiate stores
    submissionStore := stores.NewSubmissionStore(dbClient)
    unitStore := stores.NewUnitStore(dbClient)
    venueStore := stores.NewVenueStore(dbClient)

    // Initiate submission service
    submissionService := services.NewSubmissionService(submissionStore, unitStore, venueStore)
    
    // Initiate handlers
    healthHandler := handlers.NewHealthHandler(dbClient)
    venueHandler := handlers.NewVenueHandler(venueStore)
    submissionHandler := handlers.NewSubmissionHandler(submissionService)
    // Unit handler no longer has any methods after moving insertion logic to the submission service.
    // Will implement some getter methods, like fetching every unit for a specific venue id etc. 
    // unitHandler := handlers.NewUnitHandler(unitStore)   
    
    mux := http.NewServeMux()

    // Client facing
    mux.HandleFunc("POST /submission/create", submissionHandler.CreateSubmission)   // Create a new submission
	mux.HandleFunc("GET /database/venue/{id}", venueHandler.GetByID)                // Fetch all data for one venue, including location
    
    // Admin facing
	mux.HandleFunc("GET /admin/health", healthHandler.Health)
    mux.HandleFunc("GET /admin/submission/next", submissionHandler.GetOldestPending)    // Fetches the oldest pending submission, meant for reviewing submissions in order of creation date
    mux.HandleFunc("GET /admin/submission/{id}", submissionHandler.GetByID)             // Fetch an existing submission by id
    mux.HandleFunc("GET /admin/submission/list", submissionHandler.ListSubmissions)     // Fetches a list of submissions without payloads, only metadata. Can be filtered on status
    mux.HandleFunc("POST /admin/submission/{id}/accept", submissionHandler.Accept)      // Accepts a submission, inserting it into the database and updating the submission status        
    mux.HandleFunc("POST /admin/submission/{id}/reject", submissionHandler.Reject)      // Rejects a submission        
    

    // Default to 8081
	port := os.Getenv("PORT")
    if port == "" {
        port = "8081"
    }

    srv := &http.Server{
        Addr:    ":" + port,
        Handler: mux,
    }

    log.Println("Starting server on", port)
    log.Fatal(srv.ListenAndServe())
}