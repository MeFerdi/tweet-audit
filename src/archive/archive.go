package archive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ferdinand/tweet-audit/src/model"
)

const twitterTimeLayout = "Mon Jan 02 15:04:05 -0700 2006"

type archiveTweet struct {
	Tweet archiveTweetPayload `json:"tweet"`
}

type archiveTweetPayload struct {
	IDStr     string `json:"id_str"`
	FullText  string `json:"full_text"`
	CreatedAt string `json:"created_at"`
}

func LoadTweets(archivePath, username string) ([]model.Tweet, error) {
	path := filepath.Join(archivePath, "data", "tweets.js")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read archive tweets file: %w", err)
	}

	rawJSON, err := stripArchivePrefix(string(data))
	if err != nil {
		return nil, err
	}

	var payload []archiveTweet
	if err := json.Unmarshal([]byte(rawJSON), &payload); err != nil {
		return nil, fmt.Errorf("parse archive tweets JSON: %w", err)
	}

	tweets := make([]model.Tweet, 0, len(payload))
	for _, item := range payload {
		createdAt, err := time.Parse(twitterTimeLayout, item.Tweet.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse tweet created_at for %s: %w", item.Tweet.IDStr, err)
		}

		tweets = append(tweets, model.Tweet{
			ID:        item.Tweet.IDStr,
			FullText:  item.Tweet.FullText,
			CreatedAt: createdAt,
			URL:       fmt.Sprintf("https://x.com/%s/status/%s", username, item.Tweet.IDStr),
		})
	}

	return tweets, nil
}

func stripArchivePrefix(content string) (string, error) {
	idx := strings.Index(content, "=")
	if idx == -1 {
		return "", fmt.Errorf("archive tweets file did not contain '=' separator")
	}

	jsonPart := strings.TrimSpace(content[idx+1:])
	jsonPart = strings.TrimSuffix(jsonPart, ";")
	jsonPart = strings.TrimSpace(jsonPart)

	if !strings.HasPrefix(jsonPart, "[") {
		return "", fmt.Errorf("archive tweets payload is not a JSON array")
	}

	return jsonPart, nil
}
