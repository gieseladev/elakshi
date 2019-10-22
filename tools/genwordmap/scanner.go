package main

import (
	"bufio"
	"io"
	"strings"
)

// ScanLines is the same bufio.ScanLines but strips comments from tokens.
func ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advance, token, err = bufio.ScanLines(data, atEOF)
	if err != nil || len(token) == 0 {
		return
	}

	if i := strings.IndexByte(string(token), '#'); i > -1 {
		token = token[:i]
	}

	return
}

// NewScanner creates a new bufio.Scanner for the given reader and assigns the
// ScanLines splitter to it.
func NewScanner(reader io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(reader)
	scanner.Split(ScanLines)

	return scanner
}
