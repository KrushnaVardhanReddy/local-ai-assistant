import sys
with open('wails-app/app.go', 'r') as f:
    content = f.read()

content = content.replace('\tsessionManager *session.SessionManager\n', '')
content = content.replace('a.sessionManager', 'a.engine.GetSessionManager()')

with open('wails-app/app.go', 'w') as f:
    f.write(content)

with open('wails-app/core/engine/engine.go', 'r') as f:
    engine_content = f.read()

engine_content += '\n// GetSessionManager returns the session manager instance.\nfunc (e *StealthEngine) GetSessionManager() *session.SessionManager {\n\treturn e.sessionMgr\n}\n'

with open('wails-app/core/engine/engine.go', 'w') as f:
    f.write(engine_content)
