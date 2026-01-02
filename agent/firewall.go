package agent

import (
	"net"
	"strings"
)

func (resp *Response) FirewallOpen(c net.Conn, ports []string) error {
	for _, port := range ports {
		status, err := RunCommand("ufw", "allow", port)
		if err != nil {
			return err
		}
		result := string(status)
		if !strings.Contains(result, "existing rule") {
			resp.Say(result)
		}
	}

	status, err := RunCommand("ufw", "status")
	if err != nil {
		return err
	}
	result := string(status)

	if strings.Contains(result, "Status: active") {
		if _, err := RunCommand("ufw", "reload"); err != nil {
			return err
		}
		resp.Say("✅ firewall is active")
		return nil
	}

	status, err = RunCommand("ufw", "enable")
	if err != nil {
		return err
	}
	resp.Say(string(result))

	return nil
}
