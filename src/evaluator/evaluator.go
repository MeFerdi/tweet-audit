package evaluator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ferdinand/tweet-audit/src/config"
	"github.com/ferdinand/tweet-audit/src/model"
)

type Analyzer interface {
	EvaluateTweets(ctx context.Context, criteria config.Criteria, tweets []model.Tweet) ([]model.FlaggedTweet, error)
}

type GeminiClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewGeminiClient(apiKey, modelName string, httpClient *http.Client) *GeminiClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 45 * time.Second}
	}

	return &GeminiClient{
		apiKey:     apiKey,
		model:      modelName,
		httpClient: httpClient,
	}
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

type flaggedOutput struct {
	Flagged []struct {
		TweetURL string `json:"tweet_url"`
		Reason   string `json:"reason"`
	} `json:"flagged"`
}

func (g *GeminiClient) EvaluateTweets(ctx context.Context, criteria config.Criteria, tweets []model.Tweet) ([]model.FlaggedTweet, error) {
	promptPayload, err := buildPrompt(criteria, tweets)
	if err != nil {
		return nil, err
	}

	requestBody, err := json.Marshal(geminiRequest{
		Contents: []geminiContent{{
			Parts: []geminiPart{{Text: promptPayload}},
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal Gemini request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", g.model, g.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("build Gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send Gemini request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read Gemini response: %w", err)
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Gemini returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return parseGeminiResponse(body)
}

func buildPrompt(criteria config.Criteria, tweets []model.Tweet) (string, error) {
	criteriaJSON, err := json.Marshal(criteria)
	if err != nil {
		return "", fmt.Errorf("marshal criteria: %w", err)
	}

	type promptTweet struct {
		ID        string `json:"id"`
		URL       string `json:"url"`
		Text      string `json:"text"`
		CreatedAt string `json:"created_at"`
	}

	payload := make([]promptTweet, 0, len(tweets))
	for _, tweet := range tweets {
		payload = append(payload, promptTweet{
			ID:        tweet.ID,
			URL:       tweet.URL,
			Text:      tweet.FullText,
			CreatedAt: tweet.CreatedAt.Format(time.RFC3339),
		})
	}

	tweetsJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal tweets for prompt: %w", err)
	}

	return fmt.Sprintf(`Review these tweets against the alignment criteria.
Return only JSON in this shape:
{"flagged":[{"tweet_url":"https://x.com/...","reason":"short explanation"}]}

Criteria:
%s

Tweets:
%s`, string(criteriaJSON), string(tweetsJSON)), nil
}

func parseGeminiResponse(body []byte) ([]model.FlaggedTweet, error) {
	var response geminiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode Gemini response envelope: %w", err)
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("Gemini response did not contain content")
	}

	raw := strings.TrimSpace(response.Candidates[0].Content.Parts[0].Text)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var parsed flaggedOutput
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("decode Gemini flagged tweet JSON: %w", err)
	}

	flagged := make([]model.FlaggedTweet, 0, len(parsed.Flagged))
	for _, item := range parsed.Flagged {
		flagged = append(flagged, model.FlaggedTweet{
			TweetURL: item.TweetURL,
			Deleted:  false,
			Reason:   item.Reason,
		})
	}

	return flagged, nil
}
