package utils

import (
	"os"
	"strings"
)

func EnsureDirExists(path string) error {
	_, err := os.ReadDir(path)
	if err != nil {
		if !strings.Contains(err.Error(), "no such") {
			return err
		}
	} else {
		return nil
	}

	var acc strings.Builder
	segs := strings.Split(path, "/")
	if strings.Index(path, "~") != 0 {
		acc.WriteString("/")
	}	

	for _, s := range segs {
		seg := strings.Trim(s, " ")
		if seg == "" {
			continue
		}
		
		acc.WriteString(seg + "/")
		
		_, err := os.ReadDir(acc.String())
		if err != nil {
			if !strings.Contains(err.Error(), "no such") {
				return err
			}
		} else {
			continue
		}
		
		err = os.Mkdir(acc.String(), 0o774)
		if err != nil {
			return err
		}
	}

	return nil
}