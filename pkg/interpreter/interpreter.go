package interpreter

import (
	"os"
	"os/exec"

	"github.com/ivf8/simp-shell/pkg/ast"
	"github.com/ivf8/simp-shell/pkg/eiene_errors"
	"github.com/ivf8/simp-shell/pkg/token"
)

// Built in commands
var (
	BUILTINS     = []string{"exit"}
	BUILTINS_MAP = SliceToMap(BUILTINS)
)

type Interpreter struct {
	cmds        []ast.Cmd
	eieneErrors *eiene_errors.EieneErrors
}

func NewInterpreter(cmds []ast.Cmd, e *eiene_errors.EieneErrors) *Interpreter {
	return &Interpreter{
		cmds:        cmds,
		eieneErrors: e,
	}
}

func (i *Interpreter) Interpret() {
	for _, cmd := range i.cmds {
		cmd.Accept(i)

		if c, ok := cmd.(*ast.LogicalCmd); ok {
			if c.Operator.Type == token.OR && !i.eieneErrors.HadError {
				i.eieneErrors.HadInterpreterError = false
			}
		}

		if i.eieneErrors.HadExitError {
			break
		}

		i.eieneErrors.ResetErrors()
	}
}

func (i *Interpreter) VisitLogicalCmd(cmd *ast.LogicalCmd) any {
	switch cmd.Operator.Type {
	case token.AND:
		cmd.Left.Accept(i)
		if i.eieneErrors.HadError {
			return nil
		}

		cmd.Right.Accept(i)
		break

	case token.OR:
		cmd.Left.Accept(i)
		// Return if no error was encountered except if the error is
		// ExitError which is raised by the exit command
		if !i.eieneErrors.HadError || i.eieneErrors.HadExitError {
			return nil
		}

		i.eieneErrors.ResetErrors()
		cmd.Right.Accept(i)
	}

	return nil
}

func (i *Interpreter) VisitPrimaryCmd(cmd *ast.PrimaryCmd) any {
	var args []string
	for _, arg := range cmd.Arguments {
		args = append(args, arg.Lexeme)
	}

	if BUILTINS_MAP[cmd.ProgramName.Lexeme] {
		switch cmd.ProgramName.Lexeme {
		case "exit":
			i.exit()
		}

		return nil
	}

	_cmd := exec.Command(cmd.ProgramName.Lexeme, args...)
	_cmd.Stdin = os.Stdin
	_cmd.Stdout = os.Stdout
	_cmd.Stderr = os.Stderr

	err := _cmd.Run()

	if err != nil {
		i.eieneErrors.InterpreterError(err.Error())
	}

	return nil
}

// Execute exit builtin command
func (i *Interpreter) exit() {
	i.eieneErrors.ExitError()
}

// Creates a map from a slice
// The created map uses the slice values as keys and sets the value of each key to true.
// The produced map can be used to check if a certain value is found in the parent slice.
func SliceToMap[T comparable](arr []T) map[T]bool {
	m := make(map[T]bool)

	for _, v := range arr {
		m[v] = true
	}

	return m
}
