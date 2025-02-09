package pkg

import "strings"

func NameToId(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), " ", "-")
}
