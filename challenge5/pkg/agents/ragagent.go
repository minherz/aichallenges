package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	dimensionality = 768
)

type RagAgent struct {
	embedding *Embedding
	model     *GenAIModel
	connector *BQConnector
}

type RagAgentRequest struct {
	Message string `json:"message"`
}

type RagAgentResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

func NewRagAgent(ctx context.Context) (agent *RagAgent, err error) {
	embedding, err := NewEmbedding(ctx)
	if err != nil {
		return nil, err
	}
	model, err := NewGenAIModel(ctx)
	if err != nil {
		return nil, err
	}
	connector, err := NewBigQueryConnector(ctx)
	if err != nil {
		return nil, err
	}
	agent = &RagAgent{embedding: embedding, model: model, connector: connector}
	return
}

func (c *RagAgent) Close() {
	if c.embedding != nil {
		c.embedding.Close()
	}
	if c.model != nil {
		c.model.Close()
	}
	if c.connector != nil {
		c.connector.Close()
	}
}

func (c *RagAgent) Handler(ectx echo.Context) error {
	r := &RagAgentRequest{}
	if err := ectx.Bind(&r); err != nil {
		return echoError(ectx, http.StatusBadRequest, fmt.Errorf("invalid input: %w", err))
	}
	if r.Message == "" {
		return echoError(ectx, http.StatusBadRequest, fmt.Errorf("request message is empty"))
	}
	ctx := ectx.Request().Context()
	evector, err := c.embedding.Embed(ctx, r.Message)
	if err != nil {
		return echoError(ectx, http.StatusInternalServerError, err)
	}
	hotels, err := c.connector.matchEmbedding(ctx, evector)
	if err != nil {
		return echoError(ectx, http.StatusInternalServerError, err)
	}
	prompts := []string{
		r.Message,
		"Use the following list of hotels for suggestions.",
		"Information about each hotel is JSON record with the following fields:",
		"* name - the name of the hotel",
		"* description - the description of the hotel",
		"* address - the location of the hotel",
		"* attractions - the list of attractions near the hotel",
		"",
	}
	for _, hotel := range hotels {
		record, _ := json.Marshal(hotel)
		prompts = append(prompts, string(record))
	}
	response, err := c.model.Inference(ctx, strings.Join(prompts, "\n"))
	if err != nil {
		return echoError(ectx, http.StatusInternalServerError, err)
	}
	return ectx.JSON(http.StatusOK, RagAgentResponse{Message: response})
}

func echoError(ectx echo.Context, code int, err error) error {
	msg := err.Error()
	slog.Error(msg, "response_code", code)
	return ectx.JSON(code, RagAgentResponse{Error: msg})
}
