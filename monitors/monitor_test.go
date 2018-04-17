package monitors

import "testing"

func TestMStatus(t *testing.T) {
	RunMStatus([]string{"stage", "production", "testing"})
}
