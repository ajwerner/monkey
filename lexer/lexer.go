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
	cur token.Token
	err error

	state
}

// New creates a new Lexer for an input string.
func New(input string) *Lexer {
	var l Lexer
	initState(&l.state, input)
	return &l
}

var zeroToken = token.Token{}

func (l *Lexer) Next() bool {
	if l.err == nil && l.cur.Type != token.EOF {
		l.cur, l.err = lexNext(&l.state)
	}
	return l.err == nil
}

func (l *Lexer) Token() token.Token {
	return l.cur
}

func (l *Lexer) Err() error {
	return l.err
}

////////////////////////////////////////////////////////////////////////////////
// lexFuncs
////////////////////////////////////////////////////////////////////////////////

type lexFunc func(l *state) (cur token.Token, err error)

func lexNext(s *state) (token.Token, error) {
	next, err := s.skipWhitespace()
	if err != nil {
		return token.Token{}, err
	}
	if f := lexFuncs[next]; f != nil {
		return f(s)
	}
	return lexDefault(s)
}

func lexDefault(s *state) (token.Token, error) {
	next, _ := s.peek()
	if isLetter(next) {
		return lexIdentifier(s)
	}
	if next == '.' || isDecimal(next) {
		return lexNumber(s)
	}
	return token.Token{}, fmt.Errorf("Illegal token %q at position %d", s.curLit(), s.tokPos)
}

func nextTok(typ token.TokenType) lexFunc {
	return eat(litTok(typ))
}

func litTok(typ token.TokenType) lexFunc {
	return func(s *state) (token.Token, error) {
		return newToken(typ, s.curLit()), nil
	}
}

func eat(f lexFunc) lexFunc {
	return func(s *state) (token.Token, error) {
		s.readRune()
		return f(s)
	}
}

var (
	assign = litTok(token.ASSIGN)
	bang   = litTok(token.BANG)
	eq     = nextTok(token.EQ)
	neq    = nextTok(token.NEQ)
)

var lexFuncs = map[rune]lexFunc{
	'+': nextTok(token.PLUS),
	'-': nextTok(token.MINUS),
	'/': nextTok(token.SLASH),
	'*': nextTok(token.STAR),
	'<': nextTok(token.LT),
	'>': nextTok(token.GT),
	';': nextTok(token.SEMICOLON),
	',': nextTok(token.COMMA),
	'(': nextTok(token.LPAREN),
	')': nextTok(token.RPAREN),
	'{': nextTok(token.LBRACE),
	'}': nextTok(token.RBRACE),
	'[': nextTok(token.LBRACKET),
	']': nextTok(token.RBRACKET),
	':': nextTok(token.COLON),
	'"': lexString,
	0: func(s *state) (token.Token, error) {
		return token.Token{Type: token.EOF}, nil
	},
	'=': func(s *state) (token.Token, error) {
		next, err := s.readRune()
		if err != nil {
			return token.Token{}, err
		}
		if next == '=' {
			return eq(s)
		}
		return assign(s)
	},
	'!': func(s *state) (token.Token, error) {
		next, err := s.readRune()
		if err != nil {
			return token.Token{}, err
		}
		if next == '=' {
			return neq(s)
		}
		return bang(s)
	},
}

func lexNumber(s *state) (token.Token, error) {
	next, err := s.readDecimals()
	typ := token.INT
	if next == '.' {
		typ = token.FLOAT
		if next, err = s.readRune(); err != nil {
			return token.Token{}, err
		}
		if !isDecimal(next) {
			return token.Token{}, fmt.Errorf("Illegal character '%c' after . at position %d", next, s.readPos)
		}
		if next, err = s.readDecimals(); err != nil {
			return token.Token{}, err
		}
	}
	if next == 'e' || next == 'E' {
		typ = token.FLOAT
		if next, err = s.readRune(); err != nil {
			return token.Token{}, err
		}
		if next == '+' || next == '-' {
			if next, err = s.readRune(); err != nil {
				return token.Token{}, err
			}
		}
		if _, err = s.readDecimals(); err != nil {
			return token.Token{}, err
		}
	}
	return token.Token{
		Type:    typ,
		Literal: s.curLit(),
	}, nil
}

func lexString(s *state) (token.Token, error) {
	for {
		next, err := s.readRune()
		if err != nil {
			return token.Token{}, err
		}
		if next == 0 {
			return token.Token{}, fmt.Errorf("unterminated string at position %d:%d", s.tokPos, s.readPos)
		}
		if next == '"' {
			if _, err = s.readRune(); err != nil {
				return token.Token{}, err
			}
			break
		}
	}
	return token.Token{
		Type:    token.STRING,
		Literal: s.input[s.tokPos+1 : s.runePos],
	}, nil
}

func lexIdentifier(s *state) (token.Token, error) {
	next, err := s.peek()
	for err == nil && isLetter(next) {
		next, err = s.readRune()
	}
	if err != nil {
		return token.Token{}, err
	}
	return token.Token{
		Type:    token.LookupIdent(s.curLit()),
		Literal: s.curLit(),
	}, nil
}

////////////////////////////////////////////////////////////////////////////////
// state
////////////////////////////////////////////////////////////////////////////////

type state struct {
	input  string
	tokPos int

	rune     rune
	runeSize int
	runePos  int

	peeked   bool
	peekRune rune
	peekSize int

	readPos int
}

func initState(s *state, input string) {
	*s = state{
		input: input,
	}
}

func (s *state) skipWhitespace() (next rune, err error) {
	next, err = s.readWhitespace()
	if err == nil {
		s.reset()
	}
	return next, err
}

func (s *state) reset() {
	s.tokPos = s.readPos
	s.runePos = s.readPos
	s.runeSize = 0
	s.rune = 0
}

func (s *state) readWhitespace() (next rune, err error) {
	next, err = s.peek()
	for err == nil && unicode.IsSpace(next) {
		next, err = s.readRune()
	}
	return next, err
}

func (l *state) readDecimals() (next rune, err error) {
	next, err = l.peek()
	for err == nil && isDecimal(next) {
		next, err = l.readRune()
	}
	return
}

func (s *state) peek() (rune, error) {
	if s.peeked {
		return s.peekRune, nil
	}
	if s.readPos >= len(s.input) {
		return 0, nil
	}
	s.peekRune, s.peekSize = utf8.DecodeRuneInString(s.input[s.readPos:])
	s.peeked = true
	if s.peekRune == utf8.RuneError {
		return utf8.RuneError, fmt.Errorf("failed to decode from utf8 at position %d", s.readPos)
	}
	return s.peekRune, nil
}

// readRune reads consumes the next rune as part of the current token and
// returns the next rune plus any error.
func (s *state) readRune() (rune, error) {
	if p, err := s.peek(); err != nil {
		return p, err
	}
	s.rune = s.peekRune
	s.runePos = s.readPos
	s.readPos += s.peekSize
	s.runeSize = s.peekSize
	s.peeked = false
	s.peekRune = 0
	s.peekSize = 0
	return s.peek()
}

func (s *state) curLit() string {
	return s.input[s.tokPos:s.readPos]
}

////////////////////////////////////////////////////////////////////////////////
// helpers
////////////////////////////////////////////////////////////////////////////////

func isLetter(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isDecimal(r rune) bool {
	return unicode.IsDigit(r)
}

func newToken(tokenType token.TokenType, lit string) token.Token {
	return token.Token{Type: tokenType, Literal: lit}
}
