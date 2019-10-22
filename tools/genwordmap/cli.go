package main

import (
	"flag"
	"fmt"
	"github.com/dave/jennifer/jen"
	"os"
	"strings"
	"time"
)

func main() {
	filename := flag.String("file", "labels.go", "name of the output file")
	packageName := flag.String("package", "", "package of the output")

	flag.Parse()

	if *packageName == "" {
		panic("package is a required argument")
	}

	f := jen.NewFile(*packageName)
	f.HeaderComment(fmt.Sprintf(
		"Code generated using tools/genwordmap on %s! DO NOT EDIT.",
		time.Now().UTC().Format(time.RFC822),
	))

	f.HeaderComment("Args: " + strings.Join(os.Args[1:], " "))

	if err := genCode(f, flag.Args()); err != nil {
		panic(err)
	}

	if err := f.Save(*filename); err != nil {
		panic(err)
	}
}
