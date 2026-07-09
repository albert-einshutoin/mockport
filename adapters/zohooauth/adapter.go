package zohooauth

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/adapter/httpx"
	"github.com/albert-einshutoin/mockport/internal/security"
	"github.com/albert-einshutoin/mockport/internal/state"
)

const adapterName = "zoho-oauth"

// authScheme is the Zoho-specific Authorization scheme used by the user info
// endpoint. Zoho uses "Zoho-oauthtoken" instead of "Bearer".
const authScheme = "Zoho-oauthtoken "

const (
	defaultUserEmail = "mockport@example.test"
	defaultUserName  = "Mockport User"
)

type Adapter struct{}

func New() Adapter { return Adapter{} }

func (a Adapter) Name() string { return "zoho-oauth" }

func (a Adapter) Register(mux *http.ServeMux, cfg adapter.Config) error {
	basePath := resolveBasePath(cfg.BasePath)
	r := &routes{basePath: basePath, cfg: cfg, store: state.NewStore()}
	mux.HandleFunc(basePath+"/", r.handle)
	return nil
}

func (a Adapter) FakeEnv(cfg adapter.Config) map[string]string {
	basePath := resolveBasePath(cfg.BasePath)
	return map[string]string{
		"ZOHO_AUTH_BASE_URL":       adapter.LocalBaseURL(basePath),
		"ZOHO_OAUTH_CLIENT_ID":     "mockport_zoho_client",
		"ZOHO_OAUTH_CLIENT_SECRET": "mockport_zoho_secret",
		"ZOHO_USER_EMAIL":          defaultUserEmail,
		"ZOHO_USER_NAME":           defaultUserName,
	}
}

func (a Adapter) Metadata() adapter.Metadata {
	return adapter.Metadata{
		Name:            adapterName,
		Maturity:        adapter.MaturityWorkflowCompatible,
		ProviderVersion: "oauth-v2",
		ClientEvidence:  []string{"oauth-client-contract"},
		Levels:          []adapter.Level{adapter.LevelWire, adapter.LevelClient, adapter.LevelWorkflow, adapter.LevelState, adapter.LevelError},
		Capabilities:    []string{"oauth_authorize", "oauth_token", "user_profile"},
		StatefulResources: []string{
			"oauth_code",
			"oauth_token",
		},
		Reset: true,
		Scenarios: []adapter.Scenario{
			{Name: "oauth_success", Supported: true},
			{Name: "invalid_code", Supported: true, Category: adapter.ScenarioCategoryError},
			{Name: "invalid_token", Supported: true, Category: adapter.ScenarioCategoryError},
		},
		Endpoints: []adapter.Endpoint{
			{Method: http.MethodGet, Path: "/zoho/oauth/v2/auth", SupportedScenarios: []string{"oauth_success"}, Notes: "Zoho OAuth-like authorize redirect (no login screen)"},
			{Method: http.MethodPost, Path: "/zoho/oauth/v2/token", SupportedScenarios: []string{"oauth_success", "invalid_code"}, Notes: "Zoho OAuth-like token exchange"},
			{Method: http.MethodGet, Path: "/zoho/oauth/user/info", SupportedScenarios: []string{"oauth_success", "invalid_token"}, Notes: "Zoho-like user info (Zoho-oauthtoken auth scheme)"},
			{Method: http.MethodPost, Path: "/zoho/test/reset", SupportedScenarios: []string{"oauth_success", "invalid_code", "invalid_token"}, Notes: "Clears state for test isolation"},
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
	case req.Method == http.MethodGet && path == "/oauth/v2/auth":
		r.authorize(w, req)
	case req.Method == http.MethodPost && path == "/oauth/v2/token":
		r.token(w, req)
	case req.Method == http.MethodGet && path == "/oauth/user/info":
		r.userInfo(w, req)
	case req.Method == http.MethodPost && path == "/test/reset":
		r.handleReset(w, req)
	default:
		http.NotFound(w, req)
	}
}

// handleReset clears adapter state for test isolation. It is restricted to
// loopback callers, matching the other adapters' reset endpoints.
func (r *routes) handleReset(w http.ResponseWriter, req *http.Request) {
	if !security.IsLoopbackRemoteAddr(req.RemoteAddr) {
		writeError(w, http.StatusForbidden, "loopback_required")
		return
	}
	resourceTypes := r.store.ResetAll(adapterName)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"reset":          true,
		"adapter":        adapterName,
		"resource_types": resourceTypes,
	})
}

// authorize emulates GET /oauth/v2/auth: it never shows a login screen and
// immediately redirects to redirect_uri with a generated code and the echoed
// state.
func (r *routes) authorize(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	clientID := strings.TrimSpace(q.Get("client_id"))
	if clientID == "" {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	redirectURI := q.Get("redirect_uri")
	if strings.TrimSpace(redirectURI) == "" {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	if !security.IsSafeOAuthRedirectURL(redirectURI) {
		writeError(w, http.StatusBadRequest, "invalid_redirect_uri")
		return
	}

	resource, err := r.store.Create(adapterName, "oauth_code", map[string]any{
		"client_id":    clientID,
		"redirect_uri": redirectURI,
		"email":        firstNonEmpty(q.Get("mock_email"), os.Getenv("ZOHO_USER_EMAIL"), defaultUserEmail),
		"name":         firstNonEmpty(q.Get("mock_name"), os.Getenv("ZOHO_USER_NAME"), defaultUserName),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error")
		return
	}
	redirectWithQuery(w, req, redirectURI, map[string]string{
		"code":  resource.ID,
		"state": q.Get("state"),
	})
}

// token emulates POST /oauth/v2/token. On success it returns {"access_token": ...}.
// On failure (bad grant_type or unknown/invalid code) it returns {"error": ...}.
// Zoho returns HTTP 200 even for these errors; the client inspects the error
// field, not the status code.
func (r *routes) token(w http.ResponseWriter, req *http.Request) {
	if normalizeScenario(r.cfg.Scenario) == "invalid_code" {
		writeTokenError(w, "invalid_code")
		return
	}
	if !parseForm(w, req) {
		return
	}
	if req.Form.Get("grant_type") != "authorization_code" {
		writeTokenError(w, "invalid_grant_type")
		return
	}
	code := req.Form.Get("code")
	if strings.TrimSpace(code) == "" {
		writeTokenError(w, "invalid_code")
		return
	}
	// Atomically consume the authorization code first so it cannot be exchanged
	// twice even under concurrent requests, then validate the consumed code.
	// A failed validation still spends the code, matching real OAuth one-time
	// code semantics.
	codeResource, ok := r.store.Take(adapterName, "oauth_code", code)
	if !ok {
		writeTokenError(w, "invalid_code")
		return
	}
	if want, _ := codeResource.Data["redirect_uri"].(string); want != "" && req.Form.Get("redirect_uri") != "" && req.Form.Get("redirect_uri") != want {
		writeTokenError(w, "redirect_uri_mismatch")
		return
	}
	if !clientIDMatches(codeResource, req.Form.Get("client_id")) {
		writeTokenError(w, "invalid_client")
		return
	}
	token, err := r.store.Create(adapterName, "oauth_token", map[string]any{
		"client_id": codeResource.Data["client_id"],
		"email":     codeResource.Data["email"],
		"name":      codeResource.Data["name"],
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, tokenResponse{AccessToken: token.ID})
}

// userInfo emulates GET /oauth/user/info. It requires the Zoho-specific
// "Authorization: Zoho-oauthtoken <access_token>" header and returns the
// configured Email/Display_Name on success.
func (r *routes) userInfo(w http.ResponseWriter, req *http.Request) {
	if normalizeScenario(r.cfg.Scenario) == "invalid_token" {
		writeError(w, http.StatusUnauthorized, "invalid_token")
		return
	}
	token := zohoOAuthToken(req)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "invalid_token")
		return
	}
	resource, ok := r.store.Get(adapterName, "oauth_token", token)
	if !ok {
		writeError(w, http.StatusUnauthorized, "invalid_token")
		return
	}
	email, _ := resource.Data["email"].(string)
	name, _ := resource.Data["name"].(string)
	httpx.WriteJSON(w, http.StatusOK, userInfoResponse{Email: email, DisplayName: name})
}

// zohoOAuthToken extracts the token from "Authorization: Zoho-oauthtoken <token>".
// It returns "" when the scheme is missing or not Zoho-oauthtoken (e.g. Bearer).
func zohoOAuthToken(req *http.Request) string {
	value := strings.TrimSpace(req.Header.Get("Authorization"))
	if !strings.HasPrefix(value, authScheme) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(value, authScheme))
}

func resolveBasePath(basePath string) string {
	if basePath == "" {
		basePath = "/zoho"
	}
	return strings.TrimRight(basePath, "/")
}

func normalizeScenario(s string) string {
	if s == "" {
		return "oauth_success"
	}
	return s
}

func clientIDMatches(resource state.Resource, got string) bool {
	want, _ := resource.Data["client_id"].(string)
	return strings.TrimSpace(want) != "" && strings.TrimSpace(got) != "" && got == want
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func redirectWithQuery(w http.ResponseWriter, req *http.Request, redirectURI string, values map[string]string) {
	parsed, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
		return
	}
	query := parsed.Query()
	for key, value := range values {
		if value != "" {
			query.Set(key, value)
		}
	}
	parsed.RawQuery = query.Encode()
	http.Redirect(w, req, parsed.String(), http.StatusFound)
}

func writeTokenError(w http.ResponseWriter, reason string) {
	// Zoho returns HTTP 200 with an error field for token-exchange failures.
	httpx.WriteJSON(w, http.StatusOK, errorResponse{Error: reason})
}

func writeError(w http.ResponseWriter, status int, reason string) {
	httpx.WriteJSON(w, status, errorResponse{Error: reason})
}

func parseForm(w http.ResponseWriter, req *http.Request) bool {
	if err := req.ParseForm(); err != nil {
		if httpx.IsRequestBodyTooLarge(err) {
			writeError(w, http.StatusRequestEntityTooLarge, "request_too_large")
			return false
		}
		writeError(w, http.StatusBadRequest, "invalid_request")
		return false
	}
	return true
}
