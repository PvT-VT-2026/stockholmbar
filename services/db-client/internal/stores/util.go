package stores

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// Takes a base64 encoded image and returns the raw byte data
func decodeBase64Image(raw string) ([]byte, error) {
	if idx := strings.Index(raw, ","); idx != -1 {
		raw = raw[idx+1:]
	}
	return base64.StdEncoding.DecodeString(raw)
}


// Returns a sha256 hash for a json blob.
// This is used for the uniqueness constraints of submissions,
// i.e. the same user cant submit the exact same submission more than once.
func hashPayload(payload json.RawMessage) (string, error) {
    // JSON can be semantically identical but differ in key ordering
    // or whitespace, so we normalize it first by unmarshalling and re-marshalling
    var normalized any
    if err := json.Unmarshal(payload, &normalized); err != nil {
        return "", err
    }

    canonical, err := json.Marshal(normalized)
    if err != nil {
        return "", err
    }

    hash := sha256.Sum256(canonical)
    return fmt.Sprintf("%x", hash), nil
}