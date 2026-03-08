package utils

import "testing"

func TestCleanText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal text",
			input:    "odc. 434 – Pochopna decyzja",
			expected: "odc. 434 – Pochopna decyzja",
		},
		{
			name:     "Trailing tab (real TVP API data)",
			input:    "odc. 434 – Pochopna decyzja\t",
			expected: "odc. 434 – Pochopna decyzja",
		},
		{
			name:     "Multiple tabs",
			input:    "odc. 433\t-\tPozytywny pacjent\t",
			expected: "odc. 433 - Pozytywny pacjent",
		},
		{
			name:     "Leading tab",
			input:    "\todc. 123 - Title",
			expected: "odc. 123 - Title",
		},
		{
			name:     "Multiple spaces",
			input:    "Multiple   Spaces",
			expected: "Multiple Spaces",
		},
		{
			name:     "Leading and trailing whitespace",
			input:    "  Leading and trailing  ",
			expected: "Leading and trailing",
		},
		{
			name:     "Newlines",
			input:    "Line1\nLine2",
			expected: "Line1 Line2",
		},
		{
			name:     "Carriage returns",
			input:    "Text\rwith\rCR",
			expected: "Text with CR",
		},
		{
			name:     "Mixed whitespace",
			input:    "Mixed\t  \n  \rWhitespace",
			expected: "Mixed Whitespace",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only whitespace",
			input:    "  \t\n\r  ",
			expected: "",
		},
		{
			name:     "Double dash normalization (slug)",
			input:    "odc-434--pochopna-decyzja",
			expected: "odc-434-pochopna-decyzja",
		},
		{
			name:     "Multiple consecutive dashes",
			input:    "test---with----many-----dashes",
			expected: "test-with-many-dashes",
		},
		{
			name:     "Mixed tabs and double dashes",
			input:    "odc-434--pochopna-decyzja\t",
			expected: "odc-434-pochopna-decyzja",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanText(tt.input)
			if result != tt.expected {
				t.Errorf("CleanText(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
