package aiagent

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/minherz/aichallenges/challenge1/pkg/utils"
	opt "google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	endpointIDEnvVar       = "ENDPOINT_ID"
	endpointLocationEnvVar = "REGION_NAME"
)

type Agent struct {
	c           *aiplatform.PredictionClient
	endpointUri string
	sessions    map[string]*ChatSession
}

type ChatSession struct {
	id       string
	messages *Chat
}

func NewAgent(ctx context.Context, e *echo.Echo) (*Agent, error) {
	var (
		c   *aiplatform.PredictionClient
		err error
	)
	projectID := utils.GetenvWithDefault("PROJECT_ID", "")
	if projectID == "" {
		if projectID, err = utils.ProjectID(ctx); err != nil {
			return nil, fmt.Errorf("could not retrieve current project ID: %w", err)
		}
	}
	region := utils.GetenvWithDefault(endpointLocationEnvVar, "")
	if region == "" {
		return nil, fmt.Errorf("could not retrieve model location from environment")
	}
	modelEndpointID := utils.GetenvWithDefault(endpointIDEnvVar, "")
	if modelEndpointID == "" {
		return nil, fmt.Errorf("could not retrieve model endpoint ID from environment")
	}
	endpointUrl := fmt.Sprintf("%s-aiplatform.googleapis.com:443", region)
	c, err = aiplatform.NewPredictionClient(ctx, opt.WithEndpoint(endpointUrl))
	if err != nil {
		return nil, fmt.Errorf("could not initialize AI client: %w", err)
	}
	endpointUri := fmt.Sprintf("projects/%s/locations/%s/endpoints/%s", projectID, region, modelEndpointID)
	agent := &Agent{c: c, endpointUri: endpointUri, sessions: make(map[string]*ChatSession)}
	slog.Debug("initialized ai agent", "project", projectID, "region", region, "endpoint_id", modelEndpointID)

	// setup handlers
	e.POST("/ask", agent.onAsk)

	return agent, nil
}

func (a *Agent) Close() {
	if a.c != nil {
		a.Close()
	}
}

func (a *Agent) SendMessage(ctx context.Context, msg string) (string, error) {
	instance, err := structpb.NewValue(map[string]interface{}{
		"prompt":      structpb.NewStringValue(msg),
		"temperature": 0.2,
		"max_tokens":  512,
	})
	if err != nil {
		return "", fmt.Errorf("failed to setup request parameters: %w", err)
	}
	r := &aiplatformpb.PredictRequest{
		Endpoint:  a.endpointUri,
		Instances: []*structpb.Value{instance},
	}
	resp, err := a.c.Predict(ctx, r)
	if err != nil {
		return "", fmt.Errorf("model failed to respond: %w", err)
	}
	if len(resp.Predictions) == 0 {
		return "", fmt.Errorf("model returned empty response: %v", resp)
	}
	return resp.Predictions[0].GetStringValue(), nil
}

func (a *Agent) getOrCreateSession(id string) *ChatSession {
	s, ok := a.sessions[id]
	if !ok {
		s = &ChatSession{id: id, messages: NewChat("", a.SendMessage)}
		a.sessions[id] = s
	}
	return s
}

type BaseResponse struct {
	Error string `json:"error,omitempty"`
}

type AskRequest struct {
	SessionID string `json:"session,omitempty"`
	Message   string `json:"message,omitempty"`
	Location  string `json:"loc,omitempty"`
	Company   string `json:"company,omitempty"`
}

type AskResponse struct {
	BaseResponse
	SessionID string `json:"session,omitempty"`
	Response  string `json:"response,omitempty"`
}

func (a *Agent) onAsk(ectx echo.Context) error {
	r := &AskRequest{}
	if err := ectx.Bind(&r); err != nil {
		return reportError(ectx, http.StatusBadRequest, fmt.Errorf("invalid input: %w", err))
	}
	if r.Message == "" {
		return reportError(ectx, http.StatusBadRequest, fmt.Errorf("request message is empty"))
	}
	if r.SessionID == "" {
		id, err := newID()
		if err != nil {
			return reportError(ectx, http.StatusBadRequest, fmt.Errorf("cannot generate session ID: %w", err))
		}
		r.SessionID = id
	}
	s := a.getOrCreateSession(r.SessionID)
	response, err := s.messages.SendMessage(ectx.Request().Context(), r.Message)
	if err != nil {
		return reportError(ectx, http.StatusInternalServerError, fmt.Errorf("chat response error: %w", err))
	}
	return ectx.JSON(http.StatusOK, AskResponse{SessionID: r.SessionID, Response: response})
}

func newID() (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cannot generate session ID: %w", err)
	}
	return uuid.String(), nil
}

func reportError(ectx echo.Context, code int, err error) error {
	msg := err.Error()
	slog.Error(msg, "response_code", code)
	return ectx.JSON(code, AskResponse{BaseResponse: BaseResponse{Error: msg}})
}
