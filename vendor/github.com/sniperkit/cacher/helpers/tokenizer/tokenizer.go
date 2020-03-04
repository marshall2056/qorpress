package tokenizer

import (
	"io"
)

type Tokenizer interface {
	GetTokens() []string
}

type EnglishTokenizer struct {
	input string
	// Current position in input (points to current char)
	position int
	// Current reading position in input (after current char)
	readPosition int
	// Current character being examined
	char byte
	// All the tokens that have been created
	tokens []string
}

// New creates a new EnglishTokenizer to be used
func NewEnglish(input string) *EnglishTokenizer {
	tokenizer := &EnglishTokenizer{input: input}
	tokenizer.readChar()
	return tokenizer
}

func (e *EnglishTokenizer) GetTokens() []string {
	var result []string

	if e.tokens != nil {
		copy(result, e.tokens)
	} else {
		token, err := e.nextToken()
		for err == nil {
			result = append(result, token)
			token, err = e.nextToken()
		}

		e.tokens = result
	}

	return result
}

func (e *EnglishTokenizer) readChar() {
	// Check if we are at the end of input
	if e.readPosition >= len(e.input) {
		e.char = 0
	} else {
		e.char = e.input[e.readPosition]
	}

	// Advance the position
	e.position = e.readPosition
	e.readPosition++
}

func (e *EnglishTokenizer) peekChar() byte {
	if e.readPosition >= len(e.input) {
		return 0
	}

	return e.input[e.readPosition]
}

func (e *EnglishTokenizer) nextToken() (string, error) {
	var tok string
	var err error

	e.skipNonLetters()

	switch e.char {
	case 0:
		tok = ""
		err = io.EOF
	default:
		tok = e.readWord()
	}

	e.readChar()

	return tok, err
}

func (e *EnglishTokenizer) readWord() string {
	position := e.position
	for isLetter(e.char) {
		e.readChar()
	}

	return e.input[position:e.position]
}

func (e *EnglishTokenizer) skipNonLetters() {
	for e.char != 0 && !isLetter(e.char) {
		e.readChar()
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}
