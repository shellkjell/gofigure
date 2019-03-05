package main

import (
	"io/ioutil"
	"strings"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
)

var PutinLexer = lexer.Must(lexer.Regexp(
	`(?m)` +
		`(\s+)` +
		`|(?P<Include>#include)` +
		`|([#;].*$)` +
		`|(?P<MLString>("""(?:\\.|[^(""")])*""")|('''(?:\\.|[^(''')])*'''))` +
		`|(?P<String>("(?:\\.|[^"])*")|('(?:\\.|[^'])*'))` +
		`|(?P<Ident>(\\.|(@|[[:alpha:]_][[:alpha:]\d_]+)(\.(@|[[:alpha:]\d_]+))*\.?))` +
		`|(?P<Float>-?\d+\.\d+)` +
		`|(?P<Int>-?\d+)` +
		`|(?P<Punct>[][{},. :%])` +
		`|(?P<FileEnd>\z)`,
))

type CONFIG struct {
	Entries []*Entry `(@@)*`

	Pos lexer.Position
}

type Entry struct {
	Include *Include `@@`
	Field   *Field   `| @@`
	Section *Section `| @@`

	Pos lexer.Position
}

type Include struct {
	Includes []string `"#include" @String (","? @String)*`
	Parsed   []*CONFIG

	Pos lexer.Position
}

type Section struct {
	Identifier []string `"[" (@(Ident " ")+ | (@(Ident? ("%" "{" (","? Ident)* "}")? "."? Ident? ("%" "{" (","? Ident)* "}")?)!)* ) "]"`
	Fields     []*Field ` (@@)* ("[]"|FileEnd)?`

	Pos lexer.Position
}

type Field struct {
	Key   string `@Ident (":"`
	Value *Value `@@)?`

	Pos lexer.Position
}

type UnprocessedString struct {
	String *string `@MLString`
}

type Value struct {
	String          *string            `@String`
	MultilineString *UnprocessedString `| @@`
	Integer         *int64             `| @Int`
	Float           *float64           `| @Float`
	Identifier      *string            `| @Ident `
	List            []*Value           `| "[" ((@@)*)? "]"`
	Map             []*Field           `| "{" ((@@)*)? "}"`

	Reassigns bool

	Pos lexer.Position
}

func checkFileError(err error, filename string) {
	if err != nil {
		panic(strings.Replace(err.Error(), "<source>", filename, -1))
	}
}

func BuildParser() (parser *participle.Parser) {
	parser, err := participle.Build(
		&CONFIG{},
		participle.Lexer(PutinLexer),
		participle.Unquote("String"),
	)

	check(err)

	return
}

func ParseFile(filename string, parser *participle.Parser) (config *CONFIG) {
	config = &CONFIG{}
	if parser == nil {
		parser = BuildParser()
	}

	// Read file
	data, err := ioutil.ReadFile(filename)
	check(err)
	dataString := string(data)

	// Parse the config
	err = parser.ParseString(dataString, config)
	checkFileError(err, filename)

	return
}
