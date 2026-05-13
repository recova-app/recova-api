package auth

import "testing"

func TestExtractBearerToken_Valid(t *testing.T) {
	token, err := ExtractBearerToken("Bearer abc.def.ghi")
	if err != nil {
		t.Fatalf("extract bearer token: %v", err)
	}
	if token != "abc.def.ghi" {
		t.Fatalf("unexpected token: %s", token)
	}
}

func TestExtractBearerToken_Invalid(t *testing.T) {
	cases := []string{"", "Token abc", "Bearer   "}
	for _, value := range cases {
		if _, err := ExtractBearerToken(value); err == nil {
			t.Fatalf("expected invalid bearer header for %q", value)
		}
	}
}
