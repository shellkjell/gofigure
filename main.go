package main

import (
	//"fmt"
	"os"
	"io/ioutil"
	"encoding/json"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"github.com/alecthomas/repr"
)

var iniLexer = lexer.Unquote(lexer.Must(lexer.Regexp(
	`(?m)` +
		`(\s+)` +
		`|(^[#;].*$)` +
		`|(?P<Ident>[a-zA-Z][a-zA-Z_\d]*(\.[a-zA-Z][a-zA-Z_\d]*)*)` +
		`|(?P<String>"(?:\\.|[^"])*")` +
		`|(?P<Float>\d+(?:\.\d+)?)` +
		`|(?P<Punct>[][:])` +
		`|(?P<MapStart>[][{])` +
		`|(?P<MapStart>[][}])`,
)))

type CONFIG struct {
	Properties []*Property `{ @@ }`
	Sections   []*Section  `{ @@ }`
}

type Section struct {
	Identifier string      `"[" @Ident "]"`
	Properties []*Property `{ @@ }`
}

type Property struct {
	Key   string `@Ident { @"." @Ident } ":"`
	Value *Value `@@`
}

type Value struct {
	String *string        `  @String`
	Number *float64       `| @Float`
	Identifier string     `| @Ident `
	Map []*Property       `| "{" { @@ [ "," ] } "}"`
	List []*Value         `| "[" { @@ [ "," ] } "]"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	// Initialize the root
	config := &CONFIG{}

	// Build the parser
	parser, err := participle.Build(&CONFIG{}, iniLexer)

	// Read file
	data, err := ioutil.ReadFile("config.txt")
	check(err)
	dataString := string(data)

	// Parse the config
	err = parser.ParseString(dataString, config)
	check(err)

	json.NewEncoder(os.Stdout).Encode(config)
	repr.Println(config, repr.Indent("  "), repr.OmitEmpty(true))
}