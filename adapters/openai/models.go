package openai

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type chatCompletion struct {
	Object  string       `json:"object"`
	Choices []chatChoice `json:"choices"`
	Model   any          `json:"model,omitempty"`
	Status  string       `json:"status,omitempty"`
	Output  []outputItem `json:"output,omitempty"`
}

type chatChoice struct {
	Index   int         `json:"index"`
	Message chatMessage `json:"message"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseBody struct {
	Object     string       `json:"object"`
	Choices    []chatChoice `json:"choices"`
	OutputText string       `json:"output_text"`
	Status     string       `json:"status"`
	Output     []outputItem `json:"output"`
	Model      any          `json:"model,omitempty"`
}

type outputItem struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Status  string          `json:"status"`
	Role    string          `json:"role"`
	Content []outputContent `json:"content"`
}

type outputContent struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	Annotations []any  `json:"annotations"`
}

type embeddingResponse struct {
	ID     string          `json:"id"`
	Object string          `json:"object"`
	Data   []embeddingData `json:"data"`
	Model  any             `json:"model"`
	Usage  usage           `json:"usage"`
}

type embeddingData struct {
	Object    string `json:"object"`
	Index     int    `json:"index"`
	Embedding any    `json:"embedding"`
}

type usage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type chatCompletionChunk struct {
	ID                string                      `json:"id"`
	Object            string                      `json:"object"`
	Created           int64                       `json:"created"`
	Model             string                      `json:"model"`
	SystemFingerprint string                      `json:"system_fingerprint"`
	Choices           []chatCompletionChunkChoice `json:"choices"`
}

type chatCompletionChunkChoice struct {
	Index        int                      `json:"index"`
	Delta        chatCompletionChunkDelta `json:"delta"`
	FinishReason *string                  `json:"finish_reason"`
}

type chatCompletionChunkDelta struct {
	Role    string  `json:"role,omitempty"`
	Content *string `json:"content,omitempty"`
}
