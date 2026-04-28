package main

import (
	"db-client/internal/db"
	"db-client/internal/handlers"
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


    venueStore := stores.NewVenueStore(dbClient)
    unitStore := stores.NewUnitStore(dbClient)


    // Bind handlers
    mux := http.NewServeMux()
    healthHandler := handlers.NewHealthHandler(dbClient)
    venueHandler := handlers.NewVenueHandler(venueStore)
    unitHandler := handlers.NewUnitHandler(unitStore)

	mux.HandleFunc("GET /health", healthHandler.Health)
	mux.HandleFunc("GET /venue/{id}", venueHandler.GetByID)
    mux.HandleFunc("POST /venue/create", venueHandler.Create)
    mux.HandleFunc("POST /unit/create", unitHandler.CreateUnits)

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