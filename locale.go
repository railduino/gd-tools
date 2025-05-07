package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/leonelquinteros/gotext"
	"golang.org/x/text/language"
)

//go:embed locales/**
var localeFS embed.FS

const (
	languageKey contextKey = "language"
	HelloWorld             = "hello-world"
)

type LangCode struct {
	Lang   string
	Locale *gotext.Locale
}

var (
	langCodes = []LangCode{
		{Lang: "de_DE"},
		{Lang: "en_US"},
	}
	langTags = []language.Tag{}

	toolsLocale *gotext.Locale
	localeInfo  []string
)

func LocaleGetInfo() []string {
	return localeInfo
}

// initialize the i18n system
func LocaleInit() {
	for index, code := range langCodes {
		if langCodes[index].Locale != nil {
			continue // already loaded
		}
		lang := langCodes[index].Lang

		// if the file tree is present, load from it, else use embedded
		var locale *gotext.Locale
		localePath := filepath.Join("locales", lang, "LC_MESSAGES", "messages.po")
		if _, err := os.Stat(localePath); err == nil {
			locale = gotext.NewLocale("locales", lang)
			localeInfo = append(localeInfo, lang+" loaded from file")
		} else {
			locale = gotext.NewLocaleFSWithPath(lang, localeFS, "locales")
			localeInfo = append(localeInfo, lang+" loaded from exec")
		}
		if locale == nil {
			fmt.Fprintln(os.Stderr, "Fatal: Locale could not be initialized.")
			os.Exit(1)
		}

		locale.AddDomain("messages")
		langCodes[index].Locale = locale
		langTags = append(langTags, language.Make(code.Lang))
	}

	// at this point, at least de_DE and en_US are known to exist
	if toolsLocale == nil {
		toolsLang := os.Getenv("LANG")
		if toolsLang == "" {
			toolsLang = langCodes[0].Lang
		}
		for _, code := range langCodes {
			if code.Lang == toolsLang {
				toolsLocale = code.Locale
				break
			}
		}
		if toolsLocale == nil {
			toolsLocale = langCodes[0].Locale
		}
		localeInfo = append(localeInfo, T("hello-world"))
	}
	if toolsLocale == nil {
		fmt.Fprintln(os.Stderr, "Fatal: toolsLocale is missing.")
		os.Exit(1)
	}
}

// WebT: Handle strings from the embedded webserver
func WebT(r *http.Request, key string, args ...interface{}) string {
	code := langCodes[0]
	if lang, ok := r.Context().Value(languageKey).(string); ok {
		for _, check := range langCodes {
			if check.Lang == lang {
				code = check
			}
		}
	}

	if result := code.Locale.Get(key, args...); result != key {
		return result
	}

	return "WebMSG: " + key
}

// i18n middleware for the embedded webserver
func LocaleFromRequest(r *http.Request) LangCode {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = r.Header.Get("Accept-Language")
	}

	prefs, _, _ := language.ParseAcceptLanguage(lang)
	match := language.NewMatcher(langTags)
	_, index, _ := match.Match(prefs...)

	// n.b. index is 0 if no match is found
	return langCodes[index]
}

func LocaleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "static/") {
			next.ServeHTTP(w, r)
			return
		}

		code := LocaleFromRequest(r)
		ctx := context.WithValue(r.Context(), languageKey, code.Lang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// T: Handle plain strings
func T(msg string) string {
	if toolsLocale == nil {
		LocaleInit()
	}

	if text := toolsLocale.Get(msg); text != "" && text != msg {
		return text
	} else {
		return "MSG:" + msg
	}
}

// Tf: Handle formatted strings
func Tf(format string, args ...any) string {
	if toolsLocale == nil {
		LocaleInit()
	}

	return fmt.Sprintf(T(format), args...)
}

// Tn: Handle counting strings
func Tn(singular, plural string, n int) string {
	if toolsLocale == nil {
		LocaleInit()
	}

	return toolsLocale.GetN(singular, plural, n)
}

// Tnf: Handle counting strings with args
func Tnf(singular, plural string, n int, args ...any) string {
	if toolsLocale == nil {
		LocaleInit()
	}

	return fmt.Sprintf(Tn(singular, plural, n), args...)
}

// make sure there is a valid language
func normalizeLang(lang string) string {
	for _, code := range langCodes {
		if code.Lang == lang {
			return lang
		}
	}

	return langCodes[0].Lang
}
