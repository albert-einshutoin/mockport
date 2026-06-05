package state

import (
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestIdempotencyStoreReplaysMatchingRequest(t *testing.T) {
	store := NewIdempotencyStore()
	response := IdempotentResponse{Status: 200, Body: map[string]any{"id": "cs_test_123"}}

	replayed, got, err := store.Remember("stripe", "key-1", "POST /v1/checkout/sessions amount=1200", response)
	if err != nil {
		t.Fatalf("remember first request: %v", err)
	}
	if replayed {
		t.Fatal("first request replayed = true, want false")
	}

	replayed, got, err = store.Remember("stripe", "key-1", "POST /v1/checkout/sessions amount=1200", IdempotentResponse{Status: 500})
	if err != nil {
		t.Fatalf("remember replay request: %v", err)
	}
	if !replayed {
		t.Fatal("second matching request replayed = false, want true")
	}
	if got.Status != response.Status || got.Body["id"] != "cs_test_123" {
		t.Fatalf("replayed response = %#v", got)
	}
}

func TestIdempotencyStoreZeroValueIsUsable(t *testing.T) {
	var store IdempotencyStore
	response := IdempotentResponse{Status: 200, Body: map[string]any{"id": "cs_test_123"}}

	replayed, got, err := store.Remember("stripe", "key-1", "amount=1200", response)
	if err != nil {
		t.Fatalf("remember with zero-value store: %v", err)
	}
	if replayed || got.Status != http.StatusOK {
		t.Fatalf("first remember = replayed %v response %#v", replayed, got)
	}
}

func TestIdempotencyStoreRejectsConflictingRequest(t *testing.T) {
	store := NewIdempotencyStore()
	if _, _, err := store.Remember("stripe", "key-1", "amount=1200", IdempotentResponse{Status: 200}); err != nil {
		t.Fatalf("remember first request: %v", err)
	}

	_, _, err := store.Remember("stripe", "key-1", "amount=999", IdempotentResponse{Status: 200})
	var conflict *IdempotencyConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("err = %v, want IdempotencyConflictError", err)
	}
	if conflict.Scope != "stripe" || conflict.Key != "key-1" {
		t.Fatalf("conflict = %#v", conflict)
	}
}

func TestIdempotencyStoreDoReplaysMatchingInFlightRequest(t *testing.T) {
	store := NewIdempotencyStore()
	scope := "stripe:checkout_session"
	response := IdempotentResponse{Status: http.StatusOK, Body: map[string]any{"id": "cs_test_123"}}
	started := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan idempotencyResult, 1)
	secondDone := make(chan idempotencyResult, 1)
	var calls atomic.Int64

	go func() {
		replayed, got, err := store.Do(scope, "key-1", "client_reference_id=cart_1", func() (IdempotentResponse, error) {
			calls.Add(1)
			close(started)
			<-release
			return response, nil
		})
		firstDone <- idempotencyResult{replayed: replayed, response: got, err: err}
	}()
	<-started

	go func() {
		replayed, got, err := store.Do(scope, "key-1", "client_reference_id=cart_1", func() (IdempotentResponse, error) {
			calls.Add(1)
			return IdempotentResponse{Status: http.StatusInternalServerError}, nil
		})
		secondDone <- idempotencyResult{replayed: replayed, response: got, err: err}
	}()
	close(release)

	first := <-firstDone
	second := <-secondDone
	if first.err != nil || second.err != nil {
		t.Fatalf("Do errors = first %v second %v", first.err, second.err)
	}
	if first.replayed {
		t.Fatal("first request replayed = true, want false")
	}
	if !second.replayed {
		t.Fatal("second matching request replayed = false, want true")
	}
	if first.response.Body["id"] != "cs_test_123" || second.response.Body["id"] != "cs_test_123" {
		t.Fatalf("responses = first %#v second %#v", first.response, second.response)
	}
	if calls.Load() != 1 {
		t.Fatalf("side effect calls = %d, want 1", calls.Load())
	}
}

func TestIdempotencyStoreDoRejectsConflictingInFlightRequest(t *testing.T) {
	store := NewIdempotencyStore()
	scope := "stripe:checkout_session"
	started := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan idempotencyResult, 1)
	var conflictingRunCalled atomic.Bool

	go func() {
		replayed, got, err := store.Do(scope, "key-1", "client_reference_id=cart_1", func() (IdempotentResponse, error) {
			close(started)
			<-release
			return IdempotentResponse{Status: http.StatusOK}, nil
		})
		firstDone <- idempotencyResult{replayed: replayed, response: got, err: err}
	}()
	<-started

	_, _, err := store.Do(scope, "key-1", "client_reference_id=cart_2", func() (IdempotentResponse, error) {
		conflictingRunCalled.Store(true)
		return IdempotentResponse{Status: http.StatusOK}, nil
	})
	var conflict *IdempotencyConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("err = %v, want IdempotencyConflictError", err)
	}
	if conflictingRunCalled.Load() {
		t.Fatal("conflicting request executed side effect")
	}

	close(release)
	if first := <-firstDone; first.err != nil {
		t.Fatalf("first request err = %v", first.err)
	}
}

func TestIdempotencyStoreCapsRecordsPerScope(t *testing.T) {
	store := NewIdempotencyStore()
	scope := "stripe:checkout_session"

	for i := range MaxIdempotencyRecordsPerScope {
		key := fmt.Sprintf("key-%04d", i)
		if _, _, err := store.Remember(scope, key, fmt.Sprintf("fingerprint-%04d", i), IdempotentResponse{Status: http.StatusOK}); err != nil {
			t.Fatalf("remember retained key %d: %v", i, err)
		}
	}

	replayed, _, err := store.Lookup(scope, "key-0000", "fingerprint-0000")
	if err != nil {
		t.Fatalf("lookup first retained key: %v", err)
	}
	if !replayed {
		t.Fatal("first key before overflow replayed = false, want true")
	}

	if _, _, err := store.Remember(scope, "key-overflow", "fingerprint-overflow", IdempotentResponse{Status: http.StatusOK}); err != nil {
		t.Fatalf("remember overflow key: %v", err)
	}

	if len(store.records) != MaxIdempotencyRecordsPerScope {
		t.Fatalf("record count = %d, want %d", len(store.records), MaxIdempotencyRecordsPerScope)
	}
	if replayed, _, err := store.Lookup(scope, "key-0000", "fingerprint-0000"); err != nil || replayed {
		t.Fatalf("evicted first key replayed = %v err = %v, want no replay", replayed, err)
	}
	if replayed, _, err := store.Lookup(scope, "key-0001", "fingerprint-0001"); err != nil || !replayed {
		t.Fatalf("second key replayed = %v err = %v, want replay", replayed, err)
	}
}

func TestIdempotencyStoreAllowsNewRequestAfterEvictionAndConflictsAgain(t *testing.T) {
	store := NewIdempotencyStore()
	scope := "stripe:checkout_session"
	key := "key-0000"

	if _, _, err := store.Remember(scope, key, "old-fingerprint", IdempotentResponse{Status: http.StatusOK}); err != nil {
		t.Fatalf("remember original key: %v", err)
	}
	for i := 1; i <= MaxIdempotencyRecordsPerScope; i++ {
		nextKey := fmt.Sprintf("key-%04d", i)
		if _, _, err := store.Remember(scope, nextKey, fmt.Sprintf("fingerprint-%04d", i), IdempotentResponse{Status: http.StatusOK}); err != nil {
			t.Fatalf("remember eviction filler %d: %v", i, err)
		}
	}

	replayed, _, err := store.Remember(scope, key, "new-fingerprint", IdempotentResponse{Status: http.StatusAccepted})
	if err != nil {
		t.Fatalf("remember evicted key with new fingerprint: %v", err)
	}
	if replayed {
		t.Fatal("evicted key replayed = true, want new request")
	}

	_, _, err = store.Remember(scope, key, "different-fingerprint", IdempotentResponse{Status: http.StatusAccepted})
	var conflict *IdempotencyConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("err = %v, want IdempotencyConflictError for retained replacement record", err)
	}
}

func TestIdempotencyStoreEvictionIsScoped(t *testing.T) {
	store := NewIdempotencyStore()
	primaryScope := "stripe:checkout_session"
	otherScope := "stripe:payment_intent"

	if _, _, err := store.Remember(otherScope, "shared-key", "other-fingerprint", IdempotentResponse{Status: http.StatusOK}); err != nil {
		t.Fatalf("remember other scope key: %v", err)
	}
	for i := 0; i <= MaxIdempotencyRecordsPerScope; i++ {
		key := fmt.Sprintf("key-%04d", i)
		if _, _, err := store.Remember(primaryScope, key, fmt.Sprintf("fingerprint-%04d", i), IdempotentResponse{Status: http.StatusOK}); err != nil {
			t.Fatalf("remember primary scope key %d: %v", i, err)
		}
	}

	replayed, _, err := store.Lookup(otherScope, "shared-key", "other-fingerprint")
	if err != nil {
		t.Fatalf("lookup other scope key: %v", err)
	}
	if !replayed {
		t.Fatal("other scope key replayed = false, want true")
	}
	if got, want := len(store.records), MaxIdempotencyRecordsPerScope+1; got != want {
		t.Fatalf("record count across scopes = %d, want %d", got, want)
	}
}

type idempotencyResult struct {
	replayed bool
	response IdempotentResponse
	err      error
}

func TestRequireFieldsReportsMissingRequiredFields(t *testing.T) {
	err := RequireFields(map[string]any{
		"amount":   1200,
		"currency": "",
	}, "amount", "currency", "customer")

	var missing *ValidationError
	if !errors.As(err, &missing) {
		t.Fatalf("err = %v, want ValidationError", err)
	}
	if got, want := missing.MissingFields, []string{"currency", "customer"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("missing fields = %#v, want %#v", got, want)
	}
}
