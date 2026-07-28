package mods

import "testing"

// Workshop IDs are interpolated into a launch argument, so anything that is not
// a plain decimal must be rejected before it can get near one. This is the
// guard for separator and shell-metacharacter injection into -mods=.
func TestWorkshopIDPatternRejectsNonNumeric(t *testing.T) {
	valid := []string{"1", "731604991", "12345678901234567890"}
	for _, id := range valid {
		if !workshopIDPattern.MatchString(id) {
			t.Errorf("expected %q to be accepted", id)
		}
	}

	invalid := []string{
		"",
		" 123",
		"123 ",
		"12,34",         // would inject an extra entry into -mods=
		"123;rm -rf /",  // shell metacharacters
		"123 -someflag", // would append an unrelated launch flag
		"$(id)",
		"`id`",
		"abc",
		"12.3",
		"-123",
		"123456789012345678901", // 21 digits, beyond any real publishedfileid
		"1\n2",
	}
	for _, id := range invalid {
		if workshopIDPattern.MatchString(id) {
			t.Errorf("expected %q to be rejected", id)
		}
	}
}
