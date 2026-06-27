package openai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/adapter/adaptertest"
)

func TestModels(t *testing.T) {
	rec := performRequest(t, adapter.Config{BasePath: "/openai", Scenario: "chat_success"}, http.MethodGet, "/openai/v1/models")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["object"] != "list" {
		t.Fatalf("object = %v, want list", body["object"])
	}
}

func TestChatCompletionScenarios(t *testing.T) {
	tests := []struct {
		name     string
		scenario string
		status   int
		code     string
	}{
		{"success", "chat_success", http.StatusOK, ""},
		{"rate limited", "rate_limited", http.StatusTooManyRequests, "rate_limited"},
		{"context", "context_length_exceeded", http.StatusBadRequest, "context_length_exceeded"},
		{"auth", "auth_error", http.StatusUnauthorized, "invalid_api_key"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := performRequest(t, adapter.Config{BasePath: "/openai", Scenario: tt.scenario}, http.MethodPost, "/openai/v1/chat/completions")
			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d", rec.Code, tt.status)
			}
			if tt.code != "" {
				assertErrorCode(t, rec, tt.code)
			}
		})
	}
}

func TestResponses(t *testing.T) {
	rec := performRequest(t, adapter.Config{BasePath: "/openai", Scenario: "chat_success"}, http.MethodPost, "/openai/v1/responses")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestResponsesCreateAndRetrieveState(t *testing.T) {
	mux := newOpenAIMux(t, adapter.Config{BasePath: "/openai", Scenario: "chat_success"})
	create := serveOpenAIRequest(mux, http.MethodPost, "/openai/v1/responses", `{"model":"gpt-mockport","input":"hello"}`)
	if create.Code != http.StatusOK {
		t.Fatalf("create status = %d, want %d, body=%s", create.Code, http.StatusOK, create.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if created["id"] != "openai_response_000001" {
		t.Fatalf("created id = %#v", created["id"])
	}

	retrieve := serveOpenAIRequest(mux, http.MethodGet, "/openai/v1/responses/openai_response_000001", "")
	if retrieve.Code != http.StatusOK {
		t.Fatalf("retrieve status = %d, want %d, body=%s", retrieve.Code, http.StatusOK, retrieve.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(retrieve.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode retrieve: %v", err)
	}
	if got["id"] != created["id"] || got["model"] != "gpt-mockport" {
		t.Fatalf("retrieved = %#v", got)
	}
	if _, ok := got["output"].([]any); !ok {
		t.Fatalf("response output is missing or not array: %#v", got)
	}
}

func TestChatCompletionUsesStatefulIDsAndValidatesRequiredFields(t *testing.T) {
	mux := newOpenAIMux(t, adapter.Config{BasePath: "/openai", Scenario: "chat_success"})
	first := serveOpenAIRequest(mux, http.MethodPost, "/openai/v1/chat/completions", `{"model":"gpt-mockport","messages":[{"role":"user","content":"hi"}]}`)
	second := serveOpenAIRequest(mux, http.MethodPost, "/openai/v1/chat/completions", `{"model":"gpt-mockport","messages":[{"role":"user","content":"hi again"}]}`)
	if first.Code != http.StatusOK || second.Code != http.StatusOK {
		t.Fatalf("statuses = %d/%d bodies=%s/%s", first.Code, second.Code, first.Body.String(), second.Body.String())
	}
	if !strings.Contains(first.Body.String(), `"id":"openai_chat_completion_000001"`) || !strings.Contains(second.Body.String(), `"id":"openai_chat_completion_000002"`) {
		t.Fatalf("stateful ids not found: %s / %s", first.Body.String(), second.Body.String())
	}

	missing := serveOpenAIRequest(mux, http.MethodPost, "/openai/v1/chat/completions", `{"model":"gpt-mockport"}`)
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("missing status = %d, want %d, body=%s", missing.Code, http.StatusBadRequest, missing.Body.String())
	}
	assertErrorCode(t, missing, "missing_required_field")
}

func TestChatCompletionStreamSuccessReturnsSSE(t *testing.T) {
	rec := performRequest(t, adapter.Config{BasePath: "/openai", Scenario: "stream_success"}, http.MethodPost, "/openai/v1/chat/completions")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/event-stream") {
		t.Fatalf("Content-Type = %q, want text/event-stream", contentType)
	}

	events := parseSSEDataEvents(rec.Body.String())
	if len(events) < 3 {
		t.Fatalf("data event count = %d, want >= 3", len(events))
	}
	if events[len(events)-1] != "[DONE]" {
		t.Fatalf("last event = %q, want [DONE]", events[len(events)-1])
	}

	dataEvents := events[:len(events)-1]
	streamID := ""
	var assembled strings.Builder
	for i, raw := range dataEvents {
		var chunk chatCompletionChunk
		if err := json.Unmarshal([]byte(raw), &chunk); err != nil {
			t.Fatalf("decode chunk %d: %v", i, err)
		}
		if chunk.Object != "chat.completion.chunk" {
			t.Fatalf("chunk %d object = %q, want chat.completion.chunk", i, chunk.Object)
		}
		if chunk.Model != streamCompletionModel {
			t.Fatalf("chunk %d model = %q, want %s", i, chunk.Model, streamCompletionModel)
		}
		if chunk.Created != streamCompletionCreated {
			t.Fatalf("chunk %d created = %d, want %d", i, chunk.Created, streamCompletionCreated)
		}
		if chunk.SystemFingerprint != streamCompletionFingerprint {
			t.Fatalf("chunk %d system_fingerprint = %q, want %s", i, chunk.SystemFingerprint, streamCompletionFingerprint)
		}
		if streamID == "" {
			streamID = chunk.ID
		} else if chunk.ID != streamID {
			t.Fatalf("chunk %d id = %q, want %q", i, chunk.ID, streamID)
		}
		if len(chunk.Choices) != 1 {
			t.Fatalf("chunk %d choices = %d, want 1", i, len(chunk.Choices))
		}
		choice := chunk.Choices[0]
		switch {
		case i == 0:
			if choice.Delta.Role != "assistant" {
				t.Fatalf("role chunk delta.role = %q, want assistant", choice.Delta.Role)
			}
			if choice.Delta.Content == nil || *choice.Delta.Content != "" {
				t.Fatalf("role chunk delta.content = %#v, want empty string", choice.Delta.Content)
			}
			if choice.FinishReason != nil {
				t.Fatalf("role chunk finish_reason = %#v, want null", choice.FinishReason)
			}
		case i == len(dataEvents)-1:
			if choice.Delta.Role != "" || choice.Delta.Content != nil {
				t.Fatalf("finish chunk delta = %#v, want empty", choice.Delta)
			}
			if choice.FinishReason == nil || *choice.FinishReason != "stop" {
				t.Fatalf("finish chunk finish_reason = %#v, want stop", choice.FinishReason)
			}
		default:
			if choice.Delta.Content == nil {
				t.Fatalf("content chunk %d missing delta.content", i)
			}
			assembled.WriteString(*choice.Delta.Content)
			if choice.FinishReason != nil {
				t.Fatalf("content chunk %d finish_reason = %#v, want null", i, choice.FinishReason)
			}
		}
	}
	if assembled.String() != streamCompletionFullText {
		t.Fatalf("assembled content = %q, want %q", assembled.String(), streamCompletionFullText)
	}
}

func TestResponsesStreamSuccessReturnsJSON(t *testing.T) {
	rec := performRequest(t, adapter.Config{BasePath: "/openai", Scenario: "stream_success"}, http.MethodPost, "/openai/v1/responses")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}
	if strings.Contains(rec.Body.String(), "data: [DONE]") {
		t.Fatalf("responses stream_success returned SSE: %s", rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["object"] != "response" {
		t.Fatalf("object = %v, want response", body["object"])
	}
}

func parseSSEDataEvents(body string) []string {
	lines := strings.Split(body, "\n")
	events := make([]string, 0, len(lines))
	for _, line := range lines {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		events = append(events, strings.TrimPrefix(line, "data: "))
	}
	return events
}

func TestChatCompletionStreamSuccessFlushesSSE(t *testing.T) {
	mux := newOpenAIMux(t, adapter.Config{BasePath: "/openai", Scenario: "stream_success"})
	rec := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}

	req := httptest.NewRequest(http.MethodPost, "/openai/v1/chat/completions", nil)
	mux.ServeHTTP(rec, req)

	if !rec.flushed {
		t.Fatalf("stream response did not flush")
	}
}

func TestChatCompletionSuccessStillReturnsJSON(t *testing.T) {
	rec := performRequest(t, adapter.Config{BasePath: "/openai", Scenario: "chat_success"}, http.MethodPost, "/openai/v1/chat/completions")
	if contentType := rec.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}
	if strings.Contains(rec.Body.String(), "data: [DONE]") {
		t.Fatalf("non-stream response used SSE body: %s", rec.Body.String())
	}
}

type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed bool
}

func (r *flushRecorder) Flush() {
	r.flushed = true
	r.ResponseRecorder.Flush()
}

func TestEmbeddingsFilesAndBatchesAreStateful(t *testing.T) {
	mux := newOpenAIMux(t, adapter.Config{BasePath: "/openai", Scenario: "chat_success"})

	embedding := serveOpenAIRequest(mux, http.MethodPost, "/openai/v1/embeddings", `{"model":"text-embedding-mockport","input":"hello"}`)
	if embedding.Code != http.StatusOK || !strings.Contains(embedding.Body.String(), `"object":"embedding"`) {
		t.Fatalf("embedding status/body = %d %s", embedding.Code, embedding.Body.String())
	}
	base64Embedding := serveOpenAIRequest(mux, http.MethodPost, "/openai/v1/embeddings", `{"model":"text-embedding-mockport","input":"hello","encoding_format":"base64"}`)
	if base64Embedding.Code != http.StatusOK || !strings.Contains(base64Embedding.Body.String(), `"embedding":"`) {
		t.Fatalf("base64 embedding status/body = %d %s", base64Embedding.Code, base64Embedding.Body.String())
	}

	file := serveOpenAIRequest(mux, http.MethodPost, "/openai/v1/files", `{"purpose":"batch","filename":"mockport.jsonl"}`)
	if file.Code != http.StatusOK {
		t.Fatalf("file status = %d, body=%s", file.Code, file.Body.String())
	}
	var fileBody map[string]any
	if err := json.Unmarshal(file.Body.Bytes(), &fileBody); err != nil {
		t.Fatalf("decode file: %v", err)
	}
	if fileBody["id"] != "openai_file_000001" || fileBody["purpose"] != "batch" {
		t.Fatalf("file body = %#v", fileBody)
	}

	batch := serveOpenAIRequest(mux, http.MethodPost, "/openai/v1/batches", `{"input_file_id":"openai_file_000001","endpoint":"/v1/responses","completion_window":"24h"}`)
	if batch.Code != http.StatusOK || !strings.Contains(batch.Body.String(), `"input_file_id":"openai_file_000001"`) {
		t.Fatalf("batch status/body = %d %s", batch.Code, batch.Body.String())
	}
}

func TestOpenAIValidationAndInvalidModelErrors(t *testing.T) {
	mux := newOpenAIMux(t, adapter.Config{BasePath: "/openai", Scenario: "chat_success"})

	invalidModel := serveOpenAIRequest(mux, http.MethodPost, "/openai/v1/chat/completions", `{"model":"not-real","messages":[{"role":"user","content":"hi"}]}`)
	if invalidModel.Code != http.StatusBadRequest {
		t.Fatalf("invalid model status = %d, body=%s", invalidModel.Code, invalidModel.Body.String())
	}
	assertErrorCode(t, invalidModel, "model_not_found")

	malformed := serveOpenAIRequest(mux, http.MethodPost, "/openai/v1/chat/completions", `{"model":"gpt-mockport","messages":"bad"}`)
	if malformed.Code != http.StatusBadRequest {
		t.Fatalf("malformed status = %d, body=%s", malformed.Code, malformed.Body.String())
	}
	assertErrorCode(t, malformed, "invalid_request_error")

	unsupported := serveOpenAIRequest(mux, http.MethodPost, "/openai/v1/responses", `{"model":"gpt-mockport","input":"hello","mockport_unsupported_parameter":true}`)
	if unsupported.Code != http.StatusBadRequest {
		t.Fatalf("unsupported status = %d, body=%s", unsupported.Code, unsupported.Body.String())
	}
	assertErrorCode(t, unsupported, "unsupported_parameter")
}

func TestOpenAIRejectsOversizedJSONBody(t *testing.T) {
	mux := newOpenAIMux(t, adapter.Config{BasePath: "/openai", Scenario: "chat_success"})
	body := `{"model":"gpt-mockport","messages":[{"role":"user","content":"` + strings.Repeat("x", 1<<20) + `"}]}`

	rec := serveOpenAIRequest(mux, http.MethodPost, "/openai/v1/chat/completions", body)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
	}
	assertErrorCode(t, rec, "request_too_large")
}

func TestOpenAIRejectsMalformedFileJSON(t *testing.T) {
	mux := newOpenAIMux(t, adapter.Config{BasePath: "/openai", Scenario: "chat_success"})

	rec := serveOpenAIRequest(mux, http.MethodPost, "/openai/v1/files", `{"purpose":`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	assertErrorCode(t, rec, "invalid_json")
}

func TestOpenAIRejectsTrailingJSONToken(t *testing.T) {
	mux := newOpenAIMux(t, adapter.Config{BasePath: "/openai", Scenario: "chat_success"})

	rec := serveOpenAIRequest(mux, http.MethodPost, "/openai/v1/chat/completions", `{"model":"gpt-mockport","messages":[{"role":"user","content":"hi"}]}{}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	assertErrorCode(t, rec, "invalid_json")
}

func TestMetadata(t *testing.T) {
	meta := New().Metadata()
	if meta.Name != "openai" {
		t.Fatalf("name = %q", meta.Name)
	}
	if meta.Maturity != "workflow-compatible" {
		t.Fatalf("maturity = %q", meta.Maturity)
	}
	if meta.ProviderVersion != "2025-02-01" || len(meta.SDKVersions) != 1 {
		t.Fatalf("compat metadata = %#v", meta)
	}
	if len(meta.Scenarios) < 5 || len(meta.Endpoints) < 9 {
		t.Fatalf("metadata too small: %#v", meta)
	}
	if !meta.Reset || len(meta.StatefulResources) != 5 {
		t.Fatalf("state metadata = %#v", meta)
	}
}

func TestOpenAIResetClearsState(t *testing.T) {
	mux := newOpenAIMux(t, adapter.Config{BasePath: "/openai", Scenario: "chat_success"})

	response := serveOpenAIRequest(mux, http.MethodPost, "/openai/v1/responses", `{"model":"gpt-mockport","input":"hello"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("response create status = %d, body=%s", response.Code, response.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	responseID, _ := created["id"].(string)

	lookup := serveOpenAIRequest(mux, http.MethodGet, "/openai/v1/responses/"+responseID, "")
	if lookup.Code != http.StatusOK {
		t.Fatalf("response lookup before reset status = %d, body=%s", lookup.Code, lookup.Body.String())
	}

	reset := httptest.NewRequest(http.MethodPost, "/openai/test/reset", nil)
	reset.RemoteAddr = "127.0.0.1:12345"
	resetRec := httptest.NewRecorder()
	mux.ServeHTTP(resetRec, reset)
	if resetRec.Code != http.StatusOK {
		t.Fatalf("reset status = %d, body=%s", resetRec.Code, resetRec.Body.String())
	}
	var resetBody map[string]any
	if err := json.Unmarshal(resetRec.Body.Bytes(), &resetBody); err != nil {
		t.Fatalf("decode reset response: %v", err)
	}
	if resetBody["reset"] != true || resetBody["adapter"] != "openai" {
		t.Fatalf("reset body = %#v", resetBody)
	}

	lookupAfter := serveOpenAIRequest(mux, http.MethodGet, "/openai/v1/responses/"+responseID, "")
	if lookupAfter.Code != http.StatusNotFound {
		t.Fatalf("response lookup after reset status = %d, body=%s", lookupAfter.Code, lookupAfter.Body.String())
	}

	remoteReset := httptest.NewRequest(http.MethodPost, "/openai/test/reset", nil)
	remoteReset.RemoteAddr = "192.168.0.2:12345"
	remoteResetRec := httptest.NewRecorder()
	mux.ServeHTTP(remoteResetRec, remoteReset)
	if remoteResetRec.Code != http.StatusForbidden {
		t.Fatalf("remote reset status = %d, body=%s", remoteResetRec.Code, remoteResetRec.Body.String())
	}
	if !strings.Contains(remoteResetRec.Body.String(), "loopback") {
		t.Fatalf("remote reset body = %s", remoteResetRec.Body.String())
	}
}

func performRequest(t *testing.T, cfg adapter.Config, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	mux := newOpenAIMux(t, cfg)
	return serveOpenAIRequest(mux, method, path, "")
}

func newOpenAIMux(t *testing.T, cfg adapter.Config) *http.ServeMux {
	t.Helper()
	return adaptertest.NewMux(t, New(), cfg)
}

func serveOpenAIRequest(mux http.Handler, method, path, body string) *httptest.ResponseRecorder {
	header := http.Header{}
	if body != "" {
		header.Set("Content-Type", "application/json")
	}
	return adaptertest.Serve(mux, method, path, strings.NewReader(body), header)
}

func assertErrorCode(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()
	adaptertest.AssertJSONField(t, rec, "error.code", want)
}
