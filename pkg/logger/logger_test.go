package logger

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLog_Debug(t *testing.T) {
	type args struct {
		message string
		args    []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test debug",
			args: args{
				message: "debug message",
			},
		},
		{
			name: "test debug format",
			args: args{
				message: "debug message %v",
				args:    []interface{}{"format"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg string
			InitMock(t)

			Debug(tt.args.message, tt.args.args...)

			msg = tt.args.message
			if len(tt.args.args) != 0 {
				msg = fmt.Sprintf(tt.args.message, tt.args.args...)
			}
			ent := MockEntries()
			ent.ExpLen(1)
			ent.ExpMsg(msg)
			ent.ExpStr("level", "debug")
			DestroyMock()
		})
	}
}

func TestInfo(t *testing.T) {
	type args struct {
		message string
		args    []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test info",
			args: args{
				message: "info message",
			},
		},
		{
			name: "test info format",
			args: args{
				message: "info message %v",
				args:    []interface{}{"format"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg string
			InitMock(t)
			Info(tt.args.message, tt.args.args...)

			msg = tt.args.message
			if len(tt.args.args) != 0 {
				msg = fmt.Sprintf(tt.args.message, tt.args.args...)
			}
			ent := MockEntries()
			ent.ExpLen(1)
			ent.ExpMsg(msg)
			ent.ExpStr("level", "info")
			DestroyMock()
		})
	}
}

func TestWarn(t *testing.T) {
	type args struct {
		message string
		args    []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test warn",
			args: args{
				message: "warn message",
			},
		},
		{
			name: "test warn format",
			args: args{
				message: "warn message %v",
				args:    []interface{}{"format"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg string
			InitMock(t)
			Warn(tt.args.message, tt.args.args...)

			msg = tt.args.message
			if len(tt.args.args) != 0 {
				msg = fmt.Sprintf(tt.args.message, tt.args.args...)
			}

			ent := MockEntries()
			ent.ExpLen(1)
			ent.ExpMsg(msg)
			ent.ExpStr("level", "warn")
			DestroyMock()
		})
	}
}

func TestError(t *testing.T) {
	type args struct {
		err     error
		message string
		args    []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test error",
			args: args{
				message: "error message",
				err:     fmt.Errorf("error"),
			},
		},
		{
			name: "test error format",
			args: args{
				message: "error message %v",
				args:    []interface{}{"format"},
				err:     fmt.Errorf("error format"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg string
			InitMock(t)
			Error(tt.args.err, tt.args.message, tt.args.args...)

			msg = tt.args.message
			if len(tt.args.args) != 0 {
				msg = fmt.Sprintf(tt.args.message, tt.args.args...)
			}

			ent := MockEntries()
			ent.ExpLen(1)
			ent.ExpMsg(msg)
			ent.ExpStr("level", "error")
			DestroyMock()
		})
	}
}

// os.Exit should be mocked
// func TestLog_Fatal(t *testing.T) {
// 	type args struct {
// 		err     error
// 		message string
// 		args    []interface{}
// 	}
// 	tests := []struct {
// 		name string
// 		l    *Log
// 		args args
// 	}{
// 		{
// 			name: "test fatal",
// 			args: args{
// 				message: "fatal message",
// 				err:     fmt.Errorf("fatal"),
// 			},
// 		},
// 		{
// 			name: "test fatal format",
// 			args: args{
// 				message: "fatal message %v",
// 				args:    []interface{}{"format"},
// 				err:     fmt.Errorf("fatal format"),
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			var msg string
// 			tst := zltest.New(t)
// 			tt.l = newLoggerMock(t, tst)

// 			tt.l.Fatal(tt.args.err, tt.args.message, tt.args.args...)

// 			msg = tt.args.message
// 			if len(tt.args.args) != 0 {
// 				msg = fmt.Sprintf(tt.args.message, tt.args.args...)
// 			}

// 			ent := tst.Entries()
// 			ent.ExpLen(1)
// 			ent.ExpMsg(msg)
// 			ent.ExpStr("level", "fail")
// 		})
// 	}
// }

func TestLog(t *testing.T) {
	tests := []struct {
		name string
	}{{name: "Log not nil"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, Log())
		})
	}
}

func TestErrorWithCause(t *testing.T) {
	type args struct {
		err      error
		errCause error
		message  string
		args     []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test error with cause",
			args: args{
				errCause: errors.New("root cause"),
				message:  "error with cause",
			},
		},
		{
			name: "test error with nil cause",
			args: args{
				errCause: errors.New("root cause"),
				message:  "error no cause",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitMock(t)
			tt.args.err = fmt.Errorf("another error: %w", tt.args.errCause)
			ErrorWithCause(tt.args.err, tt.args.errCause, tt.args.message, tt.args.args...)
			ent := MockEntries()
			fmt.Println(ent)
			ent.ExpStr("cause", tt.args.errCause.Error())
			ent.ExpStr("error", tt.args.err.Error())
			DestroyMock()
		})
	}
}
