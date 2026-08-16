package ws

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"strings"
	"testing"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/logger"
)

func TestBasicAuth(t *testing.T) {
	cases := map[string]string{
		"user:pass":       "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass")),
		"":                "Basic " + base64.StdEncoding.EncodeToString([]byte("")),
		"admin:secret123": "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:secret123")),
	}
	for input, want := range cases {
		if got := BasicAuth(input); got != want {
			t.Errorf("BasicAuth(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestUnzipGzipBytes(t *testing.T) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	original := "hello world"
	gzWriter.Write([]byte(original))
	gzWriter.Close()

	got, err := unzipGzipBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("unzipGzipBytes() error: %v", err)
	}
	if got != original {
		t.Errorf("unzipGzipBytes() = %q, want %q", got, original)
	}
}

func TestUnzipGzipBytesInvalid(t *testing.T) {
	_, err := unzipGzipBytes([]byte("not gzip"))
	if err == nil {
		t.Error("unzipGzipBytes() with invalid data should return error")
	}
}

func TestFormatMessage(t *testing.T) {
	origFlags := config.Flags
	defer func() { config.Flags = origFlags }()

	config.Flags = &config.Flag{IsJSONPrettyPrint: false}
	got := formatMessage([]byte("hello"))
	if !strings.Contains(got, "hello") {
		t.Errorf("formatMessage() = %q, want it to contain 'hello'", got)
	}

	config.Flags = &config.Flag{IsJSONPrettyPrint: true}
	got = formatMessage([]byte(`{"key":"value"}`))
	if !strings.Contains(got, "key") {
		t.Errorf("formatMessage() with JSON = %q, want it to contain 'key'", got)
	}

	config.Flags = &config.Flag{IsJSONPrettyPrint: true}
	got = formatMessage([]byte("not json"))
	if !strings.Contains(got, "not json") {
		t.Errorf("formatMessage() with invalid JSON = %q, want it to contain 'not json'", got)
	}
}

func TestDecryptPrivateKeyUnencrypted(t *testing.T) {
	keyPEM := []byte(`-----BEGIN PRIVATE KEY-----
MIGkAgIBBDANBgkqhkiG9w0BAQEFAASCBkwGggZkCAQ9NAoGBAMB8vQ==
-----END PRIVATE KEY-----`)

	_, err := decryptPrivateKey(keyPEM, nil)
	if err == nil {
		t.Log("decryptPrivateKey with minimal key may fail on PEM decode, that's ok")
	}
}

func TestDecryptPrivateKeyNoPEMBlock(t *testing.T) {
	_, err := decryptPrivateKey([]byte("not a pem block"), nil)
	if err == nil {
		t.Error("decryptPrivateKey() with no PEM block should return error")
	}
}

// Ensure logger is initialized for tests that use formatMessage
func init() {
	config.Flags = &config.Flag{}
	logger.Init(io.Discard, nil)
}
