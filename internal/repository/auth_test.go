package repository

import (
	"context"
	"database/sql"
	"errors"
	"family-catering/internal/model"
	"family-catering/pkg/db/postgres"
	"family-catering/pkg/db/redis"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/elliotchance/redismock/v8"
	redisV8 "github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// due to github.com/elliotchance/redismock/v8/cmdable.go:18 (*ClientMock).Pipelined use Arguments.Get(index).(error) instead of Arguments.Error(index)
// this error was created to mimics no error and should only be used for testing purpose only
// see https://github.com/elliotchance/redismock/blob/master/v8/cmdable.go#L18
// will impact to the test coverage which would not be fully covered for happy flow (method Pipelined only) because error always happens
var (
	errNoError error = errors.New("not an error")
)

func TestNewAuthRepository(t *testing.T) {
	type args struct {
		postgres postgres.PostgresClient
		redis    redis.RedisClient
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success NewAuthRepository",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// nil is sufficient for this test case (IMHO)
			got := NewAuthRepository(tt.args.postgres, tt.args.redis)

			assert.NotNil(t, got)
		})
	}
}

func Test_authRepository_Login(t *testing.T) {
	type args struct {
		ctx       context.Context
		authLogin model.Auth
	}
	type mocks struct {
		redisMock *redismock.ClientMock
		pgMock    sqlmock.Sqlmock
	}
	tests := []struct {
		name         string
		repo         *authRepository
		args         args
		wantErr      bool
		prepareMocks func(*mocks)
	}{
		{
			name: "success login",
			repo: &authRepository{},
			args: args{
				ctx: context.Background(),
				authLogin: model.Auth{
					OwnerID: 1,
					Jti:     "jti",
					Email:   "test@example.com",
					SID:     "another-random-string",
				},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("INSERT INTO auth .+").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(nil).WillReturnRows(sqlmock.NewRows([]string{"session_id"}).AddRow("sid"))
				m.redisMock.On("Pipelined").Return([]redisV8.Cmder{redisV8.NewIntResult(1, errNoError)}, errNoError)
			},
		},
		{
			name: "fail login (query error)",
			repo: &authRepository{},
			args: args{
				ctx: context.Background(),
				authLogin: model.Auth{
					OwnerID: 1,
					Jti:     "jti",
					Email:   "test@example.com",
					SID:     "another-random-string",
				},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("INSERT INTO auth .+").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("oops! query error")).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(0))
			},
			wantErr: true,
		},
		{
			name: "fail login (redis error)",
			repo: &authRepository{},
			args: args{
				ctx: context.Background(),
				authLogin: model.Auth{
					OwnerID: 1,
					Jti:     "jti",
					Email:   "test@example.com",
					SID:     "another-random-string",
				},
			},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectQuery("INSERT INTO auth .+").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(nil).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				m.redisMock.On("Pipeline").Return(redisV8.Pipeline{})
				m.redisMock.On("Pipelined").Return([]redisV8.Cmder{redisV8.NewIntResult(0, errors.New("oops! redis error"))}, errors.New("oops! redis error"))

			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisMock, err := redis.NewMockWithMiniRedisClient(t)
			if err != nil {
				panic(err)
			}
			db, pgMock, err := sqlmock.New()
			if err != nil {
				panic(err)
			}
			fMocks := mocks{
				redisMock: redisMock,
				pgMock:    pgMock,
			}

			tt.repo.redis = redisMock
			tt.repo.postgres = db

			if tt.prepareMocks != nil {
				tt.prepareMocks(&fMocks)
			}

			got := tt.repo.Login(tt.args.ctx, tt.args.authLogin)
			assert.Equal(t, tt.wantErr, (got != nil && !errors.Is(got, errNoError)), got)
		})
	}
}

func Test_authRepository_Session(t *testing.T) {
	type args struct {
		ctx context.Context
		sid string
	}
	type mocks struct {
		redisMock *redismock.ClientMock
		pgMock    sqlmock.Sqlmock
	}
	tests := []struct {
		name                   string
		repo                   *authRepository
		args                   args
		prepareMocks           func(*mocks)
		wantAuthLogoutResponse *model.Auth
		wantErr                bool
	}{
		{
			name: "success get Session (redis)",
			args: args{
				ctx: context.Background(),
				sid: "sid",
			},
			repo: &authRepository{},
			prepareMocks: func(m *mocks) {
				m.redisMock.On("HGetAll", mock.Anything, mock.Anything).Return(redisV8.NewStringStringMapResult(map[string]string{
					"owner_id": "1",
					"jti":      "jti",
					"valid":    "true",
					"email":    "test@example.com",
				}, nil)).Times(1)
			},
			wantAuthLogoutResponse: &model.Auth{
				OwnerID: 1,
				SID:     "sid", // alwasy be assigned at the end of function body
				Valid:   true,
				Jti:     "jti",
				Email:   "test@example.com",
			},
		},
		{
			name: "success get Session (postgres)",
			args: args{
				ctx: context.Background(),
				sid: "sid",
			},
			repo: &authRepository{},
			prepareMocks: func(m *mocks) {
				m.redisMock.On("HGetAll", mock.Anything, mock.Anything).Return(redisV8.NewStringStringMapResult(map[string]string{}, nil))
				m.pgMock.ExpectQuery("SELECT .+FROM auth WHERE sid.+").
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"session_id", "owner_id", "email", "jti", "refresh_token", "expired_at"}).AddRow("sid", 1, "test@example.com", "jti", "refresh-token", time.Time{})).WillReturnError(nil)
			},
			wantAuthLogoutResponse: &model.Auth{
				OwnerID:      1,
				SID:          "sid",
				Jti:          "jti",
				Email:        "test@example.com",
				RefreshToken: "refresh-token",
			},
		},
		{
			name: "fail get Session (redis)",
			args: args{
				ctx: context.Background(),
				sid: "sid",
			},
			repo: &authRepository{},
			prepareMocks: func(m *mocks) {
				m.redisMock.On("HGetAll", mock.Anything, mock.Anything).Return(redisV8.NewStringStringMapResult(map[string]string{}, errors.New("oops! error from redis")))
			},
			wantAuthLogoutResponse: nil,
			wantErr:                true,
		},
		{
			name: "success get Session (hash is empty / no entries)",
			args: args{
				ctx: context.Background(),
				sid: "sid",
			},
			repo: &authRepository{},
			prepareMocks: func(m *mocks) {
				m.redisMock.On("HGetAll", mock.Anything, mock.Anything).Return(redisV8.NewStringStringMapResult(map[string]string{}, nil))
				m.pgMock.ExpectQuery("SELECT .+FROM auth WHERE sid.+").
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"session_id", "owner_id", "email", "jti", "refresh_token", "expired_at"})).WillReturnError(sql.ErrNoRows)
			},
			wantAuthLogoutResponse: nil,
			wantErr:                true,
		},

		{
			name: "fail get Session (postgres)",
			args: args{
				ctx: context.Background(),
				sid: "sid",
			},
			repo: &authRepository{},
			prepareMocks: func(m *mocks) {
				m.redisMock.On("HGetAll", mock.Anything, mock.Anything).Return(redisV8.NewStringStringMapResult(map[string]string{}, nil))
				m.pgMock.ExpectQuery("SELECT .+FROM auth WHERE sid.+").
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"session_id", "owner_id", "email", "jti", "refresh_token", "expired_at"}).AddRow("", 0, "", "", "", time.Time{})).WillReturnError(errors.New("oops! error from postgres"))
			},
			wantAuthLogoutResponse: nil,
			wantErr:                true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisMock, err := redis.NewMockWithMiniRedisClient(t)
			if err != nil {
				panic(err)
			}
			db, pgMock, err := sqlmock.New()
			if err != nil {
				panic(err)
			}
			fMocks := mocks{
				redisMock: redisMock,
				pgMock:    pgMock,
			}
			tt.repo.postgres = db
			tt.repo.redis = redisMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&fMocks)
			}

			gotAuthLogoutResponse, errNoRow, err := tt.repo.Session(tt.args.ctx, tt.args.sid)
			assert.Equal(t, tt.wantErr, err != nil || (errNoRow != nil && gotAuthLogoutResponse == nil))
			if gotAuthLogoutResponse != nil {
				gotAuthLogoutResponse.ExpiredAt = "" // ignore the ExpiredAt value
			}
			assert.Equal(t, tt.wantAuthLogoutResponse, gotAuthLogoutResponse)
		})
	}
}

func Test_authRepository_DeleteSession(t *testing.T) {
	type args struct {
		ctx context.Context
		sid string
	}
	type mocks struct {
		redisMock *redismock.ClientMock
		pgMock    sqlmock.Sqlmock
	}
	tests := []struct {
		name         string
		repo         *authRepository
		prepareMocks func(*mocks)
		args         args
		wantErr      bool
	}{
		{
			name: "success delete session",
			args: args{
				ctx: context.Background(),
				sid: "random-sid-string",
			},
			repo: &authRepository{},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec("DELETE FROM auth WHERE sid.+").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
				m.redisMock.On("Pipelined").Return([]redisV8.Cmder{redisV8.NewIntResult(1, errNoError)}, errNoError)
			},
		},
		{
			name: "fail delete session (error postgres)",
			args: args{
				ctx: context.Background(),
				sid: "random-sid-string",
			},
			repo: &authRepository{},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec("DELETE FROM auth WHERE sid.+").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1)).WillReturnError(errors.New("oops! error from postgres"))
			},
			wantErr: true,
		},
		{
			name: "fail delete session (error redis)",
			args: args{
				ctx: context.Background(),
				sid: "random-sid-string",
			},
			repo: &authRepository{},
			prepareMocks: func(m *mocks) {
				m.pgMock.ExpectExec("DELETE FROM auth WHERE sid.+").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
				m.redisMock.On("Pipelined").Return([]redisV8.Cmder{redisV8.NewIntResult(1, errors.New("oops! error from redis"))}, errors.New("oops! error from redis"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisMock, err := redis.NewMockWithMiniRedisClient(t)
			if err != nil {
				panic(err)
			}
			db, pgMock, err := sqlmock.New()
			if err != nil {
				panic(err)
			}
			fMocks := mocks{
				redisMock: redisMock,
				pgMock:    pgMock,
			}
			tt.repo.postgres = db
			tt.repo.redis = redisMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&fMocks)
			}

			err = tt.repo.DeleteSession(tt.args.ctx, tt.args.sid)
			assert.Equal(t, tt.wantErr, (err != nil && !errors.Is(err, errNoError)), err)
		})
	}
}

func Test_authRepository_GetSessionIDByEmail(t *testing.T) {
	type args struct {
		ctx   context.Context
		email string
	}
	type mocks struct {
		redisMock *redismock.ClientMock
	}
	tests := []struct {
		name          string
		repo          *authRepository
		args          args
		prepareMocks  func(*mocks)
		wantSessionID string
		wantErr       bool
	}{
		{
			name: "success get sid",
			repo: &authRepository{},
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			prepareMocks: func(m *mocks) {
				m.redisMock.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(redisV8.NewStringResult("random-sid-string", nil)).Times(1)
			},
			wantSessionID: "random-sid-string",
		},
		{
			name: "fail get sid",
			repo: &authRepository{},
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			prepareMocks: func(m *mocks) {
				m.redisMock.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(redisV8.NewStringResult("random-sid-string", errors.New("oops! error from redis"))).Times(1)
			},
			wantSessionID: "",
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisMock, err := redis.NewMockWithMiniRedisClient(t)
			if err != nil {
				panic(err)
			}

			fMocks := mocks{
				redisMock: redisMock,
			}

			tt.repo.redis = redisMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&fMocks)
			}

			gotSessionID, err := tt.repo.GetSessionIDByEmail(tt.args.ctx, tt.args.email)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantSessionID, gotSessionID)

		})
	}
}
