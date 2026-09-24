package tts

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type EdgeTTSAdapter struct {
	mu        sync.Mutex
	activeCmd *exec.Cmd
}

func NewEdgeTTSAdapter() *EdgeTTSAdapter {
	return &EdgeTTSAdapter{}
}

func (e *EdgeTTSAdapter) Speak(text string) error {
	e.Stop() // Kill any currently playing audio

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

	e.mu.Lock()
	e.activeCmd = playCmd
	e.mu.Unlock()

	if err := playCmd.Run(); err != nil {
		// If killed, it might return an error, which is fine
		return fmt.Errorf("failed to play audio: %w", err)
	}

	e.mu.Lock()
	e.activeCmd = nil
	e.mu.Unlock()

	return nil
}

func (e *EdgeTTSAdapter) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.activeCmd != nil && e.activeCmd.Process != nil {
		_ = e.activeCmd.Process.Kill()
		e.activeCmd = nil
	}
	return nil
}
