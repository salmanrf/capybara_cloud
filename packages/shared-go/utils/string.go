package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/salmanrf/capybara-cloud/packages/shared-go/database"
)

func FullImageRef(dep database.ApplicationDeployment) string {
	return fmt.Sprintf(
		"%s/%s/%s:%s",
		dep.ContainerRegistry.String,
		dep.ContainerNamespace.String,
		dep.ContainerRepository.String,
		dep.ContainerTag.String,
	)
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