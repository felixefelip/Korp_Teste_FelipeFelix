package ai

import (
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	draftModel = "claude-sonnet-5"
	maxTokens  = 4096
)

func newClient() (anthropic.Client, bool) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return anthropic.Client{}, false
	}

	return anthropic.NewClient(option.WithAPIKey(key)), true
}
