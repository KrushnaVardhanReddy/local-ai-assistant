package driven

import "context"

// ChatMessage represents a single message in a conversation history.
type ChatMessage struct {
	Role    string // "system", "user", or "assistant"
	Content string
}

// StreamCallback is called for each token streamed from the LLM.
type StreamCallback func(token string)

// LLMPort is the driven port for language model completions.
type LLMPort interface {
	// StreamCompletion sends a question with optional history and streams the response.
	// systemPrompt is the product-specific system instruction.
	// onToken is called for each streamed token.
	// onDone is called once streaming is complete.
	StreamCompletion(
		ctx context.Context,
		question string,
		systemPrompt string,
		history []ChatMessage,
		onToken StreamCallback,
		onDone func(),
	) error

	// StreamVision sends a base64-encoded image with a prompt and streams the response.
	StreamVision(
		ctx context.Context,
		base64Image string,
		prompt string,
		onToken StreamCallback,
		onDone func(),
	) error
}
