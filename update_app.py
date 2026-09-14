import re

with open('wails-app/app.go', 'r') as f:
    content = f.read()

# Add audioCapture field to App struct
content = re.sub(
    r'(sttManager \*stt\.STTManager)',
    r'\1\n\taudioCapture *stt.AudioCapture',
    content
)

# Initialize audioCapture in NewApp
content = re.sub(
    r'(sttManager: stt\.NewSTTManager\(initialEngine\),)',
    r'\1\n\t\taudioCapture: stt.NewAudioCapture(func(samples []float32) {\n\t\t\t// Route audio to STT manager if we want continuous transcription\n\t\t\t// e.g. a.sttManager.TranscribeStream(samples)\n\t\t}),',
    content
)

# Add bindings
new_methods = """
func (a *App) GetAudioDevices() ([]stt.AudioDevice, error) {
	return a.audioCapture.GetAudioDevices()
}

func (a *App) SetAudioDevice(id string, isLoopback bool) error {
	return a.audioCapture.SetAudioDevice(id, isLoopback)
}
"""

if "GetAudioDevices" not in content:
    content += "\n" + new_methods

with open('wails-app/app.go', 'w') as f:
    f.write(content)
