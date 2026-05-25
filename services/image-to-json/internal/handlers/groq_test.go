package handlers

import(
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fakeGroqResponse(content string) string {
    resp := map[string]any{
        "choices": []map[string]any{
            {
                "message": map[string]string{
                    "content": content,
                },
            },
        },
    }
    b, _ := json.Marshal(resp)
    return string(b)
}

func TestGetJsonFromMenuImage_ReturnsDrinks(t *testing.T) {
    expectedJSON := `[{"drink":"Heineken","type":"beer","price":89,"size":"33cl","tap":false}]`

    fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(fakeGroqResponse(expectedJSON)))
    }))
    defer fake.Close()

    groqURL = fake.URL
    httpClient = fake.Client()
    defer func() {
        groqURL = "https://api.groq.com/openai/v1/chat/completions"
        httpClient = http.DefaultClient
    }()

    imageData := []byte("fake image data")

    result, err := getJsonFromMenuImage(imageData)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if result != expectedJSON {
        t.Errorf("want %s\ngot  %s", expectedJSON, result)
    }
}

func TestGetJsonFromMenuImage_InvalidGroqResponse(t *testing.T) {
    fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("not json at all"))
    }))
    defer fake.Close()

    groqURL = fake.URL
    httpClient = fake.Client()
    defer func() {
        groqURL = "https://api.groq.com/openai/v1/chat/completions"
        httpClient = http.DefaultClient
    }()

    _, err := getJsonFromMenuImage([]byte("fake image"))
    if err == nil {
        t.Error("expected an error for invalid JSON response, got nil")
    }
}

func TestHandleConvertImageToJSON_RejectsGet(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/imagetojson", nil)
    w := httptest.NewRecorder()

    HandleConvertImageToJSON(w, req)

    if w.Code != http.StatusBadRequest {
        t.Errorf("want 400, got %d", w.Code)
    }
}