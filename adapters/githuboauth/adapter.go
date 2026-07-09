package githuboauth

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/adapter/httpx"
	"github.com/albert-einshutoin/mockport/internal/security"
	"github.com/albert-einshutoin/mockport/internal/state"
)

type Adapter struct{}

func New() Adapter { return Adapter{} }

func (a Adapter) Name() string { return "github-oauth" }

func (a Adapter) Register(mux *http.ServeMux, cfg adapter.Config) error {
	basePath := cfg.BasePath
	if basePath == "" {
		basePath = "/github"
	}
	r := &routes{basePath: strings.TrimRight(basePath, "/"), cfg: cfg, store: state.NewStore()}
	mux.HandleFunc(r.basePath+"/", r.handle)
	return nil
}

func (a Adapter) FakeEnv(cfg adapter.Config) map[string]string {
	basePath := cfg.BasePath
	if basePath == "" {
		basePath = "/github"
	}
	return map[string]string{
		"GITHUB_OAUTH_BASE_URL":      adapter.LocalBaseURL(basePath),
		"GITHUB_OAUTH_CLIENT_ID":     "mockport_github_client",
		"GITHUB_OAUTH_CLIENT_SECRET": "mockport_github_secret",
	}
}

func (a Adapter) Metadata() adapter.Metadata {
	return adapter.Metadata{
		Name:            "github-oauth",
		Maturity:        adapter.MaturityWorkflowCompatible,
		ProviderVersion: "2022-11-28",
		ClientEvidence:  []string{"oauth-client-contract"},
		Levels:          []adapter.Level{adapter.LevelWire, adapter.LevelClient, adapter.LevelWorkflow, adapter.LevelState, adapter.LevelError},
		Capabilities:    []string{"oauth_authorize", "oauth_token", "user_profile", "user_emails", "user_orgs"},
		StatefulResources: []string{
			"oauth_code",
			"oauth_token",
			"user_identity",
		},
		Reset: true,
		Scenarios: []adapter.Scenario{
			{Name: "oauth_success", Supported: true},
			{Name: "invalid_code", Supported: true},
			{Name: "expired_token", Supported: true},
			{Name: "scope_missing", Supported: true},
			{Name: "redirect_uri_mismatch", Supported: true},
		},
		Endpoints: []adapter.Endpoint{
			{Method: http.MethodGet, Path: "/github/login/oauth/authorize", SupportedScenarios: []string{"oauth_success"}, Notes: "GitHub OAuth-like authorize redirect"},
			{Method: http.MethodPost, Path: "/github/login/oauth/access_token", SupportedScenarios: []string{"oauth_success", "invalid_code", "expired_token", "scope_missing", "redirect_uri_mismatch"}, Notes: "GitHub OAuth-like token exchange"},
			{Method: http.MethodGet, Path: "/github/user", SupportedScenarios: []string{"oauth_success", "expired_token", "scope_missing"}, Notes: "GitHub-like user profile"},
			{Method: http.MethodGet, Path: "/github/user/emails", SupportedScenarios: []string{"oauth_success", "scope_missing"}, Notes: "GitHub-like user emails subset"},
			{Method: http.MethodGet, Path: "/github/user/orgs", SupportedScenarios: []string{"oauth_success", "scope_missing"}, Notes: "GitHub-like user orgs subset"},
			{Method: http.MethodPost, Path: "/github/test/reset", SupportedScenarios: []string{"oauth_success", "invalid_code", "expired_token", "scope_missing", "redirect_uri_mismatch"}, Notes: "Clears state for test isolation"},
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
	case req.Method == http.MethodGet && path == "/login/oauth/authorize":
		clientID := req.URL.Query().Get("client_id")
		if strings.TrimSpace(clientID) == "" {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "client_id is required")
			return
		}
		redirectURI := req.URL.Query().Get("redirect_uri")
		if redirectURI == "" {
			redirectURI = "http://localhost/callback"
		}
		if !security.IsSafeOAuthRedirectURL(redirectURI) {
			writeOAuthError(w, http.StatusBadRequest, "redirect_uri_mismatch", "redirect_uri must be a loopback URL in Mockport")
			return
		}
		scope := req.URL.Query().Get("scope")
		if unsupportedScope(scope) {
			redirectWithQuery(w, req, redirectURI, map[string]string{
				"error":             "unsupported_scope",
				"error_description": "Mockport simulated unsupported scope",
				"state":             req.URL.Query().Get("state"),
			})
			return
		}
		code, err := r.createCode(req, clientID, redirectURI)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, err.Error())
			return
		}
		redirectWithQuery(w, req, redirectURI, map[string]string{"code": code, "state": req.URL.Query().Get("state")})
	case req.Method == http.MethodPost && path == "/login/oauth/access_token":
		r.writeToken(w, req)
	case req.Method == http.MethodGet && path == "/user":
		r.writeUser(w, req)
	case req.Method == http.MethodGet && path == "/user/emails":
		r.writeEmails(w, req)
	case req.Method == http.MethodGet && path == "/user/orgs":
		r.writeOrgs(w, req)
	case req.Method == http.MethodPost && path == "/test/reset":
		r.handleReset(w, req)
	default:
		http.NotFound(w, req)
	}
}

func (r *routes) handleReset(w http.ResponseWriter, req *http.Request) {
	if !security.IsLoopbackRemoteAddr(req.RemoteAddr) {
		writeAPIError(w, http.StatusForbidden, "github reset can only be triggered from loopback")
		return
	}
	resourceTypes := r.store.ResetAll("github-oauth")
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"reset":          true,
		"adapter":        "github-oauth",
		"resource_types": resourceTypes,
	})
}

func (r *routes) createCode(req *http.Request, clientID, redirectURI string) (string, error) {
	scope := req.URL.Query().Get("scope")
	if scope == "" {
		scope = "read:user"
	}
	resource, err := r.store.Create("github-oauth", "oauth_code", map[string]any{
		"client_id":    clientID,
		"redirect_uri": redirectURI,
		"scope":        scope,
		"user":         "mockport-user",
		"expires_at":   "2999-01-01T00:00:00Z",
	})
	if err != nil {
		return "", err
	}
	return resource.ID, nil
}

func (r *routes) writeToken(w http.ResponseWriter, req *http.Request) {
	switch normalizeScenario(r.cfg.Scenario) {
	case "invalid_code":
		writeOAuthError(w, http.StatusBadRequest, "bad_verification_code", "The code passed is incorrect or expired.")
	case "expired_token":
		writeOAuthError(w, http.StatusUnauthorized, "expired_token", "Mockport simulated expired token.")
	case "scope_missing":
		writeOAuthError(w, http.StatusForbidden, "scope_missing", "Mockport simulated missing scope.")
	case "redirect_uri_mismatch":
		writeOAuthError(w, http.StatusBadRequest, "redirect_uri_mismatch", "The redirect_uri does not match the authorization request.")
	default:
		if !parseOAuthForm(w, req) {
			return
		}
		code := req.Form.Get("code")
		if code == "" {
			writeOAuthError(w, http.StatusBadRequest, "bad_verification_code", "The code passed is incorrect or expired.")
			return
		}
		codeResource, ok := r.store.Get("github-oauth", "oauth_code", code)
		if !ok {
			writeOAuthError(w, http.StatusBadRequest, "bad_verification_code", "The code passed is incorrect or expired.")
			return
		}
		if want, _ := codeResource.Data["redirect_uri"].(string); want != "" && req.Form.Get("redirect_uri") != "" && req.Form.Get("redirect_uri") != want {
			writeOAuthError(w, http.StatusBadRequest, "redirect_uri_mismatch", "The redirect_uri does not match the authorization request.")
			return
		}
		if !clientIDMatches(codeResource, req.Form.Get("client_id")) {
			writeOAuthError(w, http.StatusUnauthorized, "incorrect_client_credentials", "The client_id does not match the authorization request.")
			return
		}
		codeResource, ok = r.store.Take("github-oauth", "oauth_code", code)
		if !ok {
			writeOAuthError(w, http.StatusBadRequest, "bad_verification_code", "The code passed is incorrect or expired.")
			return
		}
		token, err := r.store.Create("github-oauth", "oauth_token", map[string]any{
			"client_id":  codeResource.Data["client_id"],
			"scope":      codeResource.Data["scope"],
			"user":       codeResource.Data["user"],
			"expires_at": "2999-01-01T00:00:00Z",
		})
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, err.Error())
			return
		}
		httpx.WriteJSON(w, http.StatusOK, tokenResponse{AccessToken: token.ID, TokenType: "bearer", Scope: codeResource.Data["scope"]})
	}
}

func (r *routes) writeUser(w http.ResponseWriter, req *http.Request) {
	switch normalizeScenario(r.cfg.Scenario) {
	case "expired_token":
		writeAPIError(w, http.StatusUnauthorized, "Bad credentials")
	case "scope_missing":
		writeAPIError(w, http.StatusForbidden, "Resource not accessible by token")
	default:
		resource, ok := r.tokenResource(req)
		if !ok {
			writeAPIError(w, http.StatusUnauthorized, "Bad credentials")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, userResponse{
			Login: "mockport-user",
			ID:    43101,
			Name:  "Mockport User",
			Email: "mockport@example.test",
			Scope: resource.Data["scope"],
		})
	}
}

func (r *routes) writeEmails(w http.ResponseWriter, req *http.Request) {
	resource, ok := r.tokenResource(req)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "Bad credentials")
		return
	}
	if !hasScope(resource.Data["scope"], "user:email") {
		writeAPIError(w, http.StatusForbidden, "Resource not accessible by token")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, []emailResponse{{Email: "mockport@example.test", Primary: true, Verified: true, Visibility: "public"}})
}

func (r *routes) writeOrgs(w http.ResponseWriter, req *http.Request) {
	resource, ok := r.tokenResource(req)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "Bad credentials")
		return
	}
	if !hasScope(resource.Data["scope"], "read:org") {
		writeAPIError(w, http.StatusForbidden, "Resource not accessible by token")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, []orgResponse{{Login: "mockport-org", ID: 431010, Description: "Mockport fake organization"}})
}

func (r *routes) tokenResource(req *http.Request) (state.Resource, bool) {
	token := bearerToken(req)
	if token == "" {
		return state.Resource{}, false
	}
	return r.store.Get("github-oauth", "oauth_token", token)
}

func bearerToken(r *http.Request) string {
	value := r.Header.Get("Authorization")
	return strings.TrimSpace(strings.TrimPrefix(value, "Bearer "))
}

func unsupportedScope(value string) bool {
	for _, scope := range splitScopes(value) {
		switch scope {
		case "", "read:user", "user:email", "read:org":
		default:
			return true
		}
	}
	return false
}

func hasScope(value any, want string) bool {
	for _, scope := range splitScopes(value) {
		if scope == want {
			return true
		}
	}
	return false
}

func splitScopes(value any) []string {
	scope, _ := value.(string)
	return strings.FieldsFunc(scope, func(r rune) bool {
		return r == ' ' || r == ','
	})
}

func clientIDMatches(resource state.Resource, got string) bool {
	want, _ := resource.Data["client_id"].(string)
	return strings.TrimSpace(want) != "" && strings.TrimSpace(got) != "" && got == want
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

func normalizeScenario(s string) string {
	if s == "" {
		return "oauth_success"
	}
	return s
}

func writeOAuthError(w http.ResponseWriter, status int, code, description string) {
	httpx.WriteJSON(w, status, oauthErrorResponse{
		Error:            code,
		ErrorDescription: description,
		ErrorURI:         "https://docs.github.com/apps/oauth-apps/maintaining-oauth-apps/troubleshooting-oauth-app-access-token-request-errors",
	})
}

func writeAPIError(w http.ResponseWriter, status int, message string) {
	httpx.WriteJSON(w, status, apiErrorResponse{
		Message:          message,
		DocumentationURL: "https://docs.github.com/rest",
		Status:           http.StatusText(status),
	})
}

func parseOAuthForm(w http.ResponseWriter, req *http.Request) bool {
	if err := req.ParseForm(); err != nil {
		if httpx.IsRequestBodyTooLarge(err) {
			writeOAuthError(w, http.StatusRequestEntityTooLarge, "request_too_large", "Request body is too large.")
			return false
		}
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "Request body is invalid.")
		return false
	}
	return true
}
