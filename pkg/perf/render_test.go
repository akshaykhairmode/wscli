package perf

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// fakeScreen is a minimal tcell.Screen for off-screen rendering.
type fakeScreen struct {
	tcell.Screen
	w, h  int
	cells map[[2]int]rune
}

func newFakeScreen(w, h int) *fakeScreen {
	return &fakeScreen{w: w, h: h, cells: make(map[[2]int]rune)}
}

func (f *fakeScreen) Init() error                { return nil }
func (f *fakeScreen) Fini()                      {}
func (f *fakeScreen) Clear()                     {}
func (f *fakeScreen) Fill(r rune, s tcell.Style) {}
func (f *fakeScreen) SetContent(x, y int, m rune, c []rune, s tcell.Style) {
	if x < 0 || y < 0 || x >= f.w || y >= f.h {
		return
	}
	f.cells[[2]int{x, y}] = m
}
func (f *fakeScreen) GetContent(x, y int) (rune, []rune, tcell.Style, int) {
	if x < 0 || y < 0 || x >= f.w || y >= f.h {
		return 0, nil, tcell.StyleDefault, 1
	}
	return f.cells[[2]int{x, y}], nil, tcell.StyleDefault, 1
}
func (f *fakeScreen) SetStyle(s tcell.Style)                {}
func (f *fakeScreen) Size() (int, int)                      { return f.w, f.h }
func (f *fakeScreen) Show()                                 {}
func (f *fakeScreen) Sync()                                 {}
func (f *fakeScreen) Colors() int                           { return 256 }
func (f *fakeScreen) CharacterSet() string                  { return "UTF-8" }
func (f *fakeScreen) CanDisplay(r rune, b bool) bool        { return true }
func (f *fakeScreen) HasKey(k tcell.Key) bool               { return true }
func (f *fakeScreen) HasMouse() bool                        { return false }
func (f *fakeScreen) EnableMouse(...tcell.MouseFlags)       {}
func (f *fakeScreen) DisableMouse()                         {}
func (f *fakeScreen) EnablePaste()                          {}
func (f *fakeScreen) DisablePaste()                         {}
func (f *fakeScreen) RegisterRuneFallback(r rune, s string) {}
func (f *fakeScreen) UnregisterRuneFallback(r rune)         {}
func (f *fakeScreen) Resize(a, b, c, d int)                 {}
func (f *fakeScreen) Beep() error                           { return nil }
func (f *fakeScreen) Suspend() error                        { return nil }
func (f *fakeScreen) Resume() error                         { return nil }
func (f *fakeScreen) SetSize(a, b int)                      {}
func (f *fakeScreen) LockRegion(a, b, c, d int, l bool)     {}
func (f *fakeScreen) Tty() (tcell.Tty, bool)                { return nil, false }
func (f *fakeScreen) PollEvent() tcell.Event                { return nil }
func (f *fakeScreen) HasPendingEvent() bool                 { return false }
func (f *fakeScreen) PostEvent(ev tcell.Event) error        { return nil }
func (f *fakeScreen) SetCell(x, y int, s tcell.Style, ch ...rune) {
	if len(ch) > 0 {
		f.SetContent(x, y, ch[0], nil, s)
	}
}

func renderLine(t *testing.T, s string, width int) string {
	t.Helper()
	screen := newFakeScreen(120, 30)
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetWrap(false)
	tv.SetText(s)
	tv.SetRect(0, 0, width, 5)
	tv.Draw(screen)

	var line []rune
	for x := 0; x < width; x++ {
		c, _, _, _ := screen.GetContent(x, 0)
		if c == 0 {
			break
		}
		line = append(line, c)
	}
	return string(line)
}

func TestProgressBarRenderClean(t *testing.T) {
	got := renderLine(t, progressBar(200, 500, 20), 60)
	if strings.ContainsAny(got, "#[]") {
		t.Errorf("progress bar leaks style-tag text: %q", got)
	}
	if !strings.Contains(got, "40.0%") {
		t.Errorf("progress bar missing percentage: %q", got)
	}
}
