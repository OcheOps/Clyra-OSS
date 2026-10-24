package provider

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

func TestBedrock_Name(t *testing.T) {
	b := &Bedrock{modelID: "test"}
	if b.Name() != "managed_bedrock" {
		t.Errorf("Name() = %q, want %q", b.Name(), "managed_bedrock")
	}
}

func TestBedrock_Managed(t *testing.T) {
	b := &Bedrock{modelID: "test"}
	if !b.Managed() {
		t.Error("Managed() should return true for Bedrock")
	}
}

func TestBedrock_Defaults(t *testing.T) {
	maxTokens, temp := applyBedrockDefaults(ProviderConfig{})
	if maxTokens != 4096 {
		t.Errorf("default maxTokens = %d, want 4096", maxTokens)
	}
	if temp != 0.3 {
		t.Errorf("default temperature = %f, want 0.3", temp)
	}

	maxTokens, temp = applyBedrockDefaults(ProviderConfig{MaxTokens: 8192, Temperature: 0.7})
	if maxTokens != 8192 {
		t.Errorf("custom maxTokens = %d, want 8192", maxTokens)
	}
	if temp != 0.7 {
		t.Errorf("custom temperature = %f, want 0.7", temp)
	}
}

func TestBedrock_ExtractText(t *testing.T) {
	diagramJSON := `{"nodes":[{"id":"n1","label":"Web","type":"aws.ec2"}],"edges":[],"groups":[]}`

	// Simulate a Converse output with text content
	output := &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role: types.ConversationRoleAssistant,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{Value: diagramJSON},
				},
			},
		},
	}

	text := extractBedrockText(output)
	if text == "" {
		t.Fatal("extractBedrockText returned empty")
	}

	var parsed GenerateResponse
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse diagram JSON: %v", err)
	}

	if len(parsed.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(parsed.Nodes))
	}
	if parsed.Nodes[0].ID != "n1" {
		t.Errorf("expected node id n1, got %q", parsed.Nodes[0].ID)
	}
	if parsed.Nodes[0].Type != "aws.ec2" {
		t.Errorf("expected type aws.ec2, got %q", parsed.Nodes[0].Type)
	}
}

func TestBedrock_ExtractText_Empty(t *testing.T) {
	// nil output
	output := &bedrockruntime.ConverseOutput{Output: nil}
	if text := extractBedrockText(output); text != "" {
		t.Errorf("expected empty, got %q", text)
	}

	// empty content blocks
	output2 := &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role:    types.ConversationRoleAssistant,
				Content: []types.ContentBlock{},
			},
		},
	}
	if text := extractBedrockText(output2); text != "" {
		t.Errorf("expected empty for no content, got %q", text)
	}
}

func TestBedrock_ExtractText_WithCodeFences(t *testing.T) {
	diagramJSON := `{"nodes":[{"id":"n1","label":"App","type":"aws.lambda"}],"edges":[],"groups":[]}`
	fenced := "```json\n" + diagramJSON + "\n```"

	output := &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role: types.ConversationRoleAssistant,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{Value: fenced},
				},
			},
		},
	}

	text := extractBedrockText(output)
	text = stripCodeFences(text)

	var parsed GenerateResponse
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse fenced JSON: %v", err)
	}
	if len(parsed.Nodes) != 1 || parsed.Nodes[0].Type != "aws.lambda" {
		t.Errorf("unexpected parse result: %+v", parsed.Nodes)
	}
}
