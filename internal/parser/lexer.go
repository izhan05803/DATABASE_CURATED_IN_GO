package parser

import (
	"unicode"
)

// TokenType represents the type of a token
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenWord
	TokenString
	TokenError
)

// Token represents a lexical token
type Token struct {
	Type  TokenType
	Value string
}

// Lexer tokenizes input strings
type Lexer struct {
	input   []rune
	pos     int
	current rune
}

// NewLexer creates a new lexer for the given input
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input: []rune(input),
		pos:   0,
	}
	if len(l.input) > 0 {
		l.current = l.input[0]
	}
	return l
}

// advance moves to the next character
func (l *Lexer) advance() {
	l.pos++
	if l.pos < len(l.input) {
		l.current = l.input[l.pos]
	} else {
		l.current = 0
	}
}

// skipWhitespace skips whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.current != 0 && unicode.IsSpace(l.current) {
		l.advance()
	}
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.current == 0 {
		return Token{Type: TokenEOF}
	}

	// Handle quoted strings
	if l.current == '"' {
		return l.readString()
	}

	// Handle words
	if unicode.IsLetter(l.current) || unicode.IsDigit(l.current) || l.current == ':' || l.current == '*' || l.current == '_' || l.current == '-' {
		return l.readWord()
	}

	// Unknown character
	ch := l.current
	l.advance()
	return Token{Type: TokenError, Value: string(ch)}
}

// readString reads a quoted string
func (l *Lexer) readString() Token {
	l.advance() // skip opening quote
	start := l.pos

	for l.current != 0 && l.current != '"' {
		l.advance()
	}

	value := string(l.input[start:l.pos])

	if l.current == '"' {
		l.advance() // skip closing quote
	}

	return Token{Type: TokenString, Value: value}
}

// readWord reads an unquoted word
func (l *Lexer) readWord() Token {
	start := l.pos

	for l.current != 0 && !unicode.IsSpace(l.current) && l.current != '"' {
		l.advance()
	}

	return Token{Type: TokenWord, Value: string(l.input[start:l.pos])}
}

// Tokenize returns all tokens from the input
func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		token := l.NextToken()
		tokens = append(tokens, token)
		if token.Type == TokenEOF {
			break
		}
	}
	return tokens
}
