package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/dave/jennifer/jen"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const bidiBracketsURL = "https://unicode.org/Public/UCD/latest/ucd/BidiBrackets.txt"

var (
	outputFile = flag.String("out", "pairs.go", "output file")
	outPackage = flag.String("package", "bracket", "output package")
	pairsFile  = flag.String("pairs", bidiBracketsURL, "pairs file")
)

type BracketPair struct {
	Opening rune
	Closing rune
	Type    string
}

func parseUnicodeCodePoint(u string) (rune, error) {
	if u == "<none>" {
		return -1, nil
	}

	code, err := strconv.ParseInt(u, 16, 0)
	if err != nil {
		return 0, err
	}

	return rune(code), nil
}

var ErrLineEmpty = errors.New("empty line")

func parsePairLine(line string) (BracketPair, error) {
	commentIndex := strings.IndexRune(line, '#')
	if commentIndex != -1 {
		line = line[:commentIndex]
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return BracketPair{}, ErrLineEmpty
	}

	fields := strings.SplitN(line, "; ", 3)
	if len(fields) != 3 {
		return BracketPair{}, errors.New("invalid amount of fields")
	}

	opening, err := parseUnicodeCodePoint(fields[0])
	if err != nil {
		return BracketPair{}, err
	}
	closing, err := parseUnicodeCodePoint(fields[1])
	if err != nil {
		return BracketPair{}, err
	}

	return BracketPair{
		Opening: opening,
		Closing: closing,
		Type:    fields[2],
	}, nil
}

func getPairsFileReader() (io.ReadCloser, error) {
	path := *pairsFile
	if _, err := url.Parse(path); err == nil {
		log.Println("detected pairs file url")
		resp, err := http.Get(path)
		if err != nil {
			return nil, err
		}

		return resp.Body, nil
	}

	return os.Open(path)
}

func parsePairsFile() ([]BracketPair, error) {
	f, err := getPairsFileReader()
	if err != nil {
		return nil, err
	}

	defer func() { _ = f.Close() }()

	var pairs []BracketPair

	s := bufio.NewScanner(f)
	for s.Scan() {
		pair, err := parsePairLine(s.Text())
		if err == ErrLineEmpty {
			continue
		} else if err != nil {
			return nil, err
		}

		pairs = append(pairs, pair)
	}

	return pairs, s.Err()
}

func orderPairsByType(pairs []BracketPair) map[string][]BracketPair {
	byType := map[string][]BracketPair{}
	for _, pair := range pairs {
		byType[pair.Type] = append(byType[pair.Type], pair)
	}

	return byType
}

func addPairMaps(f *jen.File) error {
	pairs, err := parsePairsFile()
	if err != nil {
		return err
	}

	for name, pairs := range orderPairsByType(pairs) {
		mapValues := jen.Dict{}
		for _, pair := range pairs {
			mapValues[jen.LitRune(pair.Opening)] = jen.LitRune(pair.Closing)
		}

		f.Var().Id(fmt.Sprintf("%sPairs", name)).
			Op("=").
			Map(jen.Rune()).Rune().
			Values(mapValues)
	}

	return nil
}

func main() {
	flag.Parse()

	f := jen.NewFile(*outPackage)

	f.HeaderComment(fmt.Sprintf(
		"Code generated using tools/genbidi on %s! DO NOT EDIT.",
		time.Now().UTC().Format(time.Stamp),
	))

	if err := addPairMaps(f); err != nil {
		panic(err)
	}

	if err := f.Save(*outputFile); err != nil {
		panic(err)
	}
}
