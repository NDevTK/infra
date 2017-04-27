// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testexpectations

// Package testexpectations provides a parser for layout test expectation
// files.

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// ExpectationStatement represents a statement (one line) from a layout test
// expectation file.
type ExpectationStatement struct {
	// LineNumber from the original input file.
	LineNumber int
	// Comment, if any.
	Comment string
	// Bugs associated with the test expectation.
	Bugs []string
	// Modifiers (optional) for the test expectation/
	Modifiers []string
	// TestName identifies the test file or test directory.
	TestName string
	// Expectations is a list of expected test results.
	Expectations []string
	// Original line content.
	Original string
}

func (e *ExpectationStatement) String() string {
	if e.TestName == "" {
		return e.Comment
	}

	ret := strings.Join(e.Bugs, " ")

	if len(e.Modifiers) > 0 {
		ret = fmt.Sprintf("%s [ %s ]", ret, strings.Join(e.Modifiers, " "))
	}

	return fmt.Sprintf("%s %s [ %s ]", ret, e.TestName, strings.Join(e.Expectations, " "))
}

type token int

// Constants for internal token types.
const (
	ILLEGAL token = iota
	EOF
	WS
	IDENT
	LB
	RB
	HASH
)

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isIdentStart(ch rune) bool {
	return ch != ' ' && ch != '[' && ch != ']' && ch != '#'
}

var eof = rune(0)

type scanner struct {
	r *bufio.Reader
}

func newScanner(r io.Reader) *scanner {
	return &scanner{r: bufio.NewReader(r)}
}

func (s *scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *scanner) unread() { _ = s.r.UnreadRune() }

func (s *scanner) scan() (tok token, lit string) {
	ch := s.read()

	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isIdentStart(ch) {
		s.unread()
		return s.scanIdent()
	}

	switch ch {
	case eof:
		return EOF, ""
	case '#':
		return HASH, string(ch)
	case '[':
		return LB, string(ch)
	case ']':
		return RB, string(ch)
	}

	return ILLEGAL, string(ch)
}

func (s *scanner) scanWhitespace() (tok token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return WS, buf.String()
}

func (s *scanner) scanIdent() (tok token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if ch == '[' || ch == ']' || ch == '#' || isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return IDENT, buf.String()
}

// Parser parses layout test expectation files.
type Parser struct {
	s   *scanner
	buf struct {
		tok token
		lit string
		n   int
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: newScanner(r)}
}

func (p *Parser) scan() (tok token, lit string) {
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	tok, lit = p.s.scan()

	p.buf.tok, p.buf.lit = tok, lit

	return
}

func (p *Parser) unscan() { p.buf.n = 1 }

func (p *Parser) scanIgnoreWhitespace() (tok token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

func isBug(s string) bool {
	return strings.HasPrefix(s, "crbug.com/") || strings.HasPrefix(s, "Bug(")
}

// Parse parses a *line* of input to produce an ExpectationStatement, or error.
func (p *Parser) Parse() (*ExpectationStatement, error) {
	stmt := &ExpectationStatement{}
	tok, lit := p.scanIgnoreWhitespace()

	// Exit early for a blank line.
	if lit == string(eof) {
		return stmt, nil
	}

	if tok != HASH && tok != IDENT {
		return nil, fmt.Errorf("expected HASH (comment) or IDENT (start of expectation rule) but found %q", lit)
	}

	// Check for optional: HASH comment, return early with the entire line.
	if tok == HASH {
		stmt.Comment = lit
		ch := p.s.read()
		for ; ch != eof; ch = p.s.read() {
			stmt.Comment = stmt.Comment + string(ch)
		}
		return stmt, nil
	}

	// Check for IDENT bugs.
	if tok == IDENT && isBug(lit) {
		for {
			stmt.Bugs = append(stmt.Bugs, lit)
			tok, lit = p.scanIgnoreWhitespace()
			if tok != IDENT || !isBug(lit) {
				p.unscan()
				break
			}
		}
	}

	tok, lit = p.scanIgnoreWhitespace()

	// Check for optional: LB modifiers RB
	if tok == LB {
		for {
			tok, lit = p.scanIgnoreWhitespace()
			if tok == IDENT {
				stmt.Modifiers = append(stmt.Modifiers, lit)
			} else if tok == RB {
				// Scan past the RB so testname can parse.
				tok, lit = p.scanIgnoreWhitespace()
				break
			} else {
				return nil, fmt.Errorf("expected IDENT or RB for modifiers, but found %q", lit)
			}
		}
	}

	if tok == IDENT {
		// Check for IDENT testname
		stmt.TestName = lit
	}

	tok, lit = p.scanIgnoreWhitespace()
	// check for LB expectations RB
	if tok == LB {
		for {
			tok, lit = p.scanIgnoreWhitespace()
			if tok == IDENT {
				stmt.Expectations = append(stmt.Expectations, lit)
			} else if tok == RB {
				break
			} else {
				return nil, fmt.Errorf("expected IDENT or RB for expectations, but found %q", lit)
			}
		}
	} else if lit != string(eof) {
		return nil, fmt.Errorf("expected LB or IDENT for expectations, but found %q", lit)
	}

	return stmt, nil
}
