package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type AppConfig struct {
	LLMProvider             string `env:"LLM_PROVIDER" envDefault:"openai"`
	LLMModel                string `env:"LLM_MODEL" envDefault:"gpt-4o"`
	LLMBaseURL              string `env:"LLM_BASE_URL" envDefault:"https://api.openai.com/v1"`
	OpenAIAPIKey            string `env:"OPENAI_API_KEY"`
	GroqAPIKey              string `env:"GROQ_API_KEY"`
	StealthMode             bool   `env:"STEALTH_MODE" envDefault:"false"`
	RagEnabled              bool   `env:"RAG_ENABLED" envDefault:"true"`
	WhisperModelPath        string `env:"WHISPER_MODEL_PATH"`
	STTProvider             string `env:"STT_PROVIDER"`
	STTModel                string `env:"STT_MODEL"`
	LocalSTTEngine          string `env:"LOCAL_STT_ENGINE"`
	SupabaseEdgeURL         string `env:"SUPABASE_EDGE_URL" envDefault:"https://api.barnowl.ai/v1/functions/llm-proxy"`
	PublicSupabaseURL       string `env:"PUBLIC_SUPABASE_URL"`
	MinWords                string `env:"MIN_WORDS"`
	SilenceThresholdSeconds string `env:"SILENCE_THRESHOLD_SECONDS"`
	BuddyModeEnabled        bool   `env:"BUDDY_MODE_ENABLED" envDefault:"false"`
	BuddyModePort           int    `env:"BUDDY_MODE_PORT"    envDefault:"8765"`
}

func LoadConfig() (*AppConfig, error) {
	cfg := &AppConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment: %w", err)
	}

	if cfg.LLMProvider == "groq" && cfg.GroqAPIKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY environment variable is not set")
	}

	if cfg.LLMProvider == "openai" && cfg.OpenAIAPIKey == "" {
		// Only fail if we are strictly using openai provider without API key
		// Some users might use demo proxy, so validation will be more context-specific.
		// return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	return cfg, nil
}
