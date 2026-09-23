package upload

import (
	"net"
	"testing"
)

func TestValidateRemoteURL(t *testing.T) {
	tests := []struct {
		url   string
		valid bool
	}{
		{"https://example.com/file.png", true},
		{"http://example.com/file", true},
		{"file:///etc/passwd", false},
		{"https://user:pass@example.com/file", false},
		{"//example.com/file", false},
	}
	for _, tt := range tests {
		_, err := validateRemoteURL(tt.url)
		if (err == nil) != tt.valid {
			t.Errorf("validateRemoteURL(%q) valid=%v err=%v", tt.url, tt.valid, err)
		}
	}
}

func TestIsPublicIP(t *testing.T) {
	privateAddresses := []string{"127.0.0.1", "10.0.0.1", "172.16.0.1", "192.168.1.1", "169.254.169.254", "::1", "fc00::1"}
	for _, raw := range privateAddresses {
		if isPublicIP(net.ParseIP(raw)) {
			t.Errorf("内网或保留地址不应允许: %s", raw)
		}
	}
	if !isPublicIP(net.ParseIP("8.8.8.8")) {
		t.Error("公网地址应允许")
	}
}
