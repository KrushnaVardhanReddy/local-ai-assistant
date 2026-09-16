import sys
with open('wails-app/app.go', 'r') as f:
    content = f.read()

# Since we need to delete qaCache and we can't easily proxy all these methods through CachePort without modifying the Port, we can either:
# 1. Modify the Port (but that's not in the instructions and breaks hexagonal encapsulation for UI specific things)
# 2. Expose the raw adapter or cast it? No.
# Wait, the instruction says:
# "Remove the now-redundant fields: latestTranscript, latestResponse, latestThinking, stateMu, llmBusy, qaCache, sessionManager (they live inside StealthEngine now)."
# This means we should add a getter in engine?
# Actually, the instructions say "Remove the now-redundant fields ...".
# What if we just make GetCacheStats use GetState()? We already have `s := a.engine.GetState()` giving `CachedPairs`.
content = content.replace('''// GetCacheStats returns the number of cached Q&A pairs and estimated tokens saved
func (a *App) GetCacheStats() map[string]interface{} {
	if a.qaCache == nil {
		return map[string]interface{}{
			"cached_pairs":           0,
			"estimated_tokens_saved": 0,
		}
	}
	count := a.qaCache.GetCount()
	return map[string]interface{}{
		"cached_pairs":           count,
		"estimated_tokens_saved": count * 250,
	}
}''', '''// GetCacheStats returns the number of cached Q&A pairs and estimated tokens saved
func (a *App) GetCacheStats() map[string]interface{} {
	s := a.engine.GetState()
	return map[string]interface{}{
		"cached_pairs":           s.CachedPairs,
		"estimated_tokens_saved": s.CachedPairs * 250,
	}
}''')

# For GetCacheItems, DeleteCacheItems, ClearCache... where are they used?
# If we must keep them, we can add a CachePort getter to the engine.
content = content.replace('a.qaCache', 'a.engine.GetCache()')

with open('wails-app/app.go', 'w') as f:
    f.write(content)
