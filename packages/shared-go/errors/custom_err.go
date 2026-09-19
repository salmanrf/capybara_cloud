package errors

import (
	"log/slog"

	pkgerr "github.com/pkg/errors"
)

type StackTracer interface {
	error
	StackTrace() pkgerr.StackTrace
}

type ErrWithAttrs struct {
	error
	attrs []slog.Attr
}

func (e *ErrWithAttrs) Unwrap() error {
	return e.error
}

func (e *ErrWithAttrs) Attrs() []slog.Attr {
	return e.attrs
}

func WithAttrs(err error, args ...any) error {
	return &ErrWithAttrs{
		error: err,
		attrs: argsToAttr(args),
	}
}

// argsToAttr turns a list of typed or untyped values into a slice of [slog.Attr].
// args[i] is treated as a key if it is a string or an [slog.Attr]; otherwise, it
// is treated as a value with key "!BADKEY".
func argsToAttr(args []any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(args))
	for i := 0; i < len(args); {
		switch key := args[i].(type) {
		case slog.Attr:
			attrs = append(attrs, key)
			i++
		case string:
			if i+1 >= len(args) {
				attrs = append(attrs, slog.String("!BADKEY", key))
				i++
			} else {
				attrs = append(attrs, slog.Any(key, args[i+1]))
				i += 2
			}
		default:
			attrs = append(attrs, slog.Any("!BADKEY", args[i]))
			i++
		}
	}
	
	return attrs
}

type AttrError interface {
	error
	Attrs() []slog.Attr
}