import sys

content = open("wails-app/app_test.go").read()

import re
matches = re.finditer(r"func TestSetProxyToken\(t \*testing\.T\) \{", content)

count = 0
start_idx = -1
for match in matches:
    count += 1
    if count == 2:
        start_idx = match.start()
        break

if start_idx != -1:
    # remove from start_idx to end
    content = content[:start_idx]
    open("wails-app/app_test.go", "w").write(content)
