package main

import (
	"github.com/dave/jennifer/jen"
	"os"
)

func genToLower(name string) jen.Code {
	return jen.Id(name).Op("=").Qual("strings", "ToLower").Call(
		jen.Id(name),
	)
}

func genIndexLabelFunc(f *jen.File, parsed Parsed) {
	tokensID := parsed.TokensID()
	regExpsID := parsed.RegExpsID()

	// TODO Return all found substrings, but when there are multiple
	//  overlapping ones, only return the longest!
	// 		Or alternatively, iterate through them in descending order of length

	f.Func().Id("Index" + parsed.Name).Params(jen.Id("s").String()).Index().Int().
		BlockFunc(func(g *jen.Group) {
			g.Add(genToLower("s"))

			if tokensID != "" {
				g.For(jen.Id("token, _").Op(":=").Range().Id(tokensID)).
					BlockFunc(func(g *jen.Group) {
						g.If(
							jen.Id("i").Op(":=").Qual("strings", "Index").Call(
								jen.Id("s"), jen.Id("token"),
							),
							jen.Id("i").Op(">").Lit(-1),
						).Block(
							jen.Return().Index().Int().Values(
								jen.Id("i"),
								jen.Id("i").Op("+").Len(jen.Id("token")),
							),
						)
					})
			}

			if regExpsID != "" {
				g.For(jen.Id("_, re").Op(":=").Range().Id(regExpsID)).
					Block(
						jen.If(
							jen.Id("loc").Op(":=").Id("re").Dot("FindStringIndex").Call(jen.Id("s")),
							jen.Id("loc").Op("!=").Nil(),
						).Block(
							jen.Return(jen.Id("loc")),
						),
					)
			}

			g.Return(jen.Nil())
		})
}

func genIsLabelFunc(f *jen.File, parsed Parsed) {
	tokensID := parsed.TokensID()
	regExpsID := parsed.RegExpsID()

	f.Func().Id("Is" + parsed.Name).Params(jen.Id("s").String()).Bool().
		BlockFunc(func(g *jen.Group) {
			g.Add(genToLower("s"))

			if tokensID != "" {
				g.If(
					jen.Id("_, found").Op(":=").Id(tokensID).Index(jen.Id("s")),
					jen.Id("found"),
				).Block(
					jen.Return(jen.True()),
				)
			}

			if regExpsID != "" {
				g.For(jen.Id("_, re").Op(":=").Range().Id(regExpsID)).
					Block(
						jen.If(jen.Id("re").Dot("MatchString").Call(jen.Id("s"))).
							Block(
								jen.Return(jen.True()),
							),
					)
			}

			g.Return(jen.False())
		})
}

func genForParsed(f *jen.File, parsed Parsed) {
	if tokensID := parsed.TokensID(); tokensID != "" {
		tokenDict := jen.Dict{}
		for _, str := range parsed.Tokens {
			tokenDict[jen.Lit(str)] = jen.Values()
		}

		f.Var().Id(tokensID).
			Op("=").
			Map(jen.String()).Struct().
			Values(tokenDict)

	}

	if regExpsID := parsed.RegExpsID(); regExpsID != "" {
		regexps := make([]jen.Code, len(parsed.RegExps))
		for i, exp := range parsed.RegExps {
			regexps[i] = jen.Qual("regexp", "MustCompile").
				Call(jen.Lit(exp))
		}

		f.Var().Id(regExpsID).
			Op("=").
			Index().Op("*").Qual("regexp", "Regexp").
			Values(regexps...)
	}

	genIsLabelFunc(f, parsed)
	genIndexLabelFunc(f, parsed)
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

	genForParsed(f, parsed)
	return nil
}

func genCode(f *jen.File, filenames []string) error {
	for _, filename := range filenames {
		if err := genCodeFile(f, filename); err != nil {
			return err
		}
	}

	return nil
}
