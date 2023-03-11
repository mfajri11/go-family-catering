package handler

import (
	"errors"
	"family-catering/internal/model"
	"family-catering/internal/service"
	"family-catering/pkg/apperrors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestNewMenuHandler(t *testing.T) {
	type args struct {
		menuService service.MenuService
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success NewMenuHandler",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, NewMenuHandler(tt.args.menuService))
		})
	}
}

func Test_menuHandler_GetByID(t *testing.T) {
	type mocks struct {
		r *http.Request
		// w               *httptest.ResponseRecorder
		rctx            *chi.Context
		menuServiceMock *service.MockMenuService
	}
	type params struct {
		id string
	}
	tests := []struct {
		name           string
		handler        *menuHandler
		params         params
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit api/v1/menu/{Id} [get] 'ok'",
			handler: &menuHandler{},
			params:  params{id: "1"},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().GetByID(m.r.Context(), int64(1)).
					Return(&model.GetMenuResponse{ID: 1, Name: "sate", Price: 25_000, Categories: "Indonesian food"}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody: `{
				"success": true,
				"status": "success",
				"data": {
				  "menu": {
					"id": 1,
					"name": "sate",
					"price":25000,
					"categories":"Indonesian food"
				  }
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api/v1/menu/{Id} [get] 'invalid path params'",
			handler: &menuHandler{},
			params:  params{id: "one"},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "one")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody: `{
				"success": false,
				"status": "fail",
				"error": {
				  "message": "oops! error"
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api/v1/menu/{Id} [get] 'not found'",
			handler: &menuHandler{},
			params:  params{id: "1000000000000000"},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1000000000000000")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().GetByID(m.r.Context(), int64(1_000_000_000_000_000)).
					Return(nil, apperrors.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody: `{
				"success": false,
				"status": "fail",
				"error": {
				  "message": "oops! error"
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api/v1/menu/{Id} [get] 'not authorized'",
			handler: &menuHandler{},
			params:  params{id: "1"},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().GetByID(m.r.Context(), int64(1)).
					Return(nil, apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody: `{
				"success": false,
				"status": "fail",
				"error": {
				  "message": "oops! error"
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api/v1/menu/{Id} [get] 'internal server error'",
			handler: &menuHandler{},
			params:  params{id: "1"},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().GetByID(m.r.Context(), int64(1)).
					Return(nil, errors.New("oops! error internal server"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody: `{
				"success": false,
				"status": "error",
				"error": {
				  "message": "oops! error"
				},
				"process_time": 0
			  }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			menuServiceMock := service.NewMockMenuService(ctrl)
			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/menu/%s", tt.params.id), nil)
			w := httptest.NewRecorder()
			rctx := chi.NewRouteContext()
			m := &mocks{r: r, rctx: rctx, menuServiceMock: menuServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}
			tt.handler.menuService = m.menuServiceMock

			handler := tt.handler.GetByID()

			handler(w, r)

			// resetting processing time to 0 & error message to a unchanged string
			resp := w.Result()
			respBodyStr := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".*"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, respBodyStr)
		})
	}
}

func Test_menuHandler_GetByName(t *testing.T) {
	type mocks struct {
		r *http.Request

		rctx            *chi.Context
		menuServiceMock *service.MockMenuService
	}
	type params struct {
		name string
	}
	tests := []struct {
		name           string
		handler        *menuHandler
		params         params
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit api/v1/menu/{name} [get] 'ok'",
			handler: &menuHandler{},
			params:  params{name: "sate"},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("name", "sate")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().GetByName(m.r.Context(), "sate").
					Return(&model.GetMenuResponse{ID: 1, Name: "sate", Price: 25_000, Categories: "Indonesian food"}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody: `{
				"success": true,
				"status": "success",
				"data": {
				  "menu": {
					"id": 1,
					"name": "sate",
					"price":25000,
					"categories":"Indonesian food"
				  }
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api/v1/menu/{name} [get] 'invalid path params'",
			handler: &menuHandler{},
			params:  params{name: "sate"},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("names", "sate") // should be "name"
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody: `{
				"success": false,
				"status": "fail",
				"error": {
				  "message": "oops! error"
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api/v1/menu/{name} [get] 'not found'",
			handler: &menuHandler{},
			params:  params{name: "not-exists-food"},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("name", "not-exists-food")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().GetByName(m.r.Context(), "not-exists-food").
					Return(nil, apperrors.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody: `{
				"success": false,
				"status": "fail",
				"error": {
				  "message": "oops! error"
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api/v1/menu/{name} [get] 'not authorized'",
			handler: &menuHandler{},
			params:  params{name: "sate"},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "")
				m.rctx.URLParams.Add("name", "sate")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().GetByName(m.r.Context(), "sate").
					Return(nil, apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody: `{
				"success": false,
				"status": "fail",
				"error": {
				  "message": "oops! error"
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api/v1/menu/{name} [get] 'internal server error'",
			handler: &menuHandler{},
			params:  params{name: "sate"},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("name", "sate")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().GetByName(m.r.Context(), "sate").
					Return(nil, errors.New("oops! error internal server"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody: `{
				"success": false,
				"status": "error",
				"error": {
				  "message": "oops! error"
				},
				"process_time": 0
			  }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			menuServiceMock := service.NewMockMenuService(ctrl)
			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/menu/%s", tt.params.name), nil)
			w := httptest.NewRecorder()
			rctx := chi.NewRouteContext()
			m := &mocks{r: r, rctx: rctx, menuServiceMock: menuServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}
			tt.handler.menuService = m.menuServiceMock

			handler := tt.handler.GetByName()

			handler(w, r)

			// resetting processing time to 0 & error message to a unchanged string
			resp := w.Result()
			respBodyStr := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".*"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, respBodyStr)
		})
	}
}

func Test_menuHandler_List(t *testing.T) {
	type mocks struct {
		r               *http.Request
		rctx            *chi.Context
		menuServiceMock *service.MockMenuService
	}
	type params struct {
		offset, limit string
	}
	tests := []struct {
		name           string
		handler        *menuHandler
		params         params
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit api /api/v1/menu?offset={offset}&limit={limit} [get] 'ok'",
			handler: &menuHandler{},
			params: params{
				offset: "1",
				limit:  "2",
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("limit", "2")
				m.rctx.URLParams.Add("offset", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().List(m.r.Context(), 2, 1).
					Return([]*model.GetMenuResponse{
						{ID: 1, Name: "sate", Price: 25_000, Categories: "Indonesian food"},
						{ID: 2, Name: "soto babat", Price: 30_000, Categories: "Indonesian food"},
					}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody: `{
				"success": true,
				"status": "success",
				"data": {
				  "menu": [
					{
					  "id": 1,
					  "name": "sate",
					  "price": 25000,
					  "categories": "Indonesian food"
					},
					{
					  "id": 2,
					  "name": "soto babat",
					  "price": 30000,
					  "categories": "Indonesian food"
					}
				  ]
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "success hit api /api/v1/menu?offset={offset}&limit={limit} [get] 'no row/data'",
			handler: &menuHandler{},
			params: params{
				offset: "1",
				limit:  "2",
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("limit", "2")
				m.rctx.URLParams.Add("offset", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().List(m.r.Context(), 2, 1).
					Return([]*model.GetMenuResponse{}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody: `{
				"success": true,
				"status": "success",
				"data": {
				  "menu": []
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api /api/v1/menu?offset={offset}&limit={limit} [get] 'invalid params'",
			handler: &menuHandler{},
			params: params{
				offset: "not-number",
				limit:  "2",
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("limit", "2")
				m.rctx.URLParams.Add("offset", "not-number")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody: `{
				"success": false,
				"status": "fail",
				"error": {
				  "message": "oops! error"
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api /api/v1/menu?offset={offset}&limit={limit} [get] 'no auth'",
			handler: &menuHandler{},
			params: params{
				offset: "1",
				limit:  "2",
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "")
				m.rctx.URLParams.Add("limit", "2")
				m.rctx.URLParams.Add("offset", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().List(m.r.Context(), 2, 1).Return(nil, apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody: `{
				"success": false,
				"status": "fail",
				"error": {
				  "message": "oops! error"
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api /api/v1/menu?offset={offset}&limit={limit} [get] 'internal server error'",
			handler: &menuHandler{},
			params: params{
				offset: "1",
				limit:  "2",
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("limit", "2")
				m.rctx.URLParams.Add("offset", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().List(m.r.Context(), 2, 1).Return(nil, errors.New("oops! internal server error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody: `{
				"success": false,
				"status": "error",
				"error": {
				  "message": "oops! error"
				},
				"process_time": 0
			  }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			menuServiceMock := service.NewMockMenuService(ctrl)
			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/menu?offset=%s&limit=%s", tt.params.offset, tt.params.limit), nil)
			w := httptest.NewRecorder()
			rctx := chi.NewRouteContext()
			m := &mocks{r: r, rctx: rctx, menuServiceMock: menuServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}
			tt.handler.menuService = m.menuServiceMock

			handler := tt.handler.List()

			handler(w, r)

			// resetting processing time to 0 & error message to a unchanged string
			resp := w.Result()
			respBodyStr := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".*"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, respBodyStr)
		})
	}
}

func Test_menuHandler_Create(t *testing.T) {
	type mocks struct {
		r               *http.Request
		rctx            *chi.Context
		menuServiceMock *service.MockMenuService
	}
	type params struct {
		payload string
	}
	tests := []struct {
		name           string
		handler        *menuHandler
		params         params
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit api /api/v1/menu [post] 'ok'",
			handler: &menuHandler{},
			params: params{
				payload: `{
					"name":"sate",
					"price":25000,
					"categories":"Indonesian food"
				  }`,
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.menuServiceMock.EXPECT().
					Create(m.r.Context(), gomock.AssignableToTypeOf(model.CreateMenuRequest{})).
					Return(&model.CreateMenuResponse{ID: 1, Name: "sate", Price: 25_000, Categories: "Indonesian food"}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody: `{
				"success": true,
				"status": "success",
				"data": {
				  "menu": 
					{
					  "id": 1,
					  "name": "sate",
					  "price": 25000,
					  "categories": "Indonesian food"
					}
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api /api/v1/menu [post] 'invalid token'",
			handler: &menuHandler{},
			params: params{
				payload: `{
					"name":"sate",
					"price":25000,
					"categories":"Indonesian food"
				  }`,
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "invalid-token")
				m.menuServiceMock.EXPECT().
					Create(m.r.Context(), gomock.AssignableToTypeOf(model.CreateMenuRequest{})).
					Return(nil, apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/menu [post] 'error unmarshal request payload'",
			handler: &menuHandler{},
			params: params{
				payload: `{
					"name-bad-key-not-enclosed-by-double-quoted:"sate",
					"price":25000,
					"categories":"Indonesian food"
				  }`,
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "invalid-token")
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/menu [post] 'invalid request (missing required params)'",
			handler: &menuHandler{},
			params: params{
				payload: `{
					"categories":"Indonesian food"
				  }`,
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.menuServiceMock.EXPECT().
					Create(m.r.Context(), gomock.AssignableToTypeOf(model.CreateMenuRequest{})).
					Return(nil, apperrors.ErrFieldValidationRequired)
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/menu [post] 'invalid request (error validation)'",
			handler: &menuHandler{},
			params: params{
				payload: `{
					"name":"sate",
					"price":0.04,
					"categories":"Indonesian food"
				  }`,
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.menuServiceMock.EXPECT().
					Create(m.r.Context(), gomock.AssignableToTypeOf(model.CreateMenuRequest{})).
					Return(nil, apperrors.ErrFieldValidation)
			},
			wantStatusCode: http.StatusUnprocessableEntity,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/menu [post] 'internal server error'",
			handler: &menuHandler{},
			params: params{
				payload: `{
					"name":"sate",
					"price":25000,
					"categories":"Indonesian food"
				  }`,
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.menuServiceMock.EXPECT().
					Create(m.r.Context(), gomock.AssignableToTypeOf(model.CreateMenuRequest{})).
					Return(nil, errors.New("oops! internal server error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"success":false,"status":"error","error":{"message":"oops! error"},"process_time":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			menuServiceMock := service.NewMockMenuService(ctrl)
			r := httptest.NewRequest(http.MethodPost, "/api/v1/menu", strings.NewReader(tt.params.payload))
			w := httptest.NewRecorder()
			rctx := chi.NewRouteContext()
			m := &mocks{r: r, rctx: rctx, menuServiceMock: menuServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}
			tt.handler.menuService = m.menuServiceMock

			handler := tt.handler.Create()

			handler(w, r)

			// resetting processing time to 0 & error message to a unchanged string
			resp := w.Result()
			respBodyStr := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".*"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, respBodyStr)
		})
	}

}

func Test_menuHandler_Update(t *testing.T) {
	type mocks struct {
		r               *http.Request
		rctx            *chi.Context
		menuServiceMock *service.MockMenuService
	}
	type params struct {
		id      string
		payload string
	}
	tests := []struct {
		name           string
		handler        *menuHandler
		params         params
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit api /api/v1/menu/{id} [put] 'ok'",
			handler: &menuHandler{},
			params:  params{id: "1", payload: `{"name":"sate padang", "price":30000, "categories":"Indonesian food"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().
					Update(m.r.Context(), int64(1), gomock.AssignableToTypeOf(model.UpdateMenuRequest{})).
					Return(&model.UpdateMenuResponse{ID: 1, Name: "sate padang", Price: 30_000, Categories: "Indonesian food"}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody: `{
				"success": true,
				"status": "success",
				"data": {
				  "menu": 
					{
					  "id": 1,
					  "name": "sate padang",
					  "price": 30000,
					  "categories": "Indonesian food"
					}
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api /api/v1/menu/{id} [put] 'unmarshal error'",
			handler: &menuHandler{},
			params:  params{id: "1", payload: `{"name-bad-key-not-enclosed-by-double-quoted:"sate padang", "price":30000, "categories":"Indonesian food"}`},
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
			name:    "fail hit api /api/v1/menu/{id} [put] 'invalid-token'",
			handler: &menuHandler{},
			params:  params{id: "1", payload: `{"name":"sate padang", "price":30000, "categories":"Indonesian food"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "invalid-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().
					Update(m.r.Context(), int64(1), gomock.AssignableToTypeOf(model.UpdateMenuRequest{})).
					Return(nil, apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/menu/{id} [put] 'invalid request (missing required params)'",
			handler: &menuHandler{},
			params:  params{id: "1", payload: `{"price":30000, "categories":"Indonesian food"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().
					Update(m.r.Context(), int64(1), gomock.AssignableToTypeOf(model.UpdateMenuRequest{})).
					Return(nil, apperrors.ErrFieldValidationRequired)
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/menu/{id} [put] 'invalid request (error validation)'",
			handler: &menuHandler{},
			params:  params{id: "1", payload: `{"name":"sate padang", "price":-10000, "categories":"Indonesian food"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().
					Update(m.r.Context(), int64(1), gomock.AssignableToTypeOf(model.UpdateMenuRequest{})).
					Return(nil, apperrors.ErrFieldValidation)
			},
			wantStatusCode: http.StatusUnprocessableEntity,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/menu/{id} [put] 'not found'",
			handler: &menuHandler{},
			params:  params{id: "0", payload: `{"price":30000, "categories":"Indonesian food"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "0")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().
					Update(m.r.Context(), int64(0), gomock.AssignableToTypeOf(model.UpdateMenuRequest{})).
					Return(nil, apperrors.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/menu/{id} [put] 'internal server error'",
			handler: &menuHandler{},
			params:  params{id: "1", payload: `{"price":30000, "categories":"Indonesian food"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().
					Update(m.r.Context(), int64(1), gomock.AssignableToTypeOf(model.UpdateMenuRequest{})).
					Return(nil, errors.New("oops! internal server error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"success":false,"status":"error","error":{"message":"oops! error"},"process_time":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			menuServiceMock := service.NewMockMenuService(ctrl)
			r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/menu/%s", tt.params.id), strings.NewReader(tt.params.payload))
			w := httptest.NewRecorder()
			rctx := chi.NewRouteContext()
			m := &mocks{r: r, rctx: rctx, menuServiceMock: menuServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}
			tt.handler.menuService = m.menuServiceMock

			handler := tt.handler.Update()

			handler(w, r)

			// resetting processing time to 0 & error message to a unchanged string
			resp := w.Result()
			respBodyStr := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".*"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, respBodyStr)
		})
	}
}

func Test_menuHandler_Delete(t *testing.T) {
	type mocks struct {
		r               *http.Request
		rctx            *chi.Context
		menuServiceMock *service.MockMenuService
	}
	type params struct {
		id string
	}
	tests := []struct {
		name           string
		handler        *menuHandler
		params         params
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit api /api/v1/menu/{id} [delete] 'ok'",
			handler: &menuHandler{},
			params: params{
				id: "1",
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().Delete(m.r.Context(), int64(1)).Return(int64(1), nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"success":true,"status":"success","process_time":0}`,
		},
		{
			name:    "success hit api /api/v1/menu/{id} [delete] 'not found'",
			handler: &menuHandler{},
			params: params{
				id: "1",
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "0")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().Delete(m.r.Context(), int64(0)).Return(int64(0), apperrors.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "success hit api /api/v1/menu/{id} [delete] 'not found'",
			handler: &menuHandler{},
			params: params{
				id: "1",
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().Delete(m.r.Context(), int64(1)).Return(int64(0), apperrors.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "success hit api /api/v1/menu/{id} [delete] 'internal server error'",
			handler: &menuHandler{},
			params: params{
				id: "1",
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Authorization", "Bearer access-token")
				m.rctx.URLParams.Add("id", "1")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.menuServiceMock.EXPECT().Delete(m.r.Context(), int64(1)).Return(int64(0), errors.New("oops! internal server error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"success":false,"status":"error","error":{"message":"oops! error"},"process_time":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			menuServiceMock := service.NewMockMenuService(ctrl)
			r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/menu/%s", tt.params.id), nil)
			w := httptest.NewRecorder()
			rctx := chi.NewRouteContext()
			m := &mocks{r: r, rctx: rctx, menuServiceMock: menuServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}
			tt.handler.menuService = m.menuServiceMock

			handler := tt.handler.Delete()

			handler(w, r)

			// resetting processing time to 0 & error message to a unchanged string
			resp := w.Result()
			respBodyStr := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".*"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, respBodyStr)

		})
	}
}
