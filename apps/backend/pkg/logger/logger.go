package logger

import (
	"bufio"
	"log/slog"
	"os"
)

type LogEntry struct {
	Hostname string `json:"hostname"`
	Method	 string `json:"method"`
	Status	 int 		`json:"status"`
	Level 	 string `json:"level"`
	Time  	 string `json:"time"`
	Msg   	 string `json:"msg"`
	Env 		 string `json:"env"`
}

type FileUtilDeps interface {
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
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