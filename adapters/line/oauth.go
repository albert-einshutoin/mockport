package line

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/albert-einshutoin/mockport/internal/adapter/httpx"
	"github.com/albert-einshutoin/mockport/internal/security"
	"github.com/albert-einshutoin/mockport/internal/state"
)

func (r *routes) writeAuthorize(w http.ResponseWriter, req *http.Request) {
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
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "redirect_uri must be a loopback URL in Mockport")
		return
	}
	scenario, ok := r.resolveScenarioOAuth(w, req)
	if !ok {
		return
	}
	if scenario == "invalid_request" {
		redirectWithQuery(w, req, redirectURI, map[string]string{
			"error":             "invalid_request",
			"error_description": "Mockport simulated invalid LINE Login request",
			"state":             req.URL.Query().Get("state"),
		})
		return
	}
	scope := req.URL.Query().Get("scope")
	if scope == "" {
		scope = "profile openid"
	}
	resource, err := r.store.Create("line", "oauth_code", map[string]any{
		"client_id":    clientID,
		"redirect_uri": redirectURI,
		"scope":        scope,
		"user_id":      "Umockport",
		"expires_at":   "2999-01-01T00:00:00Z",
	})
	if err != nil {
		writeLINEError(w, http.StatusInternalServerError, err.Error())
		return
	}
	redirectWithQuery(w, req, redirectURI, map[string]string{"code": resource.ID, "state": req.URL.Query().Get("state")})
}

func (r *routes) writeToken(w http.ResponseWriter, req *http.Request) {
	scenario, ok := r.resolveScenarioOAuth(w, req)
	if !ok {
		return
	}
	switch scenario {
	case "auth_error":
		writeOAuthError(w, http.StatusUnauthorized, "invalid_client", "Mockport simulated invalid LINE Login client")
	case "invalid_request":
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "Mockport simulated invalid token request")
	default:
		if err := req.ParseForm(); err != nil {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "Request body must be form encoded")
			return
		}
		code := req.Form.Get("code")
		if code == "" {
			writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "Authorization code is invalid or expired")
			return
		}
		codeResource, ok := r.store.Get("line", "oauth_code", code)
		if !ok {
			writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "Authorization code is invalid or expired")
			return
		}
		if want, _ := codeResource.Data["redirect_uri"].(string); want != "" && req.Form.Get("redirect_uri") != "" && req.Form.Get("redirect_uri") != want {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "redirect_uri does not match")
			return
		}
		if !clientIDMatches(codeResource, req.Form.Get("client_id")) {
			writeOAuthError(w, http.StatusUnauthorized, "invalid_client", "client_id does not match authorization request")
			return
		}
		codeResource, ok = r.store.Take("line", "oauth_code", code)
		if !ok {
			writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "Authorization code is invalid or expired")
			return
		}
		token, err := r.store.Create("line", "oauth_token", map[string]any{
			"client_id":  codeResource.Data["client_id"],
			"scope":      codeResource.Data["scope"],
			"user_id":    codeResource.Data["user_id"],
			"expires_at": "2999-01-01T00:00:00Z",
		})
		if err != nil {
			writeOAuthError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		body := map[string]any{
			"access_token":  token.ID,
			"expires_in":    2592000,
			"id_token":      "mockport.line.id-token",
			"refresh_token": "mockport_line_refresh_token",
			"scope":         codeResource.Data["scope"],
			"token_type":    "Bearer",
		}
		if req.Form.Get("grant_type") == "client_credentials" {
			body["key_id"] = "mockport_line_key_id"
		}
		httpx.WriteJSON(w, http.StatusOK, body)
	}
}

func (r *routes) writeStatelessToken(w http.ResponseWriter, req *http.Request) {
	scenario, ok := r.resolveScenarioOAuth(w, req)
	if !ok {
		return
	}
	if scenario == "auth_error" {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "Invalid 'client_credentials'.")
		return
	}
	if err := req.ParseForm(); err != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "Request body must be form encoded")
		return
	}
	token, err := r.store.Create("line", "oauth_token", map[string]any{
		"scope":      "profile chat_message.write",
		"user_id":    "Umockport",
		"client_id":  firstNonEmpty(req.Form.Get("client_id"), "mockport_line_channel"),
		"expires_in": 900,
	})
	if err != nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"token_type": "Bearer", "access_token": token.ID, "expires_in": 900})
}

func (r *routes) writeShortLivedToken(w http.ResponseWriter, req *http.Request) {
	scenario, ok := r.resolveScenarioOAuth(w, req)
	if !ok {
		return
	}
	if scenario == "auth_error" {
		writeOAuthError(w, http.StatusBadRequest, "invalid_client", "invalid client_secret")
		return
	}
	if err := req.ParseForm(); err != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "Request body must be form encoded")
		return
	}
	token, err := r.store.Create("line", "oauth_token", map[string]any{
		"scope":      "profile chat_message.write",
		"user_id":    "Umockport",
		"client_id":  firstNonEmpty(req.Form.Get("client_id"), "mockport_line_channel"),
		"expires_in": 2592000,
	})
	if err != nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"access_token": token.ID, "expires_in": 2592000, "token_type": "Bearer"})
}

func (r *routes) writeVerifyToken(w http.ResponseWriter, req *http.Request) {
	token := req.URL.Query().Get("access_token")
	if token == "" {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "The access token expired")
		return
	}
	resource, ok := r.store.Get("line", "oauth_token", token)
	if !ok {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "The access token expired")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"client_id":  firstNonEmpty(resource.Data["client_id"], "mockport_line_channel"),
		"expires_in": firstNonEmpty(resource.Data["expires_in"], 2592000),
		"scope":      firstNonEmpty(resource.Data["scope"], "profile chat_message.write"),
	})
}

func (r *routes) writeVerifyOAuthToken(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "Request body must be form encoded")
		return
	}
	token := req.Form.Get("access_token")
	if token == "" {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "access_token invalid")
		return
	}
	resource, ok := r.store.Get("line", "oauth_token", token)
	if !ok {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "access_token invalid")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"client_id":  firstNonEmpty(resource.Data["client_id"], "mockport_line_channel"),
		"expires_in": firstNonEmpty(resource.Data["expires_in"], 2592000),
		"scope":      firstNonEmpty(resource.Data["scope"], "profile chat_message.write"),
	})
}

func (r *routes) writeLoginProfile(w http.ResponseWriter, req *http.Request) {
	scenario, ok := r.resolveScenario(w, req)
	if !ok {
		return
	}
	if scenario == "auth_error" {
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated invalid access token")
		return
	}
	resource, ok := r.tokenResource(req)
	if !ok {
		writeLINEError(w, http.StatusUnauthorized, "Access token is invalid")
		return
	}
	userID, _ := resource.Data["user_id"].(string)
	httpx.WriteJSON(w, http.StatusOK, lineProfile(userID))
}

func (r *routes) writeLIFFProfile(w http.ResponseWriter, req *http.Request) {
	scenario, ok := r.resolveScenario(w, req)
	if !ok {
		return
	}
	if scenario == "auth_error" {
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated LIFF access token error")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, lineProfile("Umockport"))
}

func (r *routes) writeLIFFContext(w http.ResponseWriter) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"type":        "utou",
		"userId":      "Umockport",
		"viewType":    "full",
		"endpointUrl": "http://localhost:3000",
		"liffId":      "mockport-line-liff",
	})
}

func (r *routes) tokenResource(req *http.Request) (state.Resource, bool) {
	token := bearerToken(req)
	if token == "" {
		return state.Resource{}, false
	}
	return r.store.Get("line", "oauth_token", token)
}

func redirectWithQuery(w http.ResponseWriter, req *http.Request, redirectURI string, values map[string]string) {
	target, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
		return
	}
	query := target.Query()
	for key, value := range values {
		if value != "" {
			query.Set(key, value)
		}
	}
	target.RawQuery = query.Encode()
	http.Redirect(w, req, target.String(), http.StatusFound)
}

func bearerToken(r *http.Request) string {
	value := r.Header.Get("Authorization")
	if !strings.HasPrefix(value, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(value, "Bearer "))
}

func clientIDMatches(resource state.Resource, got string) bool {
	want, _ := resource.Data["client_id"].(string)
	return strings.TrimSpace(want) != "" && strings.TrimSpace(got) != "" && got == want
}
