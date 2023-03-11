package logger

import (
	"github.com/rs/zerolog"
	"github.com/rzajac/zltest"
)

// var ent *zltest.Entries
// var tst *zltest.Tester

func InitMock(t zltest.T) {
	tst := zltest.New(t)
	writter = tst
	zlog = zerolog.New(tst).With().Caller().Timestamp().Logger()

}

// GetMock return zltest.Entries which hold log value (if InitMock already called)
// this function should be called after code under test which will logging a value to a mock/zltest.Tester
func MockEntries() *zltest.Entries {
	tst, ok := writter.(*zltest.Tester)
	if !ok {
		return nil
	}
	ent := tst.Entries()
	return &ent

}

func DestroyMock() {
	initzlog()
}
