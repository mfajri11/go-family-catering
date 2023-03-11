package handler

import (
	"errors"
	"family-catering/internal/model"
	"family-catering/internal/service"
	"family-catering/pkg/apperrors"
	"family-catering/pkg/consts"
	"family-catering/pkg/utils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthandler(t *testing.T) {
	type args struct {
		authService service.AuthService
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success NewAuthHandler",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, NewAuthandler(tt.args.authService))
		})
	}
}

func Test_authHandler_Login(t *testing.T) {
	type mocks struct {
		r               *http.Request
		w               *httptest.ResponseRecorder
		authServiceMock *service.MockAuthService
	}
	tests := []struct {
		name           string
		handler        *authHandler
		payload        string
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit api /api/v1/auth/login [post] 'ok'",
			handler: &authHandler{},
			payload: `{"email":"test@example.com", "password":"12345pass"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.authServiceMock.EXPECT().Login(gomock.Any(), gomock.Any()).Return(&model.AuthLoginResponse{SID: "sid", AccessToken: "access-token", RefreshToken: "refresh-token"}, nil)

			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"success":true,"status":"success","data":{"auth":{"access_token":"access-token","refresh_token":"refresh-token"}},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/auth/login [post] 'unmarshal request payload'",
			handler: &authHandler{},
			payload: `{"bad-key-not-enclosed-by-quote:"test@example.com", "password":"12345pass"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")

			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/auth/login [post] 'missing required params (no email)'",
			handler: &authHandler{},
			payload: `{"password":"12345pass"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.authServiceMock.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, apperrors.ErrFieldValidationRequired)

			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/auth/login [post] 'invalid payload'",
			handler: &authHandler{},
			payload: `{"email":"invalid.email@example.com","password":"12345pass"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.authServiceMock.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, apperrors.ErrFieldValidation)

			},
			wantStatusCode: http.StatusUnprocessableEntity,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/auth/login [post] 'not found email'",
			handler: &authHandler{},
			payload: `{"email":"unregistered.email@example.com","password":"12345pass"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.authServiceMock.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, apperrors.ErrNotFound)

			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/auth/login [post] 'wrong password'",
			handler: &authHandler{},
			payload: `{"email":"test@example.com","password":"wrong-password"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.authServiceMock.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, apperrors.ErrAuth)

			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "error hit api /api/v1/auth/login [post] 'internal server error'",
			handler: &authHandler{},
			payload: `{"email":"test@example.com","password":"12345pass"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.authServiceMock.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, errors.New("oops! internal server error"))

			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"success":false,"status":"error","error":{"message":"oops! error"},"process_time":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			authServiceMock := service.NewMockAuthService(ctrl)
			r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(tt.payload))
			w := httptest.NewRecorder()

			tt.handler.authService = authServiceMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{r: r, w: w, authServiceMock: authServiceMock})
			}

			handler := tt.handler.Login()
			handler(w, r)
			// resetting processing time to 0 & error message to a unchanged string
			resp := w.Result()
			respBodyStr := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".*"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, respBodyStr)

		})
	}
}

func Test_authHandler_Logout(t *testing.T) {
	type mocks struct {
		r               *http.Request
		w               *httptest.ResponseRecorder
		authServiceMock *service.MockAuthService
	}
	tests := []struct {
		name           string
		handler        *authHandler
		payload        string
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit api /api/v1/auth/logout [delete] 'ok'",
			handler: &authHandler{},
			payload: `{"password":"12345pass"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.r.AddCookie(&http.Cookie{Name: consts.CookieSID, Value: "sid"})
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Sid", "sid"))
				m.authServiceMock.EXPECT().Logout(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"success":true,"status":"success","process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/auth/logout [delete] 'error unmarshal request's payload'",
			handler: &authHandler{},
			payload: `{"password-bad-key-not-enclosed-by-double-quoted:"12345pass"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.r.AddCookie(&http.Cookie{Name: consts.CookieSID, Value: "sid"})
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Sid", "sid"))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/auth/logout [delete] 'error auth no cookie'",
			handler: &authHandler{},
			payload: `{"password":"12345pass"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.authServiceMock.EXPECT().Logout(gomock.Any(), gomock.Any()).Return(apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/auth/logout [delete] 'error auth no access token'",
			handler: &authHandler{},
			payload: `{"password":"12345pass"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "")
				m.r.AddCookie(&http.Cookie{Name: "Sid", Value: "sid"})
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Sid", "sid"))
				m.authServiceMock.EXPECT().Logout(gomock.Any(), gomock.Any()).Return(apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/auth/logout [delete] 'missing required params'",
			handler: &authHandler{},
			payload: `{}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "")
				m.r.AddCookie(&http.Cookie{Name: "Sid", Value: "sid"})
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Sid", "sid"))
				m.authServiceMock.EXPECT().Logout(gomock.Any(), gomock.Any()).Return(apperrors.ErrFieldValidationRequired)
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/auth/logout [delete] 'invalid request payload'",
			handler: &authHandler{},
			payload: `{"password":"invalid-password"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "")
				m.r.AddCookie(&http.Cookie{Name: "Sid", Value: "sid"})
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Sid", "sid"))
				m.authServiceMock.EXPECT().Logout(gomock.Any(), gomock.Any()).Return(apperrors.ErrFieldValidation)
			},
			wantStatusCode: http.StatusUnprocessableEntity,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "error hit api /api/v1/auth/logout [delete] 'internal server error'",
			handler: &authHandler{},
			payload: `{"password":"12345pass"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "")
				m.r.AddCookie(&http.Cookie{Name: "Sid", Value: "sid"})
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Sid", "sid"))
				m.authServiceMock.EXPECT().Logout(gomock.Any(), gomock.Any()).Return(errors.New("oops! internal server error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"success":false,"status":"error","error":{"message":"oops! error"},"process_time":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			authServiceMock := service.NewMockAuthService(ctrl)
			r := httptest.NewRequest(http.MethodDelete, "/api/v1/auth/login", strings.NewReader(tt.payload))
			w := httptest.NewRecorder()

			tt.handler.authService = authServiceMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{r: r, w: w, authServiceMock: authServiceMock})
			}

			handler := tt.handler.Logout()
			handler(w, r)
			// resetting processing time to 0 & error message to a unchanged string
			resp := w.Result()
			respBodyStr := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".*"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, respBodyStr)
		})
	}
}

func Test_authHandler_ForgotPassword(t *testing.T) {
	type mocks struct {
		r               *http.Request
		w               *httptest.ResponseRecorder
		authServiceMock *service.MockAuthService
	}
	tests := []struct {
		name           string
		handler        *authHandler
		payload        string
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit api /api/v1/auth/forgot-password [put] 'ok'",
			handler: &authHandler{},
			payload: `{"email":"test@example.com"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.authServiceMock.EXPECT().ForgotPassword(gomock.Any(), gomock.Any()).Return("password-token", nil)
				http.SetCookie(m.w, &http.Cookie{
					Name:  consts.CookieResetPasswordToken,
					Value: "password-token",
				})
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"success":true,"status":"success","process_time":0}`,
		},
		{
			name:           "fail hit api /api/v1/auth/forgot-password [put] 'error unmarshal request'",
			handler:        &authHandler{},
			payload:        `{"email-bad-key-not-enclosed-by-double-quoted:"test@example.com"}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/auth/forgot-password [put] 'missing required params'",
			handler: &authHandler{},
			payload: `{}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Content-Type", "application/json")
				m.authServiceMock.EXPECT().ForgotPassword(gomock.Any(), gomock.Any()).Return("", apperrors.ErrFieldValidationRequired)
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/auth/forgot-password [put] 'invalid request payload'",
			handler: &authHandler{},
			payload: `{"email":"invalid.email@.com"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Content-Type", "application/json")
				m.authServiceMock.EXPECT().ForgotPassword(gomock.Any(), gomock.Any()).Return("", apperrors.ErrFieldValidation)
			},
			wantStatusCode: http.StatusUnprocessableEntity,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/auth/forgot-password [put] 'email not found'",
			handler: &authHandler{},
			payload: `{"email":"invalid.email@.com"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.authServiceMock.EXPECT().ForgotPassword(gomock.Any(), gomock.Any()).Return("", apperrors.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/auth/forgot-password [put] 'internal server error'",
			handler: &authHandler{},
			payload: `{"email":"invalid.email@.com"}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.authServiceMock.EXPECT().ForgotPassword(gomock.Any(), gomock.Any()).Return("", errors.New("oops! internal server error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"success":false,"status":"error","error":{"message":"oops! error"},"process_time":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			authServiceMock := service.NewMockAuthService(ctrl)
			r := httptest.NewRequest(http.MethodPut, "/api/v1/auth/login", strings.NewReader(tt.payload))
			w := httptest.NewRecorder()

			tt.handler.authService = authServiceMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{r: r, w: w, authServiceMock: authServiceMock})
			}

			handler := tt.handler.ForgotPassword()
			handler(w, r)
			// resetting processing time to 0 & error message to a unchanged string
			resp := w.Result()
			respBodyStr := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".+"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, respBodyStr)
		})
	}
}

func Test_authHandler_RenewAccessToken(t *testing.T) {
	type mocks struct {
		r               *http.Request
		w               *httptest.ResponseRecorder
		authServiceMock *service.MockAuthService
	}
	tests := []struct {
		name           string
		handler        *authHandler
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit /api/v1/auth/renew-access-token [get] 'ok'",
			handler: &authHandler{},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Session", model.AuthSessionResponse{Valid: true}))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "refresh-token"))
				m.authServiceMock.EXPECT().RenewAccessToken(gomock.Any()).Return(&model.AuthRenewAccessTokenResponse{AccessToken: "access-token", ExpiredAt: "2023-02-20:00.00"}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"success":true,"status":"success","data":{"auth":{"access_token":"access-token","expired_at":"2023-02-20:00.00"}},"process_time":0}`,
		},
		{
			name:    "fail hit /api/v1/auth/renew-access-token [get] 'error auth no invalid refresh-token'",
			handler: &authHandler{},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Session", model.AuthSessionResponse{Valid: true}))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "invalid-refresh-token"))
				m.authServiceMock.EXPECT().RenewAccessToken(gomock.Any()).Return(nil, apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit /api/v1/auth/renew-access-token [get] 'error auth no sid/invalid sid'",
			handler: &authHandler{},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "invalid-refresh-token"))
				m.authServiceMock.EXPECT().RenewAccessToken(gomock.Any()).Return(nil, apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "error hit /api/v1/auth/renew-access-token [get] 'internal server error'",
			handler: &authHandler{},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "invalid-refresh-token"))
				m.authServiceMock.EXPECT().RenewAccessToken(gomock.Any()).Return(nil, errors.New("oops! internal server error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"success":false,"status":"error","error":{"message":"oops! error"},"process_time":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			authServiceMock := service.NewMockAuthService(ctrl)
			r := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
			w := httptest.NewRecorder()

			tt.handler.authService = authServiceMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{r: r, w: w, authServiceMock: authServiceMock})
			}

			handler := tt.handler.RenewAccessToken()
			handler(w, r)

			resp := w.Result()
			gotRespBody := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".+"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, gotRespBody)
		})
	}
}
