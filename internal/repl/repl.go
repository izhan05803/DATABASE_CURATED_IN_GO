package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/izhan05803/gofromscratchdb/internal/engine"
	"github.com/izhan05803/gofromscratchdb/internal/parser"
	"github.com/izhan05803/gofromscratchdb/pkg/types"
)

const prompt = "godb> "
const defaultDBPath = "data/database.godb"
const historyCapacity = 100

// Start starts the REPL loop
func Start(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	db, err := engine.NewPersistent(defaultDBPath)
	if err != nil {
		return fmt.Errorf("start persistent engine: %w", err)
	}

	history := NewHistory(historyCapacity)

	for {
		fmt.Fprint(out, prompt)
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		p := parser.NewParser(line)
		cmd, err := p.Parse()
		if err != nil {
			fmt.Fprintf(out, "❌ %s\n", formatErrorMessage("parsing error", err.Error()))
			continue
		}

		if err := parser.ValidateCommand(cmd); err != nil {
			fmt.Fprintf(out, "❌ %s\n", formatValidationError(cmd.Type, err.Error()))
			continue
		}

		// Add to history only on successful parse
		history.Add(line)

		result := execute(db, cmd)
		printResult(out, result)

		if cmd.Type == "EXIT" {
			break
		}
	}

	return db.Close()
}

// execute runs a command against the engine
func execute(db *engine.Engine, cmd *types.Command) *types.Result {
	switch cmd.Type {
	case "GET":
		value, err := db.Get(cmd.Args[0])
		if err != nil {
			return &types.Result{Success: false, Message: "❌ GET key not found"}
		}
		return &types.Result{Success: true, Message: fmt.Sprintf("%q", string(value))}

	case "SET":
		err := db.Set(cmd.Args[0], []byte(cmd.Args[1]))
		if err != nil {
			return &types.Result{Success: false, Message: fmt.Sprintf("❌ SET failed: %v", err)}
		}
		return &types.Result{Success: true, Message: "✅ OK"}

	case "DELETE":
		err := db.Delete(cmd.Args[0])
		if err != nil {
			return &types.Result{Success: false, Message: fmt.Sprintf("❌ DELETE failed: %v", err)}
		}
		return &types.Result{Success: true, Message: "✅ OK"}

	case "KEYS":
		keys := db.Keys(cmd.Args[0])
		if len(keys) == 0 {
			return &types.Result{Success: true, Message: "(no keys matched)"}
		}
		// Format as pretty table with indices and alignment
		var sb strings.Builder
		maxLen := 0
		for _, k := range keys {
			if len(k) > maxLen {
				maxLen = len(k)
			}
		}
		for i, k := range keys {
			sb.WriteString(fmt.Sprintf(" %d) %-*s\n", i+1, maxLen, k))
		}
		return &types.Result{
			Success: true,
			Message: fmt.Sprintf("(%d keys matched)\n%s", len(keys), strings.TrimSuffix(sb.String(), "\n")),
		}

	case "INFO":
		info := db.Info()
		message := formatInfoOutput(info)
		return &types.Result{
			Success: true,
			Message: message,
		}

	case "SAVE":
		if err := db.Save(); err != nil {
			return &types.Result{Success: false, Message: fmt.Sprintf("❌ SAVE failed: %v", err)}
		}
		return &types.Result{Success: true, Message: "✅ OK - Data persisted to disk"}

	case "LOAD":
		if err := db.Load(); err != nil {
			return &types.Result{Success: false, Message: fmt.Sprintf("❌ LOAD failed: %v", err)}
		}
		return &types.Result{Success: true, Message: "✅ OK - Data loaded from disk"}

	case "HELP":
		help := formatHelpOutput()
		return &types.Result{Success: true, Message: help}

	case "EXIT":
		return &types.Result{Success: true, Message: "👋 Goodbye!"}

	default:
		return &types.Result{Success: false, Message: fmt.Sprintf("❌ Unknown command: %s", cmd.Type)}
	}
}

// printResult outputs a result
func printResult(out io.Writer, result *types.Result) {
	fmt.Fprintln(out, result.Message)
}

// formatInfoOutput creates a pretty-printed INFO output
func formatInfoOutput(info map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔════════════════════════════════════════╗\n")
	sb.WriteString("║      Database Server Information       ║\n")
	sb.WriteString("╚════════════════════════════════════════╝\n\n")

	// Storage & Records
	sb.WriteString("📊 Storage\n")
	sb.WriteString("───────────────────────────────────\n")
	if records, ok := info["records"].(int); ok {
		sb.WriteString(fmt.Sprintf("  • Total Records:      %d\n", records))
	}
	if memKB, ok := info["memory_usage_kb"].(int); ok {
		sb.WriteString(fmt.Sprintf("  • Memory Usage:       %d KB\n", memKB))
	}
	if fileBytes, ok := info["file_size_bytes"].(int64); ok {
		sb.WriteString(fmt.Sprintf("  • File Size:          %d KB\n", fileBytes/1024))
	}
	if persisted, ok := info["persisted"].(bool); ok {
		persistStr := "No"
		if persisted {
			persistStr = "Yes"
		}
		sb.WriteString(fmt.Sprintf("  • Persistent:         %s\n", persistStr))
	}

	sb.WriteString("\n")

	// Operations
	sb.WriteString("⚡ Operations\n")
	sb.WriteString("───────────────────────────────────\n")
	if gets, ok := info["total_gets"].(int64); ok {
		sb.WriteString(fmt.Sprintf("  • GET operations:     %d\n", gets))
	}
	if sets, ok := info["total_sets"].(int64); ok {
		sb.WriteString(fmt.Sprintf("  • SET operations:     %d\n", sets))
	}
	if deletes, ok := info["total_deletes"].(int64); ok {
		sb.WriteString(fmt.Sprintf("  • DELETE operations:  %d\n", deletes))
	}
	if total, ok := info["total_operations"].(int64); ok {
		sb.WriteString(fmt.Sprintf("  • Total Operations:   %d\n", total))
	}

	sb.WriteString("\n")

	// Cache Performance
	sb.WriteString("💾 Cache Performance\n")
	sb.WriteString("───────────────────────────────────\n")
	if hits, ok := info["cache_hits"].(int64); ok {
		sb.WriteString(fmt.Sprintf("  • Cache Hits:         %d\n", hits))
	}
	if misses, ok := info["cache_misses"].(int64); ok {
		sb.WriteString(fmt.Sprintf("  • Cache Misses:       %d\n", misses))
	}
	if hitRate, ok := info["cache_hit_rate_pct"].(string); ok {
		sb.WriteString(fmt.Sprintf("  • Hit Rate:           %s\n", hitRate))
	}

	sb.WriteString("\n")

	// System Info
	sb.WriteString("🔧 System\n")
	sb.WriteString("───────────────────────────────────\n")
	if uptime, ok := info["uptime_seconds"].(int64); ok {
		sb.WriteString(fmt.Sprintf("  • Uptime:             %d seconds\n", uptime))
	}
	if serverTime, ok := info["server_time"].(string); ok {
		sb.WriteString(fmt.Sprintf("  • Server Time:        %s\n", serverTime))
	}

	sb.WriteString("\n")

	return sb.String()
}

// formatErrorMessage creates a contextual error message with suggestion
func formatErrorMessage(context, message string) string {
	return fmt.Sprintf("%s: %s", context, message)
}

// formatValidationError creates validation errors with helpful suggestions
func formatValidationError(cmd, message string) string {
	suggestions := map[string]string{
		"GET":    "Usage: GET key",
		"SET":    "Usage: SET key value",
		"DELETE": "Usage: DELETE key",
		"KEYS":   "Usage: KEYS pattern (use * for any chars, ? for single char)",
		"SAVE":   "Usage: SAVE (no arguments)",
		"LOAD":   "Usage: LOAD (no arguments)",
		"INFO":   "Usage: INFO (no arguments)",
		"HELP":   "Usage: HELP (no arguments)",
		"EXIT":   "Usage: EXIT (no arguments)",
	}

	msg := fmt.Sprintf("%s: %s", cmd, message)
	if suggestion, ok := suggestions[cmd]; ok {
		msg += fmt.Sprintf("\nℹ️  %s", suggestion)
	}
	return msg
}

// formatHelpOutput creates a pretty-formatted help message
func formatHelpOutput() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔════════════════════════════════════════╗\n")
	sb.WriteString("║         GoFromScratchDB Commands       ║\n")
	sb.WriteString("╚════════════════════════════════════════╝\n\n")

	commands := []struct {
		name  string
		usage string
		desc  string
	}{
		{"SET", "SET key value", "Store a key-value pair in the database"},
		{"GET", "GET key", "Retrieve a value by key"},
		{"DELETE", "DELETE key", "Remove a key from the database"},
		{"KEYS", "KEYS pattern", "List keys matching a glob pattern (* = any, ? = one char)"},
		{"SAVE", "SAVE", "Persist in-memory data to disk"},
		{"LOAD", "LOAD", "Reload data from disk"},
		{"INFO", "INFO", "Show database statistics and metrics"},
		{"HELP", "HELP", "Display this help message"},
		{"EXIT", "EXIT", "Exit the database"},
	}

	for _, cmd := range commands {
		sb.WriteString(fmt.Sprintf("📋 %s\n", cmd.name))
		sb.WriteString(fmt.Sprintf("   Usage: %s\n", cmd.usage))
		sb.WriteString(fmt.Sprintf("   %s\n\n", cmd.desc))
	}

	return sb.String()
}
