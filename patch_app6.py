import sys
with open('wails-app/app.go', 'r') as f:
    content = f.read()

# Replace GetState completely
get_state_original = '''func (a *App) GetState() map[string]interface{} {
	a.stateMu.RLock()
	defer a.stateMu.RUnlock()

	var count int
	if a.qaCache != nil {
		count = a.qaCache.GetCount()
	}

	return map[string]interface{}{
		"transcript":             a.latestTranscript,
		"response":               a.latestResponse,
		"thinking":               a.latestThinking,
		"cached_pairs":           count,
		"estimated_tokens_saved": count * 250,
	}
}'''

get_state_new = '''func (a *App) GetState() map[string]interface{} {
	s := a.engine.GetState()
	return map[string]interface{}{
		"transcript":             s.Transcript,
		"response":               s.Response,
		"thinking":               s.Thinking,
		"cached_pairs":           s.CachedPairs,
		"estimated_tokens_saved": s.CachedPairs * 250,
	}
}'''

content = content.replace(get_state_original, get_state_new)

# Replace ClearState completely
clear_state_original = '''func (a *App) ClearState() {
	a.stateMu.Lock()
	defer a.stateMu.Unlock()
	a.latestTranscript = ""
	a.latestResponse = ""
	a.latestThinking = false
}'''
clear_state_new = '''func (a *App) ClearState() {
	a.engine.ClearState()
}'''
content = content.replace(clear_state_original, clear_state_new)


with open('wails-app/app.go', 'w') as f:
    f.write(content)
