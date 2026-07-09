package line

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/albert-einshutoin/mockport/internal/adapter/httpx"
)

func (r *routes) writeValidateMessage(w http.ResponseWriter, req *http.Request) {
	payload, err := decodePayload(req)
	if err != nil {
		writeDecodeError(w, err)
		return
	}
	messages, _ := payload["messages"].([]any)
	if len(messages) == 0 || len(messages) > 5 {
		writeValidationDetails(w, []lineErrorDetail{{Message: "Size must be between 1 and 5", Property: "messages"}})
		return
	}
	details := make([]lineErrorDetail, 0)
	allowedMessageTypes := map[string]bool{
		"text": true, "image": true, "video": true, "audio": true, "location": true,
		"sticker": true, "template": true, "imagemap": true, "flex": true, "coupon": true,
	}
	for i, message := range messages {
		messageObject, ok := message.(map[string]any)
		if !ok {
			details = append(details, lineErrorDetail{Message: "Must be a JSON object", Property: fmt.Sprintf("messages[%d]", i)})
			continue
		}
		messageType, _ := messageObject["type"].(string)
		if !allowedMessageTypes[messageType] {
			details = append(details, lineErrorDetail{
				Message:  "Must be one of the following values: [text, image, video, audio, location, sticker, template, imagemap, flex, coupon]",
				Property: fmt.Sprintf("messages[%d].type", i),
			})
			continue
		}
		if messageType == "text" {
			text, _ := messageObject["text"].(string)
			if strings.TrimSpace(text) == "" {
				details = append(details, lineErrorDetail{Message: "May not be empty", Property: fmt.Sprintf("messages[%d].text", i)})
			}
		}
	}
	if len(details) > 0 {
		writeValidationDetails(w, details)
		return
	}
	writeEmptyJSON(w, http.StatusOK)
}

func (r *routes) writeMessage(w http.ResponseWriter, req *http.Request, includeSentMessages bool, status int) {
	scenario, ok := r.resolveScenario(w, req)
	if !ok {
		return
	}
	switch scenario {
	case "auth_error":
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated invalid channel access token")
	case "rate_limited":
		writeLINEError(w, http.StatusTooManyRequests, "Mockport simulated rate limit")
	case "invalid_request":
		writeLINEError(w, http.StatusBadRequest, "Mockport simulated invalid message request")
	default:
		payload, err := decodePayload(req)
		if err != nil {
			writeDecodeError(w, err)
			return
		}
		if len(payload) > 0 && payload["messages"] == nil {
			writeLINEError(w, http.StatusBadRequest, "messages is required")
			return
		}
		resource, err := r.store.Create("line", "message", map[string]any{
			"to":       firstNonEmpty(payload["to"], payload["replyToken"], "Umockport"),
			"messages": firstNonEmpty(payload["messages"], []any{map[string]any{"type": "text", "text": "Mockport LINE message"}}),
			"status":   "sent",
		})
		if err != nil {
			writeLINEError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !includeSentMessages {
			writeEmptyJSON(w, status)
			return
		}
		httpx.WriteJSON(w, status, map[string]any{"sentMessages": []map[string]any{{"id": resource.ID, "quoteToken": "mockport_line_quote_token"}}})
	}
}

func (r *routes) writeNarrowcastProgress(w http.ResponseWriter) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"phase":         "succeeded",
		"successCount":  len(r.store.List("line", "message")),
		"failureCount":  0,
		"targetCount":   len(r.store.List("line", "message")),
		"acceptedTime":  "2999-01-01T00:00:00.000Z",
		"completedTime": "2999-01-01T00:00:01.000Z",
	})
}

func (r *routes) writeMessagingProfile(w http.ResponseWriter, req *http.Request, userID string) {
	scenario, ok := r.resolveScenario(w, req)
	if !ok {
		return
	}
	if scenario == "auth_error" {
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated invalid channel access token")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, lineProfile(userID))
}

func (r *routes) writeContentEndpoint(w http.ResponseWriter, path string) {
	switch {
	case strings.HasSuffix(path, "/content/transcoding"):
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": "succeeded"})
	case strings.HasSuffix(path, "/content/preview"):
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mockport line preview image"))
	case strings.HasSuffix(path, "/content"):
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mockport line message content"))
	default:
		writeLINEError(w, http.StatusNotFound, "not found")
	}
}

func (r *routes) writeBotInfo(w http.ResponseWriter) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"userId":         "U00000000000000000000000000000000",
		"basicId":        "@mockport",
		"displayName":    "Mockport LINE Official Account",
		"pictureUrl":     "https://example.test/mockport-line-bot.png",
		"chatMode":       "bot",
		"markAsReadMode": "auto",
	})
}

func (r *routes) writeGroupEndpoint(w http.ResponseWriter, path string) {
	rest := strings.TrimPrefix(path, "/v2/bot/group/")
	parts := strings.Split(rest, "/")
	if len(parts) < 2 {
		writeLINEError(w, http.StatusBadRequest, "The value for the 'groupId' parameter is invalid")
		return
	}
	groupID := parts[0]
	switch {
	case len(parts) == 2 && parts[1] == "summary":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"groupId": groupID, "groupName": "Mockport Group", "pictureUrl": "https://example.test/mockport-line-group.png"})
	case len(parts) == 3 && parts[1] == "members" && parts[2] == "count":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"count": 3})
	case len(parts) == 3 && parts[1] == "members" && parts[2] == "ids":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"memberIds": []string{"Umockport", "Ulocaluser"}})
	case len(parts) == 3 && parts[1] == "member":
		httpx.WriteJSON(w, http.StatusOK, lineProfile(parts[2]))
	default:
		writeLINEError(w, http.StatusNotFound, "Not found")
	}
}

func (r *routes) writeRoomEndpoint(w http.ResponseWriter, path string) {
	rest := strings.TrimPrefix(path, "/v2/bot/room/")
	parts := strings.Split(rest, "/")
	if len(parts) < 2 {
		writeLINEError(w, http.StatusBadRequest, "The value for the 'roomId' parameter is invalid")
		return
	}
	switch {
	case len(parts) == 3 && parts[1] == "members" && parts[2] == "count":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"count": 3})
	case len(parts) == 3 && parts[1] == "members" && parts[2] == "ids":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"memberIds": []string{"Umockport", "Ulocaluser"}})
	case len(parts) == 3 && parts[1] == "member":
		httpx.WriteJSON(w, http.StatusOK, lineProfile(parts[2]))
	default:
		writeLINEError(w, http.StatusNotFound, "Not found")
	}
}
