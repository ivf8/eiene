package interpreter

import (
	"os"
	"os/exec"

	"github.com/ivf8/simp-shell/pkg/ast"
	"github.com/ivf8/simp-shell/pkg/eiene_errors"
	"github.com/ivf8/simp-shell/pkg/token"
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
		if !i.eieneErrors.HadError {
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
