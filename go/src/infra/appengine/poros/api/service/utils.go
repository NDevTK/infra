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

func uniqueStrings(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func valueInSlice(value string, slice []string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
