package parser

import (
	"testing"

	"github.com/izhan05803/gofromscratchdb/pkg/types"
)

func TestLexer_Tokenize(t *testing.T) {
	tests := []struct {
		input string
		want  []TokenType
	}{
		{"GET key", []TokenType{TokenWord, TokenWord, TokenEOF}},
		{`SET key "hello world"`, []TokenType{TokenWord, TokenWord, TokenString, TokenEOF}},
		{"", []TokenType{TokenEOF}},
		{"KEYS user:*", []TokenType{TokenWord, TokenWord, TokenEOF}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens := lexer.Tokenize()

			if len(tokens) != len(tt.want) {
				t.Errorf("got %d tokens, want %d", len(tokens), len(tt.want))
				return
			}

			for i, tok := range tokens {
				if tok.Type != tt.want[i] {
					t.Errorf("token %d: got %v, want %v", i, tok.Type, tt.want[i])
				}
			}
		})
	}
}

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		input   string
		cmdType string
		args    []string
		wantErr bool
	}{
		{"GET foo", "GET", []string{"foo"}, false},
		{`SET foo "bar"`, "SET", []string{"foo", "bar"}, false},
		{"DELETE foo", "DELETE", []string{"foo"}, false},
		{"KEYS user:*", "KEYS", []string{"user:*"}, false},
		{"INFO", "INFO", []string{}, false},
		{"", "", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewParser(tt.input)
			cmd, err := parser.Parse()

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if cmd.Type != tt.cmdType {
				t.Errorf("cmd.Type = %s, want %s", cmd.Type, tt.cmdType)
			}

			if len(cmd.Args) != len(tt.args) {
				t.Errorf("len(cmd.Args) = %d, want %d", len(cmd.Args), len(tt.args))
				return
			}

			for i, arg := range cmd.Args {
				if arg != tt.args[i] {
					t.Errorf("cmd.Args[%d] = %s, want %s", i, arg, tt.args[i])
				}
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		cmdType string
		args    []string
		wantErr bool
	}{
		{"GET", []string{"foo"}, false},
		{"GET", []string{}, true},
		{"SET", []string{"foo", "bar"}, false},
		{"SET", []string{"foo"}, true},
		{"DELETE", []string{"foo"}, false},
		{"KEYS", []string{"*"}, false},
		{"INFO", []string{}, false},
		{"UNKNOWN", []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.cmdType, func(t *testing.T) {
			cmd := &types.Command{
				Type: tt.cmdType,
				Args: tt.args,
			}

			err := ValidateCommand(cmd)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
