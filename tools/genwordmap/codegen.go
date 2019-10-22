package main

import (
	"github.com/dave/jennifer/jen"
	"os"
)

func genFuncForParsed(f *jen.File, parsed Parsed) error {
	tokenMapName := toDromedaryCase(parsed.Name) + "Tokens"
	hasTokenMap := false

	if len(parsed.Tokens) > 0 {
		tokenDict := jen.Dict{}
		for _, str := range parsed.Tokens {
			tokenDict[jen.Lit(str)] = jen.Values()
		}

		f.Var().Id(tokenMapName).
			Op("=").
			Map(jen.String()).Struct().
			Values(tokenDict)

		hasTokenMap = true
	}

	regExpsName := toDromedaryCase(parsed.Name) + "Matchers"
	hasRegExps := false

	if len(parsed.RegExps) > 0 {
		regexps := make([]jen.Code, len(parsed.RegExps))
		for i, exp := range parsed.RegExps {
			regexps[i] = jen.Qual("regexp", "MustCompile").
				Call(jen.Lit(exp))
		}

		f.Var().Id(regExpsName).
			Op("=").
			Index().Qual("regexp", "*Regexp").
			Values(regexps...)

		hasRegExps = true
	}

	f.Func().Id("Is" + parsed.Name).Params(jen.Id("s").String()).Bool().
		BlockFunc(func(g *jen.Group) {
			g.Id("s").Op("=").Qual("strings", "ToLower").Call(jen.Id("s"))

			if hasTokenMap {
				g.If(
					jen.Id("_, found").Op(":=").Id(tokenMapName).Index(jen.Id("s")),
					jen.Id("found"),
				).Block(
					jen.Return(jen.True()),
				)
			}

			if hasRegExps {
				g.For(jen.Id("_, re").Op(":=").Range().Id(regExpsName)).
					Block(
						jen.If(jen.Id("re").Dot("MatchString").Call(jen.Id("s"))).
							Block(
								jen.Return(jen.True()),
							),
					)
			}

			g.Return(jen.False())
		})
	return nil
}

func genCodeFile(f *jen.File, filename string) error {
	fp, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer func() { _ = fp.Close() }()

	parsed, err := Parse(NewScanner(fp))
	if err != nil {
		return err
	}

	if parsed.Name == "" {
		parsed.Name = nameFromFilename(filename)
	}

	return genFuncForParsed(f, parsed)
}

func genCode(f *jen.File, filenames []string) error {
	for _, filename := range filenames {
		if err := genCodeFile(f, filename); err != nil {
			return err
		}
	}

	return nil
}
