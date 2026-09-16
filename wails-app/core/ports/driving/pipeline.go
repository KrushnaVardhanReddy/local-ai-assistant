package driving

// PipelineState holds the current state of the audio→LLM pipeline.
type PipelineState struct {
	Transcript  string
	Response    string
	Thinking    bool
	CachedPairs int
}

// PipelinePort is the primary driving port — the UI calls these methods.
type PipelinePort interface {
	// ProcessAudio takes a PCM float32 chunk and routes it through the pipeline.
	// Results are delivered asynchronously via the EventPort.
	ProcessAudio(samples []float32) error

	// AskQuestion bypasses audio and sends a direct text question to the LLM.
	AskQuestion(question string) error

	// SetStealth enables or disables OS-level screen capture exclusion.
	SetStealth(enable bool) error

	// GetState returns a snapshot of the current pipeline state for UI polling.
	GetState() PipelineState

	// ClearState resets the transcript, response, and thinking flag.
	ClearState()
}
