package agent

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

// RedisTest checks if there is work to be done
func RedisTest(req *Request) bool {
	return req != nil && req.RedisPort > 0
}

func RedisHandler(req *Request, resp *Response) error {
	if req == nil || req.RedisPort == 0 || resp == nil {
		return nil
	}
	if req.RedisPort < 0 || req.RedisPort >= 65536 {
		return fmt.Errorf("invalid redis port %d", req.RedisPort)
	}

	path := "/etc/redis/redis.conf"
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	var buf bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(content))
	found := false
	newLine := fmt.Sprintf("port %d", req.RedisPort)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "port ") {
			if strings.TrimSpace(line) == newLine {
				resp.Sayf("✅ Redis is listening on %d", req.RedisPort)
				return nil
			}
			buf.WriteString(newLine + "\n")
			found = true
		} else {
			buf.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan %s: %w", path, err)
	}

	if !found {
		buf.WriteString(newLine + "\n")
	}

	err = os.WriteFile(path, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	resp.AddService("redis")

	resp.Sayf("Redis is now listening on %d", req.RedisPort)

	return nil
}
