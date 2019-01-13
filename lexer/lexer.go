// Package lexer translates monkey text into a stream of tokens.
package lexer

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/ajwerner/monkey/token"
)

// Lexer is used to lexagraphically analyze monkey source.
type Lexer struct {
	input   string
	pos     int
	readPos int
	cur     rune
	curSize int
	err     error
}

// New creates a new Lexer for an input string.
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()
	var tok token.Token
	switch l.cur {
	case '=':
		if l.peekChar() == '=' {
			curPos := l.pos
			l.readChar()
			tok = newToken(token.EQ, l.input[curPos:l.pos+l.curSize])

		} else {
			tok = newToken(token.ASSIGN, l.curLit())
		}
	case '+':
		tok = newToken(token.PLUS, l.curLit())
	case '-':
		tok = newToken(token.MINUS, l.curLit())
	case '!':
		if l.peekChar() == '=' {
			curPos := l.pos
			l.readChar()
			tok = newToken(token.NEQ, l.input[curPos:l.pos+l.curSize])

		} else {
			tok = newToken(token.BANG, l.curLit())
		}
	case '/':
		tok = newToken(token.SLASH, l.curLit())
	case '*':
		tok = newToken(token.STAR, l.curLit())
	case '<':
		tok = newToken(token.LT, l.curLit())
	case '>':
		tok = newToken(token.GT, l.curLit())
	case ';':
		tok = newToken(token.SEMICOLON, l.curLit())
	case ',':
		tok = newToken(token.COMMA, l.curLit())
	case '(':
		tok = newToken(token.LPAREN, l.curLit())
	case ')':
		tok = newToken(token.RPAREN, l.curLit())
	case '{':
		tok = newToken(token.LBRACE, l.curLit())
	case '}':
		tok = newToken(token.RBRACE, l.curLit())
	case '[':
		tok = newToken(token.LBRACKET, l.curLit())
	case ']':
		tok = newToken(token.RBRACKET, l.curLit())
	case ':':
		tok = newToken(token.COLON, l.curLit())
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.cur) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.cur) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.curLit())
		}
	}
	l.readChar()
	return tok
}

func (l *Lexer) readString() string {
	pos := l.pos + 1
	for {
		l.readChar()
		if l.cur == '"' || l.cur == 0 {
			break
		}
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) readIdentifier() string {
	pos := l.pos
	for isLetter(l.cur) {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) readNumber() string {
	pos := l.pos
	for isDigit(l.cur) {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) peekChar() rune {
	if l.readPos >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPos:])
	return r
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.cur = 0
		l.curSize = 0
	} else {
		l.cur, l.curSize = utf8.DecodeRuneInString(l.input[l.readPos:])
		if l.cur == utf8.RuneError {
			l.err = fmt.Errorf("failed to decode from utf8 at position %d", l.readPos)
			l.cur = 0
		}
	}
	l.pos = l.readPos
	l.readPos += l.curSize
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.cur) {
		l.readChar()
	}
}

func isLetter(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

func (l *Lexer) curLit() string {
	return l.input[l.pos : l.pos+l.curSize]
}

func newToken(tokenType token.TokenType, lit string) token.Token {
	return token.Token{Type: tokenType, Literal: lit}
}
