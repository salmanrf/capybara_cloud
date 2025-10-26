package utils

import (
	"encoding/json"
)


func encode_json[T any](source *T, data *string) ([]byte, error) {
	result, err := json.Marshal(source)
	
	if err != nil {
		return nil, err
	}

	return result, nil
}

func decode_json[T any](source []byte, data *T) error {
	if err := json.Unmarshal(source, data); err != nil {
		return err
	}

	return nil
}