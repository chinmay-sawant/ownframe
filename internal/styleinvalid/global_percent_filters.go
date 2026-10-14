package styleinvalid

import "strings"

// AvoidGlobalPercentages filters changed values, dropping global percentages.
func AvoidGlobalPercentages(values map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range values {
		if IsGlobalPercentage(k, v) {
			continue
		}
		out[k] = v
	}
	return out
}

// AvoidGlobalPercentagePaths filters touched paths whose last segment is global.
func AvoidGlobalPercentagePaths(paths []StylePath) []StylePath {
	var out []StylePath
	for _, p := range paths {
		seg := p
		if i := strings.LastIndex(p, "/"); i >= 0 {
			seg = p[i+1:]
		}
		if IsGlobalPercentageProperty(seg) {
			continue
		}
		out = append(out, p)
	}
	if out == nil {
		out = []StylePath{}
	}
	return out
}

// ScopePercentage is a no-op helper returning input when local.
func ScopePercentage(path string, value any) (string, any) { return path, value }

// AvoidGlobalPercentagesForSubtree is an alias.
func AvoidGlobalPercentagesForSubtree(values map[string]any) map[string]any {
	return AvoidGlobalPercentages(values)
}

// CanScopePercentage reports whether a value avoids global percentage.
func CanScopePercentage(property string, value any) bool {
	return !IsGlobalPercentage(property, value)
}
