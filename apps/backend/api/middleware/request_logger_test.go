package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/salmanrf/capybara-cloud/apps/backend/pkg/logger"
	"github.com/salmanrf/capybara-cloud/apps/backend/tests"
)

func TestLoggingMiddleware(t *testing.T) {
	t.Run("should log with context for each incoming requests", func (t *testing.T) {
		mux := http.NewServeMux()

		mux.Handle("GET /api/tests", http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))
		mux.Handle("POST /api/tests", http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("OK"))
		}))
		mux.Handle("PUT /api/abcd", http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("OK"))
		}))
		mux.Handle("PATCH /api/masbro", http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("OK"))
		}))
		mux.Handle("DELETE /api/applications", http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte("OK"))
		}))

		pwd, err := os.Getwd()
		if err != nil {
			t.Fatal("got unexpected error Getwd", err)
		}

		requests := []struct{
			method string
			path   string
			status int
		}{
			{http.MethodGet, "/api/tests", http.StatusOK},
			{http.MethodPost, "/api/tests", http.StatusNotFound},
			{http.MethodPut, "/api/abcd", http.StatusForbidden},
			{http.MethodPatch, "/api/masbro", http.StatusUnprocessableEntity},
			{http.MethodDelete, "/api/applications", http.StatusNoContent},
		}

		want_log_path := path.Join(pwd, "apps.backend.logs")

		futils := tests.File_utils_stub{}
		futils.Open_fn = func (name string, flag int, perm os.FileMode) (*os.File, error) {
			return os.OpenFile(name, flag, perm)
		}

		slogger, log_cleanup, err := logger.InitLogger(want_log_path, &futils)
		if err != nil {
			t.Fatal("got unexpected error initializing logger", err)
		}
		middleware := CreateLoggingMiddleware(slogger)

		api := middleware(mux)

		for _, r := range requests {
			res := httptest.NewRecorder()
			req, err := http.NewRequest(r.method, r.path, nil)
			if err != nil {
				t.Fatal("got unexpected error setting up request")
			}

			api.ServeHTTP(res, req)
		}

		// ? Flush buffered logs and close writer
		log_cleanup()

		got_log_file, err := os.OpenFile(want_log_path, os.O_RDONLY, 0o644)
		if err != nil {
			t.Fatal("got log file not created")
		}
		defer os.Remove(want_log_path)

		got_lines, err := tests.ReadFileLines(got_log_file)
		if err != nil {
			t.Fatal("got unexpected error setting up log file", err)
		}

		got_len := len(got_lines)
		want_len := len(requests)
		if got_len != want_len {
			t.Errorf("got %d incoming req log lines, want %d", got_len, want_len)
		}

		for i, got_line := range got_lines {
			var got_log_entry logger.LogEntry
			got_parse_err := json.Unmarshal([]byte(got_line), &got_log_entry)
			if got_parse_err != nil {
				t.Fatal("got error parsing log line into json", got_parse_err)
			}

			got_msg := got_log_entry.Msg
			want_msg := "Served request"
			if got_msg != want_msg {
				t.Errorf("got log msg '%s', want %s", got_msg, want_msg)
			}

			got_method := got_log_entry.Method
			want_method := requests[i].method
			if got_method != want_method {
				t.Errorf("got log method '%s', want %s", got_method, want_method)
			}

			got_status_code := got_log_entry.Status
			want_status_code := requests[i].status
			if got_status_code != want_status_code {
				t.Errorf("got log status_code %d, want %d", got_status_code, want_status_code)
			}
		}
	})
}