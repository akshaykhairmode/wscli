package terminal

import (
	"io"
	"testing"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/logger"
)

func TestGetPrompt(t *testing.T) {
	origFlags := config.Flags
	defer func() { config.Flags = origFlags }()

	config.Flags = &config.Flag{NoColor: true}
	got := getPrompt("test")
	if got != "test" {
		t.Errorf("getPrompt() with NoColor = %q, want 'test'", got)
	}

	config.Flags = &config.Flag{NoColor: false}
	got = getPrompt("» ")
	if got == "» " {
		t.Errorf("getPrompt() without NoColor should have ANSI codes, got %q", got)
	}
}

func TestGetHistoryFilePath(t *testing.T) {
	path := getHistoryFilePath("testapp")
	if path == "" {
		t.Error("getHistoryFilePath() returned empty string")
	}
}

// Ensure logger is initialized for tests that use getHistoryFilePath
func init() {
	config.Flags = &config.Flag{}
	logger.Init(io.Discard, nil)
}
