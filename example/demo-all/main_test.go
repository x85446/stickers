package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func key(k tea.KeyType) tea.KeyMsg { return tea.KeyMsg{Type: k} }

func press(t *testing.T, m *model, msg tea.Msg) *model {
	t.Helper()
	next, _ := m.Update(msg)
	m, ok := next.(*model)
	if !ok {
		t.Fatalf("Update returned %T, want *model", next)
	}
	return m
}

// TestSplashThenCycle starts on the splash screen, presses a key, then
// cycles every demo in both directions. Each demo must size and render
// without panicking.
func TestSplashThenCycle(t *testing.T) {
	m := newModel()
	m = press(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	if v := m.View(); !strings.Contains(v, "press any key to start") {
		t.Fatalf("splash missing start hint:\n%s", v)
	}

	m = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !m.started {
		t.Fatal("space did not leave the splash screen")
	}
	if v := m.View(); !strings.Contains(v, demoInfos[0].name) {
		t.Fatalf("first demo header missing:\n%s", v)
	}

	for i := 0; i < len(m.demos); i++ {
		m = press(t, m, key(tea.KeyShiftTab))
		m.View()
	}
	if m.current != 0 {
		t.Fatalf("shift+tab x%d ended on %d, want 0", len(m.demos), m.current)
	}
	for i := 0; i < len(m.demos); i++ {
		m = press(t, m, key(tea.KeyTab))
		m.View()
	}
	if m.current != 0 {
		t.Fatalf("tab x%d ended on %d, want 0", len(m.demos), m.current)
	}
}

// TestSplashQuit checks q and ctrl+c quit from the splash screen.
func TestSplashQuit(t *testing.T) {
	for _, k := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'q'}},
		{Type: tea.KeyCtrlC},
	} {
		m := newModel()
		_, cmd := m.Update(k)
		if cmd == nil {
			t.Fatalf("%s on splash returned no command, want tea.Quit", k.String())
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatalf("%s on splash did not quit", k.String())
		}
	}
}
