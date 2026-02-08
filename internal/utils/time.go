package utils

import "time"

// Now returns the current time.
// Can be mocked or replaced if needed for testing.
func Now() time.Time {
	return time.Now()
}
