package stripe

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/albert-einshutoin/mockport/internal/adapter"
)

func TestStripeListResponseEnvelopeInvariants(t *testing.T) {
	mux := newStripeMux(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_success"})

	tests := []struct {
		name string
		path string
	}{
		{name: "checkout sessions", path: "/stripe/v1/checkout/sessions"},
		{name: "payment intents", path: "/stripe/v1/payment_intents"},
		{name: "customers", path: "/stripe/v1/customers"},
		{name: "products", path: "/stripe/v1/products"},
		{name: "prices", path: "/stripe/v1/prices"},
		{name: "subscriptions", path: "/stripe/v1/subscriptions"},
		{name: "invoices", path: "/stripe/v1/invoices"},
		{name: "refunds", path: "/stripe/v1/refunds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serveStripeRequest(mux, http.MethodGet, tt.path, "", nil)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
			}

			var listed struct {
				Object  string `json:"object"`
				HasMore *bool  `json:"has_more"`
				URL     string `json:"url"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
				t.Fatalf("decode list: %v", err)
			}
			if listed.Object != "list" {
				t.Fatalf("object = %#v, want list", listed.Object)
			}
			if listed.HasMore == nil {
				t.Fatal("has_more field missing")
			}
			if *listed.HasMore {
				t.Fatalf("has_more = true, want false")
			}
			if listed.URL != tt.path {
				t.Fatalf("url = %#v, want %s", listed.URL, tt.path)
			}
		})
	}
}
