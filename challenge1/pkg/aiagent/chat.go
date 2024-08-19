package aiagent

import (
	"context"
	"fmt"
	"strings"
)

const (
	startTurnUser    = "<start_of_turn>user"
	startTurnModel   = "<start_of_turn>model"
	endTurn          = "<end_of_turn>"
	chatTurnTemplate = "%s\n%s%s"
)

var (
	systemInstructions = []string{
		"Ensure your answers are concise.",
		"Return answer as html without backticks.",
	}
)

// Chat is a helper to manage the conversation state for Gemma model
type Chat struct {
	fn      SendMessage
	history []string
}

type SendMessage func(context.Context, string) (string, error)

func NewChat(fn SendMessage) *Chat {
	return &Chat{fn: fn, history: []string{}}
}

func (chat *Chat) History() string {
	if len(chat.history) == 0 {
		return ""
	}
	return strings.Join(chat.history, "\n") + "\n"
}

func (chat *Chat) SendMessage(ctx context.Context, msg string) (string, error) {
	prompt := strings.Join(systemInstructions, "") + "\n" + chat.History()
	userMsg := fmt.Sprintf(chatTurnTemplate, startTurnUser, msg, endTurn)
	prompt = fmt.Sprintf("%s%s\n%s", prompt, strings.TrimRight(userMsg, " \n"), startTurnModel)

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
	chat.history = append(chat.history, fmt.Sprintf(chatTurnTemplate, startTurnModel, response, endTurn))
	return response, nil
}
