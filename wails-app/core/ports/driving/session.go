package driving

// SessionPort manages a live interaction session (e.g., interview, presentation).
type SessionPort interface {
	// StartTurn marks the beginning of a new user input turn.
	StartTurn(input string)

	// CompleteTurn marks the end of a turn after LLM response is complete.
	CompleteTurn()

	// EndSession finalises the session and returns summary data.
	// The returned map contains at minimum: "turn_count", "turns".
	EndSession() (map[string]interface{}, error)
}
