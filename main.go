package main

import (
	"fmt"
	"io"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/ivf8/simp-shell/pkg/eiene_errors"
	"github.com/ivf8/simp-shell/pkg/interpreter"
	"github.com/ivf8/simp-shell/pkg/parser"
	"github.com/ivf8/simp-shell/pkg/scanner"
)

// Runs a single line
func run(line string, eieneErrors *eiene_errors.EieneErrors) {
	tokens := scanner.NewScanner(line, eieneErrors).ScanTokens()

	if !eieneErrors.HadError {
		cmds := parser.NewParser(tokens).Parse()

		interpreter.NewInterpreter(cmds, eieneErrors).Interpret()
	}
}

func main() {
	reader, err := readline.New("$ ")
	if err != nil {
		color.Red(err.Error())
		return
	}
	defer reader.Close()

	ctrlCClicked := false
	eieneErrors := eiene_errors.NewEieneErrors(true)

	for {
		line, err := reader.Readline()

		switch err {
		case nil:
			ctrlCClicked = false
			if line == "" {
				continue
			}
		case readline.ErrInterrupt:
			{ // ^C
				if ctrlCClicked {
					return
				}
				ctrlCClicked = true
				fmt.Println("To exit, press Ctrl+C again or Ctrl+D")
				continue
			}
		case io.EOF: // ^D
			return
		default:
			color.Red(err.Error())
			return
		}

		run(line, eieneErrors)

		eieneErrors.HadInterpreterError = false
		eieneErrors.ResetErrors()
	}
}
