package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)
// this struct holds the evaluation criteria such as forbidden words, professional check, tone, etc. It is part of the main Config struct that is loaded from the JSON file and used throughout the application to determine which tweets should be flagged based on the defined criteria.
type Criteria struct {
	ForbiddenWords    []string `json:"forbidden_words"`
	ProfessionalCheck bool     `json:"professional_check"`
	Tone              string   `json:"tone"`
	ExcludePolitics   bool     `json:"exclude_politics"`
	Notes             []string `json:"notes"`
}

// this struct holds the app settings: archive path, output path, Gemini API key and model, batch size, max workers, and the evaluation criteria.
// It is loaded from a JSON file using the Load function, which also validates the required fields and allows overriding the Gemini API key with an environment variable.
// The Config struct is used throughout the application to access these settings when loading tweets, evaluating them, and writing the output.
type Config struct {
	ArchivePath   string   `json:"archive_path"`
	OutputCSVPath string   `json:"output_csv_path"`
	Username      string   `json:"username"`
	GeminiAPIKey  string   `json:"gemini_api_key"`
	GeminiModel   string   `json:"gemini_model"`
	BatchSize     int      `json:"batch_size"`
	MaxWorkers    int      `json:"max_workers"`
	Criteria      Criteria `json:"criteria"`
}

func Load(path string) (Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config file: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config JSON: %w", err)
	}

	if envKey := os.Getenv("GEMINI_API_KEY"); envKey != "" {
		cfg.GeminiAPIKey = envKey
	}

	if err := cfg.Validate(); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.ArchivePath == "" {
		return errors.New("archive_path is required")
	}
	if c.OutputCSVPath == "" {
		return errors.New("output_csv_path is required")
	}
	if c.Username == "" {
		return errors.New("username is required")
	}
	if c.GeminiAPIKey == "" {
		return errors.New("gemini_api_key is required or set GEMINI_API_KEY")
	}
	if c.GeminiModel == "" {
		return errors.New("gemini_model is required")
	}
	if c.BatchSize <= 0 {
		return errors.New("batch_size must be greater than zero")
	}
	if c.MaxWorkers <= 0 {
		return errors.New("max_workers must be greater than zero")
	}
	return nil
}
