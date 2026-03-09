package utils

import (
	"fmt"
	"strings"
)

const (
	StatusOK      = "✅ "
	StatusError   = "❌ "
	StatusPending = "⏳ "
	StatusUnknown = "❓ "
)

// Helper function to collect lines
type LineBuffer struct {
	lines []string
}

func (lb *LineBuffer) Add(line string) {
	lb.lines = append(lb.lines, line)
}

func (lb *LineBuffer) Addf(format string, args ...any) {
	lb.lines = append(lb.lines, fmt.Sprintf(format, args...))
}

func (lb *LineBuffer) OK(line string) {
	lb.lines = append(lb.lines, line)
}

func (lb *LineBuffer) OKf(format string, args ...any) {
	lb.lines = append(lb.lines, fmt.Sprintf(StatusOK+format, args...))
}

func (lb *LineBuffer) Lines() []string {
	return lb.lines
}

func (lb *LineBuffer) Text() string {
	return strings.Join(lb.lines, "\n") + "\n"
}
