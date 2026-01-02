package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

const (
	DHParamsName = "dhparams.pem"
)

// GenerateDHParams generates dhparams.pem (2048+ bits) via openssl
func GenerateDHParams(bits int) ([]byte, error) {
	dhBytes, err := os.ReadFile(DHParamsName)
	if err == nil && len(dhBytes) > 0 {
		return dhBytes, nil
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("openssl", "dhparam", "-outform", "PEM", fmt.Sprintf("%d", bits))
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run openssl dhparam: %w\n%s", err, stderr.String())
	}
	dhBytes = stdout.Bytes()

	if err := os.WriteFile(DHParamsName, dhBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", DHParamsName, err)
	}

	return dhBytes, nil

}
