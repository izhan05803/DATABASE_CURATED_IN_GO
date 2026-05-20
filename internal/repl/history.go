package repl

// History manages command history using a ring buffer
// Supports fixed-size circular storage with navigation
type History struct {
	commands []string // Ring buffer
	size     int      // Capacity of buffer
	count    int      // Number of commands stored
	current  int      // Current position (for navigation)
	dirty    bool     // Whether current differs from buffer
}

// NewHistory creates a new command history with given capacity
func NewHistory(capacity int) *History {
	if capacity <= 0 {
		capacity = 100 // Default size
	}
	return &History{
		commands: make([]string, 0, capacity),
		size:     capacity,
		count:    0,
		current:  -1,
		dirty:    false,
	}
}

// Add adds a command to the history
func (h *History) Add(cmd string) {
	if cmd == "" {
		return // Skip empty commands
	}

	// If at capacity, remove oldest (first element)
	if len(h.commands) >= h.size {
		h.commands = h.commands[1:]
	}

	h.commands = append(h.commands, cmd)
	h.count++
	h.reset() // Reset navigation
}

// Up navigates backward in history (older commands)
// Returns command or empty string if at beginning
func (h *History) Up() string {
	if len(h.commands) == 0 {
		return ""
	}

	if h.current < 0 {
		// First time: start from end
		h.current = len(h.commands) - 1
	} else if h.current > 0 {
		// Move backward
		h.current--
	}
	// At position 0, stay there

	return h.commands[h.current]
}

// Down navigates forward in history (newer commands)
// Returns command or empty string if at end
func (h *History) Down() string {
	if len(h.commands) == 0 {
		return ""
	}

	if h.current < 0 {
		return ""
	}

	if h.current < len(h.commands)-1 {
		// Move forward
		h.current++
		return h.commands[h.current]
	}

	// At end, return empty and reset
	h.reset()
	return ""
}

// reset resets navigation to end of history
func (h *History) reset() {
	h.current = -1
	h.dirty = false
}

// List returns all commands in order (newest last)
func (h *History) List() []string {
	// Return a copy
	result := make([]string, len(h.commands))
	copy(result, h.commands)
	return result
}

// Size returns number of commands in history
func (h *History) Size() int {
	return len(h.commands)
}

// Clear empties the history
func (h *History) Clear() {
	h.commands = h.commands[:0]
	h.count = 0
	h.reset()
}
