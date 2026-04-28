package main

import (
	"get-places-data/internal/handlers"
	"log"
	"net/http"
	"os"
)


func main() {    
    // Basic sanity check, exit immediately if api key is not present
	apiKey := os.Getenv("GOOGLE_API_KEY")
    if apiKey == "" {
        panic("Failed to load api key")
    }

	env := &handlers.APIEnv{
		GoogleAPIKey: apiKey,
	}

    // Default to 8080
	port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    // Bind handlers
    mux := http.NewServeMux()
	mux.HandleFunc("/findplace", env.GetPlaceIdsHandler)
    mux.HandleFunc("/placeinfo", env.GetPlaceInfoHandler)
    mux.HandleFunc("/health", handlers.Health)

    srv := &http.Server{
        Addr:    ":" + port,
        Handler: mux,
    }

    log.Println("Starting server on", port)
    log.Fatal(srv.ListenAndServe())
}


