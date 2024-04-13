package lexer

import (
	"unicode"
	"unicode/utf8"
)

type lexer struct {
	input string // the string being scanned
	pos   int    // current position in the input
	start int    // position where did the current token start
	head  rune   // the cursor: the currently processed rune
}

func newLexer(input string) lexer {
	return lexer{input, 0, 0, utf8.RuneError}
}

// Read a token, return the status.
func (l *lexer) nextToken() (Token, error) {

	// skip whitespaces
	l.takeWhile(unicode.IsSpace)

	// reset token starting position in the input
	l.start = l.pos

	// peek first rune
	if l.next() == 0 {
		return Token{}, EoF{}
	}

	// atom, bool, etc
	if l.expect(unicode.IsLower) {
		l.takeWhile(isName)
		val, size := l.collect()
		if size == 0 {
			return Token{}, Empty{}
		}
		return Token{atomType(val), val}, nil
	}

	// variable
	if l.expect(unicode.IsUpper) {
		l.takeWhile(isName)
		return l.collectToken(Variable)
	}

	// anonymous variable
	if l.expectIs('_') {
		if l.expectNext(isName) {
			// named
			l.takeWhile(isName)
			return l.collectToken(Variable)
		}
		return Token{Dummy, "_"}, nil
	}

	// integer
	if l.expect(unicode.IsDigit) {
		l.takeWhile(unicode.IsDigit)
		return l.collectToken(Number)
	}

	// operator
	if l.expect(isOperator) {
		l.takeWhile(isOperator)
		val, _ := l.collect()
		return Token{operatorType(val), val}, nil
	}

	// string
	if l.expectIs('"') {
		l.readString()
		return l.collectToken(String)
	}

	// skip the comment
	if l.expectIs('%') {
		l.takeUntilIs('\n')
		return l.nextToken()
	}

	// other special, single-character tokens
	return literalToken(l.head)
}

// Collect the substring in the `start:pos` range.
func (l *lexer) collect() (string, int) {
	return l.input[l.start:l.pos], l.pos - l.start
}

// Collect the substring in the `start:pos` range as a token of a given type.
func (l *lexer) collectToken(typ TokenType) (Token, error) {
	val, size := l.collect()
	if size == 0 {
		return Token{}, Empty{}
	}
	return Token{typ, val}, nil
}

// Peek what is the next rune without moving the pointer.
func (l *lexer) peek() (rune, int) {
	return utf8.DecodeRuneInString(l.input[l.pos:])
}

// Move the `head` cursor one rune ahead, return size of the rune.
func (l *lexer) next() int {
	var width int
	l.head, width = l.peek()
	l.pos += width
	return width
}

// The cursor matched the condition.
func (l *lexer) expect(matches func(rune) bool) bool {
	return matches(l.head)
}

// Peek if next rune matched the condition.
func (l *lexer) expectNext(matches func(rune) bool) bool {
	r, width := l.peek()
	return width > 0 && matches(r)
}

// The cursor is a specific rune
func (l *lexer) expectIs(r rune) bool {
	return l.head == r
}

// Iterate over the input (move the cursor) while it matches the condition.
func (l *lexer) takeWhile(matches func(rune) bool) bool {
	start := l.pos
	for {
		r, width := l.peek()
		if width == 0 || !matches(r) {
			break
		}
		l.head = r
		l.pos += width
	}
	return l.pos > start
}

// Iterate over the input (move the cursor) until hitting the character.
func (l *lexer) takeUntilIs(r rune) {
	for l.next() > 0 {
		if l.head == r {
			break
		}
	}
}

// Take characters until '"' while respecting quoted characters.
func (l *lexer) readString() bool {
	for {
		r, width := l.peek()
		if width == 0 {
			return false
		}
		l.next()

		switch r {
		case '\\':
			l.next()
		case '"':
			return true
		}
	}
}
