package perf

import (
	"sync"
	"testing"
	"time"
)

func TestCalculatePercentage(t *testing.T) {
	cases := []struct {
		value int64
		total int64
		want  string
	}{
		{50, 100, "50 (50.00%)"},
		{0, 100, "0 (0.00%)"},
		{100, 100, "100 (100.00%)"},
		{0, 0, "0.00%"},
		{1, 3, "1 (33.33%)"},
	}
	for _, c := range cases {
		got := calculatePercentage(c.value, c.total)
		if got != c.want {
			t.Errorf("calculatePercentage(%d, %d) = %q, want %q", c.value, c.total, got, c.want)
		}
	}
}

func TestIntToString(t *testing.T) {
	cases := map[int64]string{
		0:   "0",
		42:  "42",
		-1:  "-1",
		100: "100",
	}
	for input, want := range cases {
		if got := intToString(input); got != want {
			t.Errorf("intToString(%d) = %q, want %q", input, got, want)
		}
	}
}

func TestDurToString(t *testing.T) {
	cases := []struct {
		input float64
		want  string
	}{
		{0, "0s"},
		{float64(time.Second), "1s"},
		{float64(500 * time.Millisecond), "500ms"},
		{float64(time.Millisecond), "1ms"},
	}
	for _, c := range cases {
		got := durToString(c.input)
		if got != c.want {
			t.Errorf("durToString(%v) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestStripTimeFromLog(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"12:34:56.789 some log", "some log"},
		{"no time prefix", "no time prefix"},
		{"01:02:03.456 test", "test"},
	}
	for _, c := range cases {
		got := stripTimeFromLog(c.input)
		if got != c.want {
			t.Errorf("stripTimeFromLog(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestErrMsgAdd(t *testing.T) {
	em := &errMsg{
		data:  make(map[string]int),
		order: []string{},
		mux:   &sync.RWMutex{},
	}

	em.Add("error1")
	em.Add("error1")
	em.Add("error2")

	em.ForEach(func(data map[string]int, order []string) {
		if data["error1"] != 2 {
			t.Errorf("error1 count = %d, want 2", data["error1"])
		}
		if data["error2"] != 1 {
			t.Errorf("error2 count = %d, want 1", data["error2"])
		}
		if len(order) != 2 {
			t.Errorf("order len = %d, want 2", len(order))
		}
		if order[0] != "error1" || order[1] != "error2" {
			t.Errorf("order = %v, want [error1 error2]", order)
		}
	})
}
