package config_test

import (
	"os"
	"testing"
	"wails-app/backend/config"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Save environment to restore later
	origEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range origEnv {
			if len(env) > 0 {
				os.Setenv(env[:len(env)/2], env[len(env)/2+1:]) // Approximate, will use a cleaner way
			}
		}
	}()

	t.Run("Default configuration", func(t *testing.T) {
		os.Clearenv()
		cfg, err := config.LoadConfig()
		assert.NoError(t, err)
		assert.Equal(t, "openai", cfg.LLMProvider)
		assert.Equal(t, "gpt-4o", cfg.LLMModel)
		assert.Equal(t, "https://api.openai.com/v1", cfg.LLMBaseURL)
		assert.False(t, cfg.StealthMode)
		assert.True(t, cfg.RagEnabled)
	})

	t.Run("Custom configuration", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("LLM_PROVIDER", "groq")
		os.Setenv("GROQ_API_KEY", "test-key")
		os.Setenv("LLM_MODEL", "llama3")
		os.Setenv("STEALTH_MODE", "true")
		os.Setenv("RAG_ENABLED", "false")

		cfg, err := config.LoadConfig()
		assert.NoError(t, err)
		assert.Equal(t, "groq", cfg.LLMProvider)
		assert.Equal(t, "llama3", cfg.LLMModel)
		assert.Equal(t, "test-key", cfg.GroqAPIKey)
		assert.True(t, cfg.StealthMode)
		assert.False(t, cfg.RagEnabled)
	})

	t.Run("Missing Groq API Key", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("LLM_PROVIDER", "groq")

		_, err := config.LoadConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GROQ_API_KEY environment variable is not set")
	})
}
