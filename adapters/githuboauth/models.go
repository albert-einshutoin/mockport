package githuboauth

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       any    `json:"scope"`
}

type oauthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorURI         string `json:"error_uri"`
}

type apiErrorResponse struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
	Status           string `json:"status"`
	// Error は Mockport 固有の機械可読なコードフィールド。実際の GitHub REST API の
	// エラー本文には存在しないため omitempty で通常応答には出さず、
	// unknown_mockport_scenario のような Mockport 共通コードを伝える場合のみ付与する。
	Error string `json:"error,omitempty"`
}

type userResponse struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Scope any    `json:"scope"`
}

type emailResponse struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

type orgResponse struct {
	Login       string `json:"login"`
	ID          int    `json:"id"`
	Description string `json:"description"`
}
