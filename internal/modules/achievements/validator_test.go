package achievements

import "testing"

func TestNormalizeCategoryQuery(t *testing.T) {
	tests := []struct {
		name      string
		input     *string
		wantNil   bool
		wantValue string
		wantErr   bool
	}{
		{name: "nil category", input: nil, wantNil: true},
		{name: "empty category", input: ptrString(" "), wantNil: true},
		{name: "valid category", input: ptrString("streak_milestone"), wantValue: "streak_milestone"},
		{name: "valid category mixed case", input: ptrString("Checkin_Consistency"), wantValue: "checkin_consistency"},
		{name: "invalid category", input: ptrString("invalid"), wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeCategoryQuery(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %q", *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected value %q, got nil", tc.wantValue)
			}
			if *got != tc.wantValue {
				t.Fatalf("expected value %q, got %q", tc.wantValue, *got)
			}
		})
	}
}

func ptrString(value string) *string {
	return &value
}
