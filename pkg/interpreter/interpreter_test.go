package interpreter_test

import (
	"testing"

	"github.com/ivf8/simp-shell/pkg/ast"
	"github.com/ivf8/simp-shell/pkg/eiene_errors"
	"github.com/ivf8/simp-shell/pkg/interpreter"
	"github.com/ivf8/simp-shell/pkg/token"
)

// Tokens
var (
	// Logical Operators
	AND_OP = newToken(token.AND, "&&")
	OR_OP  = newToken(token.OR, "||")

	// Program names
	LS           = newToken(token.PROG_NAME, "ls")
	CD           = newToken(token.PROG_NAME, "cd")
	EXIT         = newToken(token.PROG_NAME, "exit")
	INVALID_PROG = newToken(token.PROG_NAME, "invalid-prog")
)

// Commands
var (
	// Primary commands
	LS_CMD = ast.NewPrimaryCmd(LS, []token.Token{})
	CD_CMD = ast.NewPrimaryCmd(CD, []token.Token{})

	// Logical commands
	LOGICAL_AND_CMD = ast.NewLogicalCmd(LS_CMD, AND_OP, LS_CMD)
	LOGICAL_OR_CMD  = ast.NewLogicalCmd(LS_CMD, OR_OP, LS_CMD)

	// Invalid cmd
	INVALID_CMD                  = ast.NewPrimaryCmd(INVALID_PROG, []token.Token{})
	LOGICAL_AND_WITH_INVALID_CMD = ast.NewLogicalCmd(INVALID_CMD, AND_OP, LS_CMD)
	LOGICAL_OR_WITH_INVALID_CMD  = ast.NewLogicalCmd(INVALID_CMD, OR_OP, LS_CMD)

	// Exit commands
	EXIT_CMD                  = ast.NewPrimaryCmd(EXIT, []token.Token{})
	LOGICAL_AND_WITH_EXIT_CMD = ast.NewLogicalCmd(EXIT_CMD, AND_OP, LS_CMD)
	LOGICAL_OR_WITH_EXIT_CMD  = ast.NewLogicalCmd(EXIT_CMD, OR_OP, LS_CMD)
)

var EieneErrors = eiene_errors.NewEieneErrors(false)

func newToken(tokenType token.TokenType, lexeme string) token.Token {
	return token.Token{
		Type:   tokenType,
		Lexeme: lexeme,
	}
}

func interpreterHelper(cmds []ast.Cmd) {
	EieneErrors.ResetErrors()
	EieneErrors.HadInterpreterError = false

	_interpreter := interpreter.NewInterpreter(cmds, EieneErrors)
	_interpreter.Interpret()
}

func TestInterpreter(t *testing.T) {
	tests := []struct {
		cmds                     []ast.Cmd
		expectedInterpreterError bool
	}{
		{[]ast.Cmd{LS_CMD}, false},
		{[]ast.Cmd{INVALID_CMD}, true},
		{[]ast.Cmd{EXIT_CMD}, false},

		// Logical Commands
		{[]ast.Cmd{LOGICAL_OR_CMD}, false},
		{[]ast.Cmd{LOGICAL_AND_CMD}, false},
		{[]ast.Cmd{LOGICAL_OR_WITH_INVALID_CMD}, true},
		{[]ast.Cmd{LOGICAL_AND_WITH_INVALID_CMD}, true},
	}

	for _, test := range tests {
		interpreterHelper(test.cmds)

		if EieneErrors.HadInterpreterError != test.expectedInterpreterError {
			t.Errorf(
				"Error interpreting (%s) Got %v. Expected %v",
				test.cmds, EieneErrors.HadInterpreterError, test.expectedInterpreterError,
			)
		}
	}
}

func TestExitBuiltinCommand(t *testing.T) {
	tests := []struct {
		cmds              []ast.Cmd
		expectedExitError bool
	}{
		{[]ast.Cmd{EXIT_CMD}, true},
		{[]ast.Cmd{LOGICAL_AND_WITH_EXIT_CMD}, true},
		{[]ast.Cmd{LOGICAL_OR_WITH_EXIT_CMD}, true},
		{[]ast.Cmd{LS_CMD, EXIT_CMD}, true},
		{[]ast.Cmd{LS_CMD, EXIT_CMD, LS_CMD}, true},
	}

	for _, test := range tests {
		interpreterHelper(test.cmds)

		if EieneErrors.HadExitError != test.expectedExitError {
			t.Errorf(
				"Error interpreting (%s) Got %v. Expected %v",
				test.cmds, EieneErrors.HadInterpreterError, test.expectedExitError,
			)
		}
	}
}

func TestCdBuiltinCommand(t *testing.T) {

	cdCmdWithArg := func(arg string) ast.Cmd {
		return ast.NewPrimaryCmd(CD, []token.Token{{
			Type:   token.ARG,
			Lexeme: arg,
		}})
	}

	tests := []struct {
		cmds                     []ast.Cmd
		expectedInterpreterError bool
	}{
		{[]ast.Cmd{CD_CMD}, false},
		{[]ast.Cmd{cdCmdWithArg("-")}, false},
		{[]ast.Cmd{cdCmdWithArg("~")}, false},
		{[]ast.Cmd{cdCmdWithArg("/")}, false},
		{[]ast.Cmd{cdCmdWithArg(".")}, false},
		{[]ast.Cmd{cdCmdWithArg("unknown-directory-001")}, true},
	}

	for _, test := range tests {
		interpreterHelper(test.cmds)

		if EieneErrors.HadInterpreterError != test.expectedInterpreterError {
			t.Errorf(
				"Error interpreting (%s) Got %v. Expected %v",
				test.cmds, EieneErrors.HadInterpreterError, test.expectedInterpreterError,
			)
		}
	}
}
