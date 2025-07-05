package config

import (
	"fmt"
	"os"
)

type Config struct {
	OpenAIAPIKey string
	OpenAIBaseURL string
}

func Load() (*Config, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	// 如果没有设置，使用默认的 OpenAI URL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &Config{
		OpenAIAPIKey: apiKey,
		OpenAIBaseURL: baseURL,
	}, nil
}