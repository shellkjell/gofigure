// Package fig provides a parser and transformer for gofigure configuration files.
//
// Gofigure is a configuration file format with the following features:
//   - Sections with bracket notation: [section.subsection]
//   - Key-value pairs: key: "value"
//   - Expansion macros: %{a,b,c} expands to multiple entries
//   - @ selector for applying values to all roots
//   - Include directives: %include "other.fig"
//   - Multiple value types: strings, integers, floats, booleans, arrays, maps
//
// Basic usage:
//
//	result, err := fig.Parse("config.fig")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// For more control, use the Parser and Transformer types directly:
//
//	parser, err := fig.NewParser()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	config, err := parser.ParseFile("config.fig")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	transformer := fig.NewTransformer(parser)
//	result, err := transformer.Transform(config)
package fig
