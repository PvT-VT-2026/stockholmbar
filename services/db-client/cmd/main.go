package main

import (
	"context"
	"db-client/internal/db"
	"db-client/internal/handlers"
	"db-client/internal/middleware"
	"db-client/internal/services"
	"db-client/internal/stores"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/go-chi/chi/v5"
)


func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	connString := os.Getenv("SUPABASE_CONN_STRING")
	if connString == "" {
		panic("Failed to load supabase connection string")
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		panic("Failed to load supabase url")
	}
	// Default to 8081
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, connString)
	if err != nil {
		panic(fmt.Sprintf("db: %s", err))
	}
	defer pool.Close()
	

	// Initiate stores
	submissionStore := stores.NewSubmissionStore(pool)
	unitStore := stores.NewUnitStore(pool)
	venueStore := stores.NewVenueStore(pool)

	// Initiate submission service
	submissionService := services.NewSubmissionService(submissionStore, unitStore, venueStore)

	// Initiate handlers
	healthHandler := handlers.NewHealthHandler(pool)
	venueHandler := handlers.NewVenueHandler(venueStore)
	submissionHandler := handlers.NewSubmissionHandler(submissionService)
	// Unit handler no longer has any methods after moving insertion logic to the submission service.
	// Will implement some getter methods, like fetching every unit for a specific venue id etc.
	// unitHandler := handlers.NewUnitHandler(unitStore)

	// Fetch jwk signing key from supabase string. This allows for rolling a new key without updating cfg. 
	jwks, err := keyfunc.NewDefault([]string{supabaseURL + "/auth/v1/.well-known/jwks.json"})
	if err != nil {
		log.Fatalf("jwks: %s", err)
	}
	

	// Mux has been replaced with chi for easier middleware management.
	r := chi.NewRouter()
	r.Use(middleware.RequestLogger)
	

	authMW := middleware.Auth(jwks.Keyfunc)
	adminMW := middleware.RequireAdmin(pool)
	r.Get("/health", healthHandler.Health)


	r.Route("/submission", func(r chi.Router) {
		r.Use(authMW)
		r.Use(adminMW)
		r.Post("/create", submissionHandler.CreateSubmission)
	}) 
	r.Route("/database", func(r chi.Router) {
		r.Route("/venues", func(r chi.Router) {
			r.Get("/{id}", venueHandler.GetByID)
			r.Get("/list", venueHandler.List)
		})
	})

	r.Route("/admin", func(r chi.Router) {
		r.Use(authMW)
		r.Use(adminMW)

		r.Route("/submission", func(r chi.Router) {
			r.Get("/next", submissionHandler.GetOldestPending)
			r.Get("/{id}", submissionHandler.GetByID)
			r.Get("/{id}/image", submissionHandler.GetImageByID)
			r.Get("/list", submissionHandler.ListSubmissions)
			r.Post("/{id}/accept", submissionHandler.Accept)
			r.Post("/{id}/reject", submissionHandler.Reject)
		})
	})


	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	log.Println("Starting server on", port)
	log.Fatal(srv.ListenAndServe())
}
