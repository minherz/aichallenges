package utils

import (
	"os"
	"strings"

	"cloud.google.com/go/vertexai/genai"
)

func GetenvWithDefault(name, defaultValue string) string {
	v := os.Getenv(name)
	if v != "" {
		return v
	}
	return defaultValue
}

func ProcessResponse(r *genai.GenerateContentResponse) string {
	if len(r.Candidates) == 0 || r.Candidates[0] == nil {
		return "<empty>"
	}
	c := r.Candidates[0].Content
	text := make([]string, len(c.Parts))
	for _, part := range c.Parts {
		if t, ok := part.(genai.Text); !ok || len(string(t)) > 0 {
			text = append(text, string(t))
		}
	}
	return strings.Join(text, ". ")
}
