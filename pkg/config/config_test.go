package config

import (
	"strings"
	"testing"
	"time"
)

func TestPerfString(t *testing.T) {
	p := Perf{
		TotalConns:           100,
		MessageInterval:      5 * time.Second,
		WaitBeforeAuth:       1 * time.Second,
		WaitAfterAuth:        2 * time.Second,
		RampUpConnsPerSecond: 10,
		LogOutFile:           "/tmp/out.log",
		AuthMessage:          `{"type":"auth"}`,
		LoadMessage:          `{"type":"load"}`,
		ConfigPath:           "/tmp/config.yaml",
		SlowReadPercent:      5,
		SlowReadDuration:     100 * time.Millisecond,
	}

	s := p.String()
	if !strings.Contains(s, "Total Connections: 100") {
		t.Errorf("Perf.String() missing Total Connections: %q", s)
	}
	if !strings.Contains(s, "5s") {
		t.Errorf("Perf.String() missing MessageInterval: %q", s)
	}
	if !strings.Contains(s, "/tmp/out.log") {
		t.Errorf("Perf.String() missing LogOutFile: %q", s)
	}
	if !strings.Contains(s, "ConfigPath") {
		t.Errorf("Perf.String() missing ConfigPath: %q", s)
	}
}

func TestShouldProcessAsCmd(t *testing.T) {
	origFlags := Flags
	defer func() { Flags = origFlags }()

	// Case: Execute and Wait both set
	Flags = &Flag{
		Execute: []string{"cmd1"},
		Wait:    1 * time.Second,
	}
	if !Flags.ShouldProcessAsCmd() {
		t.Error("ShouldProcessAsCmd() = false, want true when Execute and Wait set")
	}

	// Case: IsSTDin
	Flags = &Flag{
		IsSTDin: true,
	}
	if !Flags.ShouldProcessAsCmd() {
		t.Error("ShouldProcessAsCmd() = false, want true when IsSTDin")
	}

	// Case: neither
	Flags = &Flag{}
	if Flags.ShouldProcessAsCmd() {
		t.Error("ShouldProcessAsCmd() = true, want false when no conditions met")
	}
}

func TestFlagString(t *testing.T) {
	f := &Flag{
		ConnectURL:  "ws://localhost:8080",
		BindAddress: "127.0.0.1",
		Auth:        "user:pass",
		IsPerf:      true,
		Perf: Perf{
			TotalConns: 50,
		},
	}

	s := f.String()
	if !strings.Contains(s, "ConnectURL: ws://localhost:8080") {
		t.Errorf("Flag.String() missing ConnectURL: %q", s)
	}
	if !strings.Contains(s, "BindAddress: 127.0.0.1") {
		t.Errorf("Flag.String() missing BindAddress: %q", s)
	}
	if !strings.Contains(s, "IsPerf: true") {
		t.Errorf("Flag.String() missing IsPerf: %q", s)
	}
	if !strings.Contains(s, "Perf Config:") {
		t.Errorf("Flag.String() missing Perf Config section: %q", s)
	}
}
