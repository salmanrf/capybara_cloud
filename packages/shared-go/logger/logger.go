package logger

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"

	errutils "github.com/salmanrf/capybara-cloud/packages/shared-go/errors"
)

type LogEntry struct {
	Error		 map[string]any `json:"error"`
	Hostname string 				`json:"hostname"`
	Method	 string 				`json:"method"`
	Status	 int 						`json:"status"`
	Level 	 string 				`json:"level"`
	Time  	 string 				`json:"time"`
	Msg   	 string 				`json:"msg"`
	Env 		 string 				`json:"env"`
}

type FileUtilDeps interface {
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
}

func Replacer(groups []string, a slog.Attr) slog.Attr {
	if(a.Key == "error") {
		err, ok := a.Value.Any().(error)
		if !ok {
			return a
		}

		attrs := []slog.Attr{
			{
				Key: "message",
				Value: slog.StringValue(err.Error()),
			},
		} 

		if stackErr, ok := errors.AsType[errutils.StackTracer](err); ok {
			attrs = append(attrs, slog.Attr{
				Key: "stack_trace",
				Value: slog.StringValue(fmt.Sprintf("%+v", stackErr.StackTrace())),
			})
		}

		if attrsError, ok := errors.AsType[errutils.AttrError](err); ok {
			attrs = append(attrs, attrsError.Attrs()...)
		}

		return slog.GroupAttrs("error", attrs...)
	}

	return a
}

func InitLogger(log_file_path string, futils FileUtilDeps) (*slog.Logger, func (), error) {
	env := os.Getenv("ENV")
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}
	if env == "" {
		env = "development"
	}

	var log_file *os.File
	var file_writer *bufio.Writer

	closer := func () {
		if file_writer == nil {
			return
		}

		file_writer.Flush()
		log_file.Close()
	}

	debug_handler := slog.NewJSONHandler(
		os.Stderr,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
			ReplaceAttr: Replacer,
		},
	)

	handlers := []slog.Handler{
		debug_handler,
	}

	if log_file_path != "" {
		log_file, err = futils.OpenFile(log_file_path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			return nil, closer, err
		}

		file_writer = bufio.NewWriterSize(log_file, 8192)
		info_handler := slog.NewJSONHandler(
			file_writer,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
				ReplaceAttr: Replacer,
			},
		)
		handlers = append(handlers, info_handler)
	}

	logger := slog.New(
		slog.NewMultiHandler(handlers...),
	)

	logger = logger.With(
		slog.String("hostname", hostname),
		slog.String("env", env),
	)

	return logger, closer, nil
}