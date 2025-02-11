package testhelper

import "testing"

// SkipIfShort is a test helper function that skips the test if it's running in short mode.
// It's typically used for integration tests that might take longer to execute.
//
// Parameters:
//   - t: testing.T pointer for test state and logging
//
// Example usage:
//
//	func TestIntegration(t *testing.T) {
//	    testhelper.SkipIfShort(t)
//	    // ... test code ...
//	}
//
// Example: Integratiion test command
// go test -v -run Integration -short ./...
func SkipIfShort(t *testing.T) {
	t.Helper()
	if !testing.Short() {
		t.Skip("Skipping integration test")
	}
}
