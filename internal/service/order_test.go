package service

import (
	"context"
	"errors"
	"family-catering/internal/model"
	"family-catering/internal/repository"
	"family-catering/pkg/utils"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewOrderService(t *testing.T) {
	type args struct {
		orderRepo repository.OrderRepository
		menuRepo  repository.MenuRepository
	}
	tests := []struct {
		name string
		args args
	}{{name: "success NewOrderService"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, NewOrderService(tt.args.orderRepo, tt.args.menuRepo))
		})
	}
}

func Test_orderService_Create(t *testing.T) {
	type args struct {
		ctx context.Context
		req model.CreateOrderRequest
	}
	type mocks struct {
		utMocks       utils.Mock
		orderRepoMock *repository.MockOrderRepository
		menuRepoMock  *repository.MockMenuRepository
	}
	tests := []struct {
		name         string
		svc          *orderService
		args         args
		prepareMocks func(*mocks)
		wantResp     *model.CreateOrderResponse
		wantErr      bool
	}{
		{
			name: "success Create",
			svc:  &orderService{},
			args: args{
				ctx: context.Background(),
				req: model.CreateOrderRequest{
					CustomerEmail: "test@example.com",
					Orders:        []model.BaseOrderRequest{{Name: "Sop Iga", Qty: 4}, {Name: "Ayam Penyet", Qty: 5}},
				},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.menuRepoMock.EXPECT().Search(context.Background(), gomock.AssignableToTypeOf(model.MenuQuery{})).
					Return([]*model.Menu{
						{ID: 83, Name: "Sop Iga", Price: 60_000, Categories: "Indonesian food"},
						{ID: 20, Name: "Ayam Penyet", Price: 20_000, Categories: "Indonesian food"},
					}, nil, nil)
				m.orderRepoMock.EXPECT().Create(context.Background(), gomock.AssignableToTypeOf([]*model.Order{})).Return(int64(2), int64(1), nil)
			},
			wantResp: &model.CreateOrderResponse{
				OrderID:       1,
				CustomerEmail: "test@example.com",
				Message:       "success create orders",
				TotalPrice:    340_000,
			},
		},
		{
			name: "fail Create (partially/all no row)",
			svc:  &orderService{},
			args: args{
				ctx: context.Background(),
				req: model.CreateOrderRequest{
					CustomerEmail: "test@example.com",
					Orders:        []model.BaseOrderRequest{{Name: "not-exists-menu-name", Qty: 4}, {Name: "Ayam Penyet", Qty: 5}},
				},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.menuRepoMock.EXPECT().Search(context.Background(), gomock.AssignableToTypeOf(model.MenuQuery{})).
					Return([]*model.Menu{
						{ID: 20, Name: "Ayam Penyet", Price: 20_000, Categories: "Indonesian food"},
					}, nil, nil)
			},
			wantErr: true,
		},
		{
			name: "fail Create (invalid/no token)",
			svc:  &orderService{},
			args: args{
				ctx: context.Background(),
				req: model.CreateOrderRequest{
					CustomerEmail: "test@example.com",
					Orders:        []model.BaseOrderRequest{{Name: "Sop Iga", Qty: 4}, {Name: "Ayam Penyet", Qty: 5}},
				},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "invalid-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return nil, errors.New("oops! invalid token")
				})
			},
			wantErr: true,
		},
		{
			name: "fail Create",
			svc:  &orderService{},
			args: args{
				ctx: context.Background(),
				req: model.CreateOrderRequest{
					CustomerEmail: "test@example.com",
					Orders:        []model.BaseOrderRequest{{Name: "Sop Iga", Qty: 4}, {Name: "Ayam Penyet", Qty: 5}},
				},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.menuRepoMock.EXPECT().Search(context.Background(), gomock.AssignableToTypeOf(model.MenuQuery{})).
					Return([]*model.Menu{
						{ID: 83, Name: "Sop Iga", Price: 60_000, Categories: "Indonesian food"},
						{ID: 20, Name: "Ayam Penyet", Price: 20_000, Categories: "Indonesian food"},
					}, nil, nil)
				m.orderRepoMock.EXPECT().Create(context.Background(), gomock.AssignableToTypeOf([]*model.Order{})).Return(int64(0), int64(0), errors.New("oops! db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			utMock := utils.InitMock()
			menuRepoMock := repository.NewMockMenuRepository(ctrl)
			orderRepoMock := repository.NewMockOrderRepository(ctrl)

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{menuRepoMock: menuRepoMock, orderRepoMock: orderRepoMock, utMocks: utMock})
			}

			tt.svc.menuRepo = menuRepoMock
			tt.svc.orderRepo = orderRepoMock

			gotResp, err := tt.svc.Create(tt.args.ctx, tt.args.req)

			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantResp, gotResp)

			utMock.UnpatchAll()
		})
	}
}

func Test_orderService_CancelUnpaidOrder(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type mocks struct {
		utMocks       utils.Mock
		orderRepoMock *repository.MockOrderRepository
	}
	tests := []struct {
		name         string
		svc          *orderService
		args         args
		prepareMocks func(*mocks)
		wantResp     *model.CancelUnpaidOrderResponse
		wantErr      bool
	}{
		{
			name: "success CancelUnpaidOrder",
			svc:  &orderService{},
			args: args{ctx: context.Background()},
			prepareMocks: func(m *mocks) {
				m.orderRepoMock.EXPECT().CancelUnpaidOrder(context.Background()).Return(int64(5), nil)
			},
			wantResp: &model.CancelUnpaidOrderResponse{
				Message:             "success cancel unpaid order",
				TotalOrderCancelled: 5,
			},
		},
		{
			name: "fail CancelUnpaidOrder",
			svc:  &orderService{},
			args: args{ctx: context.Background()},
			prepareMocks: func(m *mocks) {
				m.orderRepoMock.EXPECT().CancelUnpaidOrder(context.Background()).Return(int64(0), errors.New("oops! db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			orderRepoMock := repository.NewMockOrderRepository(ctrl)
			utMocks := utils.InitMock()

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{orderRepoMock: orderRepoMock, utMocks: utMocks})
			}

			tt.svc.orderRepo = orderRepoMock

			gotResp, err := tt.svc.CancelUnpaidOrder(tt.args.ctx)

			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantResp, gotResp)

			utMocks.UnpatchAll()
		})
	}
}

func Test_orderService_ConfirmPayment(t *testing.T) {
	type args struct {
		ctx context.Context
		req model.ConfirmPaymentRequest
	}
	type mocks struct {
		utMocks       utils.Mock
		orderRepoMock *repository.MockOrderRepository
	}
	tests := []struct {
		name         string
		svc          *orderService
		args         args
		prepareMocks func(*mocks)
		wantErr      bool
	}{
		{
			name: "success ConfirmPayment",
			svc:  &orderService{},
			args: args{ctx: context.Background(), req: model.ConfirmPaymentRequest{Email: "test@example.com"}},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.utMocks.Patch("ValidateRequest", func(s interface{}) error { return nil })
				m.orderRepoMock.EXPECT().ConfirmPayment(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf("")).Return(int64(1), nil, nil)
			},
		},
		{
			name: "fail ConfirmPayment (not rows)",
			svc:  &orderService{},
			args: args{ctx: context.Background(), req: model.ConfirmPaymentRequest{Email: "not.found@example.com"}},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.utMocks.Patch("ValidateRequest", func(s interface{}) error { return nil })
				m.orderRepoMock.EXPECT().ConfirmPayment(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf("")).Return(int64(0), errors.New("oops! error no rows"), nil)
			},
			wantErr: true,
		},
		{
			name: "fail ConfirmPayment (invalid/no token)",
			svc:  &orderService{},
			args: args{ctx: context.Background(), req: model.ConfirmPaymentRequest{Email: "test@example.com"}},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "invalid-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return nil, errors.New("oops! invalid token")
				})

			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			orderRepoMock := repository.NewMockOrderRepository(ctrl)
			utMocks := utils.InitMock()

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{orderRepoMock: orderRepoMock, utMocks: utMocks})
			}

			tt.svc.orderRepo = orderRepoMock

			err := tt.svc.ConfirmPayment(tt.args.ctx, tt.args.req)

			assert.Equal(t, tt.wantErr, err != nil)

			utMocks.UnpatchAll()
		})
	}
}

func Test_orderService_Search(t *testing.T) {
	type args struct {
		ctx context.Context
		req model.OrderQuery
	}
	type mocks struct {
		utMocks       utils.Mock
		orderRepoMock *repository.MockOrderRepository
	}
	tests := []struct {
		name         string
		svc          *orderService
		args         args
		prepareMocks func(*mocks)
		wantResp     *model.SearchOrdersResponse
		wantErr      bool
	}{
		{
			name: "success Search",
			svc:  &orderService{},
			args: args{ctx: context.Background(), req: model.OrderQuery{CustomerEmails: []string{"test1@example.com", "test2@example.com"}}},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.orderRepoMock.EXPECT().
					Search(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(model.OrderQuery{})).
					Return([]*model.Order{
						{
							CustomerEmail: "test1@example.com",
							MenuName:      "nasi kepal isi salmon",
							Price:         44_000,
							Qty:           2,
							Status:        2,
						},
						{
							CustomerEmail: "test2@example.com",
							MenuName:      "soto kambing",
							Price:         30_000,
							Qty:           1,
							Status:        2,
						},
						{
							CustomerEmail: "test2@example.com",
							MenuName:      "nasi lemak",
							Price:         20_000,
							Qty:           2,
							Status:        2,
						},
					}, nil, nil)
			},
			wantResp: &model.SearchOrdersResponse{
				Orders: []*model.SearchResponse{
					{
						CustomerEmail: "test1@example.com",
						MenuName:      "nasi kepal isi salmon",
						Price:         44_000,
						Qty:           2,
						Status:        2,
					},
					{
						CustomerEmail: "test2@example.com",
						MenuName:      "soto kambing",
						Price:         30_000,
						Qty:           1,
						Status:        2,
					},
					{
						CustomerEmail: "test2@example.com",
						MenuName:      "nasi lemak",
						Price:         20_000,
						Qty:           2,
						Status:        2,
					},
				},
				TotalPrice: 158_000,
			},
		},
		{
			name: "fail Search (partially/no rows)",
			svc:  &orderService{},
			args: args{ctx: context.Background(), req: model.OrderQuery{CustomerEmails: []string{"not.found@example.com"}}},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.orderRepoMock.EXPECT().
					Search(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(model.OrderQuery{})).
					Return(nil, errors.New("oops! no rows"), nil)
			},
			wantErr: true,
		},
		{
			name: "fail Search (invalid/no token)",
			svc:  &orderService{},
			args: args{ctx: context.Background(), req: model.OrderQuery{CustomerEmails: []string{"not.found@example.com"}}},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "invalid-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return nil, errors.New("oops! invalid token")
				})
			},
			wantErr: true,
		},
		{
			name: "fail Search ",
			svc:  &orderService{},
			args: args{ctx: context.Background(), req: model.OrderQuery{CustomerEmails: []string{"not.found@example.com"}}},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.orderRepoMock.EXPECT().
					Search(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(model.OrderQuery{})).
					Return(nil, nil, errors.New("oops! db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			orderRepoMock := repository.NewMockOrderRepository(ctrl)
			utMocks := utils.InitMock()

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{orderRepoMock: orderRepoMock, utMocks: utMocks})
			}

			tt.svc.orderRepo = orderRepoMock

			gotResp, err := tt.svc.Search(tt.args.ctx, tt.args.req)

			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantResp, gotResp)

			utMocks.UnpatchAll()
		})
	}
}
