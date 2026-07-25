package utils

import "testing"

func TestParseUint_Valid(t *testing.T) {
	v, err := ParseUint("42")
	if err != nil {
		t.Fatalf("ParseUint(42) should not error: %v", err)
	}
	if v != 42 {
		t.Errorf("ParseUint(42) = %d, want 42", v)
	}
}

func TestParseUint_Zero(t *testing.T) {
	v, err := ParseUint("0")
	if err != nil {
		t.Fatalf("ParseUint(0) should not error: %v", err)
	}
	if v != 0 {
		t.Errorf("ParseUint(0) = %d, want 0", v)
	}
}

func TestParseUint_Invalid(t *testing.T) {
	_, err := ParseUint("abc")
	if err == nil {
		t.Error("ParseUint(abc) should error")
	}
}

func TestParseUint_Negative(t *testing.T) {
	_, err := ParseUint("-1")
	if err == nil {
		t.Error("ParseUint(-1) should error")
	}
}

func TestParseUint_Overflow(t *testing.T) {
	_, err := ParseUint("999999999999999999999")
	if err == nil {
		t.Error("ParseUint(overflow) should error")
	}
}
