package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseWithSuccess(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		data         *map[string]any
		message      string
		want_status  int
		want_message string
		want_success bool
	}{
		{
			"sets json content type, status and message",
			http.StatusCreated,
			&map[string]any{"id": "abc"},
			"Project created successfully",
			http.StatusCreated,
			"Project created successfully",
			true,
		},
		{
			"falls back to a 500 json error when data cannot be encoded",
			http.StatusOK,
			&map[string]any{"bad": make(chan int)},
			"ignored",
			http.StatusInternalServerError,
			"",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			ResponseWithSuccess(rec, tt.status, tt.data, tt.message)

			got_content_type := rec.Header().Get("Content-Type")
			if got_content_type != "application/json" {
				t.Errorf("got Content-Type %q, want %q", got_content_type, "application/json")
			}
			if rec.Code != tt.want_status {
				t.Errorf("got status %d, want %d", rec.Code, tt.want_status)
			}

			var got_body BaseResponse[any]
			if err := json.Unmarshal(rec.Body.Bytes(), &got_body); err != nil {
				t.Fatalf("body is not a single JSON document: %v, body: %s", err, rec.Body.String())
			}
			if got_body.Success != tt.want_success {
				t.Errorf("got success %v, want %v", got_body.Success, tt.want_success)
			}
			if got_body.Message != tt.want_message {
				t.Errorf("got message %q, want %q", got_body.Message, tt.want_message)
			}
		})
	}
}

func TestResponseWithError(t *testing.T) {
	rec := httptest.NewRecorder()

	ResponseWithError(rec, http.StatusBadRequest, nil, "Project with this name already exists")

	got_content_type := rec.Header().Get("Content-Type")
	if got_content_type != "application/json" {
		t.Errorf("got Content-Type %q, want %q", got_content_type, "application/json")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var got_body BaseResponse[any]
	if err := json.Unmarshal(rec.Body.Bytes(), &got_body); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if got_body.ErrorDetails == nil {
		t.Fatalf("got no error details")
	}
	if got_body.ErrorDetails.Message != "Project with this name already exists" {
		t.Errorf("got error message %q", got_body.ErrorDetails.Message)
	}
}
