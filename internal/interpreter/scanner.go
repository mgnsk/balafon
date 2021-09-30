package interpreter

import (
	"bufio"
	"io"
)

// Scanner interprets input by scanning lines from an io.Reader.
type Scanner struct {
	scanner     *bufio.Scanner
	interpreter *Interpreter
	messages    []Message
	err         error
}

// NewScanner creates a scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		scanner:     bufio.NewScanner(r),
		interpreter: NewInterpreter(),
	}
}

// Err returns the first non-EOF error that was encountered by the Scanner.
func (s *Scanner) Err() error {
	if s.err != nil {
		return s.err
	}
	return s.scanner.Err()
}

// Messages returns the currently accumulated messages.
func (s *Scanner) Messages() []Message {
	return s.messages
}

// Scan the next batch of messages.
func (s *Scanner) Scan() bool {
	s.messages = nil
	s.err = nil

	for s.scanner.Scan() {
		messages, err := s.interpreter.Eval(s.scanner.Text())
		if err != nil {
			s.err = err
			return false
		}
		if messages != nil {
			s.messages = messages
			return true
		}
	}
	return false
}
