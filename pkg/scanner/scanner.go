package scanner

import (
	"strings"

	"github.com/ivf8/simp-shell/pkg/eiene_errors"
	"github.com/ivf8/simp-shell/pkg/token"
)

// &, | and ; characters. They are intepreted differently from other
// characters unless if enclosed between quotes or in a comment.
var (
	SPECIAL_CHARS     = []rune{'&', '|', ';'}
	SPECIAL_CHARS_MAP = SliceToMap(SPECIAL_CHARS)
)

type Flags struct {
	slashFound bool
	spaceFound bool
	newCmd     bool // If true next token is the program name
}

// Function for continuing to read a command from the cmd line.
type ReaderFunc func(prompt string) (string, error)

type Scanner struct {
	source  []rune        // Runes from the source string
	Tokens  []token.Token // Scanned tokens
	start   int           // Index to start indexing the current lexeme being scanned
	current int           // Index of next character to be scanned

	flags       *Flags
	eieneErrors *eiene_errors.EieneErrors

	reader ReaderFunc // Function for reading a command that is continued
}

func NewScanner(source string, e *eiene_errors.EieneErrors, reader ReaderFunc) *Scanner {
	return &Scanner{
		source:      []rune(source),
		Tokens:      []token.Token{},
		start:       0,
		current:     0,
		eieneErrors: e,

		flags: &Flags{
			slashFound: false,
			spaceFound: false,
			newCmd:     true,
		},

		reader: reader,
	}
}

// Scans for tokens in source.
func (s *Scanner) ScanTokens() []token.Token {
	s.eieneErrors.ResetErrors()

	for !s.isAtEnd() && !s.eieneErrors.HadError {
		s.start = s.current
		s.scanToken()
	}

	if s.eieneErrors.HadError {
		return nil
	}

	s.Tokens = append(s.Tokens, token.Token{
		Type:   token.EOF,
		Lexeme: "",
	})

	return s.Tokens
}

// Scan and append single tokens.
func (s *Scanner) scanToken() {
	c := s.advance()

	switch c {
	case ';':
		if SPECIAL_CHARS_MAP[s.peek()] {
			s.eieneErrors.ParseError(";" + string(s.peek()))
			return
		}
		s.addToken(token.SEMICOLON)

		s.flags.newCmd = true
	case '&':
		if s.peek() == '&' {
			s.logicalOperator(token.AND)
		} else {
			// TODO: Background task
			s.eieneErrors.NotImplementedError("Background execution (&)")
		}
	case '|':
		if s.peek() == '|' {
			s.logicalOperator(token.OR)
		} else {
			// TODO: Piping
			s.eieneErrors.NotImplementedError("Piping (|)")
		}

	// Whitespace
	case ' ',
		'\t',
		'\r':
		s.flags.spaceFound = true
		break

	case '\n':
		break

	// Comment
	case '#':
		for s.peek() != '\n' && !s.isAtEnd() {
			s.advance()
		}

	// Back slash - Escape next character or continue reading command in next line
	case '\\':
		_continueReading := false
		if s.isAtEnd() {
			s.source = s.source[:len(s.source)-1]
			_continueReading = true
		} else {
			s.source = append(s.source[:s.current-1], s.source[s.current:]...)
		}

		// Prompt for command continuation if the command ends with \
		if s.peek() == rune(0) && _continueReading {
			line, err := s.reader(">")
			if err != nil {
				s.eieneErrors.HadError = true
				s.eieneErrors.Errors = append(s.eieneErrors.Errors, err.Error())
			}
			if len(line) > 0 {
				s.source = append(s.source, []rune(line)...)
				s.current-- // Don't skip first character of next line, might be space
			}
		}

		if s.isAtEnd() && s.source[len(s.source)-1] != '\\' {
			s.current--
		}

		for s.peek() != ' ' && !s.isAtEnd() {
			if s.peek() == '\\' && s.peekNext() != rune(0) {
				// Remove the current \ and replace with next character
				s.source = append(s.source[:s.current], s.source[s.current+1:]...)
			} else if SPECIAL_CHARS_MAP[s.peek()] || s.peekNext() == rune(0) {
				break
			}
			s.advance()
		}
		s.flags.slashFound = true

		fallthrough

	// Command and arguments
	default:
		for s.peek() != ' ' && s.peek() != '\n' && s.peek() != '\\' && !SPECIAL_CHARS_MAP[s.peek()] && !s.isAtEnd() {
			s.advance()
		}

		if s.flags.newCmd {
			s.addToken(token.PROG_NAME)

			// Tokens after this one will be arguments or operators
			s.flags.newCmd = false
		} else if !s.flags.spaceFound && s.flags.slashFound {
			s.updatePreviousToken()
		} else {
			s.addToken(token.ARG)
		}

		s.flags.spaceFound = false
	}

	s.flags.slashFound = false
}

func (s *Scanner) logicalOperator(tokenType token.TokenType) {
	s.advance()

	_current := s.current
	if s.peek() == ' ' {
		_current = s.consumeWhitespace()
	}

	// Prevent reading out of s.source when _current is >= len(s.source)
	_isSpecialCharacter := false
	if _current < len(s.source) {
		_isSpecialCharacter = SPECIAL_CHARS_MAP[s.source[_current]]
	}

	// OK
	if !SPECIAL_CHARS_MAP[s.peek()] && !_isSpecialCharacter {
		s.addToken(tokenType)
		s.flags.newCmd = true

		// Continue reading if the command ends in && or ||
		if s.peek() == rune(0) {
			prompt := "cmdand>"
			if tokenType == token.OR {
				prompt = "cmdor>"
			}

			// Exit if ^C is pressed or a non-empty command is entered
			for !s.eieneErrors.HadError {
				line, err := s.reader(prompt)
				if err != nil {
					s.eieneErrors.HadError = true
					s.eieneErrors.Errors = append(s.eieneErrors.Errors, err.Error())
				}

				// Trim space and tabs to prevent whitespace only commands
				line = strings.Trim(line, " \t\r\n")
				if len(line) > 0 {
					s.source = append(s.source, []rune(line)...)
					break
				}
			}
		}

		return
	}

	// If the next character is ; eg &&; it's not an error
	// unless if the semicolon is repeated eg &&;;. This case
	// where the error occurs will be handled in the semicolon(;) case
	if s.peek() == ';' || (!s.isAtEnd() && s.source[_current] == ';') {
		s.addToken(tokenType)
		s.flags.newCmd = true
		return
	}

	// Error
	s.start = _current
	s.current = _current

	s.advance()
	if SPECIAL_CHARS_MAP[s.peek()] {
		s.advance()
	}
	error_chars := string(s.source[s.start:s.current])
	s.eieneErrors.ParseError(error_chars)
}

// Consumes whitespace from s.current to the next non-whitespace character.
// Returns the index of the next character that is not whitespace
func (s *Scanner) consumeWhitespace() int {
	idx := s.current
	for idx < len(s.source) && (s.source[idx] == ' ' || s.source[idx] == '\t' || s.source[idx] == '\r') {
		idx++
	}

	// Prevent reading out of the s.source slice
	if idx == len(s.source) {
		idx--
	}
	return idx
}

// Adds a token of Type tokenType to the s.Tokens slice
func (s *Scanner) addToken(tokenType token.TokenType) {
	value := strings.Trim(string(s.source[s.start:s.current]), " ")

	s.Tokens = append(s.Tokens, token.Token{
		Type:   tokenType,
		Lexeme: value,
	})
}

// Updates the previous token after a \ is encountered and
// no space is found between the last character in previous line and
// the first in the next line
func (s *Scanner) updatePreviousToken() {
	// Error if called with s.Tokens being empty
	if len(s.Tokens) == 0 {
		s.eieneErrors.ParseError(
			"Error in scanner. Not configured properly" +
				"function: s.updatePreviousToken",
		)
		return
	}

	value := strings.Trim(string(s.source[s.start:s.current]), " ")
	s.Tokens[len(s.Tokens)-1].Lexeme = s.Tokens[len(s.Tokens)-1].Lexeme + value
}

// Gets next character to be scanned.
// Returns character at s.current in s.source and increments current by 1.
func (s *Scanner) advance() rune {
	if s.isAtEnd() {
		return rune(0)
	}

	c := s.source[s.current]
	s.current++

	return c
}

// Looks up next character and returns it if not at EOF
// else '/0' is returned
// Does not increment s.current
func (s Scanner) peek() rune {
	if s.isAtEnd() {
		return rune(0)
	}
	return s.source[s.current]
}

// Looks up the character after current
func (s Scanner) peekNext() rune {
	if s.current+1 >= len(s.source) {
		return rune(0)
	}
	return s.source[s.current+1]
}

// Checks if s.current is at the end of s.source.
func (s Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
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
