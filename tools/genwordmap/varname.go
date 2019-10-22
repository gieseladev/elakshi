package main

import (
	"github.com/jinzhu/inflection"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

func toDromedaryCase(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[size:]
}

func screamingSnakeToPascalCase(s string) string {
	parts := strings.Split(s, "_")

	for i, part := range parts {
		part = strings.ToLower(part)

		r, size := utf8.DecodeRuneInString(part)
		upperRune := unicode.ToUpper(r)
		if upperRune != r {
			part = string(upperRune) + part[size:]
		}

		parts[i] = part
	}

	return strings.Join(parts, "")
}

func nameFromFilename(filename string) string {
	name := filepath.Base(filename)
	if i := strings.LastIndexByte(name, '.'); i > -1 {
		name = name[:i]
	}

	return screamingSnakeToPascalCase(inflection.Singular(name))
}
