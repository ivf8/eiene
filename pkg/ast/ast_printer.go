package ast

import (
	"fmt"
	"strings"
)

type AstPrinter struct {
	cmdList []Cmd
}

func NewAstPrinter(cmdList []Cmd) *AstPrinter {
	return &AstPrinter{
		cmdList: cmdList,
	}
}

func (a AstPrinter) Print() {
	fmt.Println(a.SPrint())
}

func (a AstPrinter) SPrint() string {
	s := strings.Builder{}

	for _, cmd := range a.cmdList {
		s.WriteString(cmd.Accept(a).(string))
		s.WriteString(";")
	}

	return s.String()
}

// Print a LogicalCmd enclosed in ()
func (a AstPrinter) VisitLogicalCmd(cmd *LogicalCmd) any {
	logicalCmdBuilder := strings.Builder{}

	logicalCmdBuilder.WriteString(" (")

	logicalCmdBuilder.WriteString(cmd.Left.Accept(a).(string))
	logicalCmdBuilder.WriteString(cmd.Operator.Lexeme)
	logicalCmdBuilder.WriteString(cmd.Right.Accept(a).(string))

	logicalCmdBuilder.WriteString(")")

	return logicalCmdBuilder.String()
}

// Print a PrimaryCmd
func (a AstPrinter) VisitPrimaryCmd(cmd *PrimaryCmd) any {
	primaryCmdBuilder := strings.Builder{}
	primaryCmdBuilder.WriteString(" ")
	primaryCmdBuilder.WriteString(cmd.ProgramName.Lexeme)
	primaryCmdBuilder.WriteString(" ")

	for _, arg := range cmd.Arguments {
		primaryCmdBuilder.WriteString(" " + arg.Lexeme)
	}

	return primaryCmdBuilder.String()
}
