import sys

def fix_mock(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # Replace Process with TranscribeStream
    content = content.replace('func (m *mockSTT) Process(audio []float32) (string, error) {\n	if len(m.transcripts) > 0 {\n		t := m.transcripts[0]\n		m.transcripts = m.transcripts[1:]\n		return t, nil\n	}\n	return "", nil\n}', 'func (m *mockSTT) TranscribeStream(audio []float32) (chan string, error) {\n\tch := make(chan string, 1)\n\tif len(m.transcripts) > 0 {\n\t\tch <- m.transcripts[0]\n\t\tm.transcripts = m.transcripts[1:]\n\t}\n\tclose(ch)\n\treturn ch, nil\n}')
    # Remove unused Initialize / Terminate / ProviderName
    content = content.replace('func (m *mockSTT) Initialize() error   { return nil }\n', '')
    content = content.replace('func (m *mockSTT) Terminate() error    { return nil }\n', '')
    content = content.replace('func (m *mockSTT) ProviderName() string { return "mock" }\n', '')

    with open(filepath, 'w') as f:
        f.write(content)

fix_mock('wails-app/core/engine/engine_test.go')

with open('wails-app/core/engine/engine_error_test.go', 'r') as f:
    err_content = f.read()

err_content = err_content.replace('func (m *errorSTT) Process(audio []float32) (string, error) {\n	return "", fmt.Errorf("mock error")\n}', 'func (m *errorSTT) TranscribeStream(audio []float32) (chan string, error) {\n\treturn nil, fmt.Errorf("mock error")\n}')
err_content = err_content.replace('func (m *errorSTT) Initialize() error   { return nil }\n', '')
err_content = err_content.replace('func (m *errorSTT) Terminate() error    { return nil }\n', '')
err_content = err_content.replace('func (m *errorSTT) ProviderName() string { return "error" }\n', '')

with open('wails-app/core/engine/engine_error_test.go', 'w') as f:
    f.write(err_content)
