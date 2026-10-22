package state

import "testing"

// dataMap returns the store data as a map, failing the test when the
// store holds anything else. It keeps the checked assertion in one place.
func dataMap(t *testing.T, s *Store) map[string]any {
	t.Helper()
	m, ok := s.Data().(map[string]any)
	if !ok {
		t.Fatalf("data = %T, want map[string]any", s.Data())
	}
	return m
}

// dataUser returns the nested user map, failing the test on any shape.
func dataUser(t *testing.T, s *Store) map[string]any {
	t.Helper()
	u, ok := dataMap(t, s)["user"].(map[string]any)
	if !ok {
		t.Fatal("user is not a map")
	}
	return u
}
