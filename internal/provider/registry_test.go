package provider

import (
	"context"
	"testing"
)

// stubProvider is a minimal Provider for testing registry behavior.
type stubProvider struct {
	name    string
	managed bool
}

func (s *stubProvider) Name() string  { return s.name }
func (s *stubProvider) Managed() bool { return s.managed }
func (s *stubProvider) GenerateDiagram(_ context.Context, _ GenerateRequest) (*GenerateResponse, error) {
	return nil, nil
}
func (s *stubProvider) StreamDiagram(_ context.Context, _ GenerateRequest, _ func(StreamChunk)) error {
	return nil
}
func (s *stubProvider) GenerateText(_ context.Context, _ GenerateRequest) (string, error) {
	return "", nil
}
func (s *stubProvider) GenerateStructured(_ context.Context, _ GenerateRequest, _ string) (string, error) {
	return "{}", nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	r.Register("test", &stubProvider{name: "test"})

	p, err := r.Get("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "test" {
		t.Errorf("Name() = %q, want %q", p.Name(), "test")
	}
}

func TestRegistry_GetUnknown(t *testing.T) {
	r := NewRegistry()
	_, err := r.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestRegistry_IsManaged(t *testing.T) {
	r := NewRegistry()
	r.Register("managed", &stubProvider{name: "managed", managed: true})
	r.Register("byo", &stubProvider{name: "byo", managed: false})

	if !r.IsManaged("managed") {
		t.Error("expected IsManaged=true for managed provider")
	}
	if r.IsManaged("byo") {
		t.Error("expected IsManaged=false for BYO provider")
	}
	if r.IsManaged("nonexistent") {
		t.Error("expected IsManaged=false for nonexistent provider")
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	r.Register("a", &stubProvider{name: "a"})
	r.Register("b", &stubProvider{name: "b"})
	r.Register("c", &stubProvider{name: "c"})

	names := r.List()
	if len(names) != 3 {
		t.Errorf("List() returned %d names, want 3", len(names))
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	for _, want := range []string{"a", "b", "c"} {
		if !nameSet[want] {
			t.Errorf("List() missing %q", want)
		}
	}
}

func TestRegistry_BYOProviders_NotManaged(t *testing.T) {
	// Verify the real providers return Managed() == false
	providers := []Provider{
		NewOpenAI(),
		NewAnthropic(),
		NewGemini(),
	}
	for _, p := range providers {
		if p.Managed() {
			t.Errorf("%s should not be managed", p.Name())
		}
	}
}
