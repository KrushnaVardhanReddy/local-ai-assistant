import sys
with open('wails-app/core/engine/pipeline.go', 'r') as f:
    content = f.read()

# Fix Turn field references
content = content.replace('driven.ChatMessage{Role: "user", Content: t.Transcript}', 'driven.ChatMessage{Role: "user", Content: t.InterviewerQuestion}')
content = content.replace('driven.ChatMessage{Role: "assistant", Content: t.Response}', 'driven.ChatMessage{Role: "assistant", Content: t.AISuggestion}')

# Fix AppendToken method call
# It seems there is no AppendToken method on SessionManager, we will just buffer it in our pipeline and it gets stored on complete or we use SetAISuggestion? Wait. What did the original app.go do?
# Original app.go: `a.sessionManager.AppendToken(token)` -> wait, did it have AppendToken in the old code?
# Let's check old app.go if possible. But old app.go is already overwritten! Let's check git diff.
