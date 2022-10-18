package api

import (
	"fmt"
	"unicode"
)

type Action interface {
	Parse(*Scanner) (string, bool, error)
}

type Scanner struct {
	position int
	s        []rune
}

func NewScanner(s string) Scanner {
	return Scanner{-1, []rune(s)}
}

func (s *Scanner) Parse(actions ...Action) (result []string, err error) {
	var ret string
	var capture bool

	for _, action := range actions {
		ret, capture, err = action.Parse(s)

		if err != nil {
			return
		}

		if capture {
			result = append(result, ret)
		}
	}

	return
}

func (s *Scanner) PeekRune() rune {
	return s.s[s.position+1]
}

func (s *Scanner) Next() rune {
	s.position++
	return s.s[s.position]
}

func (s *Scanner) AtEnd() bool {
	return s.position+1 == len(s.s)
}

type ExpectRune struct {
	r rune
}

func (e ExpectRune) Parse(scanner *Scanner) (s string, capture bool, err error) {
	capture = false
	if scanner.AtEnd() || scanner.Next() != e.r {
		err = fmt.Errorf("expected to find a '%c' rune", e.r)
	}
	return
}

type CaptureUpToRune struct {
	r rune
}

func (u CaptureUpToRune) Parse(scanner *Scanner) (ret string, capture bool, err error) {
	capture = true
	p := scanner.position + 1

	for !scanner.AtEnd() {
		if scanner.Next() == u.r {
			// Found our boundary
			break
		}
	}

	ret = string(scanner.s[p:scanner.position])
	return
}

type ConsumeWhitespace struct {
}

func (c ConsumeWhitespace) Parse(scanner *Scanner) (ret string, capture bool, err error) {
	capture = false
	for !scanner.AtEnd() && unicode.IsSpace(scanner.PeekRune()) {
		scanner.Next()
	}
	return
}

type ExpectString struct {
	str string
}

func (e ExpectString) Parse(scanner *Scanner) (ret string, capture bool, err error) {
	capture = false

	for i := 0; i < len(e.str); i++ {
		if scanner.AtEnd() || scanner.Next() != rune(e.str[i]) {
			err = fmt.Errorf("expected to find the string \"%s\"", e.str)
			return
		}
	}

	return
}
