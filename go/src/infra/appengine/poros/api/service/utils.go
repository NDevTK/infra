package service

import (
	"strings"
)

func snakeToPascalCase(str string) string {
	words := strings.Split(str, "_")
	key := ""
	for _, word := range words {
		key += strings.Title(word)
	}
	return key
}
