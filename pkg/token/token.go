package token

type TokenType string

type Token struct {
	Type   TokenType
	Lexeme string
}

const (
	EOF TokenType = "EOF"

	// Program run and its arguments
	PROG_NAME TokenType = "PROGRAM_NAME"
	ARG       TokenType = "ARGUMENT"

	// Separate commands
	SEMICOLON TokenType = "SEMICOLON"

	// Logical
	AND TokenType = "AND" // &&
	OR  TokenType = "OR"  // ||
)
