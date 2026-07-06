package line

import (
	"net/http"
	"strings"
	"sync"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/state"
)

type Adapter struct{}

func New() Adapter { return Adapter{} }

func (a Adapter) Name() string { return "line" }

func (a Adapter) Register(mux *http.ServeMux, cfg adapter.Config) error {
	basePath := cfg.BasePath
	if basePath == "" {
		basePath = "/line"
	}
	r := &routes{
		basePath: strings.TrimRight(basePath, "/"),
		cfg:      cfg,
		store:    state.NewStore(),
		resolver: adapter.NewScenarioResolver(cfg, "line_success", a.Metadata()),
	}
	mux.HandleFunc(r.basePath+"/", r.handle)
	return nil
}

func (a Adapter) FakeEnv(cfg adapter.Config) map[string]string {
	basePath := cfg.BasePath
	if basePath == "" {
		basePath = "/line"
	}
	return map[string]string{
		"LINE_API_BASE_URL":        adapter.LocalBaseURL(basePath),
		"LINE_CHANNEL_ID":          "mockport_line_channel",
		"LINE_CHANNEL_SECRET":      "mockport_line_secret",
		"LINE_CHANNEL_TOKEN":       "mockport_line_channel_token",
		"LINE_LIFF_ID":             "mockport-line-liff",
		"LINE_PAY_CHANNEL_ID":      "mockport_line_pay_channel",
		"LINE_PAY_CHANNEL_SECRET":  "mockport_line_pay_secret",
		"LINE_MINI_DAPP_CLIENT_ID": "mockport_line_mini_dapp_client",
	}
}

type routes struct {
	basePath string
	cfg      adapter.Config
	store    *state.Store
	resolver *adapter.ScenarioResolver

	// mu guards the singleton mutable state below. net/http dispatches
	// concurrent requests to the same routes instance, so these fields must
	// not be touched without holding the lock.
	mu                sync.RWMutex
	webhookEndpoint   string
	defaultRichMenuID string
}

func (r *routes) setWebhookEndpoint(endpoint string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.webhookEndpoint = endpoint
}

func (r *routes) getWebhookEndpoint() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.webhookEndpoint
}

func (r *routes) setDefaultRichMenu(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultRichMenuID = id
}

// deleteRichMenu removes a rich menu and clears the default pointer in the same
// critical section, so the default can never be left referencing a menu that no
// longer exists in the store.
func (r *routes) deleteRichMenu(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.store.Delete("line", "rich_menu", id) {
		return false
	}
	if r.defaultRichMenuID == id {
		r.defaultRichMenuID = ""
	}
	return true
}

func (r *routes) resetSingletonState() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.webhookEndpoint = ""
	r.defaultRichMenuID = ""
}

// setDefaultRichMenuChecked validates that the rich menu still exists with an
// uploaded image and adopts it as the default within a single lock, closing the
// check-then-set race against a concurrent delete. On success it returns
// http.StatusOK with an empty message.
func (r *routes) setDefaultRichMenuChecked(id string) (status int, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	resource, ok := r.store.Get("line", "rich_menu", id)
	if !ok {
		return http.StatusNotFound, "Not found"
	}
	if resource.Data["hasImage"] != true {
		return http.StatusBadRequest, "must upload richmenu image before applying it to user"
	}
	r.defaultRichMenuID = id
	return http.StatusOK, ""
}

// currentDefaultRichMenu returns the default rich menu id only while it still
// exists in the store, so a concurrently deleted menu is never reported.
func (r *routes) currentDefaultRichMenu() (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.defaultRichMenuID == "" {
		return "", false
	}
	if _, ok := r.store.Get("line", "rich_menu", r.defaultRichMenuID); !ok {
		return "", false
	}
	return r.defaultRichMenuID, true
}
