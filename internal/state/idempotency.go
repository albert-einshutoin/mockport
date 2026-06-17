package state

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

type IdempotentResponse struct {
	Status int            `json:"status"`
	Body   map[string]any `json:"body,omitempty"`
}

type idempotencyRecord struct {
	fingerprint string
	response    IdempotentResponse
}

type idempotencyCall struct {
	fingerprint string
	done        chan struct{}
	response    IdempotentResponse
	err         error
}

type IdempotencyStore struct {
	mu       sync.Mutex
	records  map[string]idempotencyRecord
	order    map[string][]string
	inFlight map[string]*idempotencyCall
}

const MaxIdempotencyRecordsPerScope = 1000

func NewIdempotencyStore() *IdempotencyStore {
	return &IdempotencyStore{
		records:  map[string]idempotencyRecord{},
		order:    map[string][]string{},
		inFlight: map[string]*idempotencyCall{},
	}
}

func (s *IdempotencyStore) Do(scope, key, fingerprint string, run func() (IdempotentResponse, error)) (bool, IdempotentResponse, error) {
	if strings.TrimSpace(key) == "" {
		response, err := run()
		return false, cloneResponse(response), err
	}

	recordKey := idempotencyRecordKey(scope, key)
	s.mu.Lock()
	if s.records == nil {
		s.records = map[string]idempotencyRecord{}
	}
	if s.order == nil {
		s.order = map[string][]string{}
	}
	if s.inFlight == nil {
		s.inFlight = map[string]*idempotencyCall{}
	}
	if record, ok := s.records[recordKey]; ok {
		replayed, response, err := replayRecord(scope, key, fingerprint, record)
		s.mu.Unlock()
		return replayed, response, err
	}
	if call, ok := s.inFlight[recordKey]; ok {
		if call.fingerprint != fingerprint {
			s.mu.Unlock()
			return false, IdempotentResponse{}, &IdempotencyConflictError{Scope: scope, Key: key}
		}
		done := call.done
		s.mu.Unlock()

		<-done

		s.mu.Lock()
		defer s.mu.Unlock()
		if call.err != nil {
			return false, IdempotentResponse{}, call.err
		}
		return true, cloneResponse(call.response), nil
	}

	call := &idempotencyCall{fingerprint: fingerprint, done: make(chan struct{})}
	s.inFlight[recordKey] = call
	s.mu.Unlock()

	response, err := run()

	s.mu.Lock()
	call.response = cloneResponse(response)
	call.err = err
	if err == nil {
		s.records[recordKey] = idempotencyRecord{fingerprint: fingerprint, response: cloneResponse(response)}
		s.order[scope] = append(s.order[scope], recordKey)
		s.evictOldestLocked(scope)
	}
	delete(s.inFlight, recordKey)
	close(call.done)
	s.mu.Unlock()

	if err != nil {
		return false, IdempotentResponse{}, err
	}
	return false, cloneResponse(response), nil
}

func (s *IdempotencyStore) Remember(scope, key, fingerprint string, response IdempotentResponse) (bool, IdempotentResponse, error) {
	if strings.TrimSpace(key) == "" {
		return false, cloneResponse(response), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.records == nil {
		s.records = map[string]idempotencyRecord{}
	}
	if s.order == nil {
		s.order = map[string][]string{}
	}
	if s.inFlight == nil {
		s.inFlight = map[string]*idempotencyCall{}
	}
	recordKey := idempotencyRecordKey(scope, key)
	if record, ok := s.records[recordKey]; ok {
		return replayRecord(scope, key, fingerprint, record)
	}
	s.records[recordKey] = idempotencyRecord{fingerprint: fingerprint, response: cloneResponse(response)}
	s.order[scope] = append(s.order[scope], recordKey)
	s.evictOldestLocked(scope)
	return false, cloneResponse(response), nil
}

func (s *IdempotencyStore) Lookup(scope, key, fingerprint string) (bool, IdempotentResponse, error) {
	if strings.TrimSpace(key) == "" {
		return false, IdempotentResponse{}, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[idempotencyRecordKey(scope, key)]
	if !ok {
		return false, IdempotentResponse{}, nil
	}
	return replayRecord(scope, key, fingerprint, record)
}

type IdempotencyConflictError struct {
	Scope string
	Key   string
}

func (e *IdempotencyConflictError) Error() string {
	return fmt.Sprintf("idempotency conflict for %s key %s", e.Scope, e.Key)
}

func replayRecord(scope, key, fingerprint string, record idempotencyRecord) (bool, IdempotentResponse, error) {
	if record.fingerprint != fingerprint {
		return false, IdempotentResponse{}, &IdempotencyConflictError{Scope: scope, Key: key}
	}
	return true, cloneResponse(record.response), nil
}

type ValidationError struct {
	MissingFields []string
}

func (e *ValidationError) Error() string {
	return "missing required fields: " + strings.Join(e.MissingFields, ", ")
}

func RequireFields(payload map[string]any, fields ...string) error {
	var missing []string
	for _, field := range fields {
		value, ok := payload[field]
		if !ok || isEmpty(value) {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		return &ValidationError{MissingFields: missing}
	}
	return nil
}

func (s *IdempotencyStore) evictOldestLocked(scope string) {
	ordered := s.order[scope]
	if MaxIdempotencyRecordsPerScope <= 0 || len(ordered) <= MaxIdempotencyRecordsPerScope {
		return
	}

	evictCount := len(ordered) - MaxIdempotencyRecordsPerScope
	for _, recordKey := range ordered[:evictCount] {
		delete(s.records, recordKey)
	}
	s.order[scope] = slices.Clone(ordered[evictCount:])
}

func idempotencyRecordKey(scope, key string) string {
	return scope + "\x00" + key
}

func isEmpty(value any) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(typed) == ""
	default:
		return false
	}
}

func cloneResponse(response IdempotentResponse) IdempotentResponse {
	response.Body = cloneData(response.Body)
	return response
}
