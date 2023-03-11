package service

import (
	"context"
	"errors"
	"family-catering/internal/model"
	"family-catering/internal/repository"
	"family-catering/pkg/consts"
	utils "family-catering/pkg/utils"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewOwnerService(t *testing.T) {
	type args struct {
		ownerRepo repository.OwnerRepository
	}
	tests := []struct {
		name string
		args args
	}{{name: "success create new owner service"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, NewOwnerService(tt.args.ownerRepo))
		})
	}
}

func Test_ownerService_Create(t *testing.T) {
	type args struct {
		ctx context.Context
		req model.CreateOwnerRequest
	}
	type mocks struct {
		utMocks       utils.Mock
		ownerRepoMock *repository.MockOwnerRepository
	}
	tests := []struct {
		name         string
		svc          *ownerService
		args         args
		prepareMocks func(*mocks)
		wantResp     *model.CreateOwnerResponse
		wantErr      bool
	}{
		{
			name: "success Create",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				req: model.CreateOwnerRequest{
					Name:     "test",
					Email:    "test@example.com",
					Password: "password",
				},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValidateRequest", func(s interface{}) error {
					return nil
				})
				m.ownerRepoMock.
					EXPECT().
					GetByEmail(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf("")).
					Return(nil, nil, nil)
				m.utMocks.Patch("HashPassword", func(s string) (string, error) {
					return "hashed-password", nil
				})
				m.ownerRepoMock.
					EXPECT().
					Create(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(model.Owner{})).
					Return(int64(1), nil)
			},
			wantResp: &model.CreateOwnerResponse{
				Id:       1,
				Name:     "test",
				Email:    "test@example.com",
				Password: "hashed-password",
			},
		},
		{
			name: "fail Create (error validate request)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				req: model.CreateOwnerRequest{
					Name:     "test",
					Email:    "test@example.com",
					Password: "password",
				},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValidateRequest", func(s interface{}) error {
					return errors.New("oops! error validate request")
				})
			},
			wantErr: true,
		},
		{
			name: "fail Create (error invalid password)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				req: model.CreateOwnerRequest{
					Name:     "test",
					Email:    "test@example.com",
					Password: "password",
				},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValidateRequest", func(s interface{}) error {
					return nil
				})
				m.ownerRepoMock.
					EXPECT().
					GetByEmail(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf("")).
					Return(nil, nil, nil)
				m.utMocks.Patch("HashPassword", func(s string) (string, error) {
					return "", errors.New("oops! invalid password")
				})
			},
			wantErr: true,
		},
		{
			name: "fail Create (error email already registered)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				req: model.CreateOwnerRequest{
					Name:     "test",
					Email:    "test@example.com",
					Password: "password",
				},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValidateRequest", func(s interface{}) error {
					return nil
				})
				m.ownerRepoMock.
					EXPECT().
					GetByEmail(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf("")).
					Return(&model.Owner{Name: "another-test-name", Email: "test@example.com", Password: "another-hashed-password"}, nil, nil)

			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utMocks := utils.InitMock()
			ctrl := gomock.NewController(t)
			ownerRepoMock := repository.NewMockOwnerRepository(ctrl)
			tt.svc.ownerRepo = ownerRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{utMocks: utMocks, ownerRepoMock: ownerRepoMock})
			}

			gotResp, err := tt.svc.Create(tt.args.ctx, tt.args.req)

			assert.Equal(t, tt.wantResp, gotResp)
			assert.Equal(t, err != nil, tt.wantErr)

			utMocks.UnpatchAll()

		})
	}
}

func Test_ownerService_Get(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	type mocks struct {
		ownerRepoMock *repository.MockOwnerRepository
	}
	tests := []struct {
		name         string
		svc          *ownerService
		args         args
		prepareMocks func(*mocks)
		wantResp     *model.GetOwnerResponse
		wantErr      bool
	}{
		{
			name: "success GetByID",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			prepareMocks: func(m *mocks) {
				m.ownerRepoMock.
					EXPECT().
					Get(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(int64(1))).
					Return(&model.Owner{Id: 1, Name: "test"}, nil, nil)
			},
			wantResp: &model.GetOwnerResponse{
				Id:   1,
				Name: "test",
			},
		},

		{
			name: "fail GetByID (error repo)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			prepareMocks: func(m *mocks) {
				m.ownerRepoMock.
					EXPECT().
					Get(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(int64(1))).
					Return(nil, nil, errors.New("oops! error repo"))
			},
			wantErr: true,
		},
		{
			name: "fail GetByID (no rows)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  0,
			},
			prepareMocks: func(m *mocks) {
				m.ownerRepoMock.
					EXPECT().
					Get(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(int64(1))).
					Return(nil, errors.New("oops! error no rows"), nil)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ownerRepoMock := repository.NewMockOwnerRepository(ctrl)
			tt.svc.ownerRepo = ownerRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{ownerRepoMock: ownerRepoMock})
			}

			gotResp, err := tt.svc.Get(tt.args.ctx, tt.args.id)

			assert.Equal(t, tt.wantResp, gotResp)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_ownerService_List(t *testing.T) {
	type args struct {
		ctx    context.Context
		limit  int
		offset int
	}
	type mocks struct {
		ownerRepoMock *repository.MockOwnerRepository
	}
	tests := []struct {
		name         string
		svc          *ownerService
		args         args
		prepareMocks func(*mocks)
		wantResp     []*model.GetOwnerResponse
		wantErr      bool
	}{
		{
			name: "success List",
			svc:  &ownerService{},
			args: args{
				ctx:    context.Background(),
				limit:  2,
				offset: 1,
			},
			prepareMocks: func(m *mocks) {
				m.ownerRepoMock.
					EXPECT().
					List(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(1), gomock.AssignableToTypeOf(1)).
					Return([]*model.Owner{{Id: 1, Name: "test-1"}, {Id: 2, Name: "test-2"}}, nil, nil)
			},
			wantResp: []*model.GetOwnerResponse{
				{Id: 1, Name: "test-1"},
				{Id: 2, Name: "test-2"},
			},
		},
		{
			name: "fail List (error repo)",
			svc:  &ownerService{},
			args: args{
				ctx:    context.Background(),
				limit:  1,
				offset: 2,
			},
			prepareMocks: func(m *mocks) {
				m.ownerRepoMock.
					EXPECT().
					List(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(1), gomock.AssignableToTypeOf(1)).
					Return(nil, nil, errors.New("oops! error repo"))
			},
			wantErr: true,
		},
		{
			name: "success List (but no rows/empty data)",
			svc:  &ownerService{},
			args: args{
				ctx:    context.Background(),
				limit:  2,
				offset: 1,
			},
			prepareMocks: func(m *mocks) {
				m.ownerRepoMock.
					EXPECT().
					List(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(1), gomock.AssignableToTypeOf(1)).
					Return(nil, errors.New("oops! error no rows"), nil)
			},
			wantResp: []*model.GetOwnerResponse{},
			wantErr:  false, // errNoRow does not considered as an error at ownerService.List
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ownerRepoMock := repository.NewMockOwnerRepository(ctrl)
			tt.svc.ownerRepo = ownerRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{ownerRepoMock: ownerRepoMock})
			}

			gotResp, err := tt.svc.List(tt.args.ctx, tt.args.limit, tt.args.offset)
			assert.Equal(t, tt.wantResp, gotResp)
			assert.Equal(t, tt.wantErr, err != nil)

		})
	}
}

func Test_ownerService_Update(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
		req model.UpdateOwnerRequest
	}
	type mocks struct {
		utMocks       utils.Mock
		ownerRepoMock *repository.MockOwnerRepository
	}
	tests := []struct {
		name         string
		svc          *ownerService
		args         args
		prepareMocks func(*mocks)
		wantResp     *model.UpdateOwnerResponse
		wantErr      bool
	}{
		{
			name: "success Update",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  1,
				req: model.UpdateOwnerRequest{Name: "test-updated", PhoneNumber: "123456789101"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} { return "access-token" })
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) { return &utils.JwtClaims{}, nil })
				m.utMocks.Patch("ValidateRequest", func(interface{}) error { return nil })
				m.ownerRepoMock.
					EXPECT().
					Update(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(model.Owner{})).
					Return(int64(1), nil, nil)

			},
			wantResp: &model.UpdateOwnerResponse{
				Id:          1,
				Name:        "test-updated",
				PhoneNumber: "123456789101",
			},
		},
		{
			name: "fail Update (no rows)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  0,
				req: model.UpdateOwnerRequest{Name: "test-updated", PhoneNumber: "123456789101"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} { return "access-token" })
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) { return &utils.JwtClaims{}, nil })
				m.utMocks.Patch("ValidateRequest", func(interface{}) error { return nil })
				m.ownerRepoMock.
					EXPECT().
					Update(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(model.Owner{})).
					Return(int64(0), errors.New("oops! error no rows"), nil)

			},
			wantErr: true,
		},
		{
			name: "fail Update (error invalid/no token)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  1,
				req: model.UpdateOwnerRequest{Name: "test-updated", PhoneNumber: "123456789101"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} { return "invalid-token" })
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) { return nil, errors.New("oops! invalid token") })
			},
			wantErr: true,
		},
		{
			name: "fail Update (error invalid request)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  1,
				req: model.UpdateOwnerRequest{Name: "test-updated", PhoneNumber: "not-a-number"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} { return "invalid-token" })
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) { return &utils.JwtClaims{}, nil })
				m.utMocks.Patch("ValidateRequest", func(interface{}) error { return errors.New("oops! invalid request") })

			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utMocks := utils.InitMock()
			ctrl := gomock.NewController(t)
			ownerRepoMock := repository.NewMockOwnerRepository(ctrl)
			tt.svc.ownerRepo = ownerRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{ownerRepoMock: ownerRepoMock, utMocks: utMocks})
			}

			gotResp, err := tt.svc.Update(tt.args.ctx, tt.args.id, tt.args.req)

			assert.Equal(t, tt.wantResp, gotResp)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_ownerService_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	type mocks struct {
		utMocks       utils.Mock
		ownerRepoMock *repository.MockOwnerRepository
	}
	tests := []struct {
		name          string
		svc           *ownerService
		args          args
		prepareMocks  func(*mocks)
		wantNAffected int64
		wantErr       bool
	}{
		{
			name: "success Delete",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  10,
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == consts.CtxKeyAuthorization {
						return "access-token"
					}
					if key == consts.CtxKeySession {
						return &model.AuthSessionResponse{Valid: true, OwnerID: 10, SID: "sid"}
					}
					return nil
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) { return &utils.JwtClaims{}, nil })
				m.ownerRepoMock.
					EXPECT().
					Delete(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(int64(0))).
					Return(int64(1), nil, nil)
			},
			wantNAffected: 1,
		},
		{
			name: "fail Delete (error repo)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  10,
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == consts.CtxKeyAuthorization {
						return "access-token"
					}
					if key == consts.CtxKeySession {
						return &model.AuthSessionResponse{Valid: true, OwnerID: 10, SID: "sid"}
					}
					return nil
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) { return &utils.JwtClaims{}, nil })
				m.ownerRepoMock.
					EXPECT().
					Delete(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(int64(0))).
					Return(int64(0), nil, errors.New("oops! error repo"))
			},
			wantErr: true,
		},
		{
			name: "fail Delete (error invalid/no token)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  10,
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == consts.CtxKeyAuthorization {
						return "invalid-token"
					}
					if key == consts.CtxKeySession {
						return &model.AuthSessionResponse{Valid: true, OwnerID: 10, SID: "sid"}
					}
					return nil
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) { return nil, errors.New("oops! error invalid token") })
			},
			wantErr: true,
		},
		{
			name: "fail Delete (error no rows)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  0,
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == consts.CtxKeyAuthorization {
						return "access-token"
					}
					if key == consts.CtxKeySession {
						return &model.AuthSessionResponse{Valid: true, OwnerID: 0, SID: "sid"} // ? i don't think this is possible test case, because delete must have session
					}
					return nil
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) { return &utils.JwtClaims{}, nil })
				m.ownerRepoMock.
					EXPECT().
					Delete(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(int64(0))).
					Return(int64(0), errors.New("oops! error no rows"), nil)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utMocks := utils.InitMock()
			ctrl := gomock.NewController(t)
			ownerRepoMock := repository.NewMockOwnerRepository(ctrl)
			tt.svc.ownerRepo = ownerRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{utMocks: utMocks, ownerRepoMock: ownerRepoMock})
			}
			gotNAffected, err := tt.svc.Delete(tt.args.ctx, tt.args.id)

			assert.Equal(t, tt.wantNAffected, gotNAffected)
			assert.Equal(t, tt.wantErr, err != nil)

			utMocks.UnpatchAll()

		})
	}
}

func Test_ownerService_ResetPasswordByEmail(t *testing.T) {
	type args struct {
		ctx             context.Context
		passwordResetID string
		req             model.ResetPasswordRequest
	}
	type mocks struct {
		utMocks       utils.Mock
		ownerRepoMock *repository.MockOwnerRepository
	}
	tests := []struct {
		name         string
		svc          *ownerService
		args         args
		prepareMocks func(*mocks)
		wantErr      bool
	}{
		{
			name: "success ResetPasswordByEmail",
			svc:  &ownerService{},
			args: args{
				ctx:             context.Background(),
				passwordResetID: "rpid",
				req: model.ResetPasswordRequest{
					Password:        "updated-plain-password",
					PasswordConfirm: "updated-plain-password"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} { return "password-token" })
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return utils.NewJWTClaimTesting("rpid"), nil
				})
				m.ownerRepoMock.
					EXPECT().
					UpdatePasswordByEmail(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf("")).
					Return(int64(1), nil, nil)

			},
		},
		{
			name: "fail ResetPasswordByEmail (error invalid token)",
			svc:  &ownerService{},
			args: args{
				ctx:             context.Background(),
				passwordResetID: "rpid",
				req: model.ResetPasswordRequest{
					Password:        "updated-plain-password",
					PasswordConfirm: "updated-plain-password"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} { return "password-token" })
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return nil, errors.New("oops! error invalid token")
				})
			},
			wantErr: true,
		},
		{
			name: "fail ResetPasswordByEmail (invalid claim token)",
			svc:  &ownerService{},
			args: args{
				ctx:             context.Background(),
				passwordResetID: "rpid",
				req: model.ResetPasswordRequest{
					Password:        "updated-plain-password",
					PasswordConfirm: "updated-plain-password"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} { return "password-token" })
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return utils.NewJWTClaimTesting("diff-rpid"), nil
				})
			},
			wantErr: true,
		},
		{
			name: "fail ResetPasswordByEmail (error no rows)",
			svc:  &ownerService{},
			args: args{
				ctx:             context.Background(),
				passwordResetID: "rpid",
				req: model.ResetPasswordRequest{
					Password:        "updated-plain-password",
					PasswordConfirm: "updated-plain-password"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} { return "password-token" })
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return utils.NewJWTClaimTesting("rpid"), nil
				})
				m.ownerRepoMock.
					EXPECT().
					UpdatePasswordByEmail(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf("")).
					Return(int64(0), errors.New("oops! error no rows"), nil)
			},
			wantErr: true,
		},
		{
			name: "fail ResetPasswordByEmail (error repo)",
			svc:  &ownerService{},
			args: args{
				ctx:             context.Background(),
				passwordResetID: "rpid",
				req: model.ResetPasswordRequest{
					Password:        "updated-plain-password",
					PasswordConfirm: "updated-plain-password"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} { return "password-token" })
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return utils.NewJWTClaimTesting("rpid"), nil
				})
				m.ownerRepoMock.
					EXPECT().
					UpdatePasswordByEmail(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf("")).
					Return(int64(0), nil, errors.New("oops! error repo"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utMocks := utils.InitMock()
			ctrl := gomock.NewController(t)
			ownerRepoMock := repository.NewMockOwnerRepository(ctrl)
			tt.svc.ownerRepo = ownerRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{utMocks: utMocks, ownerRepoMock: ownerRepoMock})
			}

			err := tt.svc.ResetPasswordByEmail(tt.args.ctx, tt.args.passwordResetID, tt.args.req)
			assert.Equal(t, err != nil, tt.wantErr)

			utMocks.UnpatchAll()

		})
	}
}

func Test_ownerService_ResetPasswordByID(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
		req model.ResetPasswordRequest
	}
	type mocks struct {
		utMocks       utils.Mock
		ownerRepoMock *repository.MockOwnerRepository
	}
	tests := []struct {
		name         string
		svc          *ownerService
		args         args
		prepareMocks func(*mocks)
		wantErr      bool
	}{
		{
			name: "success ResetPasswordByID",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  1,
				req: model.ResetPasswordRequest{
					Password:        "updated-plain-password",
					PasswordConfirm: "updated-plain-password"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == consts.CtxKeyAuthorization {
						return "password-token"
					}
					if key == consts.CtxKeySession {
						return &model.AuthSessionResponse{Valid: true, OwnerID: 1}
					}

					return nil
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return utils.NewJWTClaimTesting(""), nil
				})
				m.ownerRepoMock.
					EXPECT().
					UpdatePasswordByID(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf("")).
					Return(int64(1), nil, nil)

			},
		},
		{
			name: "fail ResetPasswordByID (error invalid token)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  1,
				req: model.ResetPasswordRequest{
					Password:        "updated-plain-password",
					PasswordConfirm: "updated-plain-password"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == consts.CtxKeyAuthorization {
						return "invalid-password-token"
					}
					if key == consts.CtxKeySession {
						return &model.AuthSessionResponse{Valid: true, OwnerID: 1}
					}

					return nil
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return nil, errors.New("oops! error invalid token")
				})
			},
			wantErr: true,
		},
		{
			name: "fail ResetPasswordByID (invalid claim token)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  1,
				req: model.ResetPasswordRequest{
					Password:        "updated-plain-password",
					PasswordConfirm: "updated-plain-password"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == consts.CtxKeyAuthorization {
						return "password-token"
					}
					if key == consts.CtxKeySession {
						return &model.AuthSessionResponse{Valid: false, OwnerID: 2}
					}

					return nil
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return utils.NewJWTClaimTesting(""), nil
				})
			},
			wantErr: true,
		},
		{
			name: "fail ResetPasswordByID (error no rows)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  0,
				req: model.ResetPasswordRequest{
					Password:        "updated-plain-password",
					PasswordConfirm: "updated-plain-password"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == consts.CtxKeyAuthorization {
						return "password-token"
					}
					if key == consts.CtxKeySession {
						return &model.AuthSessionResponse{Valid: true, OwnerID: 0}
					}

					return nil
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return utils.NewJWTClaimTesting(""), nil
				})
				m.ownerRepoMock.
					EXPECT().
					UpdatePasswordByID(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf("")).
					Return(int64(0), errors.New("oops! error no rows"), nil)
			},
			wantErr: true,
		},
		{
			name: "fail ResetPasswordByID (error repo)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  1,
				req: model.ResetPasswordRequest{
					Password:        "updated-plain-password",
					PasswordConfirm: "updated-plain-password"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == consts.CtxKeyAuthorization {
						return "password-token"
					}
					if key == consts.CtxKeySession {
						return &model.AuthSessionResponse{Valid: true, OwnerID: 1}
					}

					return nil
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return utils.NewJWTClaimTesting("rpid"), nil
				})
				m.ownerRepoMock.
					EXPECT().
					UpdatePasswordByID(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf("")).
					Return(int64(0), nil, errors.New("oops! error repo"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utMocks := utils.InitMock()
			ctrl := gomock.NewController(t)
			ownerRepoMock := repository.NewMockOwnerRepository(ctrl)
			tt.svc.ownerRepo = ownerRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{utMocks: utMocks, ownerRepoMock: ownerRepoMock})
			}

			err := tt.svc.ResetPasswordByID(tt.args.ctx, tt.args.id, tt.args.req)
			assert.Equal(t, err != nil, tt.wantErr)

			utMocks.UnpatchAll()
		})
	}
}

func Test_ownerService_UpdateEmailByID(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
		req model.UpdateEmailRequest
	}
	type mocks struct {
		utMocks       utils.Mock
		ownerRepoMock *repository.MockOwnerRepository
	}
	tests := []struct {
		name         string
		svc          *ownerService
		args         args
		prepareMocks func(*mocks)
		wantErr      bool
	}{
		{
			name: "success UpdateEmailByID",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  1,
				req: model.UpdateEmailRequest{Email: "test.updated@example.com"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == consts.CtxKeyAuthorization {
						return "access-token"
					}
					if key == consts.CtxKeySession {
						return &model.AuthSessionResponse{Valid: true, OwnerID: 1}
					}

					return nil
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return utils.NewJWTClaimTesting(""), nil
				})
				m.ownerRepoMock.
					EXPECT().
					UpdateEmailByID(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf("")).
					Return(int64(1), nil, nil)

			},
		},
		{
			name: "fail UpdateEmailByID (error invalid token)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  1,
				req: model.UpdateEmailRequest{Email: "test.updated@example.com"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == consts.CtxKeyAuthorization {
						return "invalid-access-token"
					}
					if key == consts.CtxKeySession {
						return &model.AuthSessionResponse{Valid: true, OwnerID: 1}
					}

					return nil
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return nil, errors.New("oops! error invalid token")
				})
			},
			wantErr: true,
		},
		{
			name: "fail UpdateEmailByID (invalid claim token)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  1,
				req: model.UpdateEmailRequest{Email: "test.updated@example.com"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == consts.CtxKeyAuthorization {
						return "access-token"
					}
					if key == consts.CtxKeySession {
						return &model.AuthSessionResponse{Valid: false, OwnerID: 2}
					}

					return nil
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return utils.NewJWTClaimTesting(""), nil
				})
			},
			wantErr: true,
		},
		{
			name: "fail ResetPasswordByID (error no rows)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  0,
				req: model.UpdateEmailRequest{Email: "test.updated@example.com"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == consts.CtxKeyAuthorization {
						return "password-token"
					}
					if key == consts.CtxKeySession {
						return &model.AuthSessionResponse{Valid: true, OwnerID: 0}
					}

					return nil
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return utils.NewJWTClaimTesting(""), nil
				})
				m.ownerRepoMock.
					EXPECT().
					UpdateEmailByID(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf("")).
					Return(int64(0), errors.New("oops! error no rows"), nil)
			},
			wantErr: true,
		},
		{
			name: "fail UpdateEmailByID (error repo)",
			svc:  &ownerService{},
			args: args{
				ctx: context.Background(),
				id:  1,
				req: model.UpdateEmailRequest{Email: "test.updated@example.com"},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == consts.CtxKeyAuthorization {
						return "access-token"
					}
					if key == consts.CtxKeySession {
						return &model.AuthSessionResponse{Valid: true, OwnerID: 1}
					}

					return nil
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return utils.NewJWTClaimTesting("rpid"), nil
				})
				m.ownerRepoMock.
					EXPECT().
					UpdateEmailByID(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(int64(0)), gomock.AssignableToTypeOf("")).
					Return(int64(0), nil, errors.New("oops! error repo"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utMocks := utils.InitMock()
			ctrl := gomock.NewController(t)
			ownerRepoMock := repository.NewMockOwnerRepository(ctrl)
			tt.svc.ownerRepo = ownerRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{utMocks: utMocks, ownerRepoMock: ownerRepoMock})
			}

			err := tt.svc.UpdateEmailByID(tt.args.ctx, tt.args.id, tt.args.req)
			assert.Equal(t, err != nil, tt.wantErr)

			utMocks.UnpatchAll()
		})
	}
}
