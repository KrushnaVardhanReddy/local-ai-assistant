import sys
with open('wails-app/core/engine/pipeline.go', 'r') as f:
    content = f.read()

# Replace a.sessionManager.AppendToken(token) with a.sessionManager.SetAISuggestion(answerBuilder.String())
# We will just remove e.sessionMgr.AppendToken(token) entirely since it doesn't exist, and call SetAISuggestion in the completion handler (or we can call it on each token but it's not strictly necessary). Let's call it on each token as it builds up.
content = content.replace('e.sessionMgr.AppendToken(token)', 'e.sessionMgr.SetAISuggestion(answerBuilder.String())')

with open('wails-app/core/engine/pipeline.go', 'w') as f:
    f.write(content)
