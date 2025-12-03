package fig

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestParser(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}

	testCases := []struct {
		name     string
		data     string
		expected string
	}{
		{
			name: "section with expansion macro",
			data: `[test.%{dev,prod}]
key:"value"`,
			expected: `{"test":{"dev":{"key":"value"},"prod":{"key":"value"}}}`,
		},
		{
			name: "section with @ selector",
			data: `[test]
key:"value"
[@.key]
key:"value"`,
			expected: `{"test":{"key":{"key":"value"}}}`,
		},
		{
			name: "multiple roots with @ selector and expansion",
			data: `[%{root1,root2}]
key:"value1"
[@.%{dev,prod,qa}]
key:"value2"`,
			expected: `{"root1":{"key":"value1","dev":{"key":"value2"},"prod":{"key":"value2"},"qa":{"key":"value2"}},"root2":{"key":"value1","dev":{"key":"value2"},"prod":{"key":"value2"},"qa":{"key":"value2"}}}`,
		},
		{
			name:     "nested expansion macro",
			data:     `[root] key:"value" [@.%{dev,prod}.test.%{sub1,sub2}] key:"value"`,
			expected: `{"root":{"key":"value","dev":{"test":{"sub1":{"key":"value"},"sub2":{"key":"value"}}},"prod":{"test":{"sub1":{"key":"value"},"sub2":{"key":"value"}}}}}`,
		},
		{
			name:     "multiple roots with comma and space separators",
			data:     `[root, root2 root3] key:"value"`,
			expected: `{"root":{"key":"value"},"root2":{"key":"value"},"root3":{"key":"value"}}`,
		},
		{
			name:     "simple key value",
			data:     `key: "value"`,
			expected: `{"key":"value"}`,
		},
		{
			name:     "integer value",
			data:     `count: 42`,
			expected: `{"count":42}`,
		},
		{
			name:     "float value",
			data:     `price: 19.99`,
			expected: `{"price":19.99}`,
		},
		{
			name:     "boolean values",
			data:     `enabled: true disabled: false`,
			expected: `{"enabled":true,"disabled":false}`,
		},
		{
			name:     "array value",
			data:     `items: ["a", "b", "c"]`,
			expected: `{"items":["a","b","c"]}`,
		},
		{
			name:     "map value",
			data:     `config: { key: "value", num: 123 }`,
			expected: `{"config":{"key":"value","num":123}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := parser.ParseString(tc.data)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}

			transformer := NewTransformer(parser)
			mappedConfig, err := transformer.Transform(config)
			if err != nil {
				t.Fatalf("Transform() error = %v", err)
			}

			marshaled, err := json.Marshal(mappedConfig)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			var actual, expected map[string]interface{}
			if err := json.Unmarshal(marshaled, &actual); err != nil {
				t.Fatalf("json.Unmarshal(actual) error = %v", err)
			}
			if err := json.Unmarshal([]byte(tc.expected), &expected); err != nil {
				t.Fatalf("json.Unmarshal(expected) error = %v", err)
			}

			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("\nGot:      %s\nExpected: %s", string(marshaled), tc.expected)
			}
		})
	}
}

func TestParseStringConvenience(t *testing.T) {
	result, err := ParseString(`key: "value"`)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	val, ok := result["key"].(*string)
	if !ok {
		t.Fatalf("Expected key to be *string, got %T", result["key"])
	}
	if *val != "value" {
		t.Errorf("Expected key='value', got key=%q", *val)
	}
}

func TestNewParserError(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() should not error: %v", err)
	}
	if parser == nil {
		t.Fatal("NewParser() returned nil parser")
	}
}

func TestParserWorkingDirectory(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}

	parser.SetWorkingDirectory("/tmp")
	if parser.WorkingDirectory() != "/tmp" {
		t.Errorf("Expected working directory '/tmp', got %q", parser.WorkingDirectory())
	}
}

func TestParserClearCache(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}

	parser.ClearCache()
}
