package main

import (
	"io/ioutil"
	"strings"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
)

// A valid indentifier part is one of the following:
// 1. an escaped character, like \"
// 2. @
// 3. a string of characters
var re_valid_ident_part = `(\\.|@|[a-zA-Z_][a-zA-Z\d_]+)`

var PutinLexer = lexer.Must(lexer.Regexp(
	`(?m)` +
		`(\s+)` +
		`|(?P<Include>#include)` +
		`|([#;].*$)` +
		`|(?P<MLString>("""(?:\\.|[^(""")])*""")|('''(?:\\.|[^(''')])*'''))` +
		`|(?P<String>("(?:\\.|[^"])*")|('(?:\\.|[^'])*'))` +
		`|(?P<Ident>` + re_valid_ident_part + `)` +
		`|(?P<Float>-?\d+\.\d+)` +
		`|(?P<Int>-?\d+)` +
		`|(?P<SectionEnd>\[\])` +
		`|(?P<Punct>[][{},. :%])`,
))

type CONFIG struct {
	Entries []*Entry `(@@)*`

	Pos lexer.Position
}

type Entry struct {
	Include *Include `@@`
	Field   *Field   `| @@`
	Section *Section `| "[" @@ (SectionEnd|EOF)?`

	Pos lexer.Position
}

type Include struct {
	Includes []string `"#include" @String (","? @String)*`

	Pos lexer.Position
}

/*
	As SectionRoot and SectionChild must have different rules for how they are parsed,
	they have to be separate structres.
*/

type Section struct {
	Roots  []SectionRoot `(@@)+ "]"`
	Fields []*Field      `(@@)*`
}

type SectionRoot struct {
	Identifier []string      `(@Ident ("," " "*|" ")? | "%" "{" (@Ident ("," " "*|" ")?)* "}")`
	Child      *SectionChild `(@@)?`

	Pos lexer.Position
}

type SectionChild struct {
	Identifier []string      `"." (@Ident ("," " "*|" ")? | "%" "{" (@Ident ("," " "*|" ")?)* "}")`
	Child      *SectionChild `(@@)?`

	Pos lexer.Position
}

type Field struct {
	Key   string `@Ident `     // Key
	Child *Field `( "." @@`    // When a child field should be created this is where it goes
	Value *Value `| ":" @@ )?` // ? == allow empty values

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
	List            []*Value           `| "[" ((@@)*)? "]"`
	Map             []*Field           `| "{" ((@@)*)? "}"`
	Identifier      *string            `| @Ident @("." Ident)*`

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
