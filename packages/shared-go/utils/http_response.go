package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorDetails struct {
	ErrorCode int `json:"error_code"`
	Message string `json:"message"`
	Context map[string]any `json:"context"`
}

type BaseResponse[T any] struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data any `json:"data"`
	ErrorDetails *ErrorDetails `json:"error"`
}

var DefaultJsonError = `{"error": "Something went wrong"}`

func create_response[T any](data *T, message string) *BaseResponse[T] {
	return &BaseResponse[T]{
		true,
		message,
		data,
		nil,
	}
}

func write_json(w http.ResponseWriter, status int, body any) error {
	buf := bytes.Buffer{}
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err := w.Write(buf.Bytes())

	return err
}

func ResponseWithSuccess[T any](w http.ResponseWriter, status int, data *T, message string) error {
	response_body := create_response(data, message)

	if err := write_json(w, status, response_body); err != nil {
		ResponseWithError(w, http.StatusInternalServerError, nil, "")
		return err
	}

	return nil
}

func ResponseWithError(w http.ResponseWriter, status int, data map[string]any, message string) error {
	response_body := create_response(&data, message)
	response_body.Success = false
	response_body.ErrorDetails = &ErrorDetails{
		ErrorCode: status,
		Message: "Internal server error",
		Context: map[string]any{},
	}

	if message != "" {
		response_body.ErrorDetails.Message = message
	}

	if data != nil {
		response_body.ErrorDetails.Context = data
	} 

	if err := write_json(w, status, response_body); err != nil {
		fmt.Println("Error encoding error response")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(DefaultJsonError))

		return err
	}

	return nil
}