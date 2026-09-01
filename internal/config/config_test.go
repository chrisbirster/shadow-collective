package config

import "testing"

func TestValidateRejectsDuplicateListeners(t *testing.T) {
	cfg := Config{
		HTTP: []HTTPService{{Name: "a", Listen: ":10001", Upstream: "http://a.internal:8080"}},
		TCP:  []TCPService{{Name: "b", Listen: ":10001", Upstream: "b.internal:5432"}},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected duplicate listener validation error")
	}
}

func TestValidateRequiresHTTPURL(t *testing.T) {
	cfg := Config{HTTP: []HTTPService{{Name: "a", Listen: ":10001", Upstream: "a.internal:8080"}}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid HTTP upstream error")
	}
}
