package pkg

import (
	"os"
	"strings"
)

// ReadSecretFile reads a single value, such as a password, from a file. The trailing line break that editors and
// "echo" add is removed. Other whitespace is kept, as it may be part of the secret.
func ReadSecretFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(data), "\r\n"), nil
}
