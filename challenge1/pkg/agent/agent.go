package agent

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/minherz/aichallenges/challenge1/pkg/utils"
)

type Agent struct {
	sessions map[string]*ChatSession
}

type ChatSession struct {
	id string
}

func NewAgent(ctx context.Context, e *echo.Echo) (*Agent, error) {
	var (
		projectID, region string
		err               error
	)
	if projectID, err = utils.ProjectID(ctx); err != nil {
		return nil, fmt.Errorf("could not retrieve current project ID: %w", err)
	}
	if region, err = utils.Region(ctx); err != nil {
		return nil, fmt.Errorf("could not retrieve current region: %w", err)
	}
	agent := &Agent{sessions: make(map[string]*ChatSession)}
	slog.Debug("initialized vertex ai", "project", projectID, "region", region)

	// setup handlers
	e.POST("/ask", agent.onAsk)

	return agent, nil
}

func (a *Agent) Close() {
	// if a.c != nil {
	// 	a.Close()
	// }
}

func (a *Agent) getOrCreateSession(id string) *ChatSession {
	s, ok := a.sessions[id]
	if !ok {
		s = &ChatSession{id: id}
		a.sessions[id] = s
	}
	return s
}

type BaseResponse struct {
	Error string `json:"error,omitempty"`
}

type AskRequest struct {
	SessionID string `json:"sessionId,omitempty"`
	Message   string `json:"message,omitempty"`
	Location  string `json:"loc,omitempty"`
	Company   string `json:"company,omitempty"`
}

type AskResponse struct {
	BaseResponse
	SessionID string `json:"sessionId,omitempty"`
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
		id, err := newSession()
		if err != nil {
			return reportError(ectx, http.StatusBadRequest, fmt.Errorf("cannot generate session ID: %w", err))
		}
		r.SessionID = id
	}
	s := a.getOrCreateSession(r.SessionID)
	/* remove the following IF condition */
	if s == nil {
		return nil
	}

	return ectx.JSON(http.StatusOK, AskResponse{SessionID: r.SessionID, Response: ""})
}

func newSession() (string, error) {
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
