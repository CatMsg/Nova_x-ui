package service

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/CatMsg/Nova_x-ui/v3/database/model"
	"github.com/CatMsg/Nova_x-ui/v3/logger"
)

const hy2ForwardScript = "scripts/hy2-forward.sh"

func (s *InboundService) syncHy2PortForwarding(oldInbound *model.Inbound, inbound *model.Inbound) error {
	if inbound != nil && inbound.Enable && isHy2Inbound(inbound) {
		listenPort, err := getInboundListenPort(inbound)
		if err != nil {
			return err
		}

		ports, err := getHy2ServerPorts(inbound.StreamSettings)
		if err != nil {
			return err
		}

		if err := runHy2ForwardScript("apply", inbound.Tag, listenPort, ports); err != nil {
			return err
		}

		if oldInbound != nil && oldInbound.Tag != inbound.Tag && isHy2Inbound(oldInbound) {
			if err := runHy2ForwardScript("remove", oldInbound.Tag, 0, nil); err != nil {
				return err
			}
		}
		return nil
	}

	if oldInbound != nil && isHy2Inbound(oldInbound) {
		return runHy2ForwardScript("remove", oldInbound.Tag, 0, nil)
	}

	return nil
}

func isHy2Inbound(inbound *model.Inbound) bool {
	if inbound == nil {
		return false
	}
	if inbound.Protocol == model.Hysteria2 {
		return true
	}
	if inbound.Protocol != model.Hysteria {
		return false
	}

	version := 2
	if len(inbound.Settings) > 0 {
		var settings map[string]any
		if err := json.Unmarshal([]byte(inbound.Settings), &settings); err == nil {
			version = parseHy2Version(settings["version"], version)
		}
	}
	return version == 2
}

func runHy2ForwardScript(action string, tag string, listenPort int, ports []int) error {
	if runtime.GOOS != "linux" {
		return nil
	}
	if tag == "" {
		return nil
	}

	args := []string{action, tag, strconv.Itoa(listenPort), joinHy2Ports(ports)}
	cmd := exec.Command("bash", append([]string{hy2ForwardScript}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		trimmed := strings.TrimSpace(string(output))
		if trimmed != "" {
			err = fmt.Errorf("%w: %s", err, trimmed)
		}
		logger.Warning("hy2 port forwarding sync failed:", err)
		return err
	}

	return nil
}

func detectHy2ForwardBackend() string {
	if runtime.GOOS != "linux" {
		return "none"
	}

	if _, err := exec.LookPath("ufw"); err == nil {
		if output, err := exec.Command("ufw", "status").CombinedOutput(); err == nil && strings.Contains(string(output), "Status: active") {
			return "ufw"
		}
	}

	if _, err := exec.LookPath("nft"); err == nil {
		return "nftables"
	}

	if _, err := exec.LookPath("iptables"); err == nil || commandExists("ip6tables") {
		return "iptables"
	}

	return "none"
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func getInboundListenPort(inbound *model.Inbound) (int, error) {
	if inbound == nil {
		return 0, fmt.Errorf("inbound is nil")
	}
	if inbound.Port < 1 || inbound.Port > 65535 {
		return 0, fmt.Errorf("invalid listen port for inbound %s", inbound.Tag)
	}
	return inbound.Port, nil
}

func getHy2ServerPorts(streamSettings string) ([]int, error) {
	if strings.TrimSpace(streamSettings) == "" {
		return nil, nil
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(streamSettings), &payload); err != nil {
		return nil, err
	}

	rawPorts := extractHy2PortsSource(payload)
	if rawPorts == nil {
		return nil, nil
	}

	return parseHy2Ports(rawPorts)
}

func extractHy2PortsSource(payload map[string]any) any {
	if payload == nil {
		return nil
	}
	if raw, ok := payload["hysteriaSettings"]; ok {
		if settings, ok := raw.(map[string]any); ok {
			if ports, exists := settings["server_ports"]; exists {
				return ports
			}
			if ports, exists := settings["serverPorts"]; exists {
				return ports
			}
		}
	}
	if ports, ok := payload["server_ports"]; ok {
		return ports
	}
	if ports, ok := payload["serverPorts"]; ok {
		return ports
	}
	return nil
}

func parseHy2Ports(rawPorts any) ([]int, error) {
	ports := make([]int, 0)
	seen := map[int]struct{}{}
	appendPort := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}
		port, err := strconv.Atoi(raw)
		if err != nil || port < 1 || port > 65535 {
			return
		}
		if _, exists := seen[port]; exists {
			return
		}
		seen[port] = struct{}{}
		ports = append(ports, port)
	}

	switch typed := rawPorts.(type) {
	case []any:
		for _, item := range typed {
			switch v := item.(type) {
			case string:
				appendPort(v)
			default:
				appendPort(fmt.Sprint(v))
			}
		}
	case []string:
		for _, item := range typed {
			appendPort(item)
		}
	case string:
		for _, item := range strings.Split(typed, ",") {
			appendPort(item)
		}
	case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, json.Number:
		appendPort(fmt.Sprint(typed))
	default:
		return nil, fmt.Errorf("unsupported server_ports format")
	}

	return ports, nil
}

func parseHy2Version(raw any, fallback int) int {
	switch v := raw.(type) {
	case float64:
		return int(v)
	case float32:
		return int(v)
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case json.Number:
		if n, err := v.Int64(); err == nil {
			return int(n)
		}
	case string:
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n
		}
	}
	return fallback
}

func joinHy2Ports(ports []int) string {
	if len(ports) == 0 {
		return ""
	}

	values := make([]string, 0, len(ports))
	for _, port := range ports {
		values = append(values, strconv.Itoa(port))
	}
	return strings.Join(values, ",")
}
