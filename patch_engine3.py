import sys
with open('wails-app/core/engine/engine.go', 'r') as f:
    content = f.read()

# Add a method to explicitly set the engine state for things like vision mode.
content += '''
// UpdateState allows modifying the pipeline state manually, e.g., for vision queries.
func (e *StealthEngine) UpdateState(transcript, response string, thinking bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.transcript = transcript
	e.response = response
	e.thinking = thinking
}
'''
with open('wails-app/core/engine/engine.go', 'w') as f:
    f.write(content)

with open('wails-app/app.go', 'r') as f:
    app_content = f.read()

app_content = app_content.replace('''	a.stateMu.Lock()
	a.latestTranscript = "📸 [Screenshot Snip Captured]"
	a.latestResponse = ""
	a.latestThinking = true
	a.stateMu.Unlock()''', '''	a.engine.UpdateState("📸 [Screenshot Snip Captured]", "", true)''')

app_content = app_content.replace('''			a.stateMu.Lock()
			a.latestResponse = answerBuilder.String()
			a.latestThinking = true
			a.stateMu.Unlock()''', '''			a.engine.UpdateState("📸 [Screenshot Snip Captured]", answerBuilder.String(), true)''')

app_content = app_content.replace('''			a.stateMu.Lock()
			a.latestThinking = false
			a.stateMu.Unlock()''', '''			a.engine.UpdateState("📸 [Screenshot Snip Captured]", answerBuilder.String(), false)''')

app_content = app_content.replace('''			a.stateMu.Lock()
			a.latestResponse = fmt.Sprintf("Vision analysis error: %v", err)
			a.latestThinking = false
			a.stateMu.Unlock()''', '''			a.engine.UpdateState("📸 [Screenshot Snip Captured]", fmt.Sprintf("Vision analysis error: %v", err), false)''')

with open('wails-app/app.go', 'w') as f:
    f.write(app_content)
