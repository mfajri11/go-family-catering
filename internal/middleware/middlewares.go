package middleware

// import (
// 	"family-catering/pkg/consts"
// 	"family-catering/pkg/logger"
// 	"family-catering/pkg/utils"
// 	"family-catering/pkg/web"
// 	"net/http"
// 	"time"
// )

// func Logger(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		start := web.RequestStartTimeFromContext(r.Context())
// 		ctx := utils.ContextWithValue(r.Context(), consts.CtxKeyRequestTime, start)
// 		zlog := logger.Log().
// 			Info().
// 			Str("name", "request").
// 			Str("protocol", "http").
// 			Str("method", r.Method).
// 			Str("host", r.Host).
// 			Str("url", r.RequestURI).
// 			Str("IP", web.RealIP(r))

// 		next.ServeHTTP(w, r.WithContext(ctx))

// 		zlog.
// 			Int("status_code", r.Response.StatusCode).
// 			Int64("process_time", int64(time.Since(start))).Msg("")
// 	})
// }
