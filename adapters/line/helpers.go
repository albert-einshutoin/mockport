package line

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/albert-einshutoin/mockport/internal/adapter/httpx"
)

func decodePayload(req *http.Request) (map[string]any, error) {
	if req.Body == nil {
		return map[string]any{}, nil
	}
	defer req.Body.Close()
	var payload map[string]any
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&payload); err != nil {
		if errors.Is(err, io.EOF) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, fmt.Errorf("unexpected trailing JSON value")
		}
		return nil, err
	}
	if payload == nil {
		payload = map[string]any{}
	}
	return payload, nil
}

func writeDecodeError(w http.ResponseWriter, err error) {
	if httpx.IsRequestBodyTooLarge(err) {
		writeLINEError(w, http.StatusRequestEntityTooLarge, "Request body is too large")
		return
	}
	writeLINEError(w, http.StatusBadRequest, "Request body must be JSON")
}

func writeValidationDetails(w http.ResponseWriter, details []lineErrorDetail) {
	httpx.WriteJSON(w, http.StatusBadRequest, struct {
		Message string            `json:"message"`
		Details []lineErrorDetail `json:"details"`
	}{
		Message: fmt.Sprintf("The request body has %d error(s)", len(details)),
		Details: details,
	})
}

func lineProfile(userID string) lineProfileResponse {
	if userID == "" {
		userID = "Umockport"
	}
	return lineProfileResponse{
		UserID:        userID,
		DisplayName:   "Mockport LINE User",
		PictureURL:    "https://example.test/mockport-line-user.png",
		StatusMessage: "Mockport local LINE profile",
	}
}

func writeLINEError(w http.ResponseWriter, status int, message string) {
	httpx.WriteJSON(w, status, map[string]any{"message": message})
}

func writeOAuthError(w http.ResponseWriter, status int, code, description string) {
	httpx.WriteJSON(w, status, map[string]any{"error": code, "error_description": description})
}

func writeEmptyJSON(w http.ResponseWriter, status int) {
	httpx.WriteJSON(w, status, map[string]any{})
}

func writeNoBody(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}

func normalizeScenario(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "line_success"
	}
	return value
}

func firstNonEmpty(values ...any) any {
	for _, value := range values {
		switch typed := value.(type) {
		case nil:
			continue
		case string:
			if strings.TrimSpace(typed) == "" {
				continue
			}
		}
		return value
	}
	return nil
}
