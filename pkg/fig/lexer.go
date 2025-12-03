package fig

import (
	"github.com/alecthomas/participle/v2/lexer"
)

// figLexer defines the lexicographic rules for parsing gofigure files.
// The order of rules matters - more specific patterns should come before
// more general ones (e.g., Float before Int, MLString before String).
var figLexer = lexer.MustSimple([]lexer.SimpleRule{
	// Comments
	{Name: "Comment", Pattern: `[#;][^\n]*`},
	{Name: "BlockComment", Pattern: `/\*[^*]*\*+(?:[^/*][^*]*\*+)*/`},

	// String literals
	{Name: "MLString", Pattern: `"""[^"]*"""|'''[^']*'''`},
	{Name: "String", Pattern: `"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'`},

	// Primitives
	{Name: "Boolean", Pattern: `\b(?:true|false)\b`},
	{Name: "Float", Pattern: `-?\d+\.\d+`},
	{Name: "Int", Pattern: `-?\d+`},

	// Section markers
	{Name: "SectionEnd", Pattern: `\[\]`},

	// Directives and operators
	{Name: "Include", Pattern: `%include`},
	{Name: "Expand", Pattern: `\.\.\.`},

	// Identifiers and punctuation
	{Name: "Ident", Pattern: `(?:\\.|[a-zA-Z_])[-a-zA-Z\d_]*`},
	{Name: "Punct", Pattern: `[][{}.,:%@]`},
	{Name: "Whitespace", Pattern: `\s+`},
})
