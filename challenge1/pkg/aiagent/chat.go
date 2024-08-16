package aiagent

import (
	"context"
	"fmt"
	"strings"
)

const (
	startTurnUser  = "<start_of_turn>user\n"
	startTurnModel = "<start_of_turn>model\n"
	endTurn        = "<end_of_turn>\n"
)

// Chat is a helper to manage the conversation state for Gemma model
type Chat struct {
	system  string
	fn      SendMessage
	history []string
}

type SendMessage func(context.Context, string) (string, error)

func NewChat(systemInstructions string, fn SendMessage) *Chat {
	return &Chat{fn: fn, system: systemInstructions, history: make([]string, 2)}
}

func (chat *Chat) Prompt() string {
	if len(chat.history) == 0 {
		return ""
	}
	return strings.Join(chat.history, "\n") + "\n"
}

func (chat *Chat) SendMessage(ctx context.Context, msg string) (string, error) {
	prompt := chat.Prompt()
	if chat.system != "" {
		prompt = fmt.Sprintf("%s\n%s", chat.system, prompt)
	}
	userMsg := fmt.Sprintf("%s%s%s", startTurnUser, msg, endTurn)
	prompt = fmt.Sprintf("%s%s", prompt, userMsg)

	response, err := chat.fn(ctx, prompt)
	if err != nil {
		return "", err
	}
	// sanitize response to exclude anything outside model's tags
	var pos int
	pos = strings.LastIndex(response, startTurnModel)
	if pos >= 0 {
		response = response[pos+len(startTurnModel):]
	}
	pos = strings.LastIndex(response, endTurn)
	if pos >= 0 {
		response = response[:pos]
	}
	// update history
	chat.history = append(chat.history, userMsg)
	chat.history = append(chat.history, fmt.Sprintf("%s%s%s", startTurnModel, response, endTurn))
	return response, nil
}
