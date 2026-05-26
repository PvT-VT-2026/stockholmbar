package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPlaceInfo_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("X-Goog-Api-Key") != "fake-api-key" {
			t.Errorf("Wrong API key")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
            "displayName": {
                "text": "Test pub"
            },
            "location": {
                "latitude": 59.3293,
                "longitude": 18.0686
            },
            "rating": 4.8
        }`))
	}))
	defer mockServer.Close()


	info, err := getPlaceInfo("test-id-99", "fake-api-key", mockServer.URL)

	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if info == nil {
		t.Fatalf("Expected PlaceInfo, but got nil")
	}

	if info.Name != "Test pub" {
		t.Errorf("Expected name 'Test pub', got '%s'", info.Name)
	}

	if info.Rating != 4.8 {
		t.Errorf("Expected rating 4.8, got %f", info.Rating)
	}
}

func TestGetPlaceIds_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("X-Goog-Api-Key") != "fake-api-key" {
			t.Errorf("Incorrect API key in header")
		}
		if r.Header.Get("X-Goog-FieldMask") != "places.id,places.displayName,places.formattedAddress" {
			t.Errorf("Incorrect FieldMask in header")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
            "places": [
                {
                    "id": "ChIJT_9-yEedX0YRM-Uu-O6uK0I",
                    "displayName": {
                        "text": "test Pub"
                    },
                    "formattedAddress": "Testgatan 1, Stockholm"
                }
            ]
        }`))
	}))
	defer mockServer.Close()

	results, err := getPlaceIds("Mocked Pub", "fake-api-key", mockServer.URL)

	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result in the list, got %d", len(results))
	}

	place := results[0]
	if place.ID != "ChIJT_9-yEedX0YRM-Uu-O6uK0I" {
		t.Errorf("Got incorrect ID: %s", place.ID)
	}
	if place.Name != "test Pub" {
		t.Errorf("Got incorrect name: %s", place.Name)
	}
	if place.Address != "Testgatan 1, Stockholm" {
		t.Errorf("Got incorrect address: %s", place.Address)
	}
}

func TestGetPlaceIds_GoogleError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{ "error": "Invalid API key" }`))
	}))
	defer mockServer.Close()

	results, err := getPlaceIds("Mocked Pub", "wrong-key", mockServer.URL)

	if err == nil {
		t.Fatal("Expected the function to return an error, but it succeeded")
	}

	if results != nil {
		t.Errorf("Expected nil results on error, but got a list")
	}
}