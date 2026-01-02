package templates

import (
	"bufio"
	"bytes"
	"embed"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/**
var templateFS embed.FS

func Load(name string, debug bool) ([]byte, error) {
	path := filepath.Join("templates", name)

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

	return content, nil
}

func Parse(name string, debug bool, data interface{}) ([]byte, error) {
	content, err := Load(name, debug)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(name).Funcs(template.FuncMap{
		// define global functions here
	}).Parse(string(content))
	if err != nil {
		return nil, err
	}

	var result bytes.Buffer
	if err := tmpl.Execute(&result, data); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func SQL(name string, debug bool, data interface{}) ([]string, error) {
	content, err := Parse(name, debug, data)
	if err != nil {
		return nil, err
	}

	var stmts []string
	for _, stmt := range strings.Split(string(content), ";") {
		var lines []string
		for _, line := range strings.Split(stmt, "\n") {
			if line = strings.TrimSpace(line); line != "" {
				lines = append(lines, line)
			}
		}
		if len(lines) > 0 {
			stmts = append(stmts, strings.Join(lines, " "))
		}
	}

	return stmts, nil
}

func Lines(name, comment string, debug bool, data interface{}) ([]string, error) {
	var content []byte
	var err error

	if data != nil {
		content, err = Parse(name, debug, data)
	} else {
		content, err = Load(name, debug)
	}
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(content)
	scanner := bufio.NewScanner(reader)
	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, comment) {
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func Export(target string) error {
	// TODO (later) err := os.CopyFS(templateFS, target)

	return nil
}
