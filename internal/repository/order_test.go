package repository

import (
	"context"
	"database/sql/driver"
	"errors"
	"family-catering/internal/model"
	"family-catering/pkg/db/postgres"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// make n sqlmock.Argument by calling sqlmock.AnyArg() and assign it returned value to slice
// will return slice of driver.Value
// should be use only for testing purpose
func makeNAnyArgs(nRow int, nCols int) []driver.Value {
	n := nRow * nCols
	args := make([]driver.Value, n)
	for i := 0; i < n; i++ {
		args[i] = sqlmock.AnyArg()
	}

	return args
}

func TestNewOrderRepository(t *testing.T) {
	type args struct {
		postgres postgres.PostgresClient
	}
	tests := []struct {
		name string
		args args
	}{{name: "success NewOrderRepository"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, NewOrderRepository(tt.args.postgres))
		})
	}
}

func Test_orderRepository_Create(t *testing.T) {
	type args struct {
		ctx    context.Context
		orders []*model.Order
	}
	type mocks struct {
		numOfOrders int
		pgMock      sqlmock.Sqlmock
	}
	tests := []struct {
		name            string
		repo            *orderRepository
		args            args
		prepareMocks    func(*mocks)
		wantBaseOrderId int64 // num of order per menu_id
		wantOrderID     int64 // order_id
		wantErr         bool
	}{
		{
			name: "success Create",
			repo: &orderRepository{},
			args: args{
				ctx: context.Background(),
				orders: []*model.Order{
					{CustomerEmail: "test@examle.com", MenuID: 1, MenuName: "Sate", Price: 25_000, Qty: 15, Status: 0},
					{CustomerEmail: "test@examle.com", MenuID: 4, MenuName: "Bebek Bakar", Price: 75_000, Qty: 3, Status: 0},
					{CustomerEmail: "test@examle.com", MenuID: 16, MenuName: "Pindang Ikan Kakap", Price: 45_000, Qty: 5, Status: 0}},
			},
			prepareMocks: func(m *mocks) {
				args := makeNAnyArgs(m.numOfOrders, 6) // 6 is number of cols inserted (see orderRepository.orderMenuInsertQuery at ./order.go )
				m.pgMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "order"`)).
					WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"base_order_id", "order_id"}).AddRow(3, 1)).
					WillReturnError(nil)
			},
			wantBaseOrderId: 3,
			wantOrderID:     1,
		},
		{
			name: "fail Create",
			repo: &orderRepository{},
			args: args{
				ctx: context.Background(),
				orders: []*model.Order{
					{CustomerEmail: "test@examle.com", MenuID: 1, MenuName: "Sate", Price: 25_000, Qty: 15, Status: 0},
					{CustomerEmail: "test@examle.com", MenuID: 4, MenuName: "Bebek Bakar", Price: 75_000, Qty: 3, Status: 0},
					{CustomerEmail: "test@examle.com", MenuID: 16, MenuName: "Pindang Ikan Kakap", Price: 45_000, Qty: 5, Status: 0}},
			},
			prepareMocks: func(m *mocks) {
				args := makeNAnyArgs(m.numOfOrders, 6) // 6 is number of cols inserted (see orderRepository.orderMenuInsertQuery at ./order.go )
				m.pgMock.ExpectQuery(`INSERT INTO "order"`).
					WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"base_order_id", "order_id"})).
					WillReturnError(errors.New("oops! db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, pgMock, err := sqlmock.New()
			if err != nil {
				panic(err)
			}

			tt.repo.postgres = db

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{pgMock: pgMock, numOfOrders: len(tt.args.orders)})
			}

			gotBaseOrderId, gotOrderID, err := tt.repo.Create(tt.args.ctx, tt.args.orders)
			assert.Equal(t, tt.wantBaseOrderId, gotBaseOrderId)
			assert.Equal(t, tt.wantOrderID, gotOrderID)
			assert.Equal(t, tt.wantErr, err != nil, err)

		})
	}
}

func Test_orderRepository_CancelUnpaidOrder(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name         string
		repo         *orderRepository
		args         args
		prepareMocks func(*mocks)
		want         int64
		wantErr      bool
	}{
		{
			name: "success CancelUnpaidOrder",
			repo: &orderRepository{},
			args: args{ctx: context.Background()},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec(`UPDATE "order".*status = 1.*`).WillReturnResult(sqlmock.NewResult(0, 10)).WillReturnError(nil)
			},
			want: 10,
		},
		{
			name: "fail CancelUnpaidOrder",
			repo: &orderRepository{},
			args: args{ctx: context.Background()},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec(`UPDATE "order".*status = 1.*`).WillReturnResult(sqlmock.NewResult(0, 0)).WillReturnError(errors.New("oop! error db"))
			},
			wantErr: true,
		},
		{
			name: "fail CancelUnpaidOrder (rows affected error)",
			repo: &orderRepository{},
			args: args{ctx: context.Background()},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec(`UPDATE "order".*status = 1.*`).WillReturnResult(sqlmock.NewErrorResult(errors.New("oop! row affected error")))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, pgMock, err := sqlmock.New()
			if err != nil {
				panic(err)
			}

			tt.repo.postgres = db

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{pgMock: pgMock})
			}

			got, err := tt.repo.CancelUnpaidOrder(tt.args.ctx)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_orderRepository_ConfirmPayment(t *testing.T) {
	type args struct {
		ctx   context.Context
		email string
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name          string
		repo          *orderRepository
		args          args
		prepareMocks  func(*mocks)
		wantNAffected int64
		wantErr       bool
	}{
		{
			name: "success ConfirmPayment",
			repo: &orderRepository{},
			args: args{ctx: context.Background(), email: "test@example.com"},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec(`UPDATE "order".*status = 2.*email.*`).WillReturnResult(sqlmock.NewResult(0, 1)).WillReturnError(nil)
			},
			wantNAffected: 1,
		},
		{
			name: "fail ConfirmPayment (no rows affected)",
			repo: &orderRepository{},
			args: args{ctx: context.Background(), email: "not.exists@example.com"},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec(`UPDATE "order".*status = 2.*email.*`).WithArgs("not.exists@example.com").WillReturnResult(sqlmock.NewResult(0, 0)).WillReturnError(nil)
			},
			wantErr: true,
		},
		{
			name: "fail ConfirmPayment (rows affected error)",
			repo: &orderRepository{},
			args: args{ctx: context.Background(), email: "not.exists@example.com"},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec(`UPDATE "order".*status = 2.*email.*`).WithArgs("not.exists@example.com").WillReturnResult(sqlmock.NewErrorResult(errors.New("oops! row affected error"))).WillReturnError(nil)
			},
			wantErr: true,
		},
		{
			name: "fail ConfirmPayment (db error)",
			repo: &orderRepository{},
			args: args{ctx: context.Background(), email: "not.exists@example.com"},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec(`UPDATE "order".*status = 2.*email.*`).WithArgs("not.exists@example.com").WillReturnResult(sqlmock.NewResult(0, 0)).WillReturnError(errors.New("oops! row affected error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, pgMock, err := sqlmock.New()
			if err != nil {
				panic(err)
			}

			tt.repo.postgres = db

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{pgMock: pgMock})
			}

			gotNAffected, errNoRow, err := tt.repo.ConfirmPayment(tt.args.ctx, tt.args.email)
			assert.Equal(t, tt.wantErr, (err != nil || errNoRow != nil))
			assert.Equal(t, tt.wantNAffected, gotNAffected)
		})
	}
}

func Test_orderRepository_Search(t *testing.T) {
	type args struct {
		ctx   context.Context
		order model.OrderQuery
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name         string
		repo         *orderRepository
		args         args
		prepareMocks func(*mocks)
		wantOrders   []*model.Order
		wantErr      bool
	}{
		{
			name: "success Search",
			repo: &orderRepository{},
			args: args{
				ctx: context.Background(),
				order: model.OrderQuery{
					MenuNames:           []string{"nasi", "soto"},
					ExactMenuNamesMatch: false,
					MinPrice:            10_000,
					MaxPrice:            100_000,
				},
			},
			prepareMocks: func(m *mocks) {
				args := makeNAnyArgs(3, 1)
				m.pgMock.ExpectQuery(`SELECT.*"order".*WHERE`).
					WithArgs(args...).
					WillReturnRows(
						sqlmock.NewRows([]string{"order_id", "base_order_id", "menu_name", "customer_email",
							"price", "qty", "created_at", "updated_at"}).
							AddRow(22, 48, "nasi lemak", "test1@example.com", 20_000, 5, "", "").
							AddRow(3, 7, "nasi goreng", "test6@example.com", 25_000, 1, "", "").
							AddRow(56, 92, "soto betawi", "test56@example.com", 40_000, 3, "", "").
							AddRow(19, 32, "soto babat", "test112@example.com", 30_000, 10, "", "")).
					WillReturnError(nil)
			},
			wantOrders: []*model.Order{
				{OrderID: 22, BaseOrderID: 48, MenuName: "nasi lemak", CustomerEmail: "test1@example.com", Price: 20_000, Qty: 5},
				{OrderID: 3, BaseOrderID: 7, MenuName: "nasi goreng", CustomerEmail: "test6@example.com", Price: 25_000, Qty: 1},
				{OrderID: 56, BaseOrderID: 92, MenuName: "soto betawi", CustomerEmail: "test56@example.com", Price: 40_000, Qty: 3},
				{OrderID: 19, BaseOrderID: 32, MenuName: "soto babat", CustomerEmail: "test112@example.com", Price: 30_000, Qty: 10},
			},
		},
		{
			name: "fail Search (no rows)",
			repo: &orderRepository{},
			args: args{
				ctx: context.Background(),
				order: model.OrderQuery{
					MenuNames: []string{"not-exists-name-1", "not-exists-name-2"},
					MinPrice:  10_000,
					MaxPrice:  100_000,
				},
			},
			prepareMocks: func(m *mocks) {
				args := makeNAnyArgs(3, 1)
				m.pgMock.ExpectQuery(`SELECT.*"order".*WHERE`).
					WithArgs(args...).
					WillReturnRows(
						sqlmock.NewRows([]string{"order_id", "base_order_id", "menu_name", "customer_email",
							"price", "qty", "created_at", "updated_at"})).
					WillReturnError(nil)
			},
			wantErr: true,
		},
		{
			name: "fail Search (db error)",
			repo: &orderRepository{},
			args: args{
				ctx: context.Background(),
				order: model.OrderQuery{
					MenuNames: []string{"nasi", "soto"},
					MinPrice:  10_000,
					MaxPrice:  100_000,
				},
			},
			prepareMocks: func(m *mocks) {
				args := makeNAnyArgs(3, 1)
				m.pgMock.ExpectQuery(`SELECT.*"order".*WHERE`).
					WithArgs(args...).
					WillReturnRows(
						sqlmock.NewRows([]string{"order_id", "base_order_id", "menu_name", "customer_email",
							"price", "qty", "created_at", "updated_at"}).
							AddRow(22, 48, "nasi lemak", "test1@example.com", 20_000, 5, "", "").
							AddRow(3, 7, "nasi goreng", "test6@example.com", 25_000, 1, "", "").
							AddRow(56, 92, "soto betawi", "test56@example.com", 40_000, 3, "", "").
							AddRow(19, 32, "soto babat", "test112@example.com", 30_000, 10, "", "")).
					WillReturnError(errors.New("oops! db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, pgMock, err := sqlmock.New()
			if err != nil {
				panic(err)
			}

			tt.repo.postgres = db

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{pgMock: pgMock})
			}

			gotOrders, errNoRow, err := tt.repo.Search(tt.args.ctx, tt.args.order)
			assert.Equal(t, tt.wantErr, (err != nil || errNoRow != nil), err)
			assert.Equal(t, tt.wantOrders, gotOrders)
		})
	}
}
