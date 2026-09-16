import sys
import re

with open('wails-app/app.go', 'r') as f:
    content = f.read()

content = content.replace('\t"wails-app/backend/filter"\n', '')

with open('wails-app/app.go', 'w') as f:
    f.write(content)
