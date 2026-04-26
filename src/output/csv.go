package output

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ferdinand/tweet-audit/src/model"
)

func WriteFlaggedTweets(path string, tweets []model.FlaggedTweet) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create output CSV: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"tweet_url", "deleted"}); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}

	for _, tweet := range tweets {
		if err := writer.Write([]string{tweet.TweetURL, fmt.Sprintf("%t", tweet.Deleted)}); err != nil {
			return fmt.Errorf("write CSV row: %w", err)
		}
	}

	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush CSV writer: %w", err)
	}

	return nil
}
