package errors

import (
	"errors"
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

type ErrMasked struct {
	msg   string
	cause error
}

func (e *ErrMasked) Error() string {
	return e.msg
}

func (e *ErrMasked) Unwrap() error {
	return e.cause
}

func (e *ErrMasked) Attrs() []slog.Attr {
	return []slog.Attr{
		slog.String("cause", e.cause.Error()),
	}
}

func Mask(cause error, msg string) error {
	if cause == nil {
		return pkgerr.New(msg)
	}

	if masked, ok := errors.AsType[*ErrMasked](cause); ok {
		cause = masked.cause
	}
	if _, ok := errors.AsType[StackTracer](cause); !ok {
		cause = pkgerr.WithStack(cause)
	}

	return &ErrMasked{
		msg:   msg,
		cause: cause,
	}
}
