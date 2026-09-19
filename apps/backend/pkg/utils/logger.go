package utils

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"sync"

	pkgerr "github.com/pkg/errors"
)

var Logger slog.Logger
var logsetup sync.Once

type stack_tracer interface {
	error
	StackTrace() pkgerr.StackTrace
}

func CreateLogger() error {
	attrs := func (groups []string, a slog.Attr) slog.Attr {
		if a.Key == "error" {
			err, ok := a.Value.Any().(error)
			if !ok {
				return a
			}
			if stackerr, ok := errors.AsType[stack_tracer](err); ok{
				return slog.GroupAttrs(
					"error",
					slog.Attr{
						Key: "message",
						Value: slog.StringValue(stackerr.Error()),
					},
					slog.Attr{
						Key: "stack_trace",
						Value: slog.StringValue(fmt.Sprintf("%+v", stackerr.StackTrace())),
					},
				)
			} else {
				return slog.String("error", fmt.Sprintf("%+v", err))
			}
		}

		return a
	}

	debug_logger := slog.NewTextHandler(
		os.Stderr,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
			ReplaceAttr: attrs,
		},
	)

	pwd, err := os.Getwd()
	log_path := path.Join(pwd, "..", "..", "apps.backend.logs")
	log_file, err := os.OpenFile(log_path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	info_logger := slog.NewJSONHandler(
		log_file,
		&slog.HandlerOptions{
			Level: slog.LevelInfo,
			ReplaceAttr: attrs,
		},
	)

	handlers := slog.NewMultiHandler(
		debug_logger,
		info_logger,
	)
	logger := slog.New(handlers)

	logsetup.Do(func () {
		Logger = *logger
	})

	return nil
}