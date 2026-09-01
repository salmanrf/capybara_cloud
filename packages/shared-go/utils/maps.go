package utils

import (
	"strings"
)

func MapFromEnvVarStrings(vars []string) map[string]any {
	pairs := [][]string{}

	for _, el := range vars {
		sepi := strings.Index(el, "=")
		if sepi < 0 {
			continue
		}
		if sepi >= len(el) - 1 {
			continue
		}
		key := el[:sepi]
		val := el[sepi + 1:]

		pairs = append(pairs, []string{key, val})
	}

	acc := make(map[string]any)
	for _, p := range pairs {
		acc[p[0]] = p[1]
	}

	return acc
}