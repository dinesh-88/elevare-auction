package usecase

import (
	"fmt"
	"math"
)

func requiredString(payload map[string]any, key string) (string, error) {
	v, ok := payload[key]
	if !ok {
		return "", fmt.Errorf("missing payload field %q", key)
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", fmt.Errorf("invalid payload field %q", key)
	}
	return s, nil
}

func requiredInt64(payload map[string]any, key string) (int64, error) {
	v, ok := payload[key]
	if !ok {
		return 0, fmt.Errorf("missing payload field %q", key)
	}
	switch n := v.(type) {
	case int:
		return int64(n), nil
	case int64:
		return n, nil
	case float64:
		if math.Trunc(n) != n {
			return 0, fmt.Errorf("invalid payload field %q", key)
		}
		return int64(n), nil
	default:
		return 0, fmt.Errorf("invalid payload field %q", key)
	}
}
