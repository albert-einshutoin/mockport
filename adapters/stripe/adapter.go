package stripe

import (
	"net/http"
	"strings"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/state"
)

type Adapter struct{}

func New() Adapter {
	return Adapter{}
}

func (a Adapter) Name() string {
	return "stripe"
}

func (a Adapter) Register(mux *http.ServeMux, cfg adapter.Config) error {
	basePath := cfg.BasePath
	if basePath == "" {
		basePath = "/stripe"
	}
	meta := a.Metadata()
	rt := &routes{
		basePath:    strings.TrimRight(basePath, "/"),
		cfg:         cfg,
		store:       state.NewStore(),
		idempotency: state.NewIdempotencyStore(),
		resolver:    adapter.NewScenarioResolver(cfg, scenarioPaymentSuccess, meta),
	}
	if rt.basePath == "" {
		rt.register(mux, "")
		return nil
	}
	rt.register(mux, rt.basePath)
	// Stripe SDK clients may hit the API root directly, but root-level test
	// helpers would collide with other adapters' /test endpoints.
	rt.registerV1Routes(mux, "")
	return nil
}

func (a Adapter) FakeEnv(cfg adapter.Config) map[string]string {
	basePath := cfg.BasePath
	if basePath == "" {
		basePath = "/stripe"
	}
	secret := cfg.FakeSecret
	if secret == "" {
		secret = "mockport_stripe_secret"
	}
	signingSecret := cfg.WebhookSigningSecret
	if signingSecret == "" {
		signingSecret = "whsec_mockport"
	}
	return map[string]string{
		"STRIPE_API_URL":        adapter.LocalBaseURL(basePath),
		"STRIPE_SECRET_KEY":     secret,
		"STRIPE_WEBHOOK_SECRET": signingSecret,
	}
}

func (a Adapter) Metadata() adapter.Metadata {
	scenarios := []adapter.Scenario{
		{Name: "payment_success", Supported: true},
		{Name: "payment_failed", Supported: true},
		{Name: "auth_error", Supported: true},
		{Name: "rate_limited", Supported: true},
		{Name: "timeout", Supported: true},
	}
	scenarioNames := []string{"payment_success", "payment_failed", "auth_error", "rate_limited", "timeout"}
	return adapter.Metadata{
		Name:            "stripe",
		Maturity:        adapter.MaturityWorkflowCompatible,
		ProviderVersion: "2025-10-29.clover",
		SDKVersions:     []adapter.SDKVersion{{Name: "stripe", Version: "22.3.0"}},
		Levels:          []adapter.Level{adapter.LevelWire, adapter.LevelSDK, adapter.LevelWorkflow, adapter.LevelState, adapter.LevelError},
		Capabilities: []string{
			"checkout_sessions",
			"payment_intents",
			"customers",
			"products",
			"prices",
			"subscriptions",
			"invoices",
			"refunds",
			"webhooks",
		},
		Scenarios: scenarios,
		StatefulResources: []string{
			"checkout_session",
			"payment_intent",
			"customer",
			"product",
			"price",
			"subscription",
			"invoice",
			"refund",
		},
		Idempotency: true,
		Reset:       true,
		Endpoints: []adapter.Endpoint{
			{Method: http.MethodPost, Path: "/stripe/v1/checkout/sessions", SupportedScenarios: scenarioNames, Notes: "Stripe-like checkout session creation"},
			{Method: http.MethodGet, Path: "/stripe/v1/checkout/sessions", SupportedScenarios: []string{"payment_success"}, Notes: "Deterministic checkout session list"},
			{Method: http.MethodGet, Path: "/stripe/v1/checkout/sessions/{id}", SupportedScenarios: []string{"payment_success"}, Notes: "Deterministic checkout session lookup"},
			{Method: http.MethodPost, Path: "/stripe/v1/payment_intents", SupportedScenarios: scenarioNames, Notes: "Stripe-like payment intent creation"},
			{Method: http.MethodGet, Path: "/stripe/v1/payment_intents", SupportedScenarios: []string{"payment_success"}, Notes: "Deterministic payment intent list"},
			{Method: http.MethodGet, Path: "/stripe/v1/payment_intents/{id}", SupportedScenarios: []string{"payment_success"}, Notes: "Deterministic payment intent lookup"},
			{Method: http.MethodPost, Path: "/stripe/v1/customers", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like customer creation"},
			{Method: http.MethodGet, Path: "/stripe/v1/customers", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like customer list"},
			{Method: http.MethodGet, Path: "/stripe/v1/customers/{id}", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like customer lookup"},
			{Method: http.MethodPost, Path: "/stripe/v1/products", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like product creation"},
			{Method: http.MethodGet, Path: "/stripe/v1/products", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like product list"},
			{Method: http.MethodGet, Path: "/stripe/v1/products/{id}", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like product lookup"},
			{Method: http.MethodPost, Path: "/stripe/v1/prices", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like price creation"},
			{Method: http.MethodGet, Path: "/stripe/v1/prices", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like price list"},
			{Method: http.MethodGet, Path: "/stripe/v1/prices/{id}", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like price lookup"},
			{Method: http.MethodPost, Path: "/stripe/v1/subscriptions", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like subscription creation"},
			{Method: http.MethodGet, Path: "/stripe/v1/subscriptions", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like subscription list"},
			{Method: http.MethodGet, Path: "/stripe/v1/subscriptions/{id}", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like subscription lookup"},
			{Method: http.MethodPost, Path: "/stripe/v1/invoices", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like invoice creation"},
			{Method: http.MethodGet, Path: "/stripe/v1/invoices", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like invoice list"},
			{Method: http.MethodGet, Path: "/stripe/v1/invoices/{id}", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like invoice lookup"},
			{Method: http.MethodPost, Path: "/stripe/v1/refunds", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like refund creation"},
			{Method: http.MethodGet, Path: "/stripe/v1/refunds", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like refund list"},
			{Method: http.MethodGet, Path: "/stripe/v1/refunds/{id}", SupportedScenarios: []string{"payment_success"}, Notes: "Stripe-like refund lookup"},
			{Method: http.MethodPost, Path: "/stripe/test/webhook/send", SupportedScenarios: []string{"payment_success", "payment_failed"}, Notes: "Sends fake signed webhook to configured target"},
			{Method: http.MethodPost, Path: "/stripe/test/reset", SupportedScenarios: []string{"payment_success", "payment_failed", "auth_error", "rate_limited", "timeout"}, Notes: "Clears state and idempotency records for test isolation"},
		},
	}
}
