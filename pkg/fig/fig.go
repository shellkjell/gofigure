package fig

// Parse parses a configuration file and returns the transformed result as a map.
// This is a convenience function that creates a parser, parses the file,
// and transforms the result in one call.
func Parse(filename string) (map[string]any, error) {
	parser, err := NewParser()
	if err != nil {
		return nil, err
	}

	config, err := parser.ParseFile(filename)
	if err != nil {
		return nil, err
	}

	transformer := NewTransformer(parser)
	return transformer.Transform(config)
}

// ParseString parses a configuration string and returns the transformed result as a map.
// This is a convenience function that creates a parser, parses the string,
// and transforms the result in one call.
func ParseString(input string) (map[string]any, error) {
	parser, err := NewParser()
	if err != nil {
		return nil, err
	}

	config, err := parser.ParseString(input)
	if err != nil {
		return nil, err
	}

	transformer := NewTransformer(parser)
	return transformer.Transform(config)
}
