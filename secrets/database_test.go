package secrets

import "testing"

func TestDatabase(t *testing.T) {
	d := NewDbSecretManager(nil)
	d.Delete("1dasd")
}
