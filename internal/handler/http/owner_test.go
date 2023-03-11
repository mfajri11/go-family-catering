package handler

import (
	"context"
	"errors"
	"family-catering/internal/model"
	"family-catering/internal/service"
	"family-catering/pkg/apperrors"
	"family-catering/pkg/consts"
	"family-catering/pkg/utils"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewOwnerHandler(t *testing.T) {
	tests := []struct {
		name string
	}{{name: "success create NewOwnerHandler"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ownerServiceMock := service.NewMockOwnerService(ctrl)
			got := NewOwnerHandler(ownerServiceMock)
			assert.NotNil(t, got)

		})
	}
}

func Test_ownerHandler_Create(t *testing.T) {
	type mocks struct {
		r                *http.Request
		ownerServiceMock *service.MockOwnerService
	}
	tests := []struct {
		name           string
		handler        *ownerHandler
		payload        string
		prepareMocks   func(*mocks)
		wantBody       string
		wantStatusCode int
	}{
		{
			name:    "success hit /api/v1/owner [post] 'ok'",
			handler: &ownerHandler{},
			payload: `{
				"name":"test",
				"email":"test@example.com",
				"phone_number": "646464",
				"password":"password"
			}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.ownerServiceMock.
					EXPECT().
					Create(m.r.Context(), gomock.AssignableToTypeOf(model.CreateOwnerRequest{})).
					Return(&model.CreateOwnerResponse{Id: 1, Name: "test", Email: "test@example.com", PhoneNumber: "646464", Password: "hashed-password"}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody: `{
				"success": true,
				"status": "success",
				"data": {
				  "owner": {
					"id": 1,
					"name": "test",
					"email":"test@example.com",
					"phone_number":"646464",
					"password":"hashed-password"
				  }
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit /api/v1/owner [post] 'unmarshal error'",
			handler: &ownerHandler{},
			payload: `{
				"name-bad-key-no-enclosed-by-double-qoutes:"test",
				"email":"test@example.com",
				"phone_number": "646464",
				"password":"password"
			}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit /api/v1/owner [post] 'email registered'",
			handler: &ownerHandler{},
			payload: `{
				"name":"test-user-exist",
				"email":"already.exist@example.com",
				"phone_number": "646464",
				"password":"password"
			}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.ownerServiceMock.
					EXPECT().
					Create(m.r.Context(), gomock.AssignableToTypeOf(model.CreateOwnerRequest{})).
					Return(nil, apperrors.ErrEmailRegistered)
			},
			wantStatusCode: http.StatusConflict,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit /api/v1/owner [post] 'missing required params'",
			handler: &ownerHandler{},
			payload: `{
				"name":"test",
				"phone_number": "646464",
				"password":"password"
			}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.ownerServiceMock.
					EXPECT().
					Create(m.r.Context(), gomock.AssignableToTypeOf(model.CreateOwnerRequest{})).
					Return(nil, apperrors.ErrFieldValidationRequired)
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit /api/v1/owner [post] 'invalid params'",
			handler: &ownerHandler{},
			payload: `{
				"name":"test",
				"email":"test@example.com",
				"phone_number": "not-a-number",
				"password":"password"
			}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.ownerServiceMock.
					EXPECT().
					Create(m.r.Context(), gomock.AssignableToTypeOf(model.CreateOwnerRequest{})).
					Return(nil, apperrors.ErrFieldValidation)
			},
			wantStatusCode: http.StatusUnprocessableEntity,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "error hit /api/v1/owner [post] 'internal server error'",
			handler: &ownerHandler{},
			payload: `{
				"name":"test",
				"email":"already.exist.email@example.com",
				"phone_number": "646464",
				"password":"password"
			}`,
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.ownerServiceMock.
					EXPECT().
					Create(m.r.Context(), gomock.AssignableToTypeOf(model.CreateOwnerRequest{})).
					Return(nil, errors.New("oops! error internal server"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"success":false,"status":"error","error":{"message":"oops! error"},"process_time":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ownerServiceMock := service.NewMockOwnerService(ctrl)
			r := httptest.NewRequest(http.MethodPost, "/api/v1/owner", strings.NewReader(tt.payload))
			w := httptest.NewRecorder()

			m := &mocks{r: r, ownerServiceMock: ownerServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}
			tt.handler.ownerService = m.ownerServiceMock

			handler := tt.handler.Create()
			handler(w, r)

			resp := w.Result()
			gotRespBody := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".+"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, gotRespBody)
		})
	}
}

func Test_ownerHandler_Get(t *testing.T) {
	type mocks struct {
		r                *http.Request
		rctx             *chi.Context
		ownerServiceMock *service.MockOwnerService
	}
	type params struct {
		id string
	}
	tests := []struct {
		name           string
		handler        *ownerHandler
		params         params
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit /api/v1/owner/{id} [get] 'ok'",
			handler: &ownerHandler{},
			params:  params{id: "1"},
			prepareMocks: func(m *mocks) {
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.ownerServiceMock.
					EXPECT().
					Get(m.r.Context(), gomock.AssignableToTypeOf(int64(0))).
					Return(&model.GetOwnerResponse{Id: 1, Name: "test", Email: "test@example.com", PhoneNumber: "646464"}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody: `{
				"success": true,
				"status": "success",
				"data": {
				  "owner": {
					"id": 1,
					"name": "test",
					"email": "test@example.com",
					"phone_number": "646464"
				  }
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit /api/v1/owner/{id} [get] 'error not found'",
			handler: &ownerHandler{},
			params:  params{id: "0"},
			prepareMocks: func(m *mocks) {
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.ownerServiceMock.
					EXPECT().
					Get(m.r.Context(), gomock.AssignableToTypeOf(int64(0))).
					Return(nil, apperrors.ErrNotFound)
			},

			wantStatusCode: http.StatusNotFound,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit /api/v1/owner/{id} [get] 'error missing path params'",
			handler: &ownerHandler{},
			params:  params{id: "not-number"},
			prepareMocks: func(m *mocks) {
				m.rctx.URLParams.Add("id", "not-number")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "error hit /api/v1/owner/{id} [get] 'internal server error'",
			params:  params{id: "1"},
			handler: &ownerHandler{},
			prepareMocks: func(m *mocks) {
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.ownerServiceMock.
					EXPECT().
					Get(m.r.Context(), gomock.AssignableToTypeOf(int64(0))).
					Return(nil, errors.New("oops! error internal server"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"success":false,"status":"error","error":{"message":"oops! error"},"process_time":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ownerServiceMock := service.NewMockOwnerService(ctrl)
			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/owner/%s", tt.params.id), nil)
			rctx := chi.NewRouteContext()
			w := httptest.NewRecorder()

			m := &mocks{r: r, rctx: rctx, ownerServiceMock: ownerServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}
			tt.handler.ownerService = m.ownerServiceMock

			handler := tt.handler.Get()
			handler(w, r)

			resp := w.Result()
			gotRespBody := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".+"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, gotRespBody)

		})
	}
}

func Test_ownerHandler_List(t *testing.T) {
	type mocks struct {
		r                *http.Request
		ownerServiceMock *service.MockOwnerService
	}
	type params struct {
		offset string
		limit  string
	}
	tests := []struct {
		name           string
		handler        *ownerHandler
		params         params
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit /api/v1/owner?limit={limit}&offset={offset} [get] 'ok'",
			handler: &ownerHandler{},
			params:  params{offset: "1", limit: "2"},
			prepareMocks: func(m *mocks) {
				m.ownerServiceMock.
					EXPECT().
					List(m.r.Context(), gomock.AssignableToTypeOf(0), gomock.AssignableToTypeOf(0)).
					Return([]*model.GetOwnerResponse{
						{Id: 1, Name: "test1", Email: "test1@example.com", PhoneNumber: "646464"},
						{Id: 2, Name: "test2", Email: "test2@example.com", PhoneNumber: "646460"}}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody: `{
				"success": true,
				"status": "success",
				"data": {
				  "owner": [
					{
					"id": 1,
					"name": "test1",
					"email": "test1@example.com",
					"phone_number": "646464"
				  },
					{
					"id": 2,
					"name": "test2",
					"email": "test2@example.com",
					"phone_number": "646460"
				  }
				  ]
				},
				"process_time": 0
			  }`,
		},
		{
			name:           "fail hit /api/v1/owner?limit={limit}&offset={offset} [get] 'error invalid query params'",
			handler:        &ownerHandler{},
			params:         params{offset: "not-number", limit: "2"},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "success hit /api/v1/owner?limit={limit}&offset={offset} [get] 'ok but no data'",
			handler: &ownerHandler{},
			params:  params{offset: "1000", limit: "2"},
			prepareMocks: func(m *mocks) {
				m.ownerServiceMock.
					EXPECT().
					List(m.r.Context(), gomock.AssignableToTypeOf(0), gomock.AssignableToTypeOf(0)).
					Return([]*model.GetOwnerResponse{}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody: `{
				"success": true,
				"status": "success",
				"data": {"owner": []},
				"process_time": 0
			  }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ownerServiceMock := service.NewMockOwnerService(ctrl)
			u, err := url.Parse("/api/v1/owner")
			if err != nil {
				panic(err)
			}
			q := u.Query()
			q.Add("limit", tt.params.limit)
			q.Add("offset", tt.params.offset)
			u.RawQuery = q.Encode()
			r := httptest.NewRequest(http.MethodGet, u.String(), nil)
			rctx := chi.NewRouteContext()
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()

			m := &mocks{r: r, ownerServiceMock: ownerServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}
			tt.handler.ownerService = m.ownerServiceMock

			handler := tt.handler.List()
			handler(w, r)
			resp := w.Result()

			gotRespBody := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".+"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, gotRespBody)

		})
	}
}

func Test_ownerHandler_Delete(t *testing.T) {
	type params struct {
		id string
	}
	type mocks struct {
		r                *http.Request
		ownerServiceMock *service.MockOwnerService
		rctx             *chi.Context
	}
	tests := []struct {
		name           string
		handler        *ownerHandler
		params         params
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit /api/v1/owner/{id} [delete] 'ok'",
			handler: &ownerHandler{},
			params:  params{id: "1"},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: true, OwnerID: 1}))
				m.ownerServiceMock.EXPECT().Delete(m.r.Context(), gomock.AssignableToTypeOf(int64(0))).Return(int64(1), nil)

			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"success":true,"status":"success","process_time":0}`,
		},
		{
			name:    "fail hit /api/v1/owner/{id} [delete] 'error invalid path params'",
			handler: &ownerHandler{},
			params:  params{id: "not-a-number"},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "not-a-number")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit /api/v1/owner/{id} [delete] 'not found'",
			handler: &ownerHandler{},
			params:  params{id: "0"},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: true, OwnerID: 1}))
				m.ownerServiceMock.EXPECT().Delete(m.r.Context(), gomock.AssignableToTypeOf(int64(0))).Return(int64(0), apperrors.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit /api/v1/owner/{id} [delete] 'error internal server'",
			handler: &ownerHandler{},
			params:  params{id: "1"},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: true, OwnerID: 1}))
				m.ownerServiceMock.EXPECT().Delete(m.r.Context(), gomock.AssignableToTypeOf(int64(0))).Return(int64(0), errors.New("oops! error internal server"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"success":false,"status":"error","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit /api/v1/owner/{id} [delete] 'error auth'",
			handler: &ownerHandler{},
			params:  params{id: "1"},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer invalid-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "invalid-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: true, OwnerID: 1}))
				m.ownerServiceMock.EXPECT().Delete(m.r.Context(), gomock.AssignableToTypeOf(int64(0))).Return(int64(0), apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ownerServiceMock := service.NewMockOwnerService(ctrl)
			r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/owner/%s", tt.params.id), nil)
			rctx := chi.NewRouteContext()
			w := httptest.NewRecorder()

			m := &mocks{r: r, rctx: rctx, ownerServiceMock: ownerServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}

			tt.handler.ownerService = ownerServiceMock

			handler := tt.handler.Delete()
			handler(w, r)
			resp := w.Result()

			gotRespBody := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".+"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, gotRespBody)

		})
	}
}

func Test_ownerHandler_Update(t *testing.T) {
	type params struct {
		id, payload string
	}
	type mocks struct {
		r                 *http.Request
		rctx              *chi.Context
		ownerServiceMocks *service.MockOwnerService
	}
	tests := []struct {
		name           string
		handler        *ownerHandler
		params         params
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit /api/v1/owner/{id} [put] 'ok'",
			handler: &ownerHandler{},
			params: params{id: "1", payload: `{
				"name":"test-updated",
				"phone_number": "646465"
			}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				m.ownerServiceMocks.
					EXPECT().
					Update(m.r.Context(), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf(model.UpdateOwnerRequest{})).
					Return(&model.UpdateOwnerResponse{Id: 1, Name: "test-updated", PhoneNumber: "646465"}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody: `{
				"success": true,
				"status": "success",
				"data": {
				  "owner": {
					"id": 1,
					"name": "test-updated",
					"phone_number": "646465"
				  }
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit /api/v1/owner/{id} [put] 'error invalid path params'",
			handler: &ownerHandler{},
			params:  params{id: "not-a-number", payload: `{"name":"test-updated","phone_number": "646465"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "not-a-number")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit /api/v1/owner/{id} [put] 'error validation request'",
			handler: &ownerHandler{},
			params:  params{id: "1", payload: `{"name":"test-updated","phone_number": "0000111-invalid-phone-number"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				m.ownerServiceMocks.
					EXPECT().
					Update(m.r.Context(), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf(model.UpdateOwnerRequest{})).
					Return(nil, apperrors.ErrFieldValidation)
			},
			wantStatusCode: http.StatusUnprocessableEntity,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit /api/v1/owner/{id} [put] 'error unmarshal request'",
			handler: &ownerHandler{},
			params:  params{id: "1", payload: `{"name-bad-key-not-enclosed-by-double-quoted:"test-updated","phone_number": "646465"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit /api/v1/owner/{id} [put] 'error not found'",
			handler: &ownerHandler{},
			params:  params{id: "0", payload: `{"name":"test-updated","phone_number": "646465"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				m.ownerServiceMocks.
					EXPECT().
					Update(m.r.Context(), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf(model.UpdateOwnerRequest{})).
					Return(nil, apperrors.ErrNotFound)

			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit /api/v1/owner/{id} [put] 'error auth'",
			handler: &ownerHandler{},
			params:  params{id: "0", payload: `{"name":"test-updated","phone_number": "646465"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer invalid-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "invalid-token"))
				m.ownerServiceMocks.
					EXPECT().
					Update(m.r.Context(), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf(model.UpdateOwnerRequest{})).
					Return(nil, apperrors.ErrAuth)

			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "error hit /api/v1/owner/{ownerId} [put] 'error internal server'",
			handler: &ownerHandler{},
			params:  params{id: "not-a-number", payload: `{"name":"test-updated","phone_number": "646465"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				m.ownerServiceMocks.
					EXPECT().
					Update(m.r.Context(), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf(model.UpdateOwnerRequest{})).
					Return(nil, errors.New("oops! error internal server"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"success":false,"status":"error","error":{"message":"oops! error"},"process_time":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ownerServiceMock := service.NewMockOwnerService(ctrl)
			r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/owner/%s", tt.params.id), strings.NewReader(tt.params.payload))
			rctx := chi.NewRouteContext()
			w := httptest.NewRecorder()

			m := &mocks{r: r, rctx: rctx, ownerServiceMocks: ownerServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}

			tt.handler.ownerService = m.ownerServiceMocks

			handler := tt.handler.Update()
			handler(w, r)
			resp := w.Result()

			gotRespBody := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".+"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, gotRespBody)
		})
	}
}

func Test_ownerHandler_ResetPasswordByEmail(t *testing.T) {
	type params struct {
		payload, rpid string
		// email will be acquired via token claim
	}
	type mocks struct {
		r                 *http.Request
		rctx              *chi.Context
		ownerServiceMocks *service.MockOwnerService
	}
	tests := []struct {
		name           string
		handler        *ownerHandler
		params         params
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit api /api/v1/owner/reset-password/{rpid} [put] 'ok'",
			handler: &ownerHandler{},
			params:  params{rpid: "rpid", payload: `{"password":"updated-plain-password", "password_confirm":"updated-plain-password"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.rctx.URLParams.Add("rpid", "rpid")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "password-token"))
				m.ownerServiceMocks.
					EXPECT().
					ResetPasswordByEmail(m.r.Context(), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(model.ResetPasswordRequest{})).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"success":true,"status":"success","process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/owner/reset-password/{rpid} [put] 'error unmarshal request'",
			handler: &ownerHandler{},
			params:  params{rpid: "rpid", payload: `{"password-bad-key-no-enclosed-by-double-quoted:"updated-plain-password", "password_confirm":"updated-plain-password"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.rctx.URLParams.Add("rpid", "rpid")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "password-token"))

			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/owner/reset-password/{rpid} [put] 'error auth'",
			handler: &ownerHandler{},
			params:  params{rpid: "rpid", payload: `{"password":"updated-plain-password", "password_confirm":"updated-plain-password"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.rctx.URLParams.Add("rpid", "rpid")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "invalid-password-token"))
				m.ownerServiceMocks.
					EXPECT().
					ResetPasswordByEmail(m.r.Context(), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(model.ResetPasswordRequest{})).
					Return(apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/owner/reset-password/{rpid} [put] 'error not found'",
			handler: &ownerHandler{},
			params:  params{rpid: "rpid", payload: `{"password":"updated-plain-password", "password_confirm":"updated-plain-password"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.rctx.URLParams.Add("rpid", "rpid")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "password-token"))
				m.ownerServiceMocks.
					EXPECT().
					ResetPasswordByEmail(m.r.Context(), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(model.ResetPasswordRequest{})).
					Return(apperrors.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/owner/reset-password/{rpid} [put] 'error internal server'",
			handler: &ownerHandler{},
			params:  params{rpid: "rpid", payload: `{"password":"updated-plain-password", "password_confirm":"updated-plain-password"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.rctx.URLParams.Add("rpid", "rpid")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "password-token"))
				m.ownerServiceMocks.
					EXPECT().
					ResetPasswordByEmail(m.r.Context(), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(model.ResetPasswordRequest{})).
					Return(errors.New("oops! error internal server"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"success":false,"status":"error","error":{"message":"oops! error"},"process_time":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ownerServiceMock := service.NewMockOwnerService(ctrl)
			r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/owner/reset-password/%s", tt.params.rpid), strings.NewReader(tt.params.payload))
			rctx := chi.NewRouteContext()
			w := httptest.NewRecorder()

			m := &mocks{r: r, rctx: rctx, ownerServiceMocks: ownerServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}

			tt.handler.ownerService = m.ownerServiceMocks

			handler := tt.handler.ResetPasswordByEmail()
			handler(w, r)
			resp := w.Result()

			gotRespBody := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".+"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, gotRespBody)
		})
	}
}

func Test_ownerHandler_ResetPasswordById(t *testing.T) {
	type params struct {
		id, payload string
	}
	type mocks struct {
		r                 *http.Request
		rctx              *chi.Context
		ownerServiceMocks *service.MockOwnerService
	}
	tests := []struct {
		name           string
		handler        *ownerHandler
		params         params
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit api /api/v1/owner/{id}/reset-password [put] 'ok'",
			handler: &ownerHandler{},
			params:  params{id: "1", payload: `{"password":"updated-plain-password", "password_confirm":"updated-plain-password"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: true, OwnerID: 1}))
				m.ownerServiceMocks.
					EXPECT().
					ResetPasswordByID(m.r.Context(), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf(model.ResetPasswordRequest{})).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"success":true,"status":"success","process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/owner/{id}/reset-password [put] 'error unmarshal request'",
			handler: &ownerHandler{},
			params:  params{id: "1", payload: `{"password-bad-key-no-enclosed-by-double-quoted:"updated-plain-password", "password_confirm":"updated-plain-password"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: true, OwnerID: 1}))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/owner/{id}/reset-password [put] 'error invalid path params'",
			handler: &ownerHandler{},
			params:  params{id: "1", payload: `{"password":"updated-plain-password", "password_confirm":"updated-plain-password"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "not-number")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: true, OwnerID: 1}))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/owner/{id}/reset-password [put] 'error auth'",
			handler: &ownerHandler{},
			params:  params{id: "1", payload: `{"password":"updated-plain-password", "password_confirm":"updated-plain-password"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "invalid-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: true, OwnerID: 1}))
				m.ownerServiceMocks.
					EXPECT().
					ResetPasswordByID(m.r.Context(), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf(model.ResetPasswordRequest{})).
					Return(apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/owner/{id}/reset-password [put] 'error not found'",
			handler: &ownerHandler{},
			params:  params{id: "0", payload: `{"password":"updated-plain-password", "password_confirm":"updated-plain-password"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "0")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: true, OwnerID: 0}))
				m.ownerServiceMocks.
					EXPECT().
					ResetPasswordByID(m.r.Context(), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf(model.ResetPasswordRequest{})).
					Return(apperrors.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/owner/{id}/reset-password [put] 'error invalid session'",
			handler: &ownerHandler{},
			params:  params{id: "1", payload: `{"password":"updated-plain-password", "password_confirm":"updated-plain-password"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: false, OwnerID: 1}))
				m.ownerServiceMocks.
					EXPECT().
					ResetPasswordByID(m.r.Context(), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf(model.ResetPasswordRequest{})).
					Return(apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/owner/{id}/reset-password [put] 'error internal server'",
			handler: &ownerHandler{},
			params:  params{id: "1", payload: `{"password":"updated-plain-password", "password_confirm":"updated-plain-password"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: true, OwnerID: 1}))
				m.ownerServiceMocks.
					EXPECT().
					ResetPasswordByID(m.r.Context(), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf(model.ResetPasswordRequest{})).
					Return(errors.New("oops! error internal server"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"success":false,"status":"error","error":{"message":"oops! error"},"process_time":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ownerServiceMock := service.NewMockOwnerService(ctrl)
			r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/owner/%s/reset-password", tt.params.id), strings.NewReader(tt.params.payload))
			rctx := chi.NewRouteContext()
			w := httptest.NewRecorder()

			m := &mocks{r: r, rctx: rctx, ownerServiceMocks: ownerServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}

			tt.handler.ownerService = m.ownerServiceMocks

			handler := tt.handler.ResetPasswordById()
			handler(w, r)
			resp := w.Result()

			gotRespBody := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".+"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, gotRespBody)
		})
	}
}

func Test_ownerHandler_UpdateEmailByID(t *testing.T) {
	type params struct {
		id, payload string
	}
	type mocks struct {
		r                 *http.Request
		rctx              *chi.Context
		ownerServiceMocks *service.MockOwnerService
	}
	tests := []struct {
		name           string
		handler        *ownerHandler
		params         params
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit api /api/v1/owner/{id}/update-email [put] 'ok'",
			handler: &ownerHandler{},
			params:  params{id: "1", payload: `{"email":"test.updated@example.com"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: true, OwnerID: 1}))
				m.ownerServiceMocks.
					EXPECT().
					UpdateEmailByID(m.r.Context(), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf(model.UpdateEmailRequest{})).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"success":true,"status":"success","process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/owner/{id}/update-email [put] 'error invalid path params'",
			handler: &ownerHandler{},
			params:  params{id: "not-number", payload: `{"email":"test.updated@example.com"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "not-number")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: true, OwnerID: 1}))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/owner/{id}/update-email [put] 'error unmarshal request'",
			handler: &ownerHandler{},
			params:  params{id: "1", payload: `{"email-bad-key-no-enclosed-by-double-quoted:"test.updated@example.com"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: true, OwnerID: 1}))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/owner/{id}/update-email [put] 'error invalid session'",
			handler: &ownerHandler{},
			params:  params{id: "1", payload: `{"email":"test.updated@example.com"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: false, OwnerID: 1}))
				m.ownerServiceMocks.
					EXPECT().
					UpdateEmailByID(m.r.Context(), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf(model.UpdateEmailRequest{})).
					Return(apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/owner/{id}/update-email [put] 'error not found'",
			handler: &ownerHandler{},
			params:  params{id: "0", payload: `{"email":"test.updated@example.com"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "0")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: false, OwnerID: 1}))
				m.ownerServiceMocks.
					EXPECT().
					UpdateEmailByID(m.r.Context(), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf(model.UpdateEmailRequest{})).
					Return(apperrors.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/owner/{id}/update-email [put] 'error internal server error'",
			handler: &ownerHandler{},
			params:  params{id: "1", payload: `{"email":"test.updated@example.com"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, "access-token"))
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), consts.CtxKeyAuthorization, &model.AuthSessionResponse{Valid: false, OwnerID: 1}))
				m.ownerServiceMocks.
					EXPECT().
					UpdateEmailByID(m.r.Context(), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf(model.UpdateEmailRequest{})).
					Return(errors.New("oops! error internal server"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"success":false,"status":"error","error":{"message":"oops! error"},"process_time":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ownerServiceMock := service.NewMockOwnerService(ctrl)
			r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/owner/%s/update-email", tt.params.id), strings.NewReader(tt.params.payload))
			rctx := chi.NewRouteContext()
			w := httptest.NewRecorder()

			m := &mocks{r: r, rctx: rctx, ownerServiceMocks: ownerServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}

			tt.handler.ownerService = m.ownerServiceMocks

			handler := tt.handler.UpdateEmailByID()
			handler(w, r)
			resp := w.Result()

			gotRespBody := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".+"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, gotRespBody)
		})
	}
}
