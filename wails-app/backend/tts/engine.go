package tts

type Engine interface {
	Speak(text string) error
	Stop() error
}
