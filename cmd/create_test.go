package cmd

import "testing"

func TestParseParams_Empty(t *testing.T) {
	got, err := parseParams(nil)
	if err != nil {
		t.Fatalf("parseParams: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestParseParams_KeyValue(t *testing.T) {
	got, err := parseParams([]string{"TOKEN=abc", "REGION=eu-west"})
	if err != nil {
		t.Fatalf("parseParams: %v", err)
	}
	if got["TOKEN"] != "abc" || got["REGION"] != "eu-west" {
		t.Errorf("got %v", got)
	}
}

func TestParseParams_ValueContainsEquals(t *testing.T) {
	// Only the first '=' is the separator.
	got, err := parseParams([]string{"URL=https://x.example/?a=1&b=2"})
	if err != nil {
		t.Fatalf("parseParams: %v", err)
	}
	if got["URL"] != "https://x.example/?a=1&b=2" {
		t.Errorf("URL = %q", got["URL"])
	}
}

func TestParseParams_Malformed(t *testing.T) {
	cases := []string{"NO_EQUALS_SIGN", "=NO_KEY"}
	for _, c := range cases {
		if _, err := parseParams([]string{c}); err == nil {
			t.Errorf("expected error for %q", c)
		}
	}
}
