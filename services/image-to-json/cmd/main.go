package main

import (
	"image-to-json/internal/handlers"
	"log"
	"net/http"
	"os"
)

// Verifies that environment variables are present,
// binds handlers to a basic mux and starts the server
func main() {    
    // Basic sanity check, exit immediately if api key is not present 
    if key := os.Getenv("GROQ_API_KEY"); key == "" {
        panic("Failed to load api key")
    }

    // Default to 8080
	port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    // Bind handlers
    mux := http.NewServeMux()
    mux.HandleFunc("/imagetojson", handlers.HandleConvertImageToJSON)
    mux.HandleFunc("/health", handlers.Health)

    srv := &http.Server{
        Addr:    ":" + port,
        Handler: mux,
    }

    log.Println("Starting server on", port)
    log.Fatal(srv.ListenAndServe())
}


