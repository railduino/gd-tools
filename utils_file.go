package main

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strings"
)

func FileCopy(src, dst string, mode int) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return out.Chmod(fs.FileMode(mode))
}

func FileGetLine(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(content)), nil
}

func FileAddLine(path, pattern, text string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	found := false
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		if re.MatchString(line) {
			found = true
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if found {
		return nil
	}

	lines = append(lines, text)
	result := strings.Join(lines, "\n")

	if err := os.WriteFile(path, []byte(result+"\n"), 0644); err != nil {
		return err
	}
	return nil
}
