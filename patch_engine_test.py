import sys

def add_close(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # Add Close to mockSTT
    content = content.replace('func (m *mockSTT) Terminate() error    { return nil }', 'func (m *mockSTT) Terminate() error    { return nil }\nfunc (m *mockSTT) Close() error        { return nil }')
    # Fix unused context
    content = content.replace('	"context"\n', '')

    with open(filepath, 'w') as f:
        f.write(content)

add_close('wails-app/core/engine/engine_test.go')

with open('wails-app/core/engine/engine_error_test.go', 'r') as f:
    err_content = f.read()

err_content = err_content.replace('func (m *errorSTT) Terminate() error    { return nil }', 'func (m *errorSTT) Terminate() error    { return nil }\nfunc (m *errorSTT) Close() error        { return nil }')
err_content = err_content.replace('sttMgr := stt.NewSTTManager(&errorSTT{})', '_ = stt.NewSTTManager(&errorSTT{})')

with open('wails-app/core/engine/engine_error_test.go', 'w') as f:
    f.write(err_content)
