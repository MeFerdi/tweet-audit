package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ferdinand/tweet-audit/src/app"
	"github.com/ferdinand/tweet-audit/src/config"
	"github.com/ferdinand/tweet-audit/src/evaluator"
)

func main() {
	configPath := flag.String("config", "./config.json", "Path to config JSON file")
	timeout := flag.Duration("timeout", 5*time.Minute, "Overall execution timeout")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	client := evaluator.NewGeminiClient(cfg.GeminiAPIKey, cfg.GeminiModel, nil)
	service := app.New(client)

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	result, err := service.Run(ctx, cfg)
	if err != nil {
		log.Fatalf("run audit: %v", err)
	}

	fmt.Fprintf(os.Stdout, "processed=%d flagged=%d output=%s\n", result.Processed, result.Flagged, result.OutputPath)
}
