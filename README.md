# tweet-audit

`tweet-audit` analyzes an exported X archive, asks Gemini AI which tweets should be reviewed for deletion, and writes a CSV of flagged tweet URLs for manual cleanup.

## What it does

1. Loads tweets from an X archive export.
2. Sends tweets to Gemini in small batches with your custom alignment criteria.
3. Writes flagged tweet URLs to a CSV file shaped like:

```csv
tweet_url,deleted
https://x.com/username/status/1234567890,false
```

## Why Go

Go is a good fit here because the workload is mostly I/O bound: archive parsing, batched HTTP calls, and CSV generation. It gives us straightforward concurrency, a fast startup path, and easy distribution as a single binary.

## Project layout

```text
tweet-audit/
├── README.md
├── TRADEOFFS.md
├── .gitignore
├── config.example.json
├── src/
│   ├── app/
│   ├── archive/
│   ├── config/
│   ├── evaluator/
│   ├── model/
│   └── output/
└── tests/
```

## Configuration

Copy `config.example.json` to `config.json` and fill in your values.

Environment variable override:

- `GEMINI_API_KEY`: overrides `gemini_api_key` in config

## Run

```bash
go run ./src/cmd/tweet-audit -config ./config.json
```

## Expected X archive file

This scaffold currently targets the common archive file:

- `data/tweets.js`

That file is expected to contain the `window.YTD.tweets.part0 = [...]` format used by X exports.

## Testing

```bash
go test ./...
```
