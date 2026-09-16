import sys
with open('wails-app/core/engine/engine_test.go', 'r') as f:
    content = f.read()

content += '''
func TestGetSessionManagerAndCache(t *testing.T) {
	cache := &MockCache{}
	eng := engine.New(engine.Config{}, nil, nil, cache, nil)
	if eng.GetCache() != cache {
		t.Errorf("Expected GetCache to return the cache")
	}
	if eng.GetSessionManager() == nil {
		t.Errorf("Expected GetSessionManager to not be nil")
	}
}
'''
with open('wails-app/core/engine/engine_test.go', 'w') as f:
    f.write(content)

with open('wails-app/core/engine/pipeline.go', 'r') as f:
    pipe_content = f.read()

# Let's check which lines in handleTranscript are missed
# If e.llmBusy.TryLock() fails
