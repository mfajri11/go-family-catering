package consts

import "time"

const (
	CtxKeyAuthorization      = "Authorization"
	CtxKeySession            = "Session"
	CtxKeySID                = "Sid"
	CtxKeyRequestTime        = "Rtime"
	CtxKeyStatusCode         = "StatusCode"
	CookieResetPasswordToken = "rpt"
	CookieSID                = "sid" // session id

	DefaultAttemptsMigration = 10
	DefaultTimeoutMigration  = time.Second

	StatusNew       = 1
	StatusPaid      = 2
	StatusCancelled = 3
)
