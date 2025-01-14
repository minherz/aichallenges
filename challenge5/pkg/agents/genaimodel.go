package agents

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"github.com/minherz/aichallenges/challenge1/pkg/utils"
)

type GenAIModel struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewGenAIModel(ctx context.Context) (*GenAIModel, error) {
	region := utils.GetEnvOrDefault("REGION_NAME", "")
	if region == "" {
		v, err := utils.Region(ctx)
		if err != nil || v == "" {
			return nil, fmt.Errorf("location is missing: %w", err)
		}
		region = v
	}
	projectID := utils.GetEnvOrDefault("PROJECT_ID", utils.GetEnvOrDefault("GOOGLE_CLOUD_PROJECT", ""))
	if projectID == "" {
		v, err := utils.ProjectID(ctx)
		if err != nil || v == "" {
			return nil, fmt.Errorf("project ID is missing: %w", err)
		}
		projectID = v
	}
	c, err := genai.NewClient(ctx, projectID, region)
	if err != nil {
		return nil, fmt.Errorf("could not initialize Vertex AI client: %w", err)
	}
	modelName := utils.GetEnvOrDefault("GENAI_MODEL", "gemini-1.5-flash-001")
	m := c.GenerativeModel(modelName)
	return &GenAIModel{client: c, model: m}, nil
}

func (m *GenAIModel) Close() {
	if m.client != nil {
		m.client.Close()
	}
}

func (m *GenAIModel) Inference(ctx context.Context, prompt string) (string, error) {
	resp, err := m.model.GenerateContent(ctx, []genai.Part{genai.Text(prompt)}...)
	if err != nil {
		return "", err
	}
	candidates := resp.Candidates
	if len(candidates) == 0 || candidates[0] == nil {
		return "", fmt.Errorf("model has no answer")
	}
	parts := candidates[0].Content.Parts
	text := make([]string, len(parts))
	for _, part := range parts {
		if t, ok := part.(genai.Text); !ok || len(string(t)) > 0 {
			text = append(text, string(t))
		}
	}
	return strings.Join(text, ". "), nil
}
