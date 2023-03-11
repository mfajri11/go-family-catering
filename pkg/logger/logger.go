package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

// type Logger interface {
// 	Debug(message string, args ...interface{})
// 	Info(message string, args ...interface{})
// 	Warn(message string, args ...interface{})
// 	Error(err error, message string, args ...interface{})
// 	Fatal(err error, message string, args ...interface{})
// }

var (
	zlog    zerolog.Logger
	writter io.Writer
)

func initzlog() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	f, err := os.Open("/etc/var/log/family-catering/access.log")
	if err != nil {
		writter = os.Stderr
	} else {
		writter = f
	}
	zlog = zerolog.New(writter).With().Timestamp().Logger()
}

func init() {
	initzlog()
}

func msg(le *zerolog.Event, message string, args ...interface{}) {
	if len(args) == 0 {
		le.Msg(message)
	} else {
		le.Msgf(message, args...)
	}
}

func Debug(message string, args ...interface{}) {
	le := zlog.Debug()
	msg(le, message, args...)
}

func Info(message string, args ...interface{}) {
	le := zlog.Info()
	msg(le, message, args...)
}

func Warn(message string, args ...interface{}) {
	le := zlog.Warn()
	msg(le, message, args...)
}

func ErrorWithCause(err, errCause error, message string, args ...interface{}) {
	le := zlog.Error().Err(err).AnErr("cause", errCause)
	msg(le, message, args...)
}

func Error(err error, message string, args ...interface{}) {
	le := zlog.Error().Err(err)
	msg(le, message, args...)
}

func Fatal(err error, message string, args ...interface{}) {
	le := zlog.Fatal().Err(err)
	msg(le, message, args...)
}

func Log() *zerolog.Logger {
	return &zlog
}
