package state

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"
)

var idInvalidChars = regexp.MustCompile(`[^a-z0-9_]+`)

type Resource struct {
	ID       string         `json:"id"`
	Adapter  string         `json:"adapter"`
	Type     string         `json:"type"`
	Data     map[string]any `json:"data"`
	Revision int64          `json:"revision"`
}

type Store struct {
	mu        sync.RWMutex
	resources map[scope]map[string]Resource
	counters  map[scope]int64
}

const MaxResourcesPerScope = 1000

type scope struct {
	adapter      string
	resourceType string
}

func NewStore() *Store {
	return &Store{
		resources: map[scope]map[string]Resource{},
		counters:  map[scope]int64{},
	}
}

func (s *Store) Create(adapter, resourceType string, data map[string]any) (Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.resources == nil {
		s.resources = map[scope]map[string]Resource{}
	}
	if s.counters == nil {
		s.counters = map[scope]int64{}
	}
	key := newScope(adapter, resourceType)
	s.counters[key]++
	resource := Resource{
		ID:       fmt.Sprintf("%s_%06d", key.idPrefix(), s.counters[key]),
		Adapter:  adapter,
		Type:     resourceType,
		Data:     cloneData(data),
		Revision: 1,
	}
	if s.resources[key] == nil {
		s.resources[key] = map[string]Resource{}
	}
	s.resources[key][resource.ID] = resource
	evictOldestLocked(s.resources[key], MaxResourcesPerScope)
	return cloneResource(resource), nil
}

func (s *Store) Get(adapter, resourceType, id string) (Resource, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resource, ok := s.resources[newScope(adapter, resourceType)][id]
	if !ok {
		return Resource{}, false
	}
	return cloneResource(resource), true
}

func (s *Store) Take(adapter, resourceType, id string) (Resource, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := newScope(adapter, resourceType)
	resource, ok := s.resources[key][id]
	if !ok {
		return Resource{}, false
	}
	delete(s.resources[key], id)
	return cloneResource(resource), true
}

func (s *Store) List(adapter, resourceType string) []Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := s.resources[newScope(adapter, resourceType)]
	ids := make([]string, 0, len(entries))
	for id := range entries {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	out := make([]Resource, 0, len(ids))
	for _, id := range ids {
		out = append(out, cloneResource(entries[id]))
	}
	return out
}

func (s *Store) Update(adapter, resourceType, id string, patch map[string]any) (Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := newScope(adapter, resourceType)
	resource, ok := s.resources[key][id]
	if !ok {
		return Resource{}, fmt.Errorf("resource not found: %s", id)
	}
	if resource.Data == nil {
		resource.Data = map[string]any{}
	}
	for name, value := range patch {
		resource.Data[name] = cloneValue(value)
	}
	resource.Revision++
	s.resources[key][id] = resource
	return cloneResource(resource), nil
}

func (s *Store) Delete(adapter, resourceType, id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := newScope(adapter, resourceType)
	if _, ok := s.resources[key][id]; !ok {
		return false
	}
	delete(s.resources[key], id)
	return true
}

func (s *Store) Reset(adapter, resourceType string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := newScope(adapter, resourceType)
	delete(s.resources, key)
	delete(s.counters, key)
}

func newScope(adapter, resourceType string) scope {
	return scope{adapter: sanitize(adapter), resourceType: sanitize(resourceType)}
}

func (s scope) idPrefix() string {
	return s.adapter + "_" + s.resourceType
}

func sanitize(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = idInvalidChars.ReplaceAllString(value, "_")
	value = strings.Trim(value, "_")
	if value == "" {
		return "resource"
	}
	return value
}

func cloneResource(resource Resource) Resource {
	resource.Data = cloneData(resource.Data)
	return resource
}

func cloneData(data map[string]any) map[string]any {
	if data == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(data))
	for name, value := range data {
		out[name] = cloneValue(value)
	}
	return out
}

func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneData(typed)
	case []any:
		out := make([]any, len(typed))
		for i, entry := range typed {
			out[i] = cloneValue(entry)
		}
		return out
	case []map[string]any:
		out := make([]map[string]any, len(typed))
		for i, entry := range typed {
			out[i] = cloneData(entry)
		}
		return out
	default:
		return value
	}
}

func evictOldestLocked(resources map[string]Resource, max int) {
	if max <= 0 || len(resources) <= max {
		return
	}
	ids := make([]string, 0, len(resources))
	for id := range resources {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for len(resources) > max {
		delete(resources, ids[0])
		ids = ids[1:]
	}
}
