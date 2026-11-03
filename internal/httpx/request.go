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
	mediaType, _, err := mime.ParseMediaType(
		r.Header.Get("Content-Type"),
	)

	if err != nil || mediaType != "application/json" {
		return errors.New(
			"content type must be application/json",
		)
	}

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		MaxJSONBodyBytes,
	)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		var maxBytesErr *http.MaxBytesError

		if errors.As(err, &maxBytesErr) {
			return fmt.Errorf(
				"request body exceeds %d bytes",
				MaxJSONBodyBytes,
			)
		}

		return fmt.Errorf("invalid JSON: %w", err)
	}

	var extra any

	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New(
				"request body must contain exactly one JSON value",
			)
		}

		return fmt.Errorf("invalid trailing JSON: %w", err)
	}

	return nil
}
