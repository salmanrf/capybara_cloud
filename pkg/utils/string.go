package utils

import (
	"regexp"
	"strings"
)

func Slugify(in string) string {
	pattern, _ := regexp.Compile("[^A-Za-z0-9 ]")
	cleaned := string(pattern.ReplaceAll([]byte(in), []byte{}))
	in = strings.ToLower(cleaned)
	parts := strings.Split(in, " ")
	partsn := len(parts)

	ret := strings.Builder{}

	for i := 0; i < partsn - 1; i++ {
		ret.Write([]byte(parts[i]))
		ret.Write([]byte("_"))
	}
	ret.Write([]byte(parts[partsn - 1]))

	return ret.String()
}