import sys
with open('wails-app/core/engine/engine_test.go', 'r') as f:
    content = f.read()

content = content.replace('eng.AskQuestion("typed question")', 'eng.AskQuestion("hello this is a valid question from the user to the bot")')

with open('wails-app/core/engine/engine_test.go', 'w') as f:
    f.write(content)
