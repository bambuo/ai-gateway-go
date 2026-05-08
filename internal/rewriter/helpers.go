package rewriter

import (
	"math/rand/v2"
	"slices"
	"strings"
)

var hopByHopHeaders = []string{
	"host", "connection", "transfer-encoding",
	"proxy-authorization", "proxy-connection",
}

func isHopByHop(key string) bool {
	return slices.Contains(hopByHopHeaders, key)
}

func toLower(s string) string {
	return strings.ToLower(s)
}

func toString(v any) string {
	switch s := v.(type) {
	case string:
		return s
	default:
		return ""
	}
}

func hasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func randomInRange(min, max int64) int64 {
	if max <= min {
		return min
	}
	return min + rand.Int64N(max-min)
}
