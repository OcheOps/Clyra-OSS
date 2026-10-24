package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// BedrockConfig holds the server-side AWS configuration for the managed provider.
type BedrockConfig struct {
	Region          string
	ModelID         string
	Profile         string
	AccessKeyID     string
	SecretAccessKey string
}

type Bedrock struct {
	client  *bedrockruntime.Client
	modelID string
}

// NewBedrock creates a managed Bedrock provider using server-side AWS credentials.
func NewBedrock(ctx context.Context, cfg BedrockConfig) (*Bedrock, error) {
	var opts []func(*awsconfig.LoadOptions) error

	opts = append(opts, awsconfig.WithRegion(cfg.Region))

	// Use named profile if provided
	if cfg.Profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(cfg.Profile))
	}

	// Use explicit credentials if provided, otherwise fall back to default chain
	// (profile, IAM role, env vars, shared credentials file, etc.)
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(awsCfg)

	modelID := cfg.ModelID
	if modelID == "" {
		// Default targets Opus 4.8 via the global cross-region inference
		// profile. Works from any region without needing a region-scoped
		// profile. Override with BEDROCK_MODEL_ID for a region-scoped
		// alternative (e.g. eu.anthropic.claude-opus-4-8 in Frankfurt).
		//
		// Confirmed via `aws bedrock list-inference-profiles` — full menu:
		//   eu.anthropic.claude-opus-4-5-20251101-v1:0
		//   eu.anthropic.claude-opus-4-6-v1
		//   eu.anthropic.claude-opus-4-7
		//   eu.anthropic.claude-opus-4-8
		//   eu.anthropic.claude-opus-5
		//   eu.anthropic.claude-sonnet-4-5-20250929-v1:0
		//   eu.anthropic.claude-sonnet-4-6
		//   eu.anthropic.claude-sonnet-5
		// (global.* variants also available). Model access must be
		// granted in the Bedrock console before use.
		modelID = "global.anthropic.claude-opus-4-8"
	}

	return &Bedrock{
		client:  client,
		modelID: modelID,
	}, nil
}

func (b *Bedrock) Name() string  { return "managed_bedrock" }
func (b *Bedrock) Managed() bool { return true }

// buildInferenceConfig assembles the Bedrock InferenceConfiguration for
// the current model, omitting fields that the specific model rejects.
//
// Newer Anthropic models (Opus 4.8+, Sonnet 5, Opus 5) removed
// `temperature` — passing it triggers ValidationException. We detect
// via the model ID prefix and skip the field when unsupported. Older
// models keep the tuned 0.3 for JSON-heavy output stability.
func (b *Bedrock) buildInferenceConfig(maxTokens int, temperature float64) *types.InferenceConfiguration {
	cfg := &types.InferenceConfiguration{
		MaxTokens: aws.Int32(int32(maxTokens)),
	}
	if supportsTemperature(b.modelID) {
		cfg.Temperature = aws.Float32(float32(temperature))
	}
	return cfg
}

// supportsTemperature returns false for Anthropic models that removed the
// parameter. Match by substring on the inference-profile / model ID so
// both region-scoped ("eu.anthropic.claude-opus-4-8") and global
// ("global.anthropic.claude-opus-4-8") variants are covered.
func supportsTemperature(modelID string) bool {
	// New Claude generation (Opus 4.8, Sonnet 5, Opus 5, and forward).
	// Add newer model markers here as they roll out.
	blocked := []string{
		"claude-opus-4-8",
		"claude-opus-5",
		"claude-sonnet-5",
	}
	for _, m := range blocked {
		if strings.Contains(modelID, m) {
			return false
		}
	}
	return true
}

func (b *Bedrock) GenerateDiagram(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	maxTokens, temperature := applyBedrockDefaults(req.Config)

	input := &bedrockruntime.ConverseInput{
		ModelId: aws.String(b.modelID),
		System: []types.SystemContentBlock{
			&types.SystemContentBlockMemberText{Value: req.SystemPrompt},
		},
		Messages: []types.Message{
			{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{Value: req.Prompt},
				},
			},
		},
		InferenceConfig: b.buildInferenceConfig(maxTokens, temperature),
	}

	output, err := b.client.Converse(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderFailed, err)
	}

	content := extractBedrockText(output)
	if content == "" {
		return nil, fmt.Errorf("%w: no text content in Bedrock response", ErrInvalidResponse)
	}

	content = stripCodeFences(content)

	var diagram GenerateResponse
	if err := json.Unmarshal([]byte(content), &diagram); err != nil {
		return nil, fmt.Errorf("%w: could not parse diagram JSON: %v", ErrInvalidResponse, err)
	}

	return &diagram, nil
}

func (b *Bedrock) GenerateText(ctx context.Context, req GenerateRequest) (string, error) {
	maxTokens, temperature := applyBedrockDefaults(req.Config)

	input := &bedrockruntime.ConverseInput{
		ModelId: aws.String(b.modelID),
		System: []types.SystemContentBlock{
			&types.SystemContentBlockMemberText{Value: req.SystemPrompt},
		},
		Messages: []types.Message{
			{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{Value: req.Prompt},
				},
			},
		},
		InferenceConfig: b.buildInferenceConfig(maxTokens, temperature),
	}

	output, err := b.client.Converse(ctx, input)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrProviderFailed, err)
	}

	content := extractBedrockText(output)
	if content == "" {
		return "", fmt.Errorf("%w: no text content in Bedrock response", ErrInvalidResponse)
	}
	return content, nil
}

func (b *Bedrock) GenerateStructured(ctx context.Context, req GenerateRequest, schema string) (string, error) {
	req.Prompt = req.Prompt + "\n\nJSON schema hint:\n" + schema
	return b.GenerateText(ctx, req)
}

func (b *Bedrock) StreamDiagram(ctx context.Context, req GenerateRequest, onChunk func(StreamChunk)) error {
	maxTokens, temperature := applyBedrockDefaults(req.Config)

	input := &bedrockruntime.ConverseStreamInput{
		ModelId: aws.String(b.modelID),
		System: []types.SystemContentBlock{
			&types.SystemContentBlockMemberText{Value: req.SystemPrompt},
		},
		Messages: []types.Message{
			{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{Value: req.Prompt},
				},
			},
		},
		InferenceConfig: b.buildInferenceConfig(maxTokens, temperature),
	}

	output, err := b.client.ConverseStream(ctx, input)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProviderFailed, err)
	}

	stream := output.GetStream()
	defer stream.Close()

	for event := range stream.Events() {
		switch v := event.(type) {
		case *types.ConverseStreamOutputMemberContentBlockDelta:
			if text, ok := v.Value.Delta.(*types.ContentBlockDeltaMemberText); ok {
				if text.Value != "" {
					onChunk(StreamChunk{Delta: text.Value})
				}
			}
		case *types.ConverseStreamOutputMemberMessageStop:
			onChunk(StreamChunk{Done: true, StopReason: string(v.Value.StopReason)})
			return nil
		}
	}

	if err := stream.Err(); err != nil {
		return fmt.Errorf("%w: stream error: %v", ErrProviderFailed, err)
	}

	onChunk(StreamChunk{Done: true})
	return nil
}

func applyBedrockDefaults(cfg ProviderConfig) (maxTokens int, temperature float64) {
	maxTokens = cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}
	temperature = cfg.Temperature
	if temperature == 0 {
		temperature = 0.3
	}
	return
}

func extractBedrockText(output *bedrockruntime.ConverseOutput) string {
	if output.Output == nil {
		return ""
	}
	msg, ok := output.Output.(*types.ConverseOutputMemberMessage)
	if !ok {
		return ""
	}
	for _, block := range msg.Value.Content {
		if text, ok := block.(*types.ContentBlockMemberText); ok {
			return text.Value
		}
	}
	return ""
}
