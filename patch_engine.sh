cat << 'INNER_EOF' >> wails-app/core/engine/engine.go

func (e *StealthEngine) GetManualMode() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.manualMode
}
INNER_EOF
