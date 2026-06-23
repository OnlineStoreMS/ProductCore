package util

import (
	"encoding/json"
	"strings"
	"time"
)

func ToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func ParseStringArray(raw string) []string {
	if raw == "" {
		return []string{}
	}
	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err == nil {
		return arr
	}
	// fallback comma-separated
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func FirstLetter(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "#"
	}
	r := []rune(s)
	return strings.ToUpper(string(r[0]))
}
