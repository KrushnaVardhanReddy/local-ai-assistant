import sys
with open('wails-app/core/engine/pipeline.go', 'r') as f:
    content = f.read()

content = content.replace('t.Transcript', 't.InterviewerQuestion')
content = content.replace('t.Response', 't.AISuggestion')

with open('wails-app/core/engine/pipeline.go', 'w') as f:
    f.write(content)
