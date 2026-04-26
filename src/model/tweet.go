package model

import "time"

type Tweet struct {
	ID        string
	FullText  string
	CreatedAt time.Time
	URL       string
}

type FlaggedTweet struct {
	TweetURL string
	Deleted  bool
	Reason   string
}
