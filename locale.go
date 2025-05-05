package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/leonelquinteros/gotext"
)

//go:embed locales/**
var localeFS embed.FS

var (
	language string
	locale   *gotext.Locale
)

// initialize the i18n system
func LocaleInit() {
	if language == "" {
		language = os.Getenv("LANG")
	}
	lang := normalizeLang(language)

	// if the file tree is present, load from it, else use embedded
	localePath := filepath.Join("locales", lang, "LC_MESSAGES", "messages.po")
	if _, err := os.Stat(localePath); err == nil {
		locale = gotext.NewLocale("locales", lang)
	} else {
		locale = gotext.NewLocaleFSWithPath(lang, localeFS, "locales")
	}
	if locale == nil {
		fmt.Fprintln(os.Stderr, "Fatal: Locale could not be initialized.")
		os.Exit(1)
	}

	locale.AddDomain("messages")
}

func SetLanguage(lang string) {
	language = lang
	LocaleInit()
}

// T: Handle plain strings
func T(msg string) string {
	if locale == nil {
		LocaleInit()
	}

	if text := locale.Get(msg); text != "" && text != msg {
		return text
	} else {
		return "MSG:" + msg
	}
}

// Tf: Handle formatted strings
func Tf(format string, args ...any) string {
	if locale == nil {
		LocaleInit()
	}

	return fmt.Sprintf(T(format), args...)
}

// Tn: Handle counting strings
func Tn(singular, plural string, n int) string {
	if locale == nil {
		LocaleInit()
	}

	return locale.GetN(singular, plural, n)
}

// Tnf: Handle counting strings with args
func Tnf(singular, plural string, n int, args ...any) string {
	if locale == nil {
		LocaleInit()
	}

	return fmt.Sprintf(Tn(singular, plural, n), args...)
}

// make sure there is a valid language
func normalizeLang(lang string) string {
	switch lang {
	case "de", "de_DE":
		return "de_DE"
	case "en", "en_US", "en_GB":
		return "en_US"
	default:
		return "de_DE"
	}

	return lang
}
