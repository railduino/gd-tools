package agent

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	languageKey contextKey = "user-language"

	DefaultLanguage = "de"
	DefaultRegion   = "DE"
	DefaultLocale   = DefaultLanguage + "_" + DefaultRegion
)

var (
	Language string
	Region   string
)

func SetLanguage(name string) {
	if name == "" {
		name = DefaultLanguage
	}
	Language = name
}

func GetLanguage() string {
	if Language == "" {
		SetLanguage("")
	}
	return Language
}

func SetRegion(name string) {
	if name == "" {
		name = DefaultRegion
	}
	Region = name
}

func GetRegion() string {
	if Region == "" {
		SetRegion("")
	}
	return Region
}

func ParseLang(raw string) string {
	if raw == "" || raw == "C" || strings.HasPrefix(raw, "C.") || raw == "POSIX" {
		return ""
	}

	// Handle "de:en"
	if parts := strings.Split(raw, ":"); len(parts) > 0 {
		raw = parts[0]
	}

	// Cut at first "-", "_", ".", or "@"
	if idx := strings.IndexAny(raw, "-_.@"); idx != -1 {
		raw = raw[:idx]
	}

	return raw
}

func RequestLanguage(r *http.Request) string {
	header := r.Header.Get("Accept-Language")
	if header == "" {
		return DefaultLanguage
	}

	// Example: "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"
	parts := strings.Split(header, ",")
	if len(parts) > 0 {
		return ParseLang(parts[0])
	}

	return DefaultLanguage
}

func LanguageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := RequestLanguage(r)
		ctx := context.WithValue(r.Context(), languageKey, lang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetLanguageCtx(ctx context.Context) string {
	if val, ok := ctx.Value(languageKey).(string); ok && val != "" {
		return val
	}
	return DefaultLanguage
}
