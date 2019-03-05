package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestJsonMarshalOfStructure(t *testing.T) {
	parser := BuildParser()
	config := &CONFIG{}

	parser.ParseString(`key:"value"`, config)

	config = config.splitAndAssociateChildren()

	mappedConfig := config.toMap()

	actual, _ := json.Marshal(mappedConfig)
	expected := []byte(`{"key":"value"}`)

	if !reflect.DeepEqual(actual, expected) { // Yes this compares array elements
		t.Errorf("Got: %s\tExpected: %s", string(actual), string(expected))
	}
}

type ActualExpected struct {
	data     string
	expected string
}

func TestJsonMarshalCases(t *testing.T) {
	parser := BuildParser()

	testCases := []ActualExpected{
		ActualExpected{
			data:     `[test.%{dev,prod}]` + "\n" + `key:"value"`,
			expected: `{"test":{"dev":{"key":"value"},"prod":{"key":"value"}}}`,
		},

		ActualExpected{
			data:     `[test] key:"value" [@.key] key:"value"`,
			expected: `{"test":{"key":{"key":"value"}}}`,
		},
	}

	for _, testCase := range testCases {
		config := &CONFIG{}
		parser.ParseString(testCase.data, config)
		config = config.splitAndAssociateChildren()
		mappedConfig := config.toMap()
		marshaled, _ := json.Marshal(mappedConfig)

		actual := map[string]interface{}{}
		expected := map[string]interface{}{}

		_ = json.Unmarshal(marshaled, actual)
		_ = json.Unmarshal([]byte(testCase.expected), expected)

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Got: %s\nExpected: %s", string(marshaled), testCase.expected)
		}
	}
}
