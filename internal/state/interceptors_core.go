package state

// AddRouteInterceptor registers an interceptor.
func (s *Store) AddRouteInterceptor(i RouteInterceptor) {
	if s == nil || i == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interceptors = append(s.interceptors, i)
}

// Deprecated: use AddRouteInterceptor.
func (s *Store) RegisterRouteInterceptor(i RouteInterceptor) { s.AddRouteInterceptor(i) }

// Deprecated: use AddRouteInterceptor.
func (s *Store) AddInterceptor(i RouteInterceptor) { s.AddRouteInterceptor(i) }
