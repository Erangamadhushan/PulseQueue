package httpx

import (
	"encoding/json"
	"errors"
	"fmt"      // Format strings
	"io"       // Read request body
	"mime"     // Parse MIME types
	"net/http" // HTTP request and response handling
)

const MaxJSONBodyBytes int64 = 1 << 20 // 1 MiB

func DecodeJSON(
	w http.ResponseWriter,
	r *http.Request,
	dst any,
) error {

	contentType := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)

	if err != nil || mediaType != "application/json" {
		return errors.New("content type must be application/json")
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxJSONBodyBytes) // Limit the size of the request body to prevent abuse

	decoder := json.NewDecoder(r.Body) // Create a new JSON decoder for the request body
	decoder.DisallowUnknownFields()    // Prevent unknown fields in the JSON payload

	if err := decoder.Decode(dst); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return fmt.Errorf("request body exceeds %d bytes", MaxJSONBodyBytes)
		}

		return fmt.Errorf("invalid JSON request body: %w", err)
	}

	var extra any // Check for any extra JSON data after the first value
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("request body must contain exactly one JSON value")
		}
		return fmt.Errorf("invalid trailing JSON data: %w", err)
	}

	return nil
}
