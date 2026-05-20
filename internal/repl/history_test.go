package repl

import (
	"testing"
)

func TestHistoryAdd(t *testing.T) {
	h := NewHistory(3)

	h.Add("SET key value")
	h.Add("GET key")
	h.Add("DELETE key")

	if h.Size() != 3 {
		t.Errorf("Size() = %d, want 3", h.Size())
	}
}

func TestHistoryAddEmpty(t *testing.T) {
	h := NewHistory(3)

	h.Add("")
	h.Add("SET key")
	h.Add("")

	if h.Size() != 1 {
		t.Errorf("Size() = %d, want 1 (empty commands skipped)", h.Size())
	}
}

func TestHistoryCapacity(t *testing.T) {
	h := NewHistory(2)

	h.Add("command 1")
	h.Add("command 2")
	h.Add("command 3") // Should evict "command 1"
	h.Add("command 4") // Should evict "command 2"

	if h.Size() != 2 {
		t.Errorf("Size() = %d, want 2", h.Size())
	}

	list := h.List()
	if list[0] != "command 3" || list[1] != "command 4" {
		t.Errorf("List() = %v, want [command 3 command 4]", list)
	}
}

func TestHistoryUp(t *testing.T) {
	h := NewHistory(5)
	h.Add("cmd1")
	h.Add("cmd2")
	h.Add("cmd3")

	// First Up should return cmd3 (newest)
	if cmd := h.Up(); cmd != "cmd3" {
		t.Errorf("Up() = %q, want %q", cmd, "cmd3")
	}

	// Second Up should return cmd2
	if cmd := h.Up(); cmd != "cmd2" {
		t.Errorf("Up() = %q, want %q", cmd, "cmd2")
	}

	// Third Up should return cmd1
	if cmd := h.Up(); cmd != "cmd1" {
		t.Errorf("Up() = %q, want %q", cmd, "cmd1")
	}

	// Further Up should stay at cmd1
	if cmd := h.Up(); cmd != "cmd1" {
		t.Errorf("Up() = %q, want %q", cmd, "cmd1")
	}
}

func TestHistoryDown(t *testing.T) {
	h := NewHistory(5)
	h.Add("cmd1")
	h.Add("cmd2")
	h.Add("cmd3")

	// Navigate up first
	h.Up()
	h.Up()
	h.Up() // At cmd1

	// Down should move forward
	if cmd := h.Down(); cmd != "cmd2" {
		t.Errorf("Down() = %q, want %q", cmd, "cmd2")
	}

	if cmd := h.Down(); cmd != "cmd3" {
		t.Errorf("Down() = %q, want %q", cmd, "cmd3")
	}

	// At end, Down returns empty
	if cmd := h.Down(); cmd != "" {
		t.Errorf("Down() = %q, want empty", cmd)
	}
}

func TestHistoryNavigationReset(t *testing.T) {
	h := NewHistory(5)
	h.Add("cmd1")
	h.Add("cmd2")

	// Navigate up
	h.Up()
	h.Up()

	// Add new command should reset navigation
	h.Add("cmd3")

	// Up should start from cmd3
	if cmd := h.Up(); cmd != "cmd3" {
		t.Errorf("Up() after Add = %q, want %q", cmd, "cmd3")
	}
}

func TestHistoryEmptyBuffer(t *testing.T) {
	h := NewHistory(5)

	if h.Up() != "" {
		t.Error("Up() on empty history should return empty string")
	}

	if h.Down() != "" {
		t.Error("Down() on empty history should return empty string")
	}

	if h.Size() != 0 {
		t.Error("Size() on empty history should be 0")
	}
}

func TestHistoryClear(t *testing.T) {
	h := NewHistory(5)
	h.Add("cmd1")
	h.Add("cmd2")

	h.Clear()

	if h.Size() != 0 {
		t.Errorf("Size() after Clear = %d, want 0", h.Size())
	}

	if h.Up() != "" {
		t.Error("Up() after Clear should return empty string")
	}
}

func TestHistoryList(t *testing.T) {
	h := NewHistory(5)
	h.Add("cmd1")
	h.Add("cmd2")
	h.Add("cmd3")

	list := h.List()
	if len(list) != 3 {
		t.Errorf("List() length = %d, want 3", len(list))
	}

	expected := []string{"cmd1", "cmd2", "cmd3"}
	for i, cmd := range list {
		if cmd != expected[i] {
			t.Errorf("List()[%d] = %q, want %q", i, cmd, expected[i])
		}
	}
}

func BenchmarkHistoryAdd(b *testing.B) {
	h := NewHistory(1000)
	for i := 0; i < b.N; i++ {
		h.Add("SET key value")
	}
}

func BenchmarkHistoryUp(b *testing.B) {
	h := NewHistory(100)
	for i := 0; i < 50; i++ {
		h.Add("command")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Up()
	}
}
