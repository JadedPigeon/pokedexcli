package main

import "testing"

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "hello world",
			expected: []string{"hello", "world"},
		},
		{
			input:    "THIS SENTENCE IS ALL CAPS",
			expected: []string{"this", "sentence", "is", "all", "caps"},
		},
		{
			input:    "  leading and trailing spaces  ",
			expected: []string{"leading", "and", "trailing", "spaces"},
		},
		{
			input:    "multiple   spaces",
			expected: []string{"multiple", "spaces"},
		},
		{
			input:    "special characters !@#$%^&*()",
			expected: []string{"special", "characters", "!@#$%^&*()"},
		},
		{
			input:    "123 456 789",
			expected: []string{"123", "456", "789"},
		},
		{
			input:    "",
			expected: []string{},
		},
		{
			input:    "   ",
			expected: []string{},
		},
		{
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			input:    "\thello\nworld\tagain ",
			expected: []string{"hello", "world", "again"},
		},
		{
			input:    "HeLLo WoRLd",
			expected: []string{"hello", "world"},
		},
	}
	for _, c := range cases {
		actual := cleanInput(c.input)
		// Check if the length of the actual slice matches the expected length
		if len(actual) != len(c.expected) {
			t.Errorf("For input %q, expected %d words, got %d", c.input, len(c.expected), len(actual))
			continue
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("For input %q, expected word %q at index %d, got %q", c.input, expectedWord, i, word)
			}
		}
	}

}
