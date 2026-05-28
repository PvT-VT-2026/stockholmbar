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
	"strings"
)

var (
	groqURL    = "https://api.groq.com/openai/v1/chat/completions"
	httpClient = http.DefaultClient
)

// HandleConvertImageToJSON accepts either:
//   - application/json body: {"url": "https://..."} — image URL passed directly to Groq
//   - any other content-type: raw image bytes (legacy behaviour)
func HandleConvertImageToJSON(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Limit body to ~80MB (covers 2-5MB images + up to 75MB high-res).
	r.Body = http.MaxBytesReader(w, r.Body, 80<<20)

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid http method", http.StatusBadRequest)
		return
	}

	var imageURL string

	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		var body struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.URL == "" {
			http.Error(w, `invalid JSON body: expected {"url": "..."}`, http.StatusBadRequest)
			return
		}
		imageURL = body.URL
	} else {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("Failed to read request body: %s\n", err.Error())
			http.Error(w, "Bad payload", http.StatusBadRequest)
			return
		}
		imageURL = "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
	}

	result, err := getJsonFromMenuImageURL(imageURL)
	if err != nil {
		fmt.Printf("Failed to generate json from image: %s\n", err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(result))
}

// getJsonFromMenuImageURL calls the Groq API with an image URL (either an https:// URL or a
// data:image/...;base64,... data URL) and returns the parsed menu JSON string.
func getJsonFromMenuImageURL(imageURL string) (string, error) {
	req, err := generateRequest(imageURL)
	if err != nil {
		return "", fmt.Errorf("unable to create request: %w", err)
	}

	fmt.Println("Executing request")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	fmt.Println("Groq responded with status " + resp.Status)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Groq response: %w", err)
	}

	if !json.Valid(data) {
		log.Println("Invalid JSON response from Groq:", string(data))
		return "", fmt.Errorf("invalid JSON from API")
	}

	var result models.Response
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal Groq response: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in Groq response")
	}

	return result.Choices[0].Message.Content, nil
}

// generateRequest builds a Groq chat-completions request using the given image URL.
// imageURL can be an https:// URL or a data:image/...;base64,... data URL.
func generateRequest(imageURL string) (*http.Request, error) {
	reqBody := models.ChatRequest{
		Model: "meta-llama/llama-4-scout-17b-16e-instruct",
		Messages: []models.Message{
			{
				Role: "system",
				Content: `You are a menu parser. Extract all items from the menu image and return them as JSON.
				Return only valid JSON, no markdown, no explanation. Extract only alcoholic beverages, ignore soft drinks and food items.
				Output format should be a list of items such as {"drink": "Carlsberg", "abv": 5, "type": "beer", "price": 89, "currency": "sek", "size": "", "volume_ml": 500, "tap": true}.
				Tap should be false by default, unless stated otherwise in the image.
				You may assume currency is sek, unless stated otherwise.
				Abv may be NULL.
				Type should be one of the following values: beer, cider, red wine, white wine, spririt, liqueur, drink, or other.
				Drink type is reserved for composite drinks, such as red bull vodka, irish coffee, etc. 
				Do not include anything in parentheses in the drink name. Example: BRISKA PÄRON (CIDER). The drink field should contain only Briska päron.
				Volume may be empty. If volume is stated in the menu, be sure to translate it to ml. For example, "Carlsberg 50cl", would have "volume_ml":500.
				If a drink is available in different sizes, such as glass/bottle for wine, they should be listed as two entries, such as:
				[{"drink": "Proverb Pinot Grigio", "abv": 5, "type": "red wine", "price": 60, "currency": "sek", "size": "glass","volume_ml":null, "tap": false}, {"drink": "Proverb Pinot Grigio", "type": "red wine", "price": 350, "size": "bottle", "tap": false}].
				Be sure to look at the whole image, and include all types of drinks (beers, wines, spirits, liquors)`,
			},
			{
				Role: "user",
				Content: []map[string]any{
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": imageURL,
						},
					},
				},
			},
		},
		Temperature:         1,
		MaxCompletionTokens: 5000,
		TopP:                1,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", groqURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("GROQ_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
