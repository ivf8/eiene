package parser

import (
	"github.com/ivf8/simp-shell/pkg/ast"
	"github.com/ivf8/simp-shell/pkg/token"
)

type Parser struct {
	tokens  []token.Token
	current int
}

func NewParser(tokens []token.Token) *Parser {
	return &Parser{
		tokens:  tokens,
		current: 0,
	}
}

func (p *Parser) Parse() []ast.Cmd {
	var cmdList []ast.Cmd

	for !p.isAtEnd() {
		cmd := p.command()
		if cmd != nil {
			cmdList = append(cmdList, cmd)
		}
	}

	return cmdList
}

func (p *Parser) command() ast.Cmd {
	cmd := p.logical()

	// Consume the semicolon
	if p.match(token.SEMICOLON) {
	}

	return cmd
}

// Parses a logical command recursively. Calls the primary method to parse
// the individual commands
// Returns a new logical command if && or || are found,
// else it just returns the primary command
func (p *Parser) logical() ast.Cmd {
	cmd := p.primary()

	// Advance to prevent infnite recursion
	if cmd == nil {
		p.advance()
	}

	var operator token.Token
	var right ast.Cmd

	if p.match(token.AND, token.OR) {
		operator = p.previous()
		right = p.logical()
	}

	if (operator.Type == token.AND || operator.Type == token.OR) && right != nil {
		return ast.NewLogicalCmd(cmd, operator, right)
	} else {
		return cmd
	}
}

// Parses individual command and its arguments
// Returns a new PrimaryCmd.
func (p *Parser) primary() ast.Cmd {
	if p.match(token.PROG_NAME) {
		programName := p.previous()
		arguments := []token.Token{}

		for p.match(token.ARG) {
			arguments = append(arguments, p.previous())
		}
		return ast.NewPrimaryCmd(programName, arguments)
	}
	return nil
}

// Checks if the current token matches either of the given tokenTypes.
// If the type matches, it also advances current.
// Returns true if a match is found, else false if no match or is at end of tokens.
func (p *Parser) match(tokenTypes ...token.TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	for _, tokenType := range tokenTypes {
		if p.peek().Type == tokenType {
			p.advance()
			return true
		}
	}

	return false
}

// Advances current and returns the next token to be parsed.
func (p *Parser) advance() token.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

// Returns the token in the current-1 position
func (p Parser) previous() token.Token {
	return p.tokens[p.current-1]
}

// Returns the token in the current position
func (p *Parser) peek() token.Token {
	return p.tokens[p.current]
}

// Checks if the EOF token has been reached.
func (p Parser) isAtEnd() bool {
	return p.peek().Type == token.EOF
}
