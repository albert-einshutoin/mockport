package openai

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/adapter/httpx"
	"github.com/albert-einshutoin/mockport/internal/security"
	"github.com/albert-einshutoin/mockport/internal/state"
)

type Adapter struct{}

func New() Adapter { return Adapter{} }

func (a Adapter) Name() string { return "openai" }

func (a Adapter) Register(mux *http.ServeMux, cfg adapter.Config) error {
	basePath := cfg.BasePath
	if basePath == "" {
		basePath = "/openai"
	}
	r := &routes{basePath: strings.TrimRight(basePath, "/"), cfg: cfg, store: state.NewStore()}
	mux.HandleFunc(r.basePath+"/", r.handle)
	return nil
}

func (a Adapter) FakeEnv(cfg adapter.Config) map[string]string {
	basePath := cfg.BasePath
	if basePath == "" {
		basePath = "/openai"
	}
	return map[string]string{
		"OPENAI_BASE_URL": "http://localhost:43101" + basePath + "/v1",
		"OPENAI_API_KEY":  "mockport_openai_key",
	}
}

func (a Adapter) Metadata() adapter.Metadata {
	scenarios := []adapter.Scenario{
		{Name: "chat_success", Supported: true},
		{Name: "stream_success", Supported: true},
		{Name: "rate_limited", Supported: true},
		{Name: "context_length_exceeded", Supported: true},
		{Name: "auth_error", Supported: true},
	}
	return adapter.Metadata{
		Name:            "openai",
		Maturity:        adapter.MaturityWorkflowCompatible,
		ProviderVersion: "2025-02-01",
		SDKVersions:     []adapter.SDKVersion{{Name: "openai", Version: "6.42.0"}},
		Levels:          []adapter.Level{adapter.LevelWire, adapter.LevelSDK, adapter.LevelWorkflow, adapter.LevelState, adapter.LevelError},
		Capabilities:    []string{"models", "chat_completions", "responses", "embeddings", "files", "batches"},
		Scenarios:       scenarios,
		StatefulResources: []string{
			"chat_completion",
			"response",
			"embedding",
			"file",
			"batch",
		},
		Reset: true,
		Endpoints: []adapter.Endpoint{
			{Method: http.MethodGet, Path: "/openai/v1/models", SupportedScenarios: []string{"chat_success"}, Notes: "OpenAI-like model list"},
			{Method: http.MethodPost, Path: "/openai/v1/chat/completions", SupportedScenarios: []string{"chat_success", "stream_success", "rate_limited", "context_length_exceeded", "auth_error"}, Notes: "OpenAI-like chat completion"},
			{Method: http.MethodPost, Path: "/openai/v1/responses", SupportedScenarios: []string{"chat_success", "rate_limited", "context_length_exceeded", "auth_error"}, Notes: "OpenAI-like responses endpoint"},
			{Method: http.MethodGet, Path: "/openai/v1/responses/{id}", SupportedScenarios: []string{"chat_success"}, Notes: "Deterministic response lookup"},
			{Method: http.MethodPost, Path: "/openai/v1/embeddings", SupportedScenarios: []string{"chat_success"}, Notes: "OpenAI-like deterministic embeddings"},
			{Method: http.MethodPost, Path: "/openai/v1/files", SupportedScenarios: []string{"chat_success"}, Notes: "OpenAI-like file creation for batch workflows"},
			{Method: http.MethodPost, Path: "/openai/v1/batches", SupportedScenarios: []string{"chat_success"}, Notes: "OpenAI-like batch creation"},
			{Method: http.MethodGet, Path: "/openai/v1/batches/{id}", SupportedScenarios: []string{"chat_success"}, Notes: "OpenAI-like batch lookup"},
			{Method: http.MethodPost, Path: "/openai/test/reset", SupportedScenarios: []string{"chat_success", "stream_success", "rate_limited", "context_length_exceeded", "auth_error"}, Notes: "Clears state for test isolation"},
		},
	}
}

type routes struct {
	basePath string
	cfg      adapter.Config
	store    *state.Store
}

func (r *routes) handle(w http.ResponseWriter, req *http.Request) {
	httpx.LimitRequestBody(w, req)
	path := strings.TrimPrefix(req.URL.Path, r.basePath)
	switch {
	case req.Method == http.MethodGet && path == "/v1/models":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"object": "list", "data": []map[string]string{{"id": "gpt-mockport", "object": "model"}}})
	case req.Method == http.MethodPost && path == "/v1/chat/completions":
		r.writeCompletion(w, req, "chat.completion")
	case req.Method == http.MethodPost && path == "/v1/responses":
		r.writeCompletion(w, req, "response")
	case req.Method == http.MethodGet && strings.HasPrefix(path, "/v1/responses/"):
		r.writeResponseLookup(w, strings.TrimPrefix(path, "/v1/responses/"))
	case req.Method == http.MethodPost && path == "/v1/embeddings":
		r.writeEmbedding(w, req)
	case req.Method == http.MethodPost && path == "/v1/files":
		r.writeFile(w, req)
	case req.Method == http.MethodPost && path == "/v1/batches":
		r.writeBatch(w, req)
	case req.Method == http.MethodGet && strings.HasPrefix(path, "/v1/batches/"):
		r.writeBatchLookup(w, strings.TrimPrefix(path, "/v1/batches/"))
	case req.Method == http.MethodPost && path == "/test/reset":
		r.handleReset(w, req)
	default:
		http.NotFound(w, req)
	}
}

func (r *routes) handleReset(w http.ResponseWriter, req *http.Request) {
	if !security.IsLoopbackRemoteAddr(req.RemoteAddr) {
		writeError(w, http.StatusForbidden, "local_request_required", "OpenAI state reset can only be triggered from loopback")
		return
	}
	resourceTypes := r.store.ResetAll("openai")
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"reset":          true,
		"adapter":        "openai",
		"resource_types": resourceTypes,
	})
}

func (r *routes) writeCompletion(w http.ResponseWriter, req *http.Request, object string) {
	switch normalizeScenario(r.cfg.Scenario) {
	case "auth_error":
		writeError(w, http.StatusUnauthorized, "invalid_api_key", "Mockport simulated invalid API key")
	case "rate_limited":
		writeError(w, http.StatusTooManyRequests, "rate_limited", "Mockport simulated rate limit")
	case "context_length_exceeded":
		writeError(w, http.StatusBadRequest, "context_length_exceeded", "Mockport simulated context length error")
	case "stream_success":
		if object == "chat.completion" {
			writeChatCompletionStream(w)
			return
		}
		r.writeStatefulCompletion(w, req, object)
	default:
		r.writeStatefulCompletion(w, req, object)
	}
}

func (r *routes) writeStatefulCompletion(w http.ResponseWriter, req *http.Request, object string) {
	payload, err := decodePayload(req)
	if err != nil {
		if httpx.IsRequestBodyTooLarge(err) {
			writeError(w, http.StatusRequestEntityTooLarge, "request_too_large", "Request body is too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be JSON")
		return
	}
	if len(payload) == 0 {
		payload["model"] = "gpt-mockport"
		if object == "chat.completion" {
			payload["messages"] = []any{map[string]any{"role": "user", "content": "Mockport request"}}
		} else {
			payload["input"] = "Mockport request"
		}
	}
	if payload["mockport_unsupported_parameter"] != nil {
		writeError(w, http.StatusBadRequest, "unsupported_parameter", "Mockport simulated unsupported parameter")
		return
	}
	if object == "chat.completion" && payload["stream"] == true {
		writeChatCompletionStream(w)
		return
	}
	if !validModel(payload["model"]) {
		writeError(w, http.StatusBadRequest, "model_not_found", "Mockport simulated invalid model")
		return
	}
	required := []string{"model"}
	resourceType := "response"
	if object == "chat.completion" {
		required = append(required, "messages")
		resourceType = "chat_completion"
	} else {
		required = append(required, "input")
	}
	if err := state.RequireFields(payload, required...); err != nil {
		writeError(w, http.StatusBadRequest, "missing_required_field", err.Error())
		return
	}
	if object == "chat.completion" && !validMessages(payload["messages"]) {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "messages must be an array")
		return
	}

	body := completionBody(object)
	body["model"] = payload["model"]
	resource, err := r.store.Create("openai", resourceType, body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "mockport_state_error", err.Error())
		return
	}
	response := resource.Data
	response["id"] = resource.ID
	httpx.WriteJSON(w, http.StatusOK, response)
}

func (r *routes) writeResponseLookup(w http.ResponseWriter, id string) {
	if resource, ok := r.store.Get("openai", "response", id); ok {
		body := resource.Data
		body["id"] = resource.ID
		httpx.WriteJSON(w, http.StatusOK, body)
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "Mockport response not found")
}

func completionBody(object string) map[string]any {
	body := dataFromStruct(chatCompletion{
		Object: object,
		Choices: []chatChoice{{
			Index:   0,
			Message: chatMessage{Role: "assistant", Content: "Mockport response"},
		}},
	})
	if object == "response" {
		body = dataFromStruct(responseBody{
			Object:     object,
			Choices:    []chatChoice{{Index: 0, Message: chatMessage{Role: "assistant", Content: "Mockport response"}}},
			OutputText: "Mockport response",
			Status:     "completed",
			Output: []outputItem{{
				ID:     "msg_mockport",
				Type:   "message",
				Status: "completed",
				Role:   "assistant",
				Content: []outputContent{{
					Type:        "output_text",
					Text:        "Mockport response",
					Annotations: []any{},
				}},
			}},
		})
	}
	return body
}

func (r *routes) writeEmbedding(w http.ResponseWriter, req *http.Request) {
	payload, err := decodePayload(req)
	if err != nil {
		if httpx.IsRequestBodyTooLarge(err) {
			writeError(w, http.StatusRequestEntityTooLarge, "request_too_large", "Request body is too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be JSON")
		return
	}
	if err := state.RequireFields(payload, "model", "input"); err != nil {
		writeError(w, http.StatusBadRequest, "missing_required_field", err.Error())
		return
	}
	if !validModel(payload["model"]) {
		writeError(w, http.StatusBadRequest, "model_not_found", "Mockport simulated invalid model")
		return
	}
	resource, err := r.store.Create("openai", "embedding", map[string]any{"model": payload["model"]})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "mockport_state_error", err.Error())
		return
	}
	embedding := any([]float64{0.01, 0.02, 0.03})
	if payload["encoding_format"] == "base64" {
		embedding = base64Embedding([]float32{0.01, 0.02, 0.03})
	}
	httpx.WriteJSON(w, http.StatusOK, embeddingResponse{
		ID:     resource.ID,
		Object: "list",
		Data:   []embeddingData{{Object: "embedding", Index: 0, Embedding: embedding}},
		Model:  payload["model"],
		Usage:  usage{PromptTokens: 1, TotalTokens: 1},
	})
}

func (r *routes) writeFile(w http.ResponseWriter, req *http.Request) {
	purpose, filename, err := fileFields(req)
	if err != nil {
		if httpx.IsRequestBodyTooLarge(err) {
			writeError(w, http.StatusRequestEntityTooLarge, "request_too_large", "Request body is too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be JSON or multipart form data")
		return
	}
	if purpose == "" {
		purpose = "batch"
	}
	if filename == "" {
		filename = "mockport.jsonl"
	}
	resource, err := r.store.Create("openai", "file", map[string]any{
		"object":   "file",
		"purpose":  purpose,
		"filename": filename,
		"bytes":    18,
		"status":   "processed",
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "mockport_state_error", err.Error())
		return
	}
	body := resource.Data
	body["id"] = resource.ID
	httpx.WriteJSON(w, http.StatusOK, body)
}

func (r *routes) writeBatch(w http.ResponseWriter, req *http.Request) {
	payload, err := decodePayload(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be JSON")
		return
	}
	if err := state.RequireFields(payload, "input_file_id", "endpoint", "completion_window"); err != nil {
		writeError(w, http.StatusBadRequest, "missing_required_field", err.Error())
		return
	}
	resource, err := r.store.Create("openai", "batch", map[string]any{
		"object":            "batch",
		"status":            "completed",
		"input_file_id":     payload["input_file_id"],
		"endpoint":          payload["endpoint"],
		"completion_window": payload["completion_window"],
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "mockport_state_error", err.Error())
		return
	}
	body := resource.Data
	body["id"] = resource.ID
	httpx.WriteJSON(w, http.StatusOK, body)
}

func (r *routes) writeBatchLookup(w http.ResponseWriter, id string) {
	if resource, ok := r.store.Get("openai", "batch", id); ok {
		body := resource.Data
		body["id"] = resource.ID
		httpx.WriteJSON(w, http.StatusOK, body)
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "Mockport batch not found")
}

func decodePayload(req *http.Request) (map[string]any, error) {
	if req.Body == nil {
		return map[string]any{}, nil
	}
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

func fileFields(req *http.Request) (string, string, error) {
	contentType := req.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := req.ParseMultipartForm(8 << 20); err != nil {
			return "", "", err
		}
		purpose := req.FormValue("purpose")
		filename := "mockport.jsonl"
		if req.MultipartForm != nil {
			for _, files := range req.MultipartForm.File {
				if len(files) > 0 && files[0].Filename != "" {
					filename = files[0].Filename
					break
				}
			}
		}
		return purpose, filename, nil
	}
	payload, err := decodePayload(req)
	if err != nil {
		return "", "", err
	}
	purpose, _ := payload["purpose"].(string)
	filename, _ := payload["filename"].(string)
	return purpose, filename, nil
}

func validModel(value any) bool {
	model, _ := value.(string)
	return model == "gpt-mockport" || model == "text-embedding-mockport"
}

func validMessages(value any) bool {
	_, ok := value.([]any)
	return ok
}

func base64Embedding(values []float32) string {
	buf := make([]byte, len(values)*4)
	for i, value := range values {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(value))
	}
	return base64.StdEncoding.EncodeToString(buf)
}

const (
	streamCompletionID          = "chatcmpl-mockport"
	streamCompletionModel       = "gpt-mockport"
	streamCompletionCreated     = int64(1700000000)
	streamCompletionFingerprint = "fp_mockport"
	streamCompletionFullText    = "Mockport simulated streaming response."
)

func writeChatCompletionStream(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	base := chatCompletionChunk{
		ID:                streamCompletionID,
		Object:            "chat.completion.chunk",
		Created:           streamCompletionCreated,
		Model:             streamCompletionModel,
		SystemFingerprint: streamCompletionFingerprint,
	}

	// OpenAI streams an initial role chunk before content deltas.
	emptyContent := ""
	writeSSEChunk(w, chatCompletionChunk{
		ID:                base.ID,
		Object:            base.Object,
		Created:           base.Created,
		Model:             base.Model,
		SystemFingerprint: base.SystemFingerprint,
		Choices: []chatCompletionChunkChoice{{
			Index:        0,
			Delta:        chatCompletionChunkDelta{Role: "assistant", Content: &emptyContent},
			FinishReason: nil,
		}},
	})

	// Split into fixed-size word groups so tests and SDK contracts see multiple chunks.
	for _, content := range splitStreamTextIntoChunks(streamCompletionFullText, 2) {
		content := content
		writeSSEChunk(w, chatCompletionChunk{
			ID:                base.ID,
			Object:            base.Object,
			Created:           base.Created,
			Model:             base.Model,
			SystemFingerprint: base.SystemFingerprint,
			Choices: []chatCompletionChunkChoice{{
				Index:        0,
				Delta:        chatCompletionChunkDelta{Content: &content},
				FinishReason: nil,
			}},
		})
	}

	stop := "stop"
	writeSSEChunk(w, chatCompletionChunk{
		ID:                base.ID,
		Object:            base.Object,
		Created:           base.Created,
		Model:             base.Model,
		SystemFingerprint: base.SystemFingerprint,
		Choices: []chatCompletionChunkChoice{{
			Index:        0,
			Delta:        chatCompletionChunkDelta{},
			FinishReason: &stop,
		}},
	})

	_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	_ = http.NewResponseController(w).Flush()
}

func writeSSEChunk(w http.ResponseWriter, chunk chatCompletionChunk) {
	data, _ := json.Marshal(chunk)
	_, _ = fmt.Fprintf(w, "data: %s\n\n", data)
	_ = http.NewResponseController(w).Flush()
}

// splitStreamTextIntoChunks groups words deterministically; trailing spaces keep chunk joins stable.
func splitStreamTextIntoChunks(text string, wordsPerChunk int) []string {
	words := strings.Fields(text)
	if len(words) == 0 || wordsPerChunk < 1 {
		return nil
	}
	chunks := make([]string, 0, (len(words)+wordsPerChunk-1)/wordsPerChunk)
	for i := 0; i < len(words); i += wordsPerChunk {
		end := i + wordsPerChunk
		if end > len(words) {
			end = len(words)
		}
		chunk := strings.Join(words[i:end], " ")
		if end < len(words) {
			chunk += " "
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}

func normalizeScenario(s string) string {
	if s == "" {
		return "chat_success"
	}
	return s
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	httpx.WriteJSON(w, status, errorBody{Error: errorDetail{Type: "mockport_error", Code: code, Message: message}})
}

func dataFromStruct(value any) map[string]any {
	encoded, _ := json.Marshal(value)
	var decoded map[string]any
	_ = json.Unmarshal(encoded, &decoded)
	return decoded
}
