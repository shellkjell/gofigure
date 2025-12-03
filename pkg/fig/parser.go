package fig

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alecthomas/participle/v2"
)

// Parser wraps the participle parser with gofigure-specific functionality.
// It maintains a file cache to avoid re-parsing the same file multiple times,
// which is especially useful when processing include directives.
type Parser struct {
	parser    *participle.Parser[Config]
	workDir   string
	fileCache map[string]*Config
}

// NewParser creates a new Parser instance configured for parsing gofigure files.
// Returns an error if the parser grammar cannot be built.
func NewParser() (*Parser, error) {
	p, err := participle.Build[Config](
		participle.Lexer(figLexer),
		participle.Unquote("String"),
		participle.Elide("Whitespace", "Comment"),
		participle.UseLookahead(2),
	)
	if err != nil {
		return nil, fmt.Errorf("building parser: %w", err)
	}

	return &Parser{
		parser:    p,
		fileCache: make(map[string]*Config),
	}, nil
}

// SetWorkingDirectory sets the base directory for resolving relative file paths.
func (p *Parser) SetWorkingDirectory(dir string) {
	p.workDir = dir
}

// WorkingDirectory returns the current working directory for include resolution.
func (p *Parser) WorkingDirectory() string {
	return p.workDir
}

// ParseFile parses a configuration file and returns the Config.
func (p *Parser) ParseFile(filename string) (*Config, error) {
	// Check cache first
	if cached, ok := p.fileCache[filename]; ok {
		return cached, nil
	}

	// Resolve path
	path := filename
	if p.workDir != "" && !filepath.IsAbs(filename) {
		path = filepath.Join(p.workDir, filename)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file %q: %w", path, err)
	}
	defer file.Close()

	config, err := p.parser.Parse(filename, file)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %w", filename, err)
	}

	p.fileCache[filename] = config
	return config, nil
}

// ParseReader parses configuration from an io.Reader.
func (p *Parser) ParseReader(name string, r io.Reader) (*Config, error) {
	config, err := p.parser.Parse(name, r)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %w", name, err)
	}
	return config, nil
}

// ParseString parses a configuration string and returns the Config.
func (p *Parser) ParseString(input string) (*Config, error) {
	return p.parser.ParseString("<string>", input)
}

// ClearCache clears the file cache, allowing files to be re-parsed.
// This is useful when files may have changed since they were first parsed.
func (p *Parser) ClearCache() {
	p.fileCache = make(map[string]*Config)
}
