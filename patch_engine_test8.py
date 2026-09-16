import sys

with open('wails-app/core/engine/engine_test.go', 'r') as f:
    content = f.read()

content += '''
func TestHandleTranscript_SessionManagerRecentTurns(t *testing.T) {
	sttEngine := &mockSTT{transcripts: []string{"hello this is a valid question with session manager"}}
	sttMgr := stt.NewSTTManager(sttEngine)

	llm := &MockLLM{}
	eng := engine.New(engine.Config{}, sttMgr, llm, nil, nil)

    // Add turns to session manager
    sess := eng.GetSessionManager()
    sess.StartTurn("old q 1")
    sess.SetCandidateResponse("old a 1")
    sess.SetAISuggestion("old s 1")
    sess.CompleteTurn()
    sess.StartTurn("old q 2")
    sess.SetCandidateResponse("old a 2")
    sess.SetAISuggestion("old s 2")
    sess.CompleteTurn()
    sess.StartTurn("old q 3")
    sess.SetCandidateResponse("old a 3")
    sess.SetAISuggestion("old s 3")
    sess.CompleteTurn()
    sess.StartTurn("old q 4")
    sess.SetCandidateResponse("old a 4")
    sess.SetAISuggestion("old s 4")
    sess.CompleteTurn()

	eng.ProcessAudio([]float32{0.1, 0.2})
	time.Sleep(100 * time.Millisecond)
}
'''

with open('wails-app/core/engine/engine_test.go', 'w') as f:
    f.write(content)
