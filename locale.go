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

	toolsCode *LangCode
)

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
		} else {
			locale = gotext.NewLocaleFSWithPath(lang, localeFS, "locales")
		}
		if locale == nil {
			fmt.Fprintln(os.Stderr, "Fatal: Locale could not be initialized.")
			os.Exit(1)
		}

		locale.AddDomain("messages")
		langCodes[index].Locale = locale
		langTags = append(langTags, language.Make(code.Lang))
		fmt.Printf("%s -> %s\n", code.Lang, locale.Get("hello-world"))
	}

	// at this point, at least de_DE and en_US are known to exist
	toolsLang := os.Getenv("LANG")
	if toolsLang == "" {
		toolsLang = langCodes[0].Lang
	}
	toolsLang = normalizeLang(toolsLang)
}

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
}
