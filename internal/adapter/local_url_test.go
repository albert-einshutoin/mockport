package adapter

import (
	"fmt"
	"testing"

	"github.com/albert-einshutoin/mockport/internal/config"
)

func TestLocalBaseURLUsesDefaultPort(t *testing.T) {
	got := LocalBaseURL("/stripe")
	want := fmt.Sprintf("http://localhost:%d/stripe", config.DefaultPort)
	if got != want {
		t.Fatalf("LocalBaseURL(/stripe) = %q, want %q", got, want)
	}
}
