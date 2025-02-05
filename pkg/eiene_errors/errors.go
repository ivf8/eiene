package eiene_errors

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type EieneErrors struct {
	HadError            bool
	HadInterpreterError bool
	HadExitError        bool
	Errors              []string
	printErrors         bool
}

func NewEieneErrors(printErrors bool) *EieneErrors {
	return &EieneErrors{
		HadError:            false,
		HadInterpreterError: false,
		HadExitError:        false,
		Errors:              []string{},
		printErrors:         printErrors,
	}
}

func (e *EieneErrors) ParseError(message string) {
	errorMessage := "Parse error near " + message

	e.Errors = append(e.Errors, errorMessage)
	e.Report(errorMessage)
}

func (e *EieneErrors) NotImplementedError(message string) {
	errorMessage := message + " Not implemented"

	e.Errors = append(e.Errors, errorMessage)
	e.Report(errorMessage)
}

func (e *EieneErrors) InterpreterError(message string) {
	errorMessage := strings.TrimPrefix(message, "exec: ")

	e.HadInterpreterError = true
	e.Errors = append(e.Errors, errorMessage)
	e.Report(errorMessage)
}

// Error raised by the exit command
func (e *EieneErrors) ExitError() {
	e.HadExitError = true
	e.HadError = true
}

func (e *EieneErrors) Report(message string) {
	e.HadError = true
	if e.printErrors {
		color.Red("eiene: %s", message)
	}
}

func (e *EieneErrors) ResetErrors() {
	e.HadError = false
	e.Errors = []string{}
}

func (e EieneErrors) Error() string {
	return fmt.Sprintf(
		"%v",
		strings.Join(e.Errors, "\n"),
	)
}
