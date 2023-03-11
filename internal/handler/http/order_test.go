package handler

import (
	"context"
	"errors"
	"family-catering/internal/model"
	"family-catering/internal/service"
	"family-catering/pkg/apperrors"
	"family-catering/pkg/utils"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewOrderHandler(t *testing.T) {
	type args struct {
		orderService service.OrderService
	}
	tests := []struct {
		name string
		args args
	}{{name: "success NewOrderHandler"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, NewOrderHandler(tt.args.orderService))
		})
	}
}

func Test_orderHandler_Create(t *testing.T) {
	type mocks struct {
		r                *http.Request
		w                *httptest.ResponseRecorder
		rctx             *chi.Context
		orderServiceMock *service.MockOrderService
	}
	type params struct {
		payload string
	}
	tests := []struct {
		name           string
		handler        *orderHandler
		params         params
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit api /api/v1/order [post] 'ok'",
			handler: &orderHandler{},
			params: params{
				payload: `{
					"customer_email":"test@example.com",
					"orders":[
						{
						  "name":"sate",
						  "qty":5
						},
					  {
						"name":"es lemon tea",
						"qty":5
					  },
					  {
						"name":"nasi pecel",
						"qty":5
					  }
					]
				  }`,
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "access-token"))
				m.orderServiceMock.EXPECT().
					Create(m.r.Context(), gomock.AssignableToTypeOf(model.CreateOrderRequest{})).
					Return(&model.CreateOrderResponse{OrderID: 1, CustomerEmail: "test@example.com", Message: "success create order", TotalPrice: 200_000}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody: `{
				"success": true,
				"status": "success",
				"data": {
				  "order": 
					{
					  "order_id": 1,
					  "customer_email": "test@example.com",
					  "message": "success create order",
					  "total_price": 200000
					}
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "success hit api /api/v1/order [post] 'error unmarshal'",
			handler: &orderHandler{},
			params: params{
				payload: `{
					"customer_email-bad-key-not-enclosed-by-double-quoted:"test@example.com",
					"orders":[
						{
						  "name":"sate",
						  "qty":5
						},
					  {
						"name":"es lemon tea",
						"qty":5
					  },
					  {
						"name":"nasi pecel",
						"qty":5
					  }
					]
				  }`,
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "access-token"))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "success hit api /api/v1/order [post] 'error auth'",
			handler: &orderHandler{},
			params: params{
				payload: `{
					"customer_email":"test@example.com",
					"orders":[
						{
						  "name":"sate",
						  "qty":5
						},
					  {
						"name":"es lemon tea",
						"qty":5
					  },
					  {
						"name":"nasi pecel",
						"qty":5
					  }
					]
				  }`,
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "invalid-token")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "access-token"))
				m.orderServiceMock.EXPECT().Create(m.r.Context(), gomock.AssignableToTypeOf(model.CreateOrderRequest{})).Return(nil, apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "success hit api /api/v1/order [post] 'error not found'",
			handler: &orderHandler{},
			params: params{
				payload: `{
					"customer_email":"test@example.com",
					"orders":[
						{
						  "name":"not-found",
						  "qty":5
						},
					  {
						"name":"es lemon tea",
						"qty":5
					  },
					  {
						"name":"nasi pecel",
						"qty":5
					  }
					]
				  }`,
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "invalid-token")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "access-token"))
				m.orderServiceMock.EXPECT().Create(m.r.Context(), gomock.AssignableToTypeOf(model.CreateOrderRequest{})).Return(nil, apperrors.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "success hit api /api/v1/order [post] 'error internal server'",
			handler: &orderHandler{},
			params: params{
				payload: `{
					"customer_email":"test@example.com",
					"orders":[
						{
						  "name":"sate",
						  "qty":5
						},
					  {
						"name":"es lemon tea",
						"qty":5
					  },
					  {
						"name":"nasi pecel",
						"qty":5
					  }
					]
				  }`,
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "invalid-token")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "access-token"))
				m.orderServiceMock.EXPECT().Create(m.r.Context(), gomock.AssignableToTypeOf(model.CreateOrderRequest{})).Return(nil, errors.New("oops! error internal server"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"success":false,"status":"error","error":{"message":"oops! error"},"process_time":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			orderServiceMock := service.NewMockOrderService(ctrl)
			r := httptest.NewRequest(http.MethodPost, "/api/v1/order", strings.NewReader(tt.params.payload))
			w := httptest.NewRecorder()
			rctx := chi.NewRouteContext()
			m := &mocks{r: r, w: w, rctx: rctx, orderServiceMock: orderServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}
			tt.handler.orderService = m.orderServiceMock

			handler := tt.handler.Create()

			handler(w, r)

			// resetting processing time to 0 & error message to a unchanged string
			resp := w.Result()
			respBodyStr := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"error":{"message":".*"`, `"error":{"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, respBodyStr)
		})
	}
}

func Test_orderHandler_ConfirmPayment(t *testing.T) {
	type mocks struct {
		r                *http.Request
		w                *httptest.ResponseRecorder
		rctx             *chi.Context
		orderServiceMock *service.MockOrderService
	}
	type params struct {
		payload string
	}
	tests := []struct {
		name           string
		handler        *orderHandler
		params         params
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:    "success hit api /api/v1/order/confirm-payment [put] 'ok'",
			handler: &orderHandler{},
			params:  params{payload: `{"email":"test@example.com"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "access-token"))
				m.rctx.URLParams.Add("email", "test@example.com")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.orderServiceMock.EXPECT().ConfirmPayment(m.r.Context(), gomock.AssignableToTypeOf(model.ConfirmPaymentRequest{})).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"success":true,"status":"success","process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/order/confirm-payment [put] 'not found'",
			handler: &orderHandler{},
			params:  params{payload: `{"email":"not.found@example.com"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "access-token"))
				m.rctx.URLParams.Add("email", "not.found@example.com")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.orderServiceMock.EXPECT().ConfirmPayment(m.r.Context(), gomock.AssignableToTypeOf(model.ConfirmPaymentRequest{})).Return(apperrors.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/order/confirm [put] 'error auth'",
			handler: &orderHandler{},
			params:  params{payload: `{"email":"test@example.com"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "invalid-token")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "access-token"))
				m.rctx.URLParams.Add("email", "test@example.com")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.orderServiceMock.EXPECT().ConfirmPayment(m.r.Context(), gomock.AssignableToTypeOf(model.ConfirmPaymentRequest{})).Return(apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       `{"success":false,"status":"fail","error":{"message":"oops! error"},"process_time":0}`,
		},
		{
			name:    "fail hit api /api/v1/order/confirm/{email} [put] 'error internal server'",
			handler: &orderHandler{},
			params:  params{payload: `{"email":"test@example.com"}`},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/json")
				m.r.Header.Set("Authorization", "Bearer access-token")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "access-token"))
				m.rctx.URLParams.Add("email", "test@example.com")
				*m.r = *m.r.WithContext(context.WithValue(m.r.Context(), chi.RouteCtxKey, m.rctx))
				m.orderServiceMock.EXPECT().ConfirmPayment(m.r.Context(), gomock.AssignableToTypeOf(model.ConfirmPaymentRequest{})).Return(errors.New("oops! internal server error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"success":false,"status":"error","error":{"message":"oops! error"},"process_time":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			orderServiceMock := service.NewMockOrderService(ctrl)
			r := httptest.NewRequest(http.MethodPut, "/api/v1/order/confirm/", strings.NewReader(tt.params.payload))
			w := httptest.NewRecorder()
			rctx := chi.NewRouteContext()
			m := &mocks{r: r, w: w, rctx: rctx, orderServiceMock: orderServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}
			tt.handler.orderService = m.orderServiceMock

			handler := tt.handler.ConfirmPayment()

			handler(w, r)

			// resetting processing time to 0 & error message to a unchanged string
			resp := w.Result()
			respBodyStr := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"error":{"message":".*"`, `"error":{"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, respBodyStr)
		})
	}
}

func Test_orderHandler_Search(t *testing.T) {

	type params struct {
		menuNames  string
		exactNames string
		emails     string
		qty        string
		// status string
		// maxPrice string
		// minPrice string
		// startDay string
		// endDay string
	}
	type mocks struct {
		r                *http.Request
		w                *httptest.ResponseRecorder
		rctx             *chi.Context
		orderServiceMock *service.MockOrderService
	}
	prepareFullURL := func(u string, q params) (string, error) {
		URL, err := url.Parse(u)
		if err != nil {
			return "", err
		}

		val := URL.Query()
		val.Add("emails", q.emails)
		val.Add("menu-names", q.menuNames)
		val.Add("exact-names", q.exactNames)
		val.Add("qty", q.qty)
		// if q.emails != "" {
		// 	val.Add("emails", q.emails)
		// }
		// if q.menuNames != "" {
		// 	val.Add("menu-names", q.menuNames)
		// }
		// if q.exactNames != "" {
		// 	val.Add("exact-names", q.exactNames)
		// }
		// if q.qty != "" {
		// 	val.Add("qty", q.qty)
		// }
		URL.RawQuery = val.Encode()
		return URL.String(), nil
	}
	tests := []struct {
		name           string
		handler        *orderHandler
		params         params
		prepareMocks   func(*mocks)
		wantStatusCode int
		wantBody       string
	}{
		{
			name: "success hit api /api/v1/order/search?<query-params> [get] 'ok'",
			// menu-names={menus-names}&emails={emails}&exact-matches-names={exact-matches-names}&max-price={max-price}&min-price={min-price}&status={status}&qty={qty}&start-day={start-day}&end-day={end-day}
			handler: &orderHandler{},
			params: params{
				menuNames:  "sate,sop buah",
				exactNames: "true",
				emails:     "test@example.com",
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/x-www-form-url-encoded")
				m.r.Header.Set("Authorization", "Bearer access-token")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "access-token"))
				m.orderServiceMock.EXPECT().Search(m.r.Context(), gomock.AssignableToTypeOf(model.OrderQuery{})).Return(&model.SearchOrdersResponse{
					TotalPrice: 125_000,
					Orders: []*model.SearchResponse{
						{
							OrderID:       1,
							CustomerEmail: "test@example.com",
							MenuName:      "sate",
							Price:         20_000,
							Status:        2,
							Qty:           4,
						},
						{
							OrderID:       2,
							CustomerEmail: "test@example.com",
							MenuName:      "sop buah",
							Price:         15_000,
							Status:        1,
							Qty:           3,
						},
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody: `{
				"success": true,
				"status": "success",
				"data": {
				  "order": {
					"total_price":125000,
					  "orders": [
						{
						  "order_id":1,
						  "customer_email":"test@example.com",
						  "qty":4,
						  "price":20000,
						  "menu_name":"sate",
						  "status":2,
						  "created_at":""
						},
						{
						  "order_id":2,
						  "customer_email":"test@example.com",
						  "qty":3,
						  "price":15000,
						  "menu_name":"sop buah",
						  "status":1,
						  "created_at":""
						}
					  ]
				  }
				},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api /api/v1/order/search?<query-params> [get] 'not found'",
			handler: &orderHandler{},
			params: params{
				menuNames:  "sate,sop buah",
				exactNames: "true",
				emails:     "not.found@example.com",
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/x-www-form-url-encoded")
				m.r.Header.Set("Authorization", "Bearer access-token")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "access-token"))
				m.orderServiceMock.EXPECT().Search(m.r.Context(), gomock.AssignableToTypeOf(model.OrderQuery{})).Return(nil, apperrors.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody: `{
				"success": false,
				"status": "fail",
				"error":{"message":"oops! error"},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api /api/v1/order/search?<query-params> [get] 'error auth'",
			handler: &orderHandler{},
			params: params{
				menuNames:  "sate,sop buah",
				exactNames: "true",
				emails:     "not.found@example.com",
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/x-www-form-url-encoded")
				m.r.Header.Set("Authorization", "Bearer access-token")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "invalid-token"))
				m.orderServiceMock.EXPECT().Search(m.r.Context(), gomock.AssignableToTypeOf(model.OrderQuery{})).Return(nil, apperrors.ErrAuth)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBody: `{
				"success": false,
				"status": "fail",
				"error":{"message":"oops! error"},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api /api/v1/order/search?<query-params> [get] 'error invalid query params'",
			handler: &orderHandler{},
			params: params{
				menuNames:  "sate,sop buah",
				exactNames: "true",
				emails:     "not.found@example.com",
				qty:        "one", // must be numeric not words/letter
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/x-www-form-url-encoded")
				m.r.Header.Set("Authorization", "Bearer access-token")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "access-token"))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody: `{
				"success": false,
				"status": "fail",
				"error":{"message":"oops! error"},
				"process_time": 0
			  }`,
		},
		{
			name:    "fail hit api /api/v1/order/search?<query-params> [get] 'error invalid query params'",
			handler: &orderHandler{},
			params: params{
				menuNames:  "sate,sop buah",
				exactNames: "true",
				emails:     "not.found@example.com",
			},
			prepareMocks: func(m *mocks) {
				m.r.Header.Set("Content-Type", "application/x-www-form-url-encoded")
				m.r.Header.Set("Authorization", "Bearer access-token")
				*m.r = *m.r.WithContext(utils.ContextWithValue(m.r.Context(), "Authorization", "access-token"))
				m.orderServiceMock.EXPECT().Search(m.r.Context(), gomock.AssignableToTypeOf(model.OrderQuery{})).Return(nil, errors.New("oops! error internal server"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody: `{
				"success": false,
				"status": "error",
				"error":{"message":"oops! error"},
				"process_time": 0
			  }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			orderServiceMock := service.NewMockOrderService(ctrl)
			endpoint, err := prepareFullURL("/api/v1/order/search", tt.params)
			if err != nil {
				panic(err)
			}
			r := httptest.NewRequest(http.MethodGet, endpoint, nil)
			w := httptest.NewRecorder()
			rctx := chi.NewRouteContext()
			m := &mocks{r: r, w: w, rctx: rctx, orderServiceMock: orderServiceMock}
			if tt.prepareMocks != nil {
				tt.prepareMocks(m)
			}
			tt.handler.orderService = m.orderServiceMock

			handler := tt.handler.Search()

			handler(w, r)

			// resetting processing time to 0 & error message to a unchanged string
			resp := w.Result()
			respBodyStr := regexReplaceAllMultiple(w.Body.String(), `"process_time":\d+`, `"process_time":0`, `"message":".*"`, `"message":"oops! error"`)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.JSONEq(t, tt.wantBody, respBodyStr)
		})
	}
}
