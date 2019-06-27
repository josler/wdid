package parser

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type lexedItem struct {
	typ lexedItemType
	val string
}

type lexedItemType int

const (
	lexItemError lexedItemType = iota
	lexItemEOF
	lexItemEq
	lexItemNe
	lexItemString
	lexItemIdentifier
	lexItemComma
)

const EOF rune = 0
const EqualSign string = "="
const NotEqualSign string = "!="
const Comma string = ","

func (i lexedItem) String() string {
	switch i.typ {
	case lexItemEOF:
		return "EOF"
	case lexItemError:
		return i.val
	}

	return fmt.Sprintf("%q", i.val)
}

type stateFn func(*lexer) stateFn

type lexer struct {
	input string
	start int
	pos   int
	width int
	items chan lexedItem
}

func lex(input string) (*lexer, chan lexedItem) {
	l := &lexer{
		input: input,
		items: make(chan lexedItem),
	}
	go l.run()
	return l, l.items
}

func (l *lexer) run() {
	for state := lexIdentifier; state != nil; {
		state = state(l)
	}
	close(l.items) // done
}

func (l *lexer) emit(t lexedItemType) {
	l.items <- lexedItem{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return EOF
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func lexIdentifier(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], EqualSign) {
			if l.pos > l.start {
				l.emit(lexItemIdentifier)
			}
			return lexEq
		} else if strings.HasPrefix(l.input[l.pos:], NotEqualSign) {
			if l.pos > l.start {
				l.emit(lexItemIdentifier)
			}
			return lexNe
		}
		if l.next() == EOF {
			break
		}
	}
	// reached EOF, emit what we have
	if l.pos > l.start {
		l.emit(lexItemIdentifier)
	}
	l.emit(lexItemEOF)
	return nil // stop loop if we hit here
}

func lexEq(l *lexer) stateFn {
	l.pos += len(EqualSign)
	l.emit(lexItemEq)
	return lexString // now inside value
}

func lexNe(l *lexer) stateFn {
	l.pos += len(NotEqualSign)
	l.emit(lexItemNe)
	return lexString
}

func lexString(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], Comma) {
			if l.pos > l.start {
				l.emit(lexItemString)
			}
			return lexComma
		}
		if l.next() == EOF {
			break
		}
	}
	// reached EOF
	if l.pos > l.start {
		l.emit(lexItemString)
	}
	l.emit(lexItemEOF)
	return nil
}

func lexComma(l *lexer) stateFn {
	l.pos += len(Comma)
	l.emit(lexItemComma)
	return lexIdentifier // now looking for another identifier
}
