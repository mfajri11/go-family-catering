package repository

import (
	"context"
	"database/sql"
	"errors"
	"family-catering/internal/model"
	"family-catering/pkg/db/postgres"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestNewOwnerRepository(t *testing.T) {
	type args struct {
		postgreClient postgres.PostgresClient
	}
	tests := []struct {
		name string
		args args
	}{{name: "success create ownerRepository"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOwnerRepository(tt.args.postgreClient)
			assert.NotNil(t, got)
		})
	}
}

func Test_ownerRepository_Create(t *testing.T) {
	type args struct {
		ctx   context.Context
		owner model.Owner
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name         string
		repo         *ownerRepository
		args         args
		prepareMocks func(*mocks)
		wantID       int64
		wantErr      bool
	}{
		{
			name: "success create to database",
			repo: &ownerRepository{},
			args: args{
				ctx: context.Background(),
				owner: model.Owner{
					Name:        "test-entity",
					Email:       "test@example.com",
					PhoneNumber: "064123456789",
					Password:    "password",
				},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("INSERT INTO owner").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
			},
			wantID: 1,
		},

		{
			name: "fail create to database",
			repo: &ownerRepository{},
			args: args{
				ctx:   context.Background(),
				owner: model.Owner{},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("INSERT INTO owner").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New("oops! error db"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				panic(err)
			}
			tt.repo.postgres = db

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{pgMock: mock})
			}

			gotID, err := tt.repo.Create(tt.args.ctx, tt.args.owner)

			assert.Equal(t, tt.wantID, gotID)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_ownerRepository_Update(t *testing.T) {
	type args struct {
		ctx   context.Context
		owner model.Owner
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name          string
		repo          *ownerRepository
		args          args
		prepareMocks  func(*mocks)
		wantNAffected int64
		wantErr       bool
	}{
		{
			name: "success update to database",
			repo: &ownerRepository{},
			args: args{
				ctx: context.Background(),
				owner: model.Owner{
					Id:   4,
					Name: "updated test",
				},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectExec(`UPDATE owner`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantNAffected: 1,
		},
		{
			name: "fail update to database (error no rows updated)",
			repo: &ownerRepository{},
			args: args{
				ctx:   context.Background(),
				owner: model.Owner{},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectExec(`UPDATE owner`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "fail update to database (error rows affected)",
			repo: &ownerRepository{},
			args: args{
				ctx:   context.Background(),
				owner: model.Owner{},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectExec(`UPDATE owner`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewErrorResult(errors.New("oops! error rows affected")))
			},
			wantErr: true,
		},
		{
			name: "fail update to database",
			repo: &ownerRepository{},
			args: args{
				ctx:   context.Background(),
				owner: model.Owner{},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectExec(`UPDATE owner`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("oops! error db"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				panic(err)
			}
			tt.repo.postgres = db

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{pgMock: mock})
			}

			gotNAffected, errNoRow, err := tt.repo.Update(tt.args.ctx, tt.args.owner)
			assert.Equal(t, tt.wantNAffected, gotNAffected)
			assert.Equal(t, tt.wantErr, err != nil || errNoRow != nil)

		})
	}
}

func Test_ownerRepository_GetByEmail(t *testing.T) {
	type args struct {
		ctx   context.Context
		email string
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name         string
		repo         *ownerRepository
		args         args
		prepareMocks func(*mocks)
		wantOwner    *model.Owner
		wantErr      bool
		// Err          error
	}{
		{
			name: "success get owner by email from database",
			repo: &ownerRepository{},
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectQuery("SELECT.*owner.*WHERE.*email").
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone_number", "date_of_birth", "password"}).
						AddRow(int64(1), "test", "test@example.com", "", nil, ""))
			},
			wantOwner: &model.Owner{
				Id:    1,
				Name:  "test",
				Email: "test@example.com",
			},
		},
		{
			name: "fail GetByEmail",
			repo: &ownerRepository{},
			args: args{
				ctx: context.Background(),
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectQuery("SELECT.*owner.*WHERE.*email").
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone_number", "date_of_birth", "password"})).
					WillReturnError(errors.New("oops! error db"))
			},
			wantErr: true,
		},
		{
			name: "fail get owner by email from database (no row)",
			repo: &ownerRepository{},
			args: args{
				ctx:   context.Background(),
				email: "not.found@example.com",
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectQuery("SELECT.*owner.*WHERE.*email").
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone_number", "date_of_birth", "password"})).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				panic(err)
			}
			tt.repo.postgres = db
			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{pgMock: mock})
			}

			gotOwner, errNoRow, err := tt.repo.GetByEmail(tt.args.ctx, tt.args.email)
			assert.Equal(t, tt.wantOwner, gotOwner)
			assert.Equal(t, err != nil || errNoRow != nil, tt.wantErr)

		})
	}
}

func Test_ownerRepository_Get(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name         string
		repo         *ownerRepository
		args         args
		prepareMocks func(*mocks)
		wantOwner    *model.Owner
		wantErr      bool
		Err          error
	}{
		{
			name: "success GetByID",
			repo: &ownerRepository{},
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectQuery("SELECT.*owner.*WHERE.*id").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone_number", "date_of_birth", "password"}).
						AddRow(int64(1), "test", "test@example.com", "", nil, ""))
			},
			wantOwner: &model.Owner{
				Id:    1,
				Name:  "test",
				Email: "test@example.com",
			},
		},
		{
			name: "fail GetByID (error db)",
			repo: &ownerRepository{},
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectQuery("SELECT.*owner.*WHERE.*id").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone_number", "date_of_birth", "password"})).
					WillReturnError(errors.New("oops! error db"))
			},
			wantErr: true,
		},
		{
			name: "fail GetByID (no rows)",
			repo: &ownerRepository{},
			args: args{
				ctx: context.Background(),
				id:  0,
			},
			wantErr: true,
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectQuery("SELECT.*owner.*WHERE.*id").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone_number", "date_of_birth", "password"})).
					WillReturnError(sql.ErrNoRows)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				panic(err)
			}
			tt.repo.postgres = db

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{pgMock: mock})
			}

			gotOwner, errNoRow, err := tt.repo.Get(tt.args.ctx, tt.args.id)

			assert.Equal(t, tt.wantOwner, gotOwner)
			assert.Equal(t, tt.wantErr, err != nil || errNoRow != nil)
		})
	}
}

func Test_ownerRepository_List(t *testing.T) {
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
		repo         *ownerRepository
		args         args
		prepareMocks func(*mocks)
		wantOwners   []*model.Owner
		wantErr      bool
		Err          error
	}{
		{
			name: "success List",
			repo: &ownerRepository{},
			args: args{
				ctx:    context.Background(),
				limit:  2,
				offset: 1,
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectQuery("SELECT.*owner.*LIMIT.*OFFSET").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone_number", "date_of_birth"}).
						AddRow(int64(1), "test1", "", "", nil).
						AddRow(int64(2), "test2", "", "", nil))
			},
			wantOwners: []*model.Owner{
				{Id: 1, Name: "test1"},
				{Id: 2, Name: "test2"},
			},
		},
		{
			name: "fail List (error db)",
			repo: &ownerRepository{},
			args: args{
				ctx:    context.Background(),
				limit:  2,
				offset: 1,
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectQuery("SELECT.*owner.*LIMIT.*OFFSET").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone_number", "date_of_birth"})).
					WillReturnError(errors.New("oops! error db"))
			},
			wantErr: true,
		},
		{
			name: "fail List (error no rows)",
			repo: &ownerRepository{},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectQuery("SELECT.*owner.*LIMIT.*OFFSET").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone_number", "date_of_birth"}))
			},
			args: args{
				ctx:    context.Background(),
				limit:  10,
				offset: 1_000_000_000_000,
			},
			wantErr: true,
		},
		{
			name: "fail List (error rows)",
			repo: &ownerRepository{},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectQuery("SELECT.*owner.*LIMIT.*OFFSET").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone_number", "date_of_birth"}).
						AddRow(int64(1), "test1", "", "", nil).RowError(0, errors.New("error rows")))
			},
			args: args{
				ctx:    context.Background(),
				limit:  2,
				offset: 1,
			},
			wantErr: true,
		},
		{
			name: "fail List (error scan)",
			repo: &ownerRepository{},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectQuery("SELECT.*owner.*LIMIT.*OFFSET").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone_number", "date_of_birth"}).
						AddRow("one", "test1", "", "", nil)) // owner_id must be int64 not int
			},
			args: args{
				ctx:    context.Background(),
				limit:  2,
				offset: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				panic(err)
			}
			tt.repo.postgres = db

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{pgMock: mock})
			}

			gotOwners, errNoRow, err := tt.repo.List(tt.args.ctx, tt.args.limit, tt.args.offset)

			assert.Equal(t, tt.wantOwners, gotOwners)
			assert.Equal(t, tt.wantErr, err != nil || errNoRow != nil, err)
		})
	}
}

func Test_ownerRepository_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name          string
		repo          *ownerRepository
		args          args
		prepareMocks  func(*mocks)
		wantNAffected int64
		wantErr       bool
	}{
		{
			name: "success Delete",
			repo: &ownerRepository{},
			args: args{
				ctx: context.Background(),
				id:  10,
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec("DELETE FROM owner").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantNAffected: 1, // number of affected rows
		},
		{
			name: "fail Delete (error db)",
			repo: &ownerRepository{},
			args: args{
				ctx: context.Background(),
				id:  -1,
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec("DELETE FROM owner").WithArgs(sqlmock.AnyArg()).WillReturnError(errors.New("oops! error db"))
			},
			wantErr: true,
		},
		{
			name: "fail Delete (error row affected)",
			repo: &ownerRepository{},
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec("DELETE FROM owner").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewErrorResult(errors.New("oops! error rows affected")))
			},
			wantErr: true,
		},
		{
			name: "fail Delete (error no rows)",
			repo: &ownerRepository{},
			args: args{
				ctx: context.Background(),
				id:  0,
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec("DELETE FROM owner").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				panic(err)
			}
			tt.repo.postgres = db

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{pgMock: mock})
			}

			gotNAffected, errNoRow, err := tt.repo.Delete(tt.args.ctx, tt.args.id)
			assert.Equal(t, tt.wantNAffected, gotNAffected)
			assert.Equal(t, tt.wantErr, err != nil || errNoRow != nil)
		})
	}
}

func Test_ownerRepository_UpdatePasswordByEmail(t *testing.T) {
	type args struct {
		ctx      context.Context
		email    string
		password string
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name          string
		repo          *ownerRepository
		args          args
		prepareMocks  func(*mocks)
		wantNAffected int64
		wantErr       bool
	}{
		{
			name: "success UpdatePasswordByEmail",
			repo: &ownerRepository{},
			args: args{ctx: context.Background(), email: "test@example.com", password: "updated-hashed-password"},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectExec(`UPDATE owner.+password.+WHERE.+email`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantNAffected: 1,
		},
		{
			name: "fail UpdatePasswordByEmail (error no rows)",
			repo: &ownerRepository{},
			args: args{ctx: context.Background(), email: "not.found@example.com", password: "updated-hashed-password"},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectExec(`UPDATE owner.+password.+WHERE.+email`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "fail UpdatePasswordByEmail (error db)",
			repo: &ownerRepository{},
			args: args{ctx: context.Background(), email: "test@example.com", password: "updated-hashed-password"},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectExec(`UPDATE owner.+password.+WHERE.+email`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("oops! error db"))
			},
			wantErr: true,
		},
		{
			name: "fail UpdatePasswordByEmail (error rows affected)",
			repo: &ownerRepository{},
			args: args{ctx: context.Background(), email: "test@example.com", password: "updated-hashed-password"},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectExec(`UPDATE owner.+password.+WHERE.+email`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewErrorResult(errors.New("oops! error rows affected")))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				panic(err)
			}
			tt.repo.postgres = db

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{pgMock: mock})
			}

			gotNAffected, errNoRow, err := tt.repo.UpdatePasswordByEmail(tt.args.ctx, tt.args.email, tt.args.password)
			assert.Equal(t, tt.wantNAffected, gotNAffected)
			assert.Equal(t, tt.wantErr, err != nil || errNoRow != nil)
		})
	}
}

func Test_ownerRepository_UpdatePasswordByID(t *testing.T) {
	type args struct {
		ctx      context.Context
		id       int64
		password string
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name          string
		repo          *ownerRepository
		args          args
		prepareMocks  func(*mocks)
		wantNAffected int64
		wantErr       bool
	}{
		{
			name: "success UpdatePasswordByID",
			repo: &ownerRepository{},
			args: args{ctx: context.Background(), id: 1, password: "updated-hashed-password"},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectExec(`UPDATE owner.+password.+WHERE.+id`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantNAffected: 1,
		},
		{
			name: "fail UpdatePasswordByID (error no rows)",
			repo: &ownerRepository{},
			args: args{ctx: context.Background(), id: 0, password: "updated-hashed-password"},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectExec(`UPDATE owner.+password.+WHERE.+id`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "fail UpdatePasswordByID (error db)",
			repo: &ownerRepository{},
			args: args{ctx: context.Background(), id: 1, password: "updated-hashed-password"},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectExec(`UPDATE owner.+password.+WHERE.+id`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("oops! error db"))
			},
			wantErr: true,
		},
		{
			name: "fail UpdatePasswordByID (error rows affected)",
			repo: &ownerRepository{},
			args: args{ctx: context.Background(), id: 1, password: "updated-hashed-password"},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectExec(`UPDATE owner.+password.+WHERE.+id`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewErrorResult(errors.New("oops! error rows affected")))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				panic(err)
			}
			tt.repo.postgres = db

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{pgMock: mock})
			}

			gotNAffected, errNoRow, err := tt.repo.UpdatePasswordByID(tt.args.ctx, tt.args.id, tt.args.password)
			assert.Equal(t, tt.wantNAffected, gotNAffected)
			assert.Equal(t, tt.wantErr, err != nil || errNoRow != nil)
		})
	}
}

func Test_ownerRepository_UpdateEmailByID(t *testing.T) {
	type args struct {
		ctx   context.Context
		id    int64
		email string
	}
	type mocks struct {
		pgMock sqlmock.Sqlmock
	}
	tests := []struct {
		name          string
		repo          *ownerRepository
		args          args
		prepareMocks  func(*mocks)
		wantNAffected int64
		wantErr       bool
	}{
		{
			name: "success UpdateEmailByID",
			repo: &ownerRepository{},
			args: args{ctx: context.Background(), id: 1, email: "test.updated@example.com"},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectExec(`UPDATE owner.+email.+WHERE.+id`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantNAffected: 1,
		},
		{
			name: "fail UpdateEmailByID (error no rows)",
			repo: &ownerRepository{},
			args: args{ctx: context.Background(), id: 0, email: "not.found@example.com"},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectExec(`UPDATE owner.+email.+WHERE.+id`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "fail UpdateEmailByID (error db)",
			repo: &ownerRepository{},
			args: args{ctx: context.Background(), id: 1, email: "test.updated@example.com"},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectExec(`UPDATE owner.+email.+WHERE.+id`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("oops! error db"))
			},
			wantErr: true,
		},
		{
			name: "fail UpdateEmailByID (error rows affected)",
			repo: &ownerRepository{},
			args: args{ctx: context.Background(), id: 1, email: "test.updated@example.com"},
			prepareMocks: func(m *mocks) {
				m.pgMock.
					ExpectExec(`UPDATE owner.+email.+WHERE.+id`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewErrorResult(errors.New("oops! error rows affected")))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				panic(err)
			}
			tt.repo.postgres = db

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{pgMock: mock})
			}

			gotNAffected, errNoRow, err := tt.repo.UpdateEmailByID(tt.args.ctx, tt.args.id, tt.args.email)
			assert.Equal(t, tt.wantNAffected, gotNAffected)
			assert.Equal(t, tt.wantErr, err != nil || errNoRow != nil)
		})
	}
}
