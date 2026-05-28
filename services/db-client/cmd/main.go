package main

import (
	"context"
	"db-client/internal/clients"
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
	"github.com/go-chi/cors"
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
	beverageStore := stores.NewBeverageStore(pool)

	// Initiate places client if URL is configured
	var placesClient *clients.PlacesClient
	if placesURL := os.Getenv("GET_PLACES_DATA_URL"); placesURL != "" {
		placesClient = clients.NewPlacesClient(placesURL)
	} else {
		log.Println("Warning: GET_PLACES_DATA_URL not set, venue business hours will not be auto-populated")
	}

	// Initiate submission service
	submissionService := services.NewSubmissionService(submissionStore, unitStore, venueStore, placesClient)

	// Initiate handlers
	healthHandler := handlers.NewHealthHandler(pool)
	venueHandler := handlers.NewVenueHandler(venueStore)
	submissionHandler := handlers.NewSubmissionHandler(submissionService)
	beverageHandler := handlers.NewBeverageHandler(beverageStore)
	placesHandler := handlers.NewPlacesHandler(placesClient)
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
	r.Use(cors.AllowAll().Handler)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))
	r.Use(middleware.RequestLogger)
	

	authMW := middleware.Auth(jwks.Keyfunc)
	adminMW := middleware.RequireAdmin(pool)
	r.Get("/health", healthHandler.Health)

	r.Route("/places", func(r chi.Router) {
		r.Get("/search", placesHandler.Search)
		r.Get("/info", placesHandler.Info)
	})

	r.Route("/submission", func(r chi.Router) {
		r.With(authMW).Post("/create", submissionHandler.CreateSubmission)
	})
	r.Route("/database", func(r chi.Router) {
		r.Route("/venues", func(r chi.Router) {
			r.Get("/search", venueHandler.SearchAll)
			r.Get("/list", venueHandler.List)
			r.Get("/{id}", venueHandler.GetByID)
			r.Get("/{id}/menu", venueHandler.GetMenu)
		})
		r.Route("/beverages", func(r chi.Router) {
			r.Get("/list", beverageHandler.List)
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

		r.Route("/venues", func(r chi.Router) {
			r.Patch("/{id}", venueHandler.Update)
			r.Delete("/{id}", venueHandler.Delete)
			r.Patch("/{id}/menu/{unitId}", venueHandler.UpdateMenuItem)
			r.Delete("/{id}/menu/{unitId}", venueHandler.DeleteMenuItem)
		})
	})


	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	log.Println("Starting server on", port)
	log.Fatal(srv.ListenAndServe())
}
