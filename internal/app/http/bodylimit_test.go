package http

import "testing"

func TestParseBodyLimitBytes_Valid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "bytes", input: "128b", want: 128},
		{name: "kilobytes", input: "2kb", want: 2048},
		{name: "megabytes", input: "1mb", want: 1024 * 1024},
		{name: "gigabytes", input: "1gb", want: 1024 * 1024 * 1024},
		{name: "no unit", input: "512", want: 512},
		{name: "spaces", input: " 3mb ", want: 3 * 1024 * 1024},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseBodyLimitBytes(tc.input)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if got != tc.want {
				t.Fatalf("unexpected bytes: got=%d want=%d", got, tc.want)
			}
		})
	}
}

func TestParseBodyLimitBytes_Invalid(t *testing.T) {
	tests := []string{
		"",
		"0mb",
		"-1mb",
		"foo",
		"10tb",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			if _, err := parseBodyLimitBytes(input); err == nil {
				t.Fatalf("expected error for input %q", input)
			}
		})
	}
}
