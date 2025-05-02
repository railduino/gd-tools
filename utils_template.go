package main

import (
	"bytes"
	"embed"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates
var templateFS embed.FS

func TemplateLoad(subdir, name string, data interface{}) ([]byte, error) {
	path := filepath.Join("templates", subdir, name)

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			content, err = templateFS.ReadFile(path)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err // any other error
		}
	}

	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return nil, err
	}

	var result bytes.Buffer
	if err := tmpl.Execute(&result, data); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func TemplateExport(target string) error {
	// TODO err := os.CopyFS(templateFS, target)

	return nil
}
