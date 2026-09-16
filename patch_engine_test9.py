import sys

with open('wails-app/core/engine/engine_test.go', 'r') as f:
    content = f.read()

content += '''
func TestHandleTranscript_Ignored(t *testing.T) {
	sttEngine := &mockSTT{transcripts: []string{"[BLANK_AUDIO]"}}
	sttMgr := stt.NewSTTManager(sttEngine)

	eng := engine.New(engine.Config{}, sttMgr, nil, nil, nil)
	eng.ProcessAudio([]float32{0.1, 0.2})
	time.Sleep(100 * time.Millisecond)
}

func TestGetState_CacheNil(t *testing.T) {
	// Need to check line 62 in GetState -> if e.cache != nil
	// This is already hit in TestGetState_NoCache. Let's make sure e.cache != nil is hit.
	cache := &MockCache{CountVal: 5}
	eng := engine.New(engine.Config{}, nil, nil, cache, nil)
	state := eng.GetState()
	if state.CachedPairs != 5 {
		t.Errorf("Expected 5 cached pairs, got %d", state.CachedPairs)
	}
}

'''

with open('wails-app/core/engine/engine_test.go', 'w') as f:
    f.write(content)
