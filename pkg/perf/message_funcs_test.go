package perf

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"text/template"
)

func TestIsFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("hello"), 0644)

	isFile, size := isFilePath(tmpFile)
	if !isFile {
		t.Error("isFilePath() = false for existing file, want true")
	}
	if size != 5 {
		t.Errorf("isFilePath() size = %d, want 5", size)
	}

	isFile, _ = isFilePath("/nonexistent")
	if isFile {
		t.Error("isFilePath() = true for nonexistent path, want false")
	}

	isFile, _ = isFilePath(tmpDir)
	if isFile {
		t.Error("isFilePath() = true for directory, want false")
	}
}

func TestRandomInt(t *testing.T) {
	for i := 0; i < 100; i++ {
		val := randomInt(10)
		if val < 0 || val >= 10 {
			t.Errorf("randomInt(10) = %d, want [0, 10)", val)
		}
	}

	val := randomInt()
	if val < 0 || val >= 10000 {
		t.Errorf("randomInt() = %d, want [0, 10000)", val)
	}
}

func TestRandomAlphaNumeric(t *testing.T) {
	got := randomAlphaNumeric(20)
	if len(got) != 20 {
		t.Errorf("randomAlphaNumeric(20) len = %d, want 20", len(got))
	}

	got = randomAlphaNumeric()
	if len(got) != 10 {
		t.Errorf("randomAlphaNumeric() len = %d, want 10", len(got))
	}
}

func TestRandomUUID(t *testing.T) {
	uuid := randomUUID()
	if len(uuid) != 36 {
		t.Errorf("randomUUID() len = %d, want 36", len(uuid))
	}
}

func TestArray(t *testing.T) {
	elems := []string{"a", "b", "c"}
	seen := map[string]bool{}
	for i := 0; i < 3; i++ {
		val := array(elems...)
		seen[val] = true
	}
	for _, e := range elems {
		if !seen[e] {
			t.Errorf("array() never returned %q", e)
		}
	}
}

func TestRandomArray(t *testing.T) {
	elems := []string{"a", "b", "c"}
	for i := 0; i < 100; i++ {
		val := randomArray(elems...)
		found := false
		for _, e := range elems {
			if val == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("randomArray() returned unexpected value %q", val)
		}
	}
}

func TestGetUint64Counter(t *testing.T) {
	c := getUint64Counter()
	if c.Load() != 0 {
		t.Errorf("getUint64Counter() = %d, want 0", c.Load())
	}

	c = getUint64Counter(42)
	if c.Load() != 42 {
		t.Errorf("getUint64Counter(42) = %d, want 42", c.Load())
	}

	c = getUint64Counter(0)
	if c.Load() != 0 {
		t.Errorf("getUint64Counter(0) = %d, want 0", c.Load())
	}
}

func TestGetUniqueSequence(t *testing.T) {
	uniqueSequenceMap = &sync.Map{}

	val1 := getUniqueSequence("test")
	val2 := getUniqueSequence("test")
	if val1 != 0 || val2 != 1 {
		t.Errorf("getUniqueSequence() returned %d, %d, want 0, 1", val1, val2)
	}
}

func TestParseTemplate(t *testing.T) {
	tmpl := newTemplate()
	err := parseTemplate(tmpl, "")
	if err != nil {
		t.Errorf("parseTemplate() with empty string error: %v", err)
	}

	err = parseTemplate(tmpl, "{{.Invalid")
	if err == nil {
		t.Error("parseTemplate() with invalid template should return error")
	}
}

func TestNewDefaultMessageGetter(t *testing.T) {
	mg, err := NewDefaultMessageGetter("hello world")
	if err != nil {
		t.Fatalf("NewDefaultMessageGetter() error: %v", err)
	}

	if mg.GetTemplateString() != "hello world" {
		t.Errorf("GetTemplateString() = %q, want 'hello world'", mg.GetTemplateString())
	}

	data, release := mg.Get(nil)
	defer release()
	if string(data) != "hello world" {
		t.Errorf("Get() = %q, want 'hello world'", string(data))
	}
}

func newTemplate() *template.Template {
	return template.New("test")
}
