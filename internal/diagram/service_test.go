package diagram

import (
	"context"
	"testing"

	"github.com/oche/diagram/internal/provider"
)

// mockProvider returns a canned diagram response.
type mockProvider struct {
	name    string
	managed bool
}

func (m *mockProvider) Name() string  { return m.name }
func (m *mockProvider) Managed() bool { return m.managed }
func (m *mockProvider) GenerateDiagram(_ context.Context, _ provider.GenerateRequest) (*provider.GenerateResponse, error) {
	return &provider.GenerateResponse{
		Nodes:  []provider.Node{{ID: "n1", Label: "Test", Type: "aws.ec2"}},
		Edges:  []provider.Edge{},
		Groups: []provider.Group{},
	}, nil
}
func (m *mockProvider) StreamDiagram(_ context.Context, _ provider.GenerateRequest, onChunk func(provider.StreamChunk)) error {
	onChunk(provider.StreamChunk{Delta: `{"nodes":[{"id":"n1","label":"Test","type":"aws.ec2"}],"edges":[],"groups":[]}`, Done: false})
	onChunk(provider.StreamChunk{Done: true})
	return nil
}
func (m *mockProvider) GenerateText(_ context.Context, _ provider.GenerateRequest) (string, error) {
	return "", nil
}
func (m *mockProvider) GenerateStructured(_ context.Context, _ provider.GenerateRequest, _ string) (string, error) {
	return "{}", nil
}

func setupTestService() *Service {
	reg := provider.NewRegistry()
	reg.Register("managed_bedrock", &mockProvider{name: "managed_bedrock", managed: true})
	reg.Register("openai", &mockProvider{name: "openai", managed: false})
	return NewService(reg, nil, nil) // nil guard + nil logger for unit tests
}

func TestPreCheck_ManagedProvider_NoAPIKey(t *testing.T) {
	svc := setupTestService()
	_, err := svc.preCheck(context.Background(), "user-1", "free", GenerateRequest{
		Prompt:   "test prompt for diagram",
		Provider: "managed_bedrock",
		APIKey:   "",
	})
	if err != nil {
		t.Errorf("managed provider should not require API key, got: %v", err)
	}
}

func TestPreCheck_BYOProvider_RequiresAPIKey(t *testing.T) {
	svc := setupTestService()
	_, err := svc.preCheck(context.Background(), "user-1", "free", GenerateRequest{
		Prompt:   "test prompt for diagram",
		Provider: "openai",
		APIKey:   "",
	})
	if err != ErrMissingAPIKey {
		t.Errorf("expected ErrMissingAPIKey, got: %v", err)
	}

	_, err = svc.preCheck(context.Background(), "user-1", "free", GenerateRequest{
		Prompt:   "test prompt for diagram",
		Provider: "openai",
		APIKey:   "sk-test",
	})
	if err != nil {
		t.Errorf("BYO provider with API key should pass, got: %v", err)
	}
}

func TestPreCheck_EmptyPrompt(t *testing.T) {
	svc := setupTestService()
	_, err := svc.preCheck(context.Background(), "user-1", "free", GenerateRequest{
		Prompt:   "",
		Provider: "managed_bedrock",
	})
	if err != ErrEmptyPrompt {
		t.Errorf("expected ErrEmptyPrompt, got: %v", err)
	}
}

func TestPreCheck_EmptyProvider(t *testing.T) {
	svc := setupTestService()
	_, err := svc.preCheck(context.Background(), "user-1", "free", GenerateRequest{
		Prompt:   "test prompt for diagram",
		Provider: "",
	})
	if err != ErrMissingProvider {
		t.Errorf("expected ErrMissingProvider, got: %v", err)
	}
}

func TestPreCheck_ModeIsOptional(t *testing.T) {
	svc := setupTestService()

	_, err := svc.preCheck(context.Background(), "user-1", "free", GenerateRequest{
		Prompt: "test prompt for diagram", Provider: "managed_bedrock", Mode: "",
	})
	if err != nil {
		t.Errorf("empty mode should be valid, got: %v", err)
	}

	_, err = svc.preCheck(context.Background(), "user-1", "free", GenerateRequest{
		Prompt: "test prompt for diagram", Provider: "managed_bedrock", Mode: "network_architecture",
	})
	if err != nil {
		t.Errorf("valid mode should pass, got: %v", err)
	}
}
