package testgen

import "testing"

// Fixtures are keyed by tests/<method>/<test>.io, so a repeated name overwrites.
func TestUniqueTestNames(t *testing.T) {
	for _, method := range AllMethods {
		seen := make(map[string]bool)
		for _, test := range method.Tests {
			if seen[test.Name] {
				t.Errorf("duplicate test name %q in %s", test.Name, method.Name)
			}
			seen[test.Name] = true
		}
	}
}
