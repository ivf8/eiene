package parser_test

import (
	"reflect"
	"testing"

	"github.com/ivf8/simp-shell/pkg/ast"
	"github.com/ivf8/simp-shell/pkg/eiene_errors"
	"github.com/ivf8/simp-shell/pkg/parser"
	"github.com/ivf8/simp-shell/pkg/token"
)

var EieneErrors = eiene_errors.NewEieneErrors(false)

func TestParsingSimpleCommand(t *testing.T) {
	tokens := []token.Token{
		newToken(token.PROG_NAME, "cd"),
		newToken(token.EOF, ""),
	}

	_parser := parser.NewParser(tokens)
	result := _parser.Parse()

	expected := []ast.Cmd{
		ast.NewPrimaryCmd(tokens[0], []token.Token{}),
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Parse(%v) got %s. Expected %s",
			tokens, cmdListToString(result), cmdListToString(expected),
		)
	}
}

func TestLogicalAndCommand(t *testing.T) {
	tokens := []token.Token{
		newToken(token.PROG_NAME, "cd"),
		newToken(token.AND, "&&"),
		newToken(token.PROG_NAME, "ls"),
		newToken(token.ARG, "-a"),
		newToken(token.EOF, ""),
	}

	_parser := parser.NewParser(tokens)
	result := _parser.Parse()

	expected := []ast.Cmd{
		ast.NewLogicalCmd(
			ast.NewPrimaryCmd(tokens[0], []token.Token{}),
			tokens[1],
			ast.NewPrimaryCmd(tokens[2], []token.Token{tokens[3]}),
		),
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Parse(%v) got %s. Expected %s",
			tokens, cmdListToString(result), cmdListToString(expected),
		)
	}
}

func TestLogicalORCommand(t *testing.T) {
	tokens := []token.Token{
		newToken(token.PROG_NAME, "cd"),
		newToken(token.OR, "||"),
		newToken(token.PROG_NAME, "ls"),
		newToken(token.ARG, "-a"),
		newToken(token.EOF, ""),
	}

	_parser := parser.NewParser(tokens)
	result := _parser.Parse()

	expected := []ast.Cmd{
		ast.NewLogicalCmd(
			ast.NewPrimaryCmd(tokens[0], []token.Token{}),
			tokens[1],
			ast.NewPrimaryCmd(tokens[2], []token.Token{tokens[3]}),
		),
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Parse(%v) got %s. Expected %s",
			tokens, cmdListToString(result), cmdListToString(expected),
		)
	}
}

func TestMultipleCommandsDelimitedBySemicolon(t *testing.T) {
	tokens := []token.Token{
		newToken(token.PROG_NAME, "cd"),
		newToken(token.OR, "||"),
		newToken(token.PROG_NAME, "ls"),
		newToken(token.ARG, "-a"),
		newToken(token.SEMICOLON, ";"),
		newToken(token.PROG_NAME, "clear"),
		newToken(token.EOF, ""),
	}

	_parser := parser.NewParser(tokens)
	result := _parser.Parse()

	expected := []ast.Cmd{
		ast.NewLogicalCmd(
			ast.NewPrimaryCmd(tokens[0], []token.Token{}),
			tokens[1],
			ast.NewPrimaryCmd(tokens[2], []token.Token{tokens[3]}),
		),
		ast.NewPrimaryCmd(tokens[5], []token.Token{}),
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Parse(%v) got %s. Expected %s",
			tokens, cmdListToString(result), cmdListToString(expected),
		)
	}
}

func TestNoProgramNameInTokens(t *testing.T) {
	tokens := []token.Token{
		newToken(token.ARG, "-a"),
		newToken(token.EOF, ""),
	}

	_parser := parser.NewParser(tokens)
	result := _parser.Parse()

	if result != nil {
		t.Errorf("Parse(%v) got %v. Expected nil", tokens, result)
	}
}

func TestProgramNameAfterArgumentsInTokens(t *testing.T) {
	tokens := []token.Token{
		newToken(token.ARG, "-a"),
		newToken(token.PROG_NAME, "ls"),
		newToken(token.EOF, ""),
	}

	_parser := parser.NewParser(tokens)
	result := _parser.Parse()

	expected := []ast.Cmd{
		ast.NewPrimaryCmd(tokens[1], []token.Token{}),
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Parse(%v) got %v. Expected %v",
			tokens, cmdListToString(result), cmdListToString(expected),
		)
	}
}

func newToken(tokenType token.TokenType, lexeme string) token.Token {
	return token.Token{
		Type:   tokenType,
		Lexeme: lexeme,
	}
}

func cmdListToString(cmdList []ast.Cmd) string {
	astPrinter := ast.NewAstPrinter(cmdList)
	return astPrinter.SPrint()
}
