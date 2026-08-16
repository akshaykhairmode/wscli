package processer

import (
	"testing"

	"github.com/akshaykhairmode/wscli/pkg/config"
)

func TestTruncateString(t *testing.T) {
	cases := []struct {
		input string
		n     int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
		{"ab", 2, "ab"},
		{"abc", 2, "ab..."},
	}
	for _, c := range cases {
		got := truncateString(c.input, c.n)
		if got != c.want {
			t.Errorf("truncateString(%q, %d) = %q, want %q", c.input, c.n, got, c.want)
		}
	}
}

func TestShouldProcessCommand(t *testing.T) {
	origFlags := config.Flags
	defer func() { config.Flags = origFlags }()

	config.Flags = &config.Flag{IsSlash: true}
	if !shouldProcessCommand("/ping", "/ping") {
		t.Error("shouldProcessCommand() = false, want true when slash enabled and prefix matches")
	}

	if shouldProcessCommand("ping", "/ping") {
		t.Error("shouldProcessCommand() = true, want false when prefix doesn't match")
	}

	config.Flags = &config.Flag{IsSlash: false}
	if shouldProcessCommand("/ping", "/ping") {
		t.Error("shouldProcessCommand() = true, want false when slash disabled")
	}
}

func TestShouldProcessAsCmd(t *testing.T) {
	origFlags := config.Flags
	defer func() { config.Flags = origFlags }()

	config.Flags = &config.Flag{IsSTDin: true}
	if !config.Flags.ShouldProcessAsCmd() {
		t.Error("ShouldProcessAsCmd() = false, want true when IsSTDin")
	}

	config.Flags = &config.Flag{Execute: []string{"cmd"}, Wait: 1}
	if !config.Flags.ShouldProcessAsCmd() {
		t.Error("ShouldProcessAsCmd() = false, want true when Execute+Wait set")
	}

	config.Flags = &config.Flag{}
	if config.Flags.ShouldProcessAsCmd() {
		t.Error("ShouldProcessAsCmd() = true, want false when no conditions")
	}
}
