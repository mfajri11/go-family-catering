package repository

import (
	"context"
	"database/sql"
	"errors"
	"family-catering/internal/model"
	"family-catering/pkg/db/postgres"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestNewMenuRepository(t *testing.T) {
	type args struct {
		postgres postgres.PostgresClient
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success NewMenuRepository",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMenuRepository(tt.args.postgres)
			assert.NotNil(t, got)
		})
	}
}

func Test_menuRepository_GetByID(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name         string
		repo         *menuRepository
		args         args
		prepareMocks func(*mocks)
		wantMenu     *model.Menu
		wantErr      bool
	}{
		{
			name: "success GetByID",
			repo: &menuRepository{},
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+FROM.+menu.+id.+").
					WithArgs(int64(1)).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "name", "price", "categories"}).
							AddRow(int64(1), "sate", float32(25_000), "Indonesian food")).WillReturnError(nil)
			},
			wantMenu: &model.Menu{
				ID:         1,
				Name:       "sate",
				Price:      25_000,
				Categories: "Indonesian food",
			},
		},
		{
			name: "fail GetByID (no row)",
			repo: &menuRepository{},
			args: args{
				ctx: context.Background(),
				id:  1_000_000_000_000_000, // assume no menu with this id
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+FROM.+menu.+id.+").
					WithArgs(int64(1_000_000_000_000_000)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "categories"})).WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "fail GetByID (db error)",
			repo: &menuRepository{},
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+FROM.+menu.+id.+").
					WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "categories"})).WillReturnError(errors.New("oops! db error"))
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

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{pgMock: pgMock})
			}

			tt.repo.postgres = db
			gotMenu, errNoRow, err := tt.repo.GetByID(tt.args.ctx, tt.args.id)

			assert.Equal(t, tt.wantErr, (err != nil || errNoRow != nil))
			assert.Equal(t, tt.wantMenu, gotMenu)
		})
	}
}

func Test_menuRepository_GetByName(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name         string
		repo         *menuRepository
		args         args
		prepareMocks func(*mocks)
		wantMenu     *model.Menu
		wantErr      bool
	}{
		{
			name: "success GetByName",
			repo: &menuRepository{},
			args: args{
				ctx:  context.Background(),
				name: "sate",
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+FROM.+menu.+name.+").
					WithArgs("sate").
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "name", "price", "categories"}).
							AddRow(int64(1), "sate", float32(25_000), "Indonesian food")).WillReturnError(nil)
			},
			wantMenu: &model.Menu{
				ID:         1,
				Name:       "sate",
				Price:      25_000,
				Categories: "Indonesian food",
			},
		},
		{
			name: "fail GetByName (no row)",
			repo: &menuRepository{},
			args: args{
				ctx:  context.Background(),
				name: "not-exists-name-in-db", // assume no menu with this id
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+FROM.+menu.+name.+").
					WithArgs("not-exists-name-in-db").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "categories"})).WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "fail GetByName (no rows)",
			repo: &menuRepository{},
			args: args{
				ctx:  context.Background(),
				name: "sate",
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+FROM.+menu.+name.+").
					WithArgs("sate").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "categories"}).CloseError(sql.ErrNoRows))
			},
			wantErr: true,
		},
		{
			name: "fail GetByName (db error)",
			repo: &menuRepository{},
			args: args{
				ctx:  context.Background(),
				name: "sate",
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+FROM.+menu.+name.+").
					WithArgs("sate").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "categories"})).WillReturnError(errors.New("oops! db error"))
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

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{pgMock: pgMock})
			}

			tt.repo.postgres = db
			gotMenu, errNoRow, err := tt.repo.GetByName(tt.args.ctx, tt.args.name)

			assert.Equal(t, tt.wantErr, (err != nil || errNoRow != nil))
			assert.Equal(t, tt.wantMenu, gotMenu)
		})
	}
}

func Test_menuRepository_List(t *testing.T) {
	type args struct {
		ctx    context.Context
		limit  int
		offset int
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name         string
		repo         *menuRepository
		args         args
		prepareMocks func(*mocks)
		wantMenu     []*model.Menu
		wantErr      bool
	}{
		{
			name: "success GetList menu",
			repo: &menuRepository{},
			args: args{
				ctx:    context.Background(),
				limit:  2,
				offset: 1,
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+menu.+LIMIT.+OFFSET").
					WithArgs(2, 1).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "name", "price", "categories"}).
							AddRow(1, "sate", "25_000", "Indonesian food").
							AddRow(2, "rendang", "35_000", "Indonesian food"),
					).
					WillReturnError(nil)
			},
			wantMenu: []*model.Menu{
				{
					ID:         1,
					Name:       "sate",
					Price:      25_000,
					Categories: "Indonesian food",
				},
				{
					ID:         2,
					Name:       "rendang",
					Price:      35_000,
					Categories: "Indonesian food",
				},
			},
		},
		{
			name: "fail GetList menu (db error)",
			repo: &menuRepository{},
			args: args{
				ctx:    context.Background(),
				limit:  -2,
				offset: -100,
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+menu.+LIMIT.+OFFSET").
					WithArgs(-2, -1).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "name", "price", "categories"})).
					WillReturnError(errors.New("oops! db error"))
			},
			wantErr: true,
		},
		{
			name: "fail GetList menu (db error row error)",
			repo: &menuRepository{},
			args: args{
				ctx:    context.Background(),
				limit:  2,
				offset: 1,
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+menu.+LIMIT.+OFFSET").
					WithArgs(2, 1).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "name", "price", "categories"}).
							AddRow(1, "sate", 25_000, "Indonesian food").
							AddRow(2, "rendang", int(35_000), "Indonesian food").
							RowError(1, errors.New("oops! dbe error unmatched type for price columns"))). // price must be float32 not int & start from zero
					WillReturnError(nil)
			},
			wantErr: true,
		},
		{
			name: "fail GetList menu (db error scan error)",
			repo: &menuRepository{},
			args: args{
				ctx:    context.Background(),
				limit:  2,
				offset: 1,
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+menu.+LIMIT.+OFFSET").
					WithArgs(2, 1).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "name", "price", "categories"}).
							AddRow(nil, nil, nil, nil)).
					WillReturnError(nil)
			},
			wantErr: true,
		},
		{
			name: "fail GetList menu (no rows)",
			repo: &menuRepository{},
			args: args{
				ctx:    context.Background(),
				limit:  2,
				offset: 1,
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+menu.+LIMIT.+OFFSET").
					WithArgs(2, 1).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "name", "price", "categories"})).
					WillReturnError(nil)
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

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{pgMock: pgMock})
			}

			tt.repo.postgres = db
			gotMenu, errNoRow, err := tt.repo.List(tt.args.ctx, tt.args.limit, tt.args.offset)

			assert.Equal(t, tt.wantErr, (err != nil || errNoRow != nil), err)
			assert.Equal(t, tt.wantMenu, gotMenu)
		})
	}
}

func Test_menuRepository_Create(t *testing.T) {
	type args struct {
		ctx  context.Context
		menu model.Menu
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name         string
		repo         *menuRepository
		args         args
		prepareMocks func(*mocks)
		wantId       int64
		wantErr      bool
	}{
		{
			name: "success Create menu",
			repo: &menuRepository{},
			args: args{
				ctx: context.Background(),
				menu: model.Menu{
					Name:       "sate",
					Price:      25_000,
					Categories: "Indonesian food",
				},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("INSERT INTO menu.+").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1))).
					WillReturnError(nil)
			},
			wantId: 1,
		},
		{
			name: "fail Create menu",
			repo: &menuRepository{},
			args: args{
				ctx: context.Background(),
				menu: model.Menu{
					Name:       "sate",
					Price:      25_000,
					Categories: "Indonesian food",
				},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("INSERT INTO menu.+").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1))).
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

			gotId, err := tt.repo.Create(tt.args.ctx, tt.args.menu)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantId, gotId)
		})
	}
}

func Test_menuRepository_Update(t *testing.T) {
	type args struct {
		ctx  context.Context
		menu model.Menu
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name          string
		repo          *menuRepository
		args          args
		prepareMocks  func(*mocks)
		wantNAffected int64
		wantErr       bool
	}{
		{
			name: "success Update menu",
			repo: &menuRepository{},
			args: args{
				ctx: context.Background(),
				menu: model.Menu{
					ID:    1,
					Name:  "sate padang",
					Price: 40_000,
				},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec("UPDATE menu.+id.+").
					WillReturnResult(sqlmock.NewResult(0, 1)).
					WillReturnError(nil)
			},
			wantNAffected: 1,
		},
		{
			name: "fail Update menu (no rows)",
			repo: &menuRepository{},
			args: args{
				ctx: context.Background(),
				menu: model.Menu{
					ID:    1_000_000_000_000, // assume no menu with this id
					Name:  "sate padang",
					Price: 40_000,
				},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec("UPDATE menu.+id.+").
					WillReturnError(nil) // no error no nAffected
			},
			wantNAffected: 0,
			wantErr:       true, // err nil and 0 nAffected consider as errNoRows
		},
		{
			name: "fail Update menu (db error)",
			repo: &menuRepository{},
			args: args{
				ctx: context.Background(),
				menu: model.Menu{
					ID:    1, // assume no menu with this id
					Name:  "sate padang",
					Price: 40_000,
				},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec("UPDATE menu.+id.+").
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

			gotNAffected, errNoRow, err := tt.repo.Update(tt.args.ctx, tt.args.menu)

			assert.Equal(t, tt.wantErr, (err != nil || errNoRow != nil))
			assert.Equal(t, tt.wantNAffected, gotNAffected)

		})
	}
}

func Test_menuRepository_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name          string
		repo          *menuRepository
		args          args
		prepareMocks  func(*mocks)
		wantNAffected int64
		wantErr       bool
	}{
		{
			name: "succes Delete menu",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			repo: &menuRepository{},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec("DELETE FROM menu.+id.+").
					WithArgs(int64(1)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantNAffected: 1,
		},
		{
			name: "fail Delete menu (no rows)",
			args: args{
				ctx: context.Background(),
				id:  1_000_000_000_000_000,
			},
			repo: &menuRepository{},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec("DELETE FROM menu.+id.+").
					WithArgs(int64(1_000_000_000_000_000)).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantNAffected: 0,
			wantErr:       true,
		},
		{
			name: "fail Delete menu (no rows)",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			repo: &menuRepository{},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec("DELETE FROM menu.+id.+").
					WithArgs(int64(1_000_000_000_000_000)).
					WillReturnResult(sqlmock.NewResult(0, 0)).
					WillReturnError(errors.New("oops! db error"))
			},
			wantNAffected: 0,
			wantErr:       true,
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

			gotNAffected, errNoRow, err := tt.repo.Delete(tt.args.ctx, tt.args.id)

			assert.Equal(t, tt.wantErr, (err != nil || errNoRow != nil))
			assert.Equal(t, tt.wantNAffected, gotNAffected)
		})
	}
}

func Test_menuRepository_Search(t *testing.T) {
	type args struct {
		ctx  context.Context
		menu model.MenuQuery
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name         string
		repo         *menuRepository
		args         args
		prepareMocks func(*mocks)
		wantMenus    []*model.Menu
		wantErr      bool
	}{
		{
			name: "success search menu",
			repo: &menuRepository{},
			args: args{
				ctx: context.Background(),
				menu: model.MenuQuery{
					Names:           []string{"sate", "nasi goreng"},
					ExactNamesMatch: false,
					MaxPrice:        100_000,
					MinPrice:        10_000,
					Categories:      "Indonesian food",
				},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+FROM menu WHERE").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "categories"}).
						AddRow(23, "nasi goreng extra pedas", float32(55_000), "Indonesian food"))
			},
			wantMenus: []*model.Menu{{ID: 23, Name: "nasi goreng extra pedas", Price: 55_000, Categories: "Indonesian food"}},
		},
		{
			name: "fail search menu (no row)",
			repo: &menuRepository{},
			args: args{
				ctx: context.Background(),
				menu: model.MenuQuery{
					Names:           []string{"not-exist-food-name", "we-don't-have-this-food"},
					ExactNamesMatch: true,
					MaxPrice:        100_000,
					MinPrice:        10_000,
					Categories:      "Spicy",
				},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+FROM menu WHERE").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "categories"})).WillReturnError(nil)
			},
			wantErr: true,
		},
		{
			name: "fail search menu (error row)",
			repo: &menuRepository{},
			args: args{
				ctx: context.Background(),
				menu: model.MenuQuery{
					Names:           []string{"nasi goreng asin", "sate"},
					ExactNamesMatch: true,
					MaxPrice:        100_000,
					MinPrice:        10_000,
					Categories:      "Indonesian food,Spicy",
				},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+FROM menu WHERE").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "name", "price", "categories"}).
							AddRow(14, "nasi goreng asin", 55_000, "Indonesian food").
							AddRow(1, "sate", int(25_000), "Indonesian food").
							RowError(1, errors.New("oops! error mismatch type of column price"))).
					WillReturnError(nil)
			},
			wantErr: true,
		},
		{
			name: "fail search menu (scan error)",
			repo: &menuRepository{},
			args: args{
				ctx: context.Background(),
				menu: model.MenuQuery{
					Names:           []string{"nasi goreng asin", "sate"},
					ExactNamesMatch: true,
					MaxPrice:        100_000,
					MinPrice:        10_000,
					Categories:      "Indonesian food,Spicy",
				},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+FROM menu WHERE").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "name", "price", "categories"}).
							AddRow(14, "nasi goreng asin", 55_000, "Indonesian food").
							AddRow(1, "sate", "twenty five thousand rupiah", "Indonesian food")).
					WillReturnError(nil)
			},
			wantErr: true,
		},
		{
			name: "fail search menu (error query)",
			repo: &menuRepository{},
			args: args{
				ctx: context.Background(),
				menu: model.MenuQuery{
					Names:           []string{"nasi goreng asin", "sate"},
					ExactNamesMatch: true,
					MaxPrice:        100_000,
					MinPrice:        10_000,
					Categories:      "Indonesian food,Spicy",
				},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("SELECT.+FROM menu WHERE").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "name", "price", "categories"})).
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

			gotMenus, errNoRow, err := tt.repo.Search(tt.args.ctx, tt.args.menu)
			assert.Equal(t, tt.wantErr, (err != nil || errNoRow != nil), err)
			assert.Equal(t, tt.wantMenus, gotMenus, fmt.Sprintf("%v", gotMenus))

		})
	}
}
