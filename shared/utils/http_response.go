package utils

import (
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

func ResponseWithSuccess[T any](w http.ResponseWriter, status int, data *T, message string) error {
	encoder := json.NewEncoder(w)
	
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")

	response_body := create_response(data, "Success")

	if err := encoder.Encode(response_body); err != nil {
		ResponseWithError(w, http.StatusInternalServerError, nil, "")
		
		return err
	}

	return nil
}

func ResponseWithError(w http.ResponseWriter, status int, data map[string]any, message string) error {
	encoder := json.NewEncoder(w)
	
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")

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

	if err := encoder.Encode(response_body); err != nil {
		fmt.Println("Error encoding error response")
		
		w.Write([]byte(DefaultJsonError))
		
		return err
	}

	return nil
}