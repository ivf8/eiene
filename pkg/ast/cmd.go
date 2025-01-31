package ast

import "github.com/ivf8/simp-shell/pkg/token"

// Interface implemented by the different commands
type Cmd interface {
	// Method called in order to perform a specific operation
	// on a certain Cmd as defined by the visitor's visiting method.
	Accept(visitor CmdVisitor) any
}

// Interface implemented by any struct that interacts with Cmd.
type CmdVisitor interface {
	VisitLogicalCmd(cmd *LogicalCmd) any
	VisitPrimaryCmd(cmd *PrimaryCmd) any
}

// Command that uses the logical operators && or ||.
// Constains two seperate commands delimited by the operator.
type LogicalCmd struct {
	Right    Cmd
	Operator token.Token
	Left     Cmd
}

func NewLogicalCmd(left Cmd, operator token.Token, right Cmd) *LogicalCmd {
	return &LogicalCmd{
		Left:     left,
		Operator: operator,
		Right:    right,
	}
}

// Implement the Cmd interface
func (l *LogicalCmd) Accept(visitor CmdVisitor) any {
	return visitor.VisitLogicalCmd(l)
}

// An individual command containing name of the program to run and Arguments
// to pass to the program
type PrimaryCmd struct {
	ProgramName token.Token
	Arguments   []token.Token
}

func NewPrimaryCmd(programName token.Token, arguments []token.Token) *PrimaryCmd {
	return &PrimaryCmd{
		ProgramName: programName,
		Arguments:   arguments,
	}
}

// Implement the Cmd interface.
func (p *PrimaryCmd) Accept(visitor CmdVisitor) any {
	return visitor.VisitPrimaryCmd(p)
}
