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

// Start starts the REPL loop
func Start(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	db := engine.New()

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
			fmt.Fprintf(out, "Error: %v\n", err)
			continue
		}

		if err := parser.ValidateCommand(cmd); err != nil {
			fmt.Fprintf(out, "Error: %v\n", err)
			continue
		}

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
			return &types.Result{Success: false, Message: "(nil)"}
		}
		return &types.Result{Success: true, Message: fmt.Sprintf("%q", string(value))}

	case "SET":
		err := db.Set(cmd.Args[0], []byte(cmd.Args[1]))
		if err != nil {
			return &types.Result{Success: false, Message: err.Error()}
		}
		return &types.Result{Success: true, Message: "OK"}

	case "DELETE":
		err := db.Delete(cmd.Args[0])
		if err != nil {
			return &types.Result{Success: false, Message: err.Error()}
		}
		return &types.Result{Success: true, Message: "OK"}

	case "KEYS":
		keys := db.Keys(cmd.Args[0])
		if len(keys) == 0 {
			return &types.Result{Success: true, Message: "(empty list)"}
		}
		var sb strings.Builder
		for i, k := range keys {
			sb.WriteString(fmt.Sprintf("%d) %s\n", i+1, k))
		}
		return &types.Result{Success: true, Message: strings.TrimSuffix(sb.String(), "\n")}

	case "INFO":
		info := db.Info()
		return &types.Result{
			Success: true,
			Message: fmt.Sprintf("Records: %d", info["records"]),
		}

	case "HELP":
		help := `Commands:
  SET key value    Store a key-value pair
  GET key          Retrieve value by key
  DELETE key       Remove a key
  KEYS pattern     List keys matching pattern (use * as wildcard)
  INFO             Show database statistics
  HELP             Show this help
  EXIT             Exit the database`
		return &types.Result{Success: true, Message: help}

	case "EXIT":
		return &types.Result{Success: true, Message: "Bye!"}

	default:
		return &types.Result{Success: false, Message: fmt.Sprintf("unknown command: %s", cmd.Type)}
	}
}

// printResult outputs a result
func printResult(out io.Writer, result *types.Result) {
	fmt.Fprintln(out, result.Message)
}
