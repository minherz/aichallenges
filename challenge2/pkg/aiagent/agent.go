package aiagent

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/minherz/aichallenges/challenge2/pkg/utils"
)

const (
	modelNameEnvVar             = "GEMINI_MODEL_NAME"
	regionEnvVar                = "REGION_NAME"
	systemInstructionPathEnvVar = "SYS_INSTRUCTION_PATH"
	systemInstructionFilePath   = "current/system_instructions.txt"
	// from https://cloud.google.com/vertex-ai/generative-ai/docs/learn/model-versions
	defaultModelName = "gemini-1.5-flash-001"
)

var (
	defaultSystemInstructions = []string{
		"You are a friendly and helpful waste sorting assistant.",
		"You help to decide what type of cart the waste should be sorted to.",
		"Answer according to cart types used in Washington state in the United States of America, unless the user explicitly specifies another location and waste collection company.",
		"Ensure your answers are concise, unless the user requests a more complete approach.",
		"When presented with inquiries seeking information, provide answers that reflect a deep understanding of the field, guaranteeing their correctness.",
		"For prompts involving reasoning, provide a clear explanation of each step in the reasoning process before presenting the final answer.",
		"For any non-English queries, respond that you understand English only.",
		"Return answer as html without backticks.",
	}
)

type Agent struct {
	c        *genai.Client
	m        *genai.GenerativeModel
	sessions map[string]*ChatSession
	w        *utils.FileWatcher
}

type ChatSession struct {
	id   string
	chat *genai.ChatSession
}

func NewAgent(ctx context.Context, e *echo.Echo) (*Agent, error) {
	var (
		c                  *genai.Client
		m                  *genai.GenerativeModel
		w                  *utils.FileWatcher
		err                error
		systemInstructions string
	)
	projectID := utils.GetenvWithDefault("PROJECT_ID", utils.GetenvWithDefault("GOOGLE_CLOUD_PROJECT", ""))
	if projectID == "" {
		if projectID, err = utils.ProjectID(ctx); err != nil {
			return nil, fmt.Errorf("could not retrieve current project ID: %w", err)
		}
	}
	region := utils.GetenvWithDefault(regionEnvVar, "")
	if region == "" {
		if region, err = utils.Region(ctx); err != nil {
			return nil, fmt.Errorf("could not retrieve location from the model: %w", err)
		}
	}
	if c, err = genai.NewClient(ctx, projectID, region); err != nil {
		return nil, fmt.Errorf("could not initialize Vertex AI client: %w", err)
	}
	modelName := utils.GetenvWithDefault(modelNameEnvVar, defaultModelName)
	m = c.GenerativeModel(modelName)
	path := utils.GetenvWithDefault(systemInstructionPathEnvVar, "")
	if path != "" {
		path := filepath.Join(path, systemInstructionFilePath)
		if text, err := os.ReadFile(path); err == nil {
			systemInstructions = string(text)
			w, _ = utils.NewFileWatcher(path)
		}
	}
	if systemInstructions == "" {
		systemInstructions = strings.Join(defaultSystemInstructions, " ")
	}
	m.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemInstructions)},
	}
	agent := &Agent{c: c, m: m, w: w, sessions: make(map[string]*ChatSession)}
	if w != nil {
		w.Watch(ctx, agent.loadSystemInstructions)
	}
	slog.Debug("initialized ai agent", "project", projectID, "region", region, "mode_name", modelName, "system_instructions", systemInstructions)

	// setup handlers
	e.POST("/ask", agent.onAsk)

	return agent, nil
}

func (a *Agent) Close() {
	if a.w != nil {
		a.w.Stop()
	}
	if a.c != nil {
		a.Close()
	}
}

func (a *Agent) getOrCreateSession(id string) *ChatSession {
	s, ok := a.sessions[id]
	if !ok {
		s = &ChatSession{id: id, chat: a.m.StartChat()}
		a.sessions[id] = s
	}
	return s
}

func (a *Agent) loadSystemInstructions(path string) {
	text, err := os.ReadFile(path)
	if err != nil {
		slog.Error("failed to read system instructions", "error", err, "path", path)
		return
	}
	instructions := string(text)
	a.m.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(instructions)},
	}
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
	Message   string `json:"message,omitempty"`
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
	response, err := s.chat.SendMessage(ectx.Request().Context(), genai.Text(r.Message))
	if err != nil {
		return reportError(ectx, http.StatusInternalServerError, fmt.Errorf("chat response error: %w", err))
	}
	msg := utils.ProcessResponse(response)
	slog.Debug("ask request processed", "session", r.SessionID, "prompt", r.Message, "response", msg)
	return ectx.JSON(http.StatusOK, AskResponse{SessionID: r.SessionID, Message: msg})
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
