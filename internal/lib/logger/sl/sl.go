package sl

import (
	"golang.org/x/exp/slog"
)

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func String(key string, value string) slog.Attr {
	return slog.Attr{
		Key:   key,
		Value: slog.StringValue(value),
	}
}

func Any(key string, value interface{}) slog.Attr {
	return slog.Attr{
		Key:   key,
		Value: slog.AnyValue(value),
	}
}
