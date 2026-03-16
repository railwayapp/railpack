package cplnauth

import (
	"testing"
)

func TestRegistryHost(t *testing.T) {
	tests := []struct {
		ref  string
		want string
	}{
		{"myorg.registry.cpln.io/gvc/workload:buildcache", "myorg.registry.cpln.io"},
		{"myorg.registry.cpln.io/gvc/workload", "myorg.registry.cpln.io"},
		{"registry.example.com/repo:tag", "registry.example.com"},
		{"", ""},
	}
	for _, tt := range tests {
		got := RegistryHost(tt.ref)
		if got != tt.want {
			t.Errorf("RegistryHost(%q) = %q, want %q", tt.ref, got, tt.want)
		}
	}
}

func TestNew(t *testing.T) {
	hosts := []string{"a.registry.cpln.io", "b.registry.cpln.io"}
	ap := New(nil, nil, hosts)
	if !ap.basicAuthHosts["a.registry.cpln.io"] {
		t.Error("expected a.registry.cpln.io in basicAuthHosts")
	}
	if !ap.basicAuthHosts["b.registry.cpln.io"] {
		t.Error("expected b.registry.cpln.io in basicAuthHosts")
	}
	if ap.basicAuthHosts["other.registry.io"] {
		t.Error("unexpected other.registry.io in basicAuthHosts")
	}
}
