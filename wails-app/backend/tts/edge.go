package tts

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type EdgeTTSAdapter struct{}

func NewEdgeTTSAdapter() *EdgeTTSAdapter {
	return &EdgeTTSAdapter{}
}

func (e *EdgeTTSAdapter) Speak(text string) error {
	_, err := exec.LookPath("edge-tts")
	if err != nil {
		log.Println("WARNING: edge-tts is not installed. Skipping TTS.")
		return nil
	}

	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "tts_temp.mp3")

	cmd := exec.Command("edge-tts", "--text", text, "--write-media", tempFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("edge-tts failed: %w", err)
	}

	// For simplicity, we use ffplay or afplay to play the audio
	player := "ffplay"
	args := []string{"-nodisp", "-autoexit", tempFile}

	_, err = exec.LookPath("ffplay")
	if err != nil {
		player = "afplay" // macOS fallback
		args = []string{tempFile}
		_, err = exec.LookPath("afplay")
		if err != nil {
			log.Println("WARNING: neither ffplay nor afplay found. Cannot play audio.")
			return nil
		}
	}

	playCmd := exec.Command(player, args...)
	if err := playCmd.Run(); err != nil {
		return fmt.Errorf("failed to play audio: %w", err)
	}

	return nil
}
