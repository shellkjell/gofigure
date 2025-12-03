package fig

import (
	"github.com/alecthomas/participle/v2/lexer"
)

// Config represents a complete gofigure configuration file.
// It contains a list of entries, where each entry can be:
// - A field (key-value pair)
// - A section (bracketed header with fields)
// - An include directive
type Config struct {
	Entries []*Entry `parser:"@@*"`
	Pos     lexer.Position
}

// Entry represents a single top-level entry in a configuration file.
type Entry struct {
	Include *Include `parser:"  @@"`
	Section *Section `parser:"| '[' @@ (SectionEnd|EOF)?"`
	Field   *Field   `parser:"| @@"`
	Pos     lexer.Position
}

// Include represents an include directive that imports other configuration files.
type Include struct {
	Includes []string `parser:"'%include' @String (','? @String)*"`
	Pos      lexer.Position
}

// Section represents a bracketed section header that provides context for fields.
type Section struct {
	Roots  []SectionRoot `parser:"@@+ ']'"`
	Fields []*Field      `parser:"@@*"`
	Pos    lexer.Position
}

// SectionRoot represents the root portion of a section path.
type SectionRoot struct {
	Identifier []string      `parser:"(@(Ident|'@') (',' ' '*|' ')? | '%' '{' (@Ident (',' ' '*|' ')?)* '}')"`
	Child      *SectionChild `parser:"@@?"`
	Pos        lexer.Position
}

// SectionChild represents child portions of a section path.
type SectionChild struct {
	Identifier []string      `parser:"'.' (@(Ident|'@'|Int) (',' ' '*|' ')? | '%' '{' (@Int @'...' @Int | (@(Ident|Int) (',' ' '*|' ')?)*) '}')"`
	Child      *SectionChild `parser:"@@?"`
	Pos        lexer.Position
}

// Field represents a key-value pair in the configuration.
type Field struct {
	Key   string      `parser:"( (@Ident|@String)"`
	Child *ChildField `parser:"  ( '.' @@"`
	Value *Value      `parser:"  | ':' @@ )?)"`

	// ArrayIndex is populated during post-processing, not during parsing.
	ArrayIndex *int64
	Pos        lexer.Position
}

// ChildField represents a nested field accessed via dot notation.
type ChildField struct {
	Key        string      `parser:"(( (@Ident|@String)"`
	ArrayIndex *int64      `parser:"|@Int)"`
	Child      *ChildField `parser:"( '.' @@"`
	Value      *Value      `parser:"| ':' @@ )?)"`
	Pos        lexer.Position
}

// UnprocessedString holds a multi-line string before processing.
type UnprocessedString struct {
	String *string `parser:"@MLString"`
	Pos    lexer.Position
}

// Bool is a custom boolean type that implements participle's Capture interface.
type Bool bool

// Capture implements the participle Capture interface for boolean parsing.
func (b *Bool) Capture(v []string) error {
	*b = v[0] == "true"
	return nil
}

// Value represents a configuration value of various types.
// Only one of the value fields will be non-nil at a time.
type Value struct {
	// Primitive types
	String          *string            `parser:"  @String"`
	MultilineString *UnprocessedString `parser:"| @@"`
	Integer         *int64             `parser:"| @Int"`
	Float           *float64           `parser:"| @Float"`
	Boolean         *Bool              `parser:"| (@'true' | @'false')"`

	// Composite types
	Map         []*Field `parser:"| '{' ((@@ ','?)*)? '}'"`
	ParsedArray []*Value `parser:"| '[' ((@@ ','?)*)? ']'"`

	// Reference to another value
	Identifier *string `parser:"| @Ident @('.' Ident)*"`

	// FinalArray is populated during post-processing for sequential numeric keys.
	FinalArray []*Value
	Pos        lexer.Position
}

// IsEmpty returns true if the Value has no value set.
func (v *Value) IsEmpty() bool {
	return v.String == nil &&
		v.MultilineString == nil &&
		v.Integer == nil &&
		v.Float == nil &&
		v.Boolean == nil &&
		v.Map == nil &&
		v.ParsedArray == nil &&
		v.Identifier == nil &&
		v.FinalArray == nil
}

// Position returns the lexer position for an entry (used for error reporting).
func (e *Entry) Position() lexer.Position { return e.Pos }

// Position returns the lexer position for a field (used for error reporting).
func (f *Field) Position() lexer.Position { return f.Pos }

// Position returns the lexer position for a value (used for error reporting).
func (v *Value) Position() lexer.Position { return v.Pos }
