package helper

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

var (
	reRefAfterR  = regexp.MustCompile(`R#\s*([0-9]{6,})`)
	reRefBeforeR = regexp.MustCompile(`#([0-9]{6,})R#`)
)

func AsMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func ToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case json.Number:
		return t.String()
	default:
		return ""
	}
}

func ToI64(v any) int64 {
	switch x := v.(type) {
	case nil:
		return 0
	case json.Number:
		n, _ := x.Int64()
		return n
	case float64:
		return int64(x)
	case int:
		return int64(x)
	case int64:
		return x
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(x), 10, 64)
		return n
	default:
		return 0
	}
}

func PickAny(payload map[string]any, data map[string]any, keys ...string) any {
	for _, k := range keys {
		if data != nil {
			if v, ok := data[k]; ok && v != nil {
				if s, ok := v.(string); ok && strings.TrimSpace(s) == "" {
					continue
				}
				return v
			}
		}
		if payload != nil {
			if v, ok := payload[k]; ok && v != nil {
				if s, ok := v.(string); ok && strings.TrimSpace(s) == "" {
					continue
				}
				return v
			}
		}
	}
	return nil
}

func PickStr(payload map[string]any, data map[string]any, keys ...string) string {
	return ToString(PickAny(payload, data, keys...))
}

func PickI64(payload map[string]any, data map[string]any, keys ...string) int64 {
	return ToI64(PickAny(payload, data, keys...))
}

func ExtractRefIDFromPesan(pesan string) string {
	pesan = strings.TrimSpace(pesan)
	if pesan == "" {
		return ""
	}
	if m := reRefAfterR.FindStringSubmatch(pesan); len(m) >= 2 {
		return m[1]
	}
	if m := reRefBeforeR.FindStringSubmatch(pesan); len(m) >= 2 {
		return m[1]
	}
	return ""
}
