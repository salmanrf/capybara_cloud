package logger

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"testing"

	pkgerr "github.com/pkg/errors"

	"github.com/salmanrf/capybara-cloud/apps/backend/tests"
	errutils "github.com/salmanrf/capybara-cloud/packages/shared-go/errors"
)

func setup(logfilename string, t *testing.T) (*os.File, func ()) {
	pwd, _ := os.Getwd() 
	log_path := path.Join(pwd, logfilename)
	log_file, err := os.OpenFile(log_path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		t.Fatal("got unexpected error setting up log file", err)
	}

	stderr := os.Stderr
	os.Stderr = log_file

	cleanup := func () {
		log_file.Close()
		os.Stderr = stderr
		os.Remove(log_path)
	}

	return log_file, cleanup
}

func TestLogger(t *testing.T) {
	futils := tests.File_utils_stub{}
	futils.Open_return = &os.File{}
	futils.Open_error = nil

	t.Run("should create structured logger that writes json lines to stderr", func (t *testing.T) {
		log_file, cleanup := setup("test.logs.stderr", t)

		defer func () {
			futils.Clear()
			cleanup()
		}()

		logger, _, err := InitLogger("", &futils)	
		if err != nil {
			t.Fatal("got unexpected error initializing logger instance", err)
		}

		logger.Debug("connected to the database!")
		logger.Debug("started listening!")
		logger.Debug("service is shutting down...")

		got_lines, err := tests.ReadFileLines(log_file)
		if err != nil {
			t.Fatal("got unexpected error setting up log file", err)
		}

		got_len := len(got_lines)
		want_len := 3
		if got_len != want_len {
			t.Fatalf("got %d log entries, want %d", got_len, want_len)
		}

		for _, got_line := range got_lines {
			var got_json_log map[string]any
			got_parse_err := json.Unmarshal([]byte(got_line), &got_json_log)
			if got_parse_err != nil {
				t.Fatal("got error parsing log line into json", got_parse_err)
			}
		}
	})

	t.Run("should have consistent structures with metada in each log entry", func (t *testing.T) {
		log_file, cleanup := setup("test.logs.structure", t)

		defer func () {
			futils.Clear()
			cleanup()
		}()
		
		logger, _, err := InitLogger("", &futils)
		if err != nil {
			t.Fatal("got unexpected error initializing logger instance", err)
		}

		logger.Debug("connected to the database!")
		logger.Debug("started listening!")
		logger.Info("received deployment request")
		logger.Debug("service is shutting down...")

		got_lines, err := tests.ReadFileLines(log_file)
		if err != nil {
			t.Fatal("got unexpected error setting up log file", err)
		}

		got_len := len(got_lines)
		want_len := 4
		if got_len != want_len {
			t.Fatalf("got %d log entries, want %d", got_len, want_len)
		}

		for _, got_line := range got_lines {
			var got_json_log LogEntry
			got_parse_err := json.Unmarshal([]byte(got_line), &got_json_log)
			if got_parse_err != nil {
				t.Fatal("got error parsing log line into json", got_parse_err)
			}

			got_level := got_json_log.Level
			if got_level == "" {
				t.Error("got log level empty string, want non-empty", got_line)
			}

			got_env := got_json_log.Env
			if got_env == "" {
				t.Error("got log env empty string, want non-empty", got_line)
			}

			got_hostname := got_json_log.Hostname
			if got_hostname == "" {
				t.Error("got log hostname empty string, want non-empty", got_line)
			}

			got_time := got_json_log.Time
			if got_time == "" {
				t.Error("got log time empty string, want non-empty", got_line)
			}

			got_msg := got_json_log.Msg
			if got_msg == "" {
				t.Error("got log msg empty string, want non-empty", got_line)
			}
		}
	})

	t.Run("should return a closer function that flush all buffered logs", func (t *testing.T) {
		_, cleanup := setup("test.logs.stderr_2", t)

		defer func () {
			futils.Clear()
			cleanup()
		}()

		futils.Open_fn = func (name string, flag int, perm os.FileMode) (*os.File, error) {
			return os.OpenFile(name, flag, perm)
		}

		// ? This is the actual designated log file, not intercepted stderr as with the above
		pwd, err := os.Getwd()
		if err != nil {
			t.Fatal("got unexpected error Getwd", err)
		}
		want_log_path := path.Join(pwd, "apps.backend.logs")

		logger, log_cleanup, err := InitLogger(want_log_path, &futils)	
		if err != nil {
			t.Fatal("got unexpected error initializing logger instance", err)
		}

		got_open_called := futils.Open_call 
		want_open_called := 1
		if got_open_called != want_open_called {
			t.Fatalf("got file open called %d time, want %d", got_open_called, want_open_called)
		}
		got_open_err := futils.Open_error 
		if got_open_err != nil {
			t.Fatal("got unexpected error creating log file", got_open_err)
		}

		logger.Info("received deployment request")

		log_cleanup()

		got_log_file, err := os.OpenFile(want_log_path, os.O_RDONLY, 0o644)
		if err != nil {
			t.Fatal("got log file not created")
		}
		defer os.Remove(want_log_path)

		got_logfile_lines, err := tests.ReadFileLines(got_log_file)
		if err != nil {
			t.Fatal("got unexpected error setting up log file", err)
		}

		got_logfile_len := len(got_logfile_lines)
		want_logfile_len := 1

		if got_logfile_len != want_logfile_len {
			t.Errorf("got %d log file entries (>= INFO only), want %d", got_logfile_len, want_logfile_len)
		}
	})

	t.Run("should write to both stderr and a log file when specified", func (t *testing.T) {
		stderr_file, cleanup := setup("test.logs.stderr_3", t)

		defer func () {
			futils.Clear()
			cleanup()
		}()

		futils.Open_fn = func (name string, flag int, perm os.FileMode) (*os.File, error) {
			return os.OpenFile(name, flag, perm)
		}

		// ? This is the actual designated log file, not intercepted stderr as with the above
		pwd, err := os.Getwd()
		if err != nil {
			t.Fatal("got unexpected error Getwd", err)
		}
		want_log_path := path.Join(pwd, "apps.backend.logs")

		logger, log_cleanup, err := InitLogger(want_log_path, &futils)	
		if err != nil {
			t.Fatal("got unexpected error initializing logger instance", err)
		}

		got_open_called := futils.Open_call 
		want_open_called := 1
		if got_open_called != want_open_called {
			t.Fatalf("got file open called %d time, want %d", got_open_called, want_open_called)
		}
		got_open_err := futils.Open_error 
		if got_open_err != nil {
			t.Fatal("got unexpected error creating log file", got_open_err)
		}

		logger.Debug("connected to the database!")
		logger.Debug("started listening!")
		logger.Info("received deployment request")
		logger.Debug("service is shutting down...")

		log_cleanup()

		got_log_file, err := os.OpenFile(want_log_path, os.O_RDONLY, 0o644)
		if err != nil {
			t.Fatal("got log file not created")
		}
		defer os.Remove(want_log_path)

		got_stderr_lines, err := tests.ReadFileLines(stderr_file)
		if err != nil {
			t.Fatal("got unexpected error setting up log file", err)
		}

		got_logfile_lines, err := tests.ReadFileLines(got_log_file)
		if err != nil {
			t.Fatal("got unexpected error setting up log file", err)
		}

		got_stderr_len := len(got_stderr_lines)
		want_stderr_len := 4
		if got_stderr_len != want_stderr_len  {
			t.Errorf("got %d stderr lines, want %d", got_stderr_len, want_stderr_len)
		}

		got_logfile_len := len(got_logfile_lines)
		want_logfile_len := 1

		if got_logfile_len != want_logfile_len {
			t.Errorf("got %d log file entries (>= INFO only), want %d", got_logfile_len, want_logfile_len)
		}
	})
}

func TestErrorLogger(t *testing.T) {
	futils := tests.File_utils_stub{}
	futils.Open_return = &os.File{}
	futils.Open_error = nil

	t.Run("should log error with nested structure if error is provided", func (t *testing.T) {
		stderr_file, cleanup := setup("test.logs.errors", t)

		defer func () {
			futils.Clear()
			cleanup()
		}()
		
		logger, _, err := InitLogger("", &futils)
		if err != nil {
			t.Fatal("got unexpected error initializing logger instance", err)
		}

		logger.Error("abcd", "error", errors.New("I'm a teapot"))

		got_stderr_lines, err := tests.ReadFileLines(stderr_file)
		if err != nil {
			t.Fatal("got unexpected error setting up log file", err)
		}

		got_stderr_len := len(got_stderr_lines)
		want_stderr_len := 1
		if got_stderr_len != want_stderr_len  {
			t.Errorf("got %d stderr lines, want %d", got_stderr_len, want_stderr_len)
		}
	})

	t.Run("should log error with arbitrary fields if error is ErrWithAttrs", func (t *testing.T) {
		stderr_file, cleanup := setup("test.logs.errors", t)

		defer func () {
			futils.Clear()
			cleanup()
		}()
		
		logger, _, err := InitLogger("", &futils)
		if err != nil {
			t.Fatal("got unexpected error initializing logger instance", err)
		}

		want_msg := "Unable to run application instance"
		want_err := errors.New("unable to find docker executable")
		want_structured_err := errutils.WithAttrs(
			want_err,
			"version",
			"1.0.0",
			"deployment_id",
			"abcd",
			"role",
			"contributor",
		)

		logger.Error(want_msg, "error", want_structured_err)

		got_stderr_lines, err := tests.ReadFileLines(stderr_file)
		if err != nil {
			t.Fatal("got unexpected error setting up log file", err)
		}

		var got_json_log LogEntry
		got_line := got_stderr_lines[0]

		got_parse_err := json.Unmarshal([]byte(got_line), &got_json_log)
		if got_parse_err != nil {
			t.Fatal("got error parsing log line into json", got_parse_err)
		}

		got_error_keys := len(got_json_log.Error)
		want_error_keys := 3 + 1 // ? arbitrary fields + message
		if got_error_keys != want_error_keys {
			t.Errorf("got %d .error keys, want %d", got_error_keys, want_error_keys)
		}
	})

	t.Run("should log error with stack_trace if error is StackTracer", func (t *testing.T) {
		stderr_file, cleanup := setup("test.logs.errors", t)

		defer func () {
			futils.Clear()
			cleanup()
		}()
		
		logger, _, err := InitLogger("", &futils)
		if err != nil {
			t.Fatal("got unexpected error initializing logger instance", err)
		}

		want_err := pkgerr.WithStack(errors.New("unable to find docker executable"))
		logger.Error("abcd", "error", want_err)

		got_stderr_lines, err := tests.ReadFileLines(stderr_file)
		if err != nil {
			t.Fatal("got unexpected error setting up log file", err)
		}

		got_stderr_len := len(got_stderr_lines)
		want_stderr_len := 1
		if got_stderr_len != want_stderr_len  {
			t.Errorf("got %d stderr lines, want %d", got_stderr_len, want_stderr_len)
		}

		var got_json_log LogEntry
		got_line := got_stderr_lines[0]

		got_parse_err := json.Unmarshal([]byte(got_line), &got_json_log)
		if got_parse_err != nil {
			t.Fatal("got error parsing log line into json", got_parse_err)
		}

		got_json_err := got_json_log.Error
		got_err_stack_trace, got_exists := got_json_log.Error["stack_trace"]
		if !got_exists {
			t.Error("got error with missing .stack_trace", got_json_err)
		}

		if stack_str, ok := got_err_stack_trace.(string); !ok || stack_str == "" {
			t.Errorf("got stack_trace empty, want non-empty string")
		}
	})

	t.Run("should log error with stack_trace and arbitrary fields if error is both StackTracer and ErrWithAttrs", func (t *testing.T) {
		stderr_file, cleanup := setup("test.logs.errors", t)

		defer func () {
			futils.Clear()
			cleanup()
		}()
		
		logger, _, err := InitLogger("", &futils)
		if err != nil {
			t.Fatal("got unexpected error initializing logger instance", err)
		}

		want_err := pkgerr.WithStack(errors.New("unable to find docker executable"))
		want_structured_err := errutils.WithAttrs(
			want_err,
			"version",
			"1.0.0",
		)

		logger.Error("abcd", "error", want_structured_err)

		got_stderr_lines, err := tests.ReadFileLines(stderr_file)
		if err != nil {
			t.Fatal("got unexpected error setting up log file", err)
		}

		got_stderr_len := len(got_stderr_lines)
		want_stderr_len := 1
		if got_stderr_len != want_stderr_len  {
			t.Errorf("got %d stderr lines, want %d", got_stderr_len, want_stderr_len)
		}

		var got_json_log LogEntry
		got_line := got_stderr_lines[0]

		got_parse_err := json.Unmarshal([]byte(got_line), &got_json_log)
		if got_parse_err != nil {
			t.Fatal("got error parsing log line into json", got_parse_err)
		}

		got_json_err := got_json_log.Error
		got_err_stack_trace, got_exists := got_json_log.Error["stack_trace"]
		if !got_exists {
			t.Error("got error with missing .stack_trace", got_json_err)
		}

		if stack_str, ok := got_err_stack_trace.(string); !ok || stack_str == "" {
			t.Errorf("got stack_trace field empty, want non-empty string")
		}

		got_err_version, got_exists := got_json_log.Error["version"]
		if version_field, ok := got_err_version.(string); !ok || version_field == "" {
			t.Errorf("got version field empty, want non-empty string")
		}
	})
}