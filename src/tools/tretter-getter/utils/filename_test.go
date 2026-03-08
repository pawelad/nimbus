package utils

import "testing"

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"odc. 34 – Wyznanie", "odc 34 – Wyznanie"},
		{"odc. 438 – Zaufaj intuicji!", "odc 438 – Zaufaj intuicji"},
		{"Test/With/Slashes", "Test With Slashes"},
		{"Test:With:Colons", "Test With Colons"},
		{"Test*With*Stars", "Test With Stars"},
		{"Test-With-Hyphens", "Test-With-Hyphens"},
		{"Test?With?Questions", "Test With Questions"},
		{"Test\"With\"Quotes", "Test With Quotes"},
		{"Test<With>Angles", "Test With Angles"},
		{"Test|With|Pipes", "Test With Pipes"},
		{"Normal Title", "Normal Title"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractEpisodeNumber(t *testing.T) {
	tests := []struct {
		title    string
		expected int
		wantErr  bool
	}{
		{"odc. 34 – Wyznanie", 34, false},
		{"odc 123 - Test", 123, false},
		{"Odc. 1 – Początek", 1, false},
		{"OdC.999 - Final", 999, false},
		{"odc.5-Short", 5, false},
		{"No episode here", 0, true},
		{"Polowanie", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			num, err := ExtractEpisodeNumber(tt.title)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for title %q, got nil", tt.title)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for title %q: %v", tt.title, err)
				return
			}

			if num != tt.expected {
				t.Errorf("for title %q: expected %d, got %d", tt.title, tt.expected, num)
			}
		})
	}
}

func TestFormatEpisodeDir(t *testing.T) {
	tests := []struct {
		episode  int
		expected string
	}{
		{1, "E.0001"},
		{34, "E.0034"},
		{123, "E.0123"},
		{1000, "E.1000"},
	}

	for _, tt := range tests {
		result := FormatEpisodeDir(tt.episode)
		if result != tt.expected {
			t.Errorf("FormatEpisodeDir(%d) = %q, want %q", tt.episode, result, tt.expected)
		}
	}
}
