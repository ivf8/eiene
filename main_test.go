package main

import (
	"testing"

	"github.com/ivf8/simp-shell/pkg/eiene_errors"
)

func TestRunScanningAndParsing(t *testing.T) {
	eieneErrors := eiene_errors.NewEieneErrors(false)

	parsingTests := []struct {
		cmd              string
		expectedHadError bool
	}{
		// Simple commands
		{"ls", false},
		{"ls -a -l", false},

		// Logical commands
		{"cd && ls", false},
		{"cd &&& ls", true},
		{"cd && || ls", true},
		{"cd || ls", false},
		{"cd ||& ls", true},
		{"cd ||| ls", true},
		{"cd || || ls", true},
		{"cd ||; ls", false},
		{"cd &&; ls", false},
		{"cd || ls && cd -", false},

		// Semicolon separated commands
		{"cd ; ls", false},
		{"cd ;; ls", true},
		{"cd ;&&; ls", true},
		{"cd ;&& ls", true},
		{"cd || ls || cd - ; clear", false},
		{";&&", true},
		{";", false},

		// Comments
		{"#this is a comment", false},
		{"cd #comment", false},
		{"cd #&&&;||", false},

		// With slash
		{"cd\\  \\.", false},
	}

	for _, test := range parsingTests {
		eieneErrors.ResetErrors()

		run(test.cmd, eieneErrors)

		if eieneErrors.HadError != test.expectedHadError {
			t.Errorf(
				"Error parsing (%s) Got %v. Expected %v",
				test.cmd, eieneErrors.HadError, test.expectedHadError,
			)
		}
	}
}

func TestRunInterpreting(t *testing.T) {
	eieneErrors := eiene_errors.NewEieneErrors(false)

	interpreterTests := []struct {
		cmd           string
		expectedError bool
	}{
		// Simple commands
		{"ls", false},
		{"ls -a -l", false},

		// Logical Commands
		{"cd && ls", false},
		{"cd || ls", false},
		{"xoo9 && ls", true},
		{"xoo9 || ls", false},

		// Semicolon separated commands
		{"cd ; ls", false},
		{"cd || ls || cd - ; clear", false},
		{";", false},

		// Comments
		{"#this is a comment", false},
		{"cd #comment", false},
		{"cd #&&&;||", false},

		{"cd\\ls", true},
		{"ls\\  \\.", true},
	}

	for _, test := range interpreterTests {
		eieneErrors.ResetErrors()

		run(test.cmd, eieneErrors)

		if eieneErrors.HadInterpreterError != test.expectedError {
			t.Errorf(
				"Error interpreting (%s) Got %v. Expected %v",
				test.cmd, eieneErrors.HadInterpreterError, test.expectedError,
			)
		}
	}
}
