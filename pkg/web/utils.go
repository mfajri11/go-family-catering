package web

import (
	"context"
	"family-catering/config"
	"family-catering/internal/model"
	"family-catering/pkg/consts"
	"family-catering/pkg/logger"
	"family-catering/pkg/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var (
	timeNow func() time.Time = time.Now
)

func PathParamInt64(r *http.Request, key string) (int64, error) {
	valStr := chi.URLParam(r, key)
	val, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return val, nil
}

func PathParamString(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func PaginationLimitOffset(r *http.Request) (limit int, offset int, err error) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// check offset
	if offsetStr == "" {
		offset = 0
	} else {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return 0, 0, err
		}
	}

	if limitStr == "" {
		limit = config.Cfg().Web.PaginationLimit
	} else {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return 0, 0, err
		} else if limit > config.Cfg().Web.PaginationLimit {
			limit = config.Cfg().Web.PaginationLimit
		}
	}

	return
}

func ParseSearchQueryParams(r *http.Request) (model.OrderQuery, error) {
	var (
		val string
		req model.OrderQuery
	)
	req = model.OrderQuery{}
	val = r.URL.Query().Get("menu-names")
	if val != "" {
		names := strings.Split(val, ",")
		req.MenuNames = names
	}
	val = r.URL.Query().Get("emails")
	if val != "" {
		emails := strings.Split(val, ",")
		req.CustomerEmails = emails
	}
	val = r.URL.Query().Get("exact-names")
	if val != "" {
		isExactMenuNamesMatch, err := strconv.ParseBool(val)
		if err != nil {
			return req, err
		}
		req.ExactMenuNamesMatch = isExactMenuNamesMatch
	}
	val = r.URL.Query().Get("qty")
	if val != "" {
		qty, err := strconv.Atoi(val)
		if err != nil {
			return req, err
		}
		req.Qty = qty
	}
	val = r.URL.Query().Get("max-price")
	if val != "" {
		price, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return req, err
		}
		req.MaxPrice = float32(price)
	}
	val = r.URL.Query().Get("min-price")
	if val != "" {
		price, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return req, err
		}
		req.MinPrice = float32(price)
	}
	val = r.URL.Query().Get("status")
	if val != "" {
		status, err := strconv.Atoi(val)
		if err != nil {
			return req, err
		}
		req.Status = status
	}
	val = r.URL.Query().Get("start-day")
	if val != "" {
		req.StartDay = val
	}
	val = r.URL.Query().Get("end-day")
	if val != "" {
		req.EndDay = val
	}
	return req, nil
}

func RequestStartTimeFromContext(ctx context.Context) time.Time {
	t := utils.ValueContext(ctx, consts.CtxKeyRequestTime)
	s, ok := t.(time.Time)
	if !ok {
		return timeNow()
	}

	return s
}

func Authorization(r *http.Request) string {
	var token string
	auth := r.Header.Get(consts.CtxKeyAuthorization)
	if strings.Contains(auth, "Bearer ") {
		token = strings.Replace(auth, "Bearer ", "", -1)
		return token
	}

	passToken, err := r.Cookie(consts.CookieResetPasswordToken)
	if err != nil {
		return ""
	}
	return passToken.Value
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var statusCode int
		w = middleware.NewWrapResponseWriter(w, 1) // http version 1.1
		start := RequestStartTimeFromContext(r.Context())
		ctx := utils.ContextWithValue(r.Context(), consts.CtxKeyRequestTime, start)
		zlog := logger.Log().
			Info().
			Str("name", "request").
			Str("protocol", "http").
			Str("url", r.URL.String()).
			Str("method", r.Method).
			Str("host", r.Host).
			Str("IP", RealIP(r))
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
		ws, ok := w.(middleware.WrapResponseWriter)
		if ok {
			statusCode = ws.Status()
		}

		zlog.
			Int("status_code", statusCode).
			Int64("process_time", int64(time.Since(start))).Msg("")
	})
}
