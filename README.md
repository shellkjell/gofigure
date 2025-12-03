# gofigure

[![Go Reference](https://pkg.go.dev/badge/github.com/shellkjell/gofigure.svg)](https://pkg.go.dev/github.com/shellkjell/gofigure)
[![Go Report Card](https://goreportcard.com/badge/github.com/shellkjell/gofigure)](https://goreportcard.com/report/github.com/shellkjell/gofigure)

A configuration file parser that converts `.fig` files to JSON. Features include sections, expansion macros, includes, and reference resolution.

## Installation

### Binary

```bash
go install github.com/shellkjell/gofigure/cmd/gofigure@latest
```

### Library

```bash
go get github.com/shellkjell/gofigure/pkg/fig
```

## Usage

### CLI

```bash
# Output to stdout
gofigure -i config.fig

# Pretty-print with indentation
gofigure -i config.fig -indent

# Write to file
gofigure -i config.fig -o config.json

# Show version
gofigure -version
```

### Library

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/shellkjell/gofigure/pkg/fig"
)

func main() {
    // Simple parsing
    result, err := fig.ParseString(`key: "value"`)
    if err != nil {
        log.Fatal(err)
    }

    // Or parse a file
    result, err = fig.Parse("config.fig")
    if err != nil {
        log.Fatal(err)
    }

    // For more control, use Parser and Transformer directly
    parser, _ := fig.NewParser()
    config, _ := parser.ParseFile("config.fig")

    transformer := fig.NewTransformer(parser)
    mapped, _ := transformer.Transform(config)

    output, _ := json.MarshalIndent(mapped, "", "  ")
    fmt.Println(string(output))
}
```

## Configuration Format

### Basic Syntax

```fig
# Comments start with # or ;
value: "string"
integerKey: 42
floatKey: 4.2
boolKey: true

# Reference another value
keyFromIdentifier: value

# Maps
mapKey: {
  key: "string"
  nested: integerKey
}

# Arrays
arrayKey: ["value1", "value2", 42]
```

### Sections

Sections provide hierarchical structure:

```fig
[server]
  host: "localhost"
  port: 8080

[database.primary]
  host: "db.example.com"
  port: 5432
```

### Expansion Macros

Generate multiple similar entries:

```fig
# Creates dev, staging, prod sections
[%{dev,staging,prod}]
  enabled: true

# Apply to all roots with @
[@.logging]
  level: "info"
```

### Includes

Include other configuration files:

```fig
%include "base.fig", "overrides.fig"
```

## Project Structure

```
gofigure/
├── cmd/
│   └── gofigure/       # CLI application
├── pkg/
│   └── fig/            # Public library
│       ├── ast.go      # AST types
│       ├── lexer.go    # Lexer definition
│       ├── parser.go   # Parser implementation
│       └── transform.go # Config transformation
├── files/              # Example config files
└── Makefile
```

## Development

```bash
# Run tests
make test

# Run tests with verbose output
make test-verbose

# Build binary
make build

# Run linters
make lint

# Format code
make fmt

# Run all checks
make check
```

## Examples

See [EXAMPLES.md](EXAMPLES.md) for more detailed examples and the [files/](files/) directory for sample configuration files.

## Built With

* [participle](https://github.com/alecthomas/participle) - Parser library

## Contributing

Issues and PRs are welcome. Please submit a short description of your work.

## License

[GNU GPLv2 License](LICENSE)
