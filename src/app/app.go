package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/ferdinand/tweet-audit/src/archive"
	"github.com/ferdinand/tweet-audit/src/config"
	"github.com/ferdinand/tweet-audit/src/evaluator"
	"github.com/ferdinand/tweet-audit/src/model"
	"github.com/ferdinand/tweet-audit/src/output"
)

type App struct {
	analyzer evaluator.Analyzer
}

type Result struct {
	Processed  int
	Flagged    int
	OutputPath string
}

func New(analyzer evaluator.Analyzer) *App {
	return &App{analyzer: analyzer}
}

func (a *App) Run(ctx context.Context, cfg config.Config) (Result, error) {
	var result Result

	tweets, err := archive.LoadTweets(cfg.ArchivePath, cfg.Username)
	if err != nil {
		return result, fmt.Errorf("load tweets: %w", err)
	}

	flagged, err := a.evaluateBatches(ctx, cfg, tweets)
	if err != nil {
		return result, err
	}

	if err := output.WriteFlaggedTweets(cfg.OutputCSVPath, flagged); err != nil {
		return result, fmt.Errorf("write output CSV: %w", err)
	}

	result = Result{
		Processed:  len(tweets),
		Flagged:    len(flagged),
		OutputPath: cfg.OutputCSVPath,
	}

	return result, nil
}

func (a *App) evaluateBatches(ctx context.Context, cfg config.Config, tweets []model.Tweet) ([]model.FlaggedTweet, error) {
	batches := splitTweets(tweets, cfg.BatchSize)

	type job struct {
		index int
		items []model.Tweet
	}

	type batchResult struct {
		index   int
		flagged []model.FlaggedTweet
		err     error
	}

	jobs := make(chan job)
	results := make(chan batchResult, len(batches))

	var wg sync.WaitGroup
	for i := 0; i < cfg.MaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range jobs {
				flagged, err := a.analyzer.EvaluateTweets(ctx, cfg.Criteria, batch.items)
				results <- batchResult{
					index:   batch.index,
					flagged: flagged,
					err:     err,
				}
			}
		}()
	}

	go func() {
		for i, batch := range batches {
			select {
			case <-ctx.Done():
				break
			case jobs <- job{index: i, items: batch}:
			}
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	ordered := make([][]model.FlaggedTweet, len(batches))
	for result := range results {
		if result.err != nil {
			return nil, fmt.Errorf("evaluate batch %d: %w", result.index, result.err)
		}
		ordered[result.index] = result.flagged
	}

	combined := make([]model.FlaggedTweet, 0)
	for _, batch := range ordered {
		combined = append(combined, batch...)
	}

	return combined, nil
}

func splitTweets(tweets []model.Tweet, batchSize int) [][]model.Tweet {
	if len(tweets) == 0 {
		return nil
	}

	batches := make([][]model.Tweet, 0, (len(tweets)+batchSize-1)/batchSize)
	for start := 0; start < len(tweets); start += batchSize {
		end := start + batchSize
		if end > len(tweets) {
			end = len(tweets)
		}
		batches = append(batches, tweets[start:end])
	}

	return batches
}
