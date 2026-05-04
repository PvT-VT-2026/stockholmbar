package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image-to-json/internal/models"
	"io"
	"log"
	"net/http"
	"os"
)

// This http handler expects an image payload. Reads the image data and returns json.
func HandleConvertImageToJSON(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	
	// Limit the amount of data that can be read to ~80MB
	// Regular images are between 2-5MB, while some high resolution cameras take images 
	// upward of 75MB. Any more than that should be considered a bad request.
	r.Body = http.MaxBytesReader(w, r.Body, 80<<20)


	if r.Method != http.MethodPost {
		http.Error(w, "Invalid http method", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
    if err != nil {
		fmt.Printf("Failed to read request body: %s\n", err.Error())
		http.Error(w, "Bad payload", http.StatusBadRequest)
        return
    }

	json, err := getJsonFromMenuImage(data)
	if err != nil {
		fmt.Printf("Failed to generate json from image: %s\n", err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(json))
}

// getJsonFromMenuImage calls the groq api and validates the response before returning it.
func getJsonFromMenuImage(imageData []byte) (string, error) {
	req, err := generateRequest(imageData)	
	if err != nil{
		return "", fmt.Errorf("Unable to create request: %s", err.Error())
	}

	fmt.Println("Executing request")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to execute request: %s", err.Error())
	}
	defer resp.Body.Close()


	fmt.Println("Grok responded with status " + resp.Status)

	fmt.Println("Reading response data")
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	if !json.Valid(data) {
		log.Println("Invalid JSON response:")
		fmt.Println(string(data))
		return "", fmt.Errorf("invalid JSON from API")
	}

	var result models.Response
	if err := json.Unmarshal(data, &result); err != nil {
		log.Fatal(err)
	}
	
	return fmt.Sprint(result.Choices[0].Message.Content), nil
}


// generateRequest takes the raw image byte data and returns a ready to execute groq api request.
func generateRequest(imageData []byte) (*http.Request, error) {
	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageData)

	reqBody := models.ChatRequest{
		Model: "meta-llama/llama-4-scout-17b-16e-instruct",
		Messages: []models.Message{
			{
				Role: "system",
				Content: `You are a menu parser. Extract all items from the menu image and return them as JSON.
				Return only valid JSON, no markdown, no explanation. Extract only alcoholic beverages, ignore soft drinks and food items.
				Output format should be a list of items such as {"drink": "Carlsberg 5%", "type": "beer", "price": 89, "size": "50cl", "tap": true}.
				Tap should be false by default, unless stated otherwise in the image.
				If a drink is vailable in different sizes, such as glass/bottle for wine, they should be listed as two entries, such as: 
				[{"drink": "Proverb Pinot Grigio", "type": "red wine", "price": 60, "size": "glass", "tap": false}, {"drink": "Proverb Pinot Grigio", "type": "red wine", "price": 350, "size": "bottle", "tap": false}].
				Be sure to look at the whole image, and include all types of drinks (beers, wines, spritits, liqours)`,
			},
			{
				Role: "user",
				Content: []map[string]any{
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": dataURL,
						},
					},
				},
			},
		},
		Temperature:         1,
		MaxCompletionTokens: 5000, // Arbitrary max limit, could be higher	
		TopP:                1,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	fmt.Println("Setting request headers")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GROQ_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

