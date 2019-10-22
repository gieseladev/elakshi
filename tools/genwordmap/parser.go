package main

import (
	"bufio"
	"strings"
)

type Parsed struct {
	Name string

	RegExps []string
	Tokens  []string
}

const (
	regexSurr = "/"
)

func Parse(scanner *bufio.Scanner) (Parsed, error) {
	var parsed Parsed

	for scanner.Scan() {
		token := strings.TrimSpace(scanner.Text())
		if token == "" {
			continue
		}

		switch {
		case len(token) > 1 && strings.HasPrefix(token, regexSurr) && strings.HasSuffix(token, regexSurr):
			parsed.RegExps = append(parsed.RegExps, token[len(regexSurr):len(token)-len(regexSurr)])
		default:
			parsed.Tokens = append(parsed.Tokens, token)
		}
	}

	return parsed, scanner.Err()
}
