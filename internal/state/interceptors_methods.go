package state

import "reflect"

// RemoveRouteInterceptor removes one interceptor.
func (s *Store) RemoveRouteInterceptor(i RouteInterceptor) bool {
	if s == nil || i == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for idx, it := range s.interceptors {
		if funcPtr(it) == funcPtr(i) {
			s.interceptors = append(s.interceptors[:idx], s.interceptors[idx+1:]...)
			return true
		}
	}
	return false
}

// RouteInterceptors returns a copy.
func (s *Store) RouteInterceptors() []RouteInterceptor {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]RouteInterceptor, len(s.interceptors))
	copy(out, s.interceptors)
	return out
}

// Interceptors is an alias.
func (s *Store) Interceptors() []RouteInterceptor { return s.RouteInterceptors() }

// InterceptRoute runs interceptors; true means handled without full load.
func (s *Store) InterceptRoute(path string) bool {
	if s == nil {
		return false
	}
	s.mu.RLock()
	its := make([]RouteInterceptor, len(s.interceptors))
	copy(its, s.interceptors)
	s.mu.RUnlock()
	for _, it := range its {
		if it != nil && it(path) {
			return true
		}
	}
	return false
}

// Intercept is an alias.
func (s *Store) Intercept(path string) bool { return s.InterceptRoute(path) }

// HandleRoute is an alias.
func (s *Store) HandleRoute(path string) bool { return s.InterceptRoute(path) }

// TryRoute is an alias.
func (s *Store) TryRoute(path string) bool { return s.InterceptRoute(path) }

func funcPtr(f RouteInterceptor) uintptr {
	return reflect.ValueOf(f).Pointer()
}
