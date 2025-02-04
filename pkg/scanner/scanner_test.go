package scanner_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ivf8/simp-shell/pkg/eiene_errors"
	"github.com/ivf8/simp-shell/pkg/scanner"
	"github.com/ivf8/simp-shell/pkg/token"
)

var EieneErrors = eiene_errors.NewEieneErrors(false)

// Generates a reader function that returns the readValues one by one
// each time it is called.
func readerFuncGenerator(readValues []string) scanner.ReaderFunc {

	c := make(chan string)

	go func() {
		defer close(c)

		for _, readValue := range readValues {
			c <- readValue
		}
	}()

	return func(prompt string) (string, error) {

		v := <-c

		if v == "^C" {
			return "", errors.New("^C pressed")
		}

		if v == "EOF" {
			return "", errors.New("EOF")
		}

		return v, nil
	}
}

// Scan tokens from commands that may span multiple lines.
func scanTokensMultilineHelper(cmd string, reader scanner.ReaderFunc) []token.Token {
	EieneErrors.ResetErrors()
	_scanner := scanner.NewScanner(cmd, EieneErrors, reader)
	return _scanner.ScanTokens()
}

// Scan tokens from single line commands
func scanTokensHelper(cmd string) []token.Token {
	EieneErrors.ResetErrors()
	_scanner := scanner.NewScanner(cmd, EieneErrors, readerFuncGenerator([]string{"EOF"}))
	return _scanner.ScanTokens()
}

func newToken(tokenType token.TokenType, lexeme string) token.Token {
	return token.Token{
		Type:   tokenType,
		Lexeme: lexeme,
	}
}

func TestSimpleCommand(t *testing.T) {
	result := scanTokensHelper("ls -a")

	expected := []token.Token{
		newToken(token.PROG_NAME, "ls"),
		newToken(token.ARG, "-a"),
		newToken(token.EOF, ""),
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Scan('ls -a') got %v. Expected %v", result, expected)
	}
}

func TestLogicalAndCommand(t *testing.T) {
	cmd := "cd && ls -a"
	result := scanTokensHelper(cmd)

	expected := []token.Token{
		newToken(token.PROG_NAME, "cd"),
		newToken(token.AND, "&&"),
		newToken(token.PROG_NAME, "ls"),
		newToken(token.ARG, "-a"),
		newToken(token.EOF, ""),
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Scan('%s') got %v. Expected %v", cmd, result, expected)
	}
}

func TestLogicalORCommand(t *testing.T) {
	cmd := "cd || ls -a"
	result := scanTokensHelper(cmd)

	expected := []token.Token{
		newToken(token.PROG_NAME, "cd"),
		newToken(token.OR, "||"),
		newToken(token.PROG_NAME, "ls"),
		newToken(token.ARG, "-a"),
		newToken(token.EOF, ""),
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Scan('%s') got %v. Expected %v", cmd, result, expected)
	}
}

func TestBothLogicalsInCommand(t *testing.T) {
	cmd := "cd && ls -a || clear"
	result := scanTokensHelper(cmd)

	expected := []token.Token{
		newToken(token.PROG_NAME, "cd"),
		newToken(token.AND, "&&"),
		newToken(token.PROG_NAME, "ls"),
		newToken(token.ARG, "-a"),
		newToken(token.OR, "||"),
		newToken(token.PROG_NAME, "clear"),
		newToken(token.EOF, ""),
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Scan('%s') got %v. Expected %v", cmd, result, expected)
	}
}

func TestLogicalCommandContinuation(t *testing.T) {
	tests := []struct {
		cmd      string
		reader   scanner.ReaderFunc
		expected []token.Token
	}{
		{
			"ls&&",
			readerFuncGenerator([]string{"ls"}),
			[]token.Token{
				newToken(token.PROG_NAME, "ls"),
				newToken(token.AND, "&&"),
				newToken(token.PROG_NAME, "ls"),
				newToken(token.EOF, ""),
			},
		},
		{
			"ls||",
			readerFuncGenerator([]string{"cd"}),
			[]token.Token{
				newToken(token.PROG_NAME, "ls"),
				newToken(token.OR, "||"),
				newToken(token.PROG_NAME, "cd"),
				newToken(token.EOF, ""),
			},
		},
		{
			"ls&&",
			readerFuncGenerator([]string{"ls&&", "cd"}),
			[]token.Token{
				newToken(token.PROG_NAME, "ls"),
				newToken(token.AND, "&&"),
				newToken(token.PROG_NAME, "ls"),
				newToken(token.AND, "&&"),
				newToken(token.PROG_NAME, "cd"),
				newToken(token.EOF, ""),
			},
		},
		{
			"ls&&",
			readerFuncGenerator([]string{"ls&&", "cd||", "cd -"}),
			[]token.Token{
				newToken(token.PROG_NAME, "ls"),
				newToken(token.AND, "&&"),
				newToken(token.PROG_NAME, "ls"),
				newToken(token.AND, "&&"),
				newToken(token.PROG_NAME, "cd"),
				newToken(token.OR, "||"),
				newToken(token.PROG_NAME, "cd"),
				newToken(token.ARG, "-"),
				newToken(token.EOF, ""),
			},
		},
		{
			// space-only command is not valid
			"ls&&",
			readerFuncGenerator([]string{"  ", "   \t\r\n", "cd -"}),
			[]token.Token{
				newToken(token.PROG_NAME, "ls"),
				newToken(token.AND, "&&"),
				newToken(token.PROG_NAME, "cd"),
				newToken(token.ARG, "-"),
				newToken(token.EOF, ""),
			},
		},
		{
			// ^C stops the reading and tokens are nil
			"ls&&",
			readerFuncGenerator([]string{"ls&&", "^C"}),
			nil,
		},
	}

	for i, test := range tests {
		result := scanTokensMultilineHelper(test.cmd, test.reader)

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("[%d] Scan('%s') got %v. Expected %v", i, test.cmd, result, test.expected)
		}
	}
}

func TestSemicolon(t *testing.T) {
	cmd := "cd ; ls -a"
	result := scanTokensHelper(cmd)

	expected := []token.Token{
		newToken(token.PROG_NAME, "cd"),
		newToken(token.SEMICOLON, ";"),
		newToken(token.PROG_NAME, "ls"),
		newToken(token.ARG, "-a"),
		newToken(token.EOF, ""),
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Scan('%s') got %v. Expected %v", cmd, result, expected)
	}
}

func TestSemicolonAndLogical(t *testing.T) {
	cmd := "cd ; ls -a && clear"
	result := scanTokensHelper(cmd)

	expected := []token.Token{
		newToken(token.PROG_NAME, "cd"),
		newToken(token.SEMICOLON, ";"),
		newToken(token.PROG_NAME, "ls"),
		newToken(token.ARG, "-a"),
		newToken(token.AND, "&&"),
		newToken(token.PROG_NAME, "clear"),
		newToken(token.EOF, ""),
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Scan('%s') got %v. Expected %v", cmd, result, expected)
	}
}

func TestLogicalAndParseError(t *testing.T) {
	errorTextPrefix := "Parse error near "

	tests := []struct {
		cmd, expectedErrorText string
	}{
		{"cd &&& ls -a", errorTextPrefix + "&"},
		{"cd &&&& ls -a", errorTextPrefix + "&&"},
		{"cd &&&&& ls -a", errorTextPrefix + "&&"},
		{"cd && &&& ls -a", errorTextPrefix + "&&"},
		{"cd &&| ls -a", errorTextPrefix + "|"},
		{"cd &&|| ls -a", errorTextPrefix + "||"},
		{"cd && ls -a &&&", errorTextPrefix + "&"},
	}

	for _, test := range tests {
		result := scanTokensHelper(test.cmd)

		if result != nil {
			t.Errorf("Scan('%s') got %v. Expected nil", test.cmd, result)
		}
		if !EieneErrors.HadError {
			t.Errorf("Scan('%s') was expected to throw an error", test.cmd)
		}

		errorText := EieneErrors.Error()
		if errorText != test.expectedErrorText {
			t.Errorf("Scan('%s') got error message %s. Expected %s.", test.cmd, errorText, test.expectedErrorText)
		}
	}
}

func TestLogicalOrParseError(t *testing.T) {
	errorTextPrefix := "Parse error near "

	tests := []struct {
		cmd, expectedErrorText string
	}{
		{"cd ||| ls -a", errorTextPrefix + "|"},
		{"cd |||| ls -a", errorTextPrefix + "||"},
		{"cd ||||| ls -a", errorTextPrefix + "||"},
		{"cd || ||| ls -a", errorTextPrefix + "||"},
		{"cd ||& ls -a", errorTextPrefix + "&"},
		{"cd ||&& ls -a", errorTextPrefix + "&&"},
		{"cd || ls -a |||", errorTextPrefix + "|"},
	}

	for _, test := range tests {
		result := scanTokensHelper(test.cmd)

		if result != nil {
			t.Errorf("Scan('%s') got %v. Expected nil", test.cmd, result)
		}
		if !EieneErrors.HadError {
			t.Errorf("Scan('%s') was expected to throw an error", test.cmd)
		}

		errorText := EieneErrors.Error()
		if errorText != test.expectedErrorText {
			t.Errorf("Scan('%s') got error message %s. Expected %s.", test.cmd, errorText, test.expectedErrorText)
		}
	}
}

func TestSemicolonParseErrors(t *testing.T) {
	errorTextPrefix := "Parse error near "

	tests := []struct {
		cmd, expectedErrorText string
	}{
		{"cd ;; ls -a", errorTextPrefix + ";;"},
		{"cd ;;; ls -a", errorTextPrefix + ";;"},
		{"cd &&;; ls -a", errorTextPrefix + ";;"},
		{"cd ||;; ls -a", errorTextPrefix + ";;"},
		{"cd ;&&;", errorTextPrefix + ";&"},
		{";&", errorTextPrefix + ";&"},
		{"cd ;|&;", errorTextPrefix + ";|"},
		{";|", errorTextPrefix + ";|"},
	}

	for _, test := range tests {
		result := scanTokensHelper(test.cmd)

		if result != nil {
			t.Errorf("Scan('%s') got %v. Expected nil", test.cmd, result)
		}
		if !EieneErrors.HadError {
			t.Errorf("Scan('%s') was expected to throw an error", test.cmd)
		}

		errorText := EieneErrors.Error()
		if errorText != test.expectedErrorText {
			t.Errorf("Scan('%s') got error message %s. Expected %s.", test.cmd, errorText, test.expectedErrorText)
		}
	}
}

func TestSingleSemicolonAfterLogicalNotError(t *testing.T) {
	tests := []struct {
		cmd, expectedErrorText string
	}{
		{"cd &&; ls -a", ""},
		{"cd ||; ls -a", ""},
	}

	for _, test := range tests {
		result := scanTokensHelper(test.cmd)

		if result == nil {
			t.Errorf("Scan('%s') got %v. Not expected to be nil", test.cmd, result)
		}
		if EieneErrors.HadError {
			t.Errorf("Scan('%s') was not expected to throw an error", test.cmd)
		}

		errorText := EieneErrors.Error()
		if errorText != test.expectedErrorText {
			t.Errorf("Scan('%s') got error message %s. Expected %s.", test.cmd, errorText, test.expectedErrorText)
		}
	}
}

func TestReadingOutOfTheSourceSlice(t *testing.T) {
	tests := []struct {
		cmd      string
		expected []token.Token
	}{
		{"cd &&   ", []token.Token{
			newToken(token.PROG_NAME, "cd"),
			newToken(token.AND, "&&"),
			newToken(token.EOF, ""),
		}},
	}

	for _, test := range tests {
		result := scanTokensHelper(test.cmd)

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Scan('%s') got %v. Expected %v", test.cmd, result, test.expected)
		}
	}
}

func TestEmptyCommand(t *testing.T) {
	tests := []struct {
		cmd      string
		expected []token.Token
	}{
		{"", []token.Token{newToken(token.EOF, "")}},
		{"   ", []token.Token{newToken(token.EOF, "")}},
		{"\t", []token.Token{newToken(token.EOF, "")}},
		{"\r", []token.Token{newToken(token.EOF, "")}},
		{"\n", []token.Token{newToken(token.EOF, "")}},
	}

	for _, test := range tests {
		result := scanTokensHelper(test.cmd)

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Scan('%s') got %v. Expected %v", test.cmd, result, test.expected)
		}
	}
}

func TestBackgroundExecution(t *testing.T) {
	expectedErrorText := "Background execution (&) Not implemented"

	tests := []struct {
		cmd string
	}{
		{"cd &"},
		{"cd && ls &"},
	}

	for _, test := range tests {
		result := scanTokensHelper(test.cmd)

		if result != nil {
			t.Errorf("Scan('%s') got %v. Expected nil", test.cmd, result)
		}

		errorText := EieneErrors.Error()
		if errorText != expectedErrorText {
			t.Errorf("Scan('%s') got error message %s. Expected %s.", test.cmd, errorText, expectedErrorText)
		}
	}
}

func TestPiping(t *testing.T) {
	expectedErrorText := "Piping (|) Not implemented"

	tests := []struct {
		cmd string
	}{
		{"cd |"},
		{"ls | cat"},
	}

	for _, test := range tests {
		result := scanTokensHelper(test.cmd)

		if result != nil {
			t.Errorf("Scan('%s') got %v. Expected nil", test.cmd, result)
		}

		errorText := EieneErrors.Error()
		if errorText != expectedErrorText {
			t.Errorf("Scan('%s') got error message %s. Expected %s.", test.cmd, errorText, expectedErrorText)
		}
	}
}

func TestCommandWithComment(t *testing.T) {
	tests := []struct {
		cmd      string
		expected []token.Token
	}{
		{"#this is a comment", []token.Token{
			newToken(token.EOF, ""),
		}},
		{"cd #this is a comment", []token.Token{
			newToken(token.PROG_NAME, "cd"),
			newToken(token.EOF, ""),
		}},
		{"cd #this is a comment\n ls", []token.Token{
			newToken(token.PROG_NAME, "cd"),
			newToken(token.ARG, "ls"),
			newToken(token.EOF, ""),
		}},
	}

	for _, test := range tests {
		result := scanTokensHelper(test.cmd)

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Scan('%s') got %v. Expected %v", test.cmd, result, test.expected)
		}
	}
}

func TestSlashCanEscapeNextCharacter(t *testing.T) {
	tests := []struct {
		cmd      string
		expected []token.Token
	}{
		{"cd \\\\", []token.Token{
			newToken(token.PROG_NAME, "cd"),
			newToken(token.ARG, "\\"),
			newToken(token.EOF, ""),
		}},
		{"cd \\&\\;", []token.Token{
			newToken(token.PROG_NAME, "cd"),
			newToken(token.ARG, "&;"),
			newToken(token.EOF, ""),
		}},
		{"cd \\\\\\one", []token.Token{
			newToken(token.PROG_NAME, "cd"),
			newToken(token.ARG, "\\one"),
			newToken(token.EOF, ""),
		}},
		{"cd \\\\\\one&&ls", []token.Token{
			newToken(token.PROG_NAME, "cd"),
			newToken(token.ARG, "\\one"),
			newToken(token.AND, "&&"),
			newToken(token.PROG_NAME, "ls"),
			newToken(token.EOF, ""),
		}},
		{"cd \\ \\\\one", []token.Token{
			newToken(token.PROG_NAME, "cd"),
			newToken(token.ARG, "\\one"),
			newToken(token.EOF, ""),
		}},
		{"cd \\   \\ls", []token.Token{
			newToken(token.PROG_NAME, "cd"),
			newToken(token.ARG, ""),
			newToken(token.ARG, "ls"),
			newToken(token.EOF, ""),
		}},
		{"cd\\ls", []token.Token{
			newToken(token.PROG_NAME, "cdls"),
			newToken(token.EOF, ""),
		}},
	}

	for _, test := range tests {
		result := scanTokensHelper(test.cmd)

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Scan('%s') got %v. Expected %v", test.cmd, result, test.expected)
		}
	}
}
