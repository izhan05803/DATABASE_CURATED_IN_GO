package engine

import (
	"testing"
)

func TestEngine_SetAndGet(t *testing.T) {
	e := New()

	err := e.Set("foo", []byte("bar"))
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, err := e.Get("foo")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if string(got) != "bar" {
		t.Errorf("Get() = %s, want bar", got)
	}
}

func TestEngine_GetNotFound(t *testing.T) {
	e := New()

	_, err := e.Get("nonexistent")
	if err == nil {
		t.Error("Get() expected error for nonexistent key")
	}
}

func TestEngine_Delete(t *testing.T) {
	e := New()

	e.Set("foo", []byte("bar"))
	err := e.Delete("foo")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = e.Get("foo")
	if err == nil {
		t.Error("Get() expected error after delete")
	}
}

func TestEngine_Keys(t *testing.T) {
	e := New()

	e.Set("user:1", []byte("alice"))
	e.Set("user:2", []byte("bob"))
	e.Set("post:1", []byte("hello"))

	tests := []struct {
		pattern string
		want    int
	}{
		{"*", 3},
		{"user:*", 2},
		{"post:*", 1},
		{"none:*", 0},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			keys := e.Keys(tt.pattern)
			if len(keys) != tt.want {
				t.Errorf("Keys(%s) = %d keys, want %d", tt.pattern, len(keys), tt.want)
			}
		})
	}
}

func TestEngine_Info(t *testing.T) {
	e := New()

	e.Set("a", []byte("1"))
	e.Set("b", []byte("2"))

	info := e.Info()
	if info["records"] != 2 {
		t.Errorf("Info() records = %v, want 2", info["records"])
	}
}

func TestEngine_Concurrent(t *testing.T) {
	e := New()
	done := make(chan bool)

	// Writer
	go func() {
		for i := 0; i < 100; i++ {
			e.Set("key", []byte("value"))
		}
		done <- true
	}()

	// Reader
	go func() {
		for i := 0; i < 100; i++ {
			e.Get("key")
		}
		done <- true
	}()

	<-done
	<-done
}
