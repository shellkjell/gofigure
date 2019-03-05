package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

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

		ActualExpected{
			data:     `[root1 root2] key:"value" [@.%{dev,prod,qa}] key:"value"`,
			expected: `{"root1":{"key":"value","dev":{"key":"value"},"prod":{"key":"value"},"qa":{"key":"value"}},"root2":{"key":"value","dev":{"key":"value"},"prod":{"key":"value"},"qa":{"key":"value"}}}`,
		},

		/* 	== config.special.txt ==
		# Wut does this mean??
		# special_case_root_key: This key will not be in the same level as dev and production keys

		# Empty value
		empty_value

		# Espace any special characters
		quote_value: \"
		array_value: \[
		*/
		// Maybe do something to ensure determinism here
		ActualExpected{
			data:     `Rick_Astley:"Never" \":"gonna" \[:"give" #include "files/config.special.txt"`,
			expected: `{"Rick_Astley":"Never","\\\"":"gonna","\\[":"give","quote_value":"gonna","array_value":"give"}`,
		},
	}

	for _, testCase := range testCases {
		config := &CONFIG{}

		parser.ParseString(testCase.data, config)

		mappedConfig := config.Transform()

		marshaled, _ := json.Marshal(mappedConfig)

		actual := map[string]interface{}{}
		expected := map[string]interface{}{}

		err := json.Unmarshal(marshaled, &actual)
		if err != nil {
			t.Errorf(err.Error())
		}

		err = json.Unmarshal([]byte(testCase.expected), &expected)
		if err != nil {
			t.Errorf(err.Error())
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("\nGot: %s\nExpected: %s", string(marshaled), testCase.expected)
		}
	}
}
