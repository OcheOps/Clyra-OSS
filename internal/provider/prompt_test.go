package provider

import (
	"strings"
	"testing"
)

func TestBuildSystemPrompt_DefaultMode(t *testing.T) {
	prompt := BuildSystemPrompt("")
	if !strings.Contains(prompt, "SOLUTION ARCHITECTURE") {
		t.Error("empty mode should default to solution architecture")
	}
	if !strings.Contains(prompt, "solution.microservice") {
		t.Error("solution prompt should mention solution primitives")
	}
	if !strings.Contains(prompt, "dot-separated canonical types") {
		t.Error("prompt should instruct dot-separated format")
	}
}

func TestBuildSystemPrompt_SolutionMode(t *testing.T) {
	prompt := BuildSystemPrompt(ModeSolutionArchitecture)
	if !strings.Contains(prompt, "SOLUTION ARCHITECTURE") {
		t.Error("should contain solution architecture header")
	}
	// Should mention key solution types
	for _, typ := range []string{
		"solution.backend_service", "solution.api_gateway",
		"solution.queue", "solution.vector_store",
		"solution.ai_service", "solution.ci_cd",
	} {
		if !strings.Contains(prompt, typ) {
			t.Errorf("solution prompt missing %q", typ)
		}
	}
}

func TestBuildSystemPrompt_NetworkMode(t *testing.T) {
	prompt := BuildSystemPrompt(ModeNetworkArchitecture)
	if !strings.Contains(prompt, "NETWORK ARCHITECTURE") {
		t.Error("should contain network architecture header")
	}
	for _, typ := range []string{
		"network.firewall", "network.bastion",
		"network.load_balancer_l7", "network.zero_trust_proxy",
		"network.private_subnet", "network.transit_hub",
	} {
		if !strings.Contains(prompt, typ) {
			t.Errorf("network prompt missing %q", typ)
		}
	}
	if strings.Contains(prompt, "SOLUTION ARCHITECTURE") {
		t.Error("network prompt should not contain solution header")
	}
}

func TestBuildSystemPrompt_EnterpriseMode(t *testing.T) {
	prompt := BuildSystemPrompt(ModeEnterpriseArchitecture)
	if !strings.Contains(prompt, "ENTERPRISE ARCHITECTURE") {
		t.Error("should contain enterprise architecture header")
	}
	for _, typ := range []string{
		"enterprise.erp", "enterprise.crm",
		"enterprise.sso", "enterprise.identity_provider",
		"enterprise.trust_boundary", "enterprise.etl",
	} {
		if !strings.Contains(prompt, typ) {
			t.Errorf("enterprise prompt missing %q", typ)
		}
	}
}

func TestBuildSystemPrompt_AllModesHaveBaseSchema(t *testing.T) {
	modes := []string{"", ModeSolutionArchitecture, ModeNetworkArchitecture, ModeEnterpriseArchitecture}
	for _, mode := range modes {
		prompt := BuildSystemPrompt(mode)
		if !strings.Contains(prompt, `"nodes"`) {
			t.Errorf("mode %q: missing nodes schema", mode)
		}
		if !strings.Contains(prompt, `"edges"`) {
			t.Errorf("mode %q: missing edges schema", mode)
		}
		if !strings.Contains(prompt, `"groups"`) {
			t.Errorf("mode %q: missing groups schema", mode)
		}
		if !strings.Contains(prompt, "dot-separated canonical types") {
			t.Errorf("mode %q: missing type format instruction", mode)
		}
		if !strings.Contains(prompt, "production-realistic") {
			t.Errorf("mode %q: missing closing rules", mode)
		}
	}
}

func TestBuildSystemPrompt_UsesNewDotFormat(t *testing.T) {
	prompt := BuildSystemPrompt("")
	// Should reference new format, not old colon format
	if strings.Contains(prompt, "aws:ec2") {
		t.Error("prompt should use aws.ec2 not aws:ec2")
	}
	if !strings.Contains(prompt, "aws.lambda") {
		t.Error("prompt should reference aws.lambda in dot format")
	}
}
