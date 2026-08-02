package csv_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"TypescriptToGolang/pkg/csv"
)

func toJSON(keys []string, row []string) map[string]string {
	object := make(map[string]string)
	for i, key := range keys {
		if i < len(row) {
			object[key] = row[i]
		}
	}
	return object
}

func TestExports(t *testing.T) {
	t.Run("libname", func(t *testing.T) {
		if csv.LibName != "@gregoranders/csv" {
			t.Errorf("Expected libname to be '@gregoranders/csv', got '%s'", csv.LibName)
		}
	})

	t.Run("libversion", func(t *testing.T) {
		if csv.LibVersion != "0.0.13" {
			t.Errorf("Expected libversion to be '0.0.13', got '%s'", csv.LibVersion)
		}
	})

	t.Run("liburl", func(t *testing.T) {
		if !strings.Contains(csv.LibURL, "https") {
			t.Errorf("Expected liburl to contain 'https', got '%s'", csv.LibURL)
		}
	})
}

func TestParse(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			text     string
			expected [][]string
		}{
			{text: "", expected: [][]string{}},
			{text: "a,b,c\n1,2,3", expected: [][]string{{"a", "b", "c"}, {"1", "2", "3"}}},
			{text: "a,\"b,\",c\n1,2,3", expected: [][]string{{"a", "b,", "c"}, {"1", "2", "3"}}},
			{text: "a,\"b\"\"\",c\n1,2,3", expected: [][]string{{"a", "b\"", "c"}, {"1", "2", "3"}}},
			{text: "a,\"b\n\",c\n1,2,3", expected: [][]string{{"a", "b\n", "c"}, {"1", "2", "3"}}},
			{text: "a,b,c\n1,2,3\n", expected: [][]string{{"a", "b", "c"}, {"1", "2", "3"}}},
			{text: "a,b,c\n1,2,", expected: [][]string{{"a", "b", "c"}, {"1", "2", ""}}},
			{text: "\"a\",\"b\",\"c\"\n\"1\",\"2\",\"\"", expected: [][]string{{"a", "b", "c"}, {"1", "2", ""}}},
		}

		for _, tc := range tests {
			name := fmt.Sprintf("%s = %v", strings.ReplaceAll(tc.text, "\n", "<NL>"), tc.expected)
			t.Run(name, func(t *testing.T) {
				parser := csv.NewParser(nil)
				parsed, err := parser.Parse(tc.text)

				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(parsed, tc.expected) {
					t.Errorf("Parse() = %v, want %v", parsed, tc.expected)
				}
				if !reflect.DeepEqual(parser.Rows(), tc.expected) {
					t.Errorf("Rows() = %v, want %v", parser.Rows(), tc.expected)
				}

				var expectedJSON []map[string]string
				if len(tc.expected) > 0 {
					expectedJSON = append(expectedJSON, toJSON(tc.expected[0], tc.expected[1]))
				} else {
					expectedJSON = []map[string]string{}
				}

				if !reflect.DeepEqual(parser.JSON(), expectedJSON) {
					t.Errorf("JSON() = %v, want %v", parser.JSON(), expectedJSON)
				}
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		tests := []struct {
			text     string
			expected [2]int
		}{
			{text: "a,\"b\"\",c\n1,2,3", expected: [2]int{0, 5}},
			{text: "a,\"b\",c\n1,\"2\"\",", expected: [2]int{1, 5}},
			{text: "a,\"b,c\n1,2,", expected: [2]int{0, 2}},			{text: "abc\"d", expected: [2]int{0, 3}},		}

		for _, tc := range tests {
			name := fmt.Sprintf("%s => ParseError(%d, %d)", strings.ReplaceAll(tc.text, "\n", "<NL>"), tc.expected[0], tc.expected[1])
			t.Run(name, func(t *testing.T) {
				parser := csv.NewParser(nil)
				_, err := parser.Parse(tc.text)

				expectedErr := fmt.Sprintf("Invalid CSV at %d:%d", tc.expected[0], tc.expected[1])
				if err == nil || err.Error() != expectedErr {
					t.Errorf("Expected error '%s', got '%v'", expectedErr, err)
				}
			})
		}
	})
}

func TestParseCustom(t *testing.T) {
	options := &csv.ParserOptions{
		FieldSeparator: ';',
		Quote:          '\'',
		LineSeparator:  '\t', // using single character per JS config
	}

	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			text     string
			expected [][]string
		}{
			{text: "", expected: [][]string{}},
			{text: "a;b;c\t1;2;3", expected: [][]string{{"a", "b", "c"}, {"1", "2", "3"}}},
			{text: "a;'b\n';c\t1;2;3", expected: [][]string{{"a", "b\n", "c"}, {"1", "2", "3"}}},
		}

		for _, tc := range tests {
			name := fmt.Sprintf("%s = %v", strings.ReplaceAll(tc.text, "\n", "<NL>"), tc.expected)
			t.Run(name, func(t *testing.T) {
				parser := csv.NewParser(options)
				parsed, err := parser.Parse(tc.text)

				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(parsed, tc.expected) {
					t.Errorf("Parse() = %v, want %v", parsed, tc.expected)
				}
				if !reflect.DeepEqual(parser.Rows(), tc.expected) {
					t.Errorf("Rows() = %v, want %v", parser.Rows(), tc.expected)
				}

				var expectedJSON []map[string]string
				if len(tc.expected) > 0 {
					expectedJSON = append(expectedJSON, toJSON(tc.expected[0], tc.expected[1]))
				} else {
					expectedJSON = []map[string]string{}
				}

				if !reflect.DeepEqual(parser.JSON(), expectedJSON) {
					t.Errorf("JSON() = %v, want %v", parser.JSON(), expectedJSON)
				}
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		tests := []struct {
			text     string
			expected [2]int
		}{
			{text: "a;\"b\"\";c\n1,2,3", expected: [2]int{0, 5}},
		}

		for _, tc := range tests {
			name := fmt.Sprintf("%s => ParseError(%d, %d)", strings.ReplaceAll(tc.text, "\n", "<NL>"), tc.expected[0], tc.expected[1])
			t.Run(name, func(t *testing.T) {
				// Only passing fieldSeparator per JS test suite behavior for this case
				parser := csv.NewParser(&csv.ParserOptions{FieldSeparator: ';'})
				_, err := parser.Parse(tc.text)

				expectedErr := fmt.Sprintf("Invalid CSV at %d:%d", tc.expected[0], tc.expected[1])
				if err == nil || err.Error() != expectedErr {
					t.Errorf("Expected error '%s', got '%v'", expectedErr, err)
				}
			})
		}
	})
}

func TestReturnedValueIsImmutable(t *testing.T) {
	// In Go, immutability is achieved by returning copies of internal slices/maps. 
	// We test this by mutating the returned value and ensuring the parser state remains untouched.

	t.Run("parse should return immutable value", func(t *testing.T) {
		parser := csv.NewParser(nil)
		rows, _ := parser.Parse("1,2,3\n,b,c")

		// Attempt to mutate the returned slice
		rows = append(rows, []string{"test"})
		rows[0] = append(rows[0], "test")
		rows[0][0] = "test"

		// Ensure internal state was unaffected
		expected := [][]string{{"1", "2", "3"}, {"", "b", "c"}}
		if !reflect.DeepEqual(parser.Rows(), expected) {
			t.Errorf("Internal state mutated! Expected %v, got %v", expected, parser.Rows())
		}
	})

	t.Run("rows should return immutable value", func(t *testing.T) {
		parser := csv.NewParser(nil)
		_, _ = parser.Parse("1,2,3\n,b,c")
		
		rows := parser.Rows()

		// Attempt to mutate the returned slice
		rows = append(rows, []string{"test"})
		rows[0] = append(rows[0], "test")
		rows[0][0] = "test"

		expected := [][]string{{"1", "2", "3"}, {"", "b", "c"}}
		if !reflect.DeepEqual(parser.Rows(), expected) {
			t.Errorf("Internal state mutated! Expected %v, got %v", expected, parser.Rows())
		}
	})

	t.Run("json should return immutable value", func(t *testing.T) {
		parser := csv.NewParser(nil)
		_, _ = parser.Parse("1,2,3\n,b,c")
		
		jsonRes := parser.JSON()

		// Attempt to mutate the returned slice and map
		jsonRes = append(jsonRes, map[string]string{"test": "test"})
		if len(jsonRes) > 0 {
			jsonRes[0]["1"] = "test"
		}

		expectedJSON := []map[string]string{{"1": "", "2": "b", "3": "c"}}
		if !reflect.DeepEqual(parser.JSON(), expectedJSON) {
			t.Errorf("Internal state mutated! Expected %v, got %v", expectedJSON, parser.JSON())
		}
	})
}