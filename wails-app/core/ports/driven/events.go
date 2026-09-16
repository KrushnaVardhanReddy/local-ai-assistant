package driven

// EventPort is the driven port for emitting UI events.
// Decouples the core engine from the Wails runtime.
type EventPort interface {
	// Emit sends an event with optional payload to the UI layer.
	// event names follow the existing convention: "on_transcript",
	// "on_response_start", "on_response_token", "on_response_end".
	Emit(event string, payload any)
}
