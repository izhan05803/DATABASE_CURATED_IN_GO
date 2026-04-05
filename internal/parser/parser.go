package parser

import (
	"fmt"
	"strings"

	"github.com/izhan05803/gofromscratchdb/pkg/types"
)

// Parser parses CLI commands
type Parser struct {
	lexer  *Lexer
	tokens []Token
	pos    int
}

// NewParser creates a new parser for the given input
func NewParser(input string) *Parser {
	lexer := NewLexer(input)
	return &Parser{
		lexer:  lexer,
		tokens: lexer.Tokenize(),
		pos:    0,
	}
}

// current returns the current token
func (p *Parser) current() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{Type: TokenEOF}
}

// advance moves to the next token
func (p *Parser) advance() {
	p.pos++
}

// Parse parses the input and returns a Command
func (p *Parser) Parse() (*types.Command, error) {
	token := p.current()
	if token.Type == TokenEOF {
		return nil, fmt.Errorf("empty command")
	}

	if token.Type != TokenWord {
		return nil, fmt.Errorf("expected command, got %v", token.Value)
	}

	cmd := &types.Command{
		Type: strings.ToUpper(token.Value),
		Args: []string{},
	}
	p.advance()

	// Collect arguments
	for {
		token := p.current()
		if token.Type == TokenEOF {
			break
		}
		if token.Type == TokenError {
			return nil, fmt.Errorf("unexpected character: %s", token.Value)
		}
		cmd.Args = append(cmd.Args, token.Value)
		p.advance()
	}

	return cmd, nil
}

// ValidateCommand checks if a command is valid
func ValidateCommand(cmd *types.Command) error {
	switch cmd.Type {
	case "GET":
		if len(cmd.Args) != 1 {
			return fmt.Errorf("GET requires exactly 1 argument")
		}
	case "SET":
		if len(cmd.Args) != 2 {
			return fmt.Errorf("SET requires exactly 2 arguments")
		}
	case "DELETE":
		if len(cmd.Args) != 1 {
			return fmt.Errorf("DELETE requires exactly 1 argument")
		}
	case "KEYS":
		if len(cmd.Args) != 1 {
			return fmt.Errorf("KEYS requires exactly 1 argument")
		}
	case "INFO", "HELP", "EXIT":
		// No arguments required
	default:
		return fmt.Errorf("unknown command: %s", cmd.Type)
	}
	return nil
}
