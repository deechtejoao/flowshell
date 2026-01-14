package core

import (
	"context"
	"strings"
	"sync"
)

type PromptRequest struct {
	Title        string
	Message      string
	DefaultValue string

	// Channels for response
	Reply  chan string
	Cancel chan struct{}
}

// Global prompt state
var (
	CurrentPrompt *PromptRequest
	promptMutex   sync.Mutex
)

// RequestPrompt makes a blocking call to prompt the user.
// It returns the user's input string, or an error if cancelled/context done.
func RequestPrompt(ctx context.Context, title, message, defaultValue string) (string, error) {
	// Create channels
	reply := make(chan string, 1)
	cancel := make(chan struct{}, 1)

	req := &PromptRequest{
		Title:        title,
		Message:      message,
		DefaultValue: defaultValue,
		Reply:        reply,
		Cancel:       cancel,
	}

	// Set global prompt
	promptMutex.Lock()
	if CurrentPrompt != nil {
		promptMutex.Unlock()
		return "", ctx.Err() // Or checks busy
	}
	CurrentPrompt = req
	promptMutex.Unlock()

	// Ensure cleanup
	defer func() {
		promptMutex.Lock()
		if CurrentPrompt == req {
			CurrentPrompt = nil
		}
		promptMutex.Unlock()
	}()

	// Wait for response
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case val := <-reply:
		return strings.TrimSpace(val), nil
	case <-cancel:
		return "", context.Canceled
	}
}

func RespondToPrompt(val string) {
	promptMutex.Lock()
	req := CurrentPrompt
	promptMutex.Unlock()

	if req != nil {
		select {
		case req.Reply <- val:
		default:
		}
	}
}

func CancelPrompt() {
	promptMutex.Lock()
	req := CurrentPrompt
	promptMutex.Unlock()

	if req != nil {
		select {
		case req.Cancel <- struct{}{}:
		default:
		}
	}
}
