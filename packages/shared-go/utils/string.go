package utils

import (
	"regexp"
	"strings"
	"time"
)

func GetDockerRepoTagFromFullName(full string) string {
	parts := strings.Split(full, "/")
	size := len(parts)
	
	return parts[size - 1]
}

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

func DockerSafeDateString(t time.Time) string {
	datestr := t.Format(time.DateTime)

	pattern, _ := regexp.Compile("[ :]")
	cleaned := string(pattern.ReplaceAll([]byte(datestr), []byte("-")))
	
	return cleaned
}