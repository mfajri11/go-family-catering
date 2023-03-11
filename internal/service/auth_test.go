package service

import (
	"errors"
	"family-catering/config"
	"family-catering/internal/model"
	"family-catering/internal/repository"
	"family-catering/pkg/consts"
	"family-catering/pkg/utils"
	"fmt"
	"strings"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestNewAuthService(t *testing.T) {
	type args struct {
		ownerRepo repository.OwnerRepository
		authRepo  repository.AuthRepository
		mailer    Mailer
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "success create auth service",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAuthService(tt.args.ownerRepo, tt.args.authRepo, tt.args.mailer)
			assert.NotNil(t, got)
		})
	}
}

func Test_authService_Login(t *testing.T) {
	type args struct {
		ctx context.Context
		req model.AuthLoginRequest
	}
	type mocks struct {
		ownerRepoMock *repository.MockOwnerRepository
		authRepoMock  *repository.MockAuthRepository
		mailerMock    *MockMailer
		utMock        *utils.Mock
		cfgMock       *config.MockConfig
	}
	tests := []struct {
		name         string
		svc          *authService
		args         args
		prepareMocks func(*mocks)
		wantResp     *model.AuthLoginResponse
		wantErr      bool
	}{
		{
			name: "success login",
			svc:  &authService{},
			args: args{ctx: context.Background(), req: model.AuthLoginRequest{Email: "test@example.com", Password: "12345pass"}},
			prepareMocks: func(m *mocks) {
				m.cfgMock.Web.RefreshTokenTTL = time.Hour
				m.cfgMock.Web.AccessTokenTTL = time.Minute
				m.utMock.Patch("ValidateRequest", func(interface{}) error {
					return nil
				})
				m.ownerRepoMock.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(&model.Owner{Email: "test@example.com", Password: "12345pass"}, nil, nil)
				m.utMock.Patch("ValidatePassword", func(string, string) error {
					return nil
				})
				m.authRepoMock.EXPECT().GetSessionIDByEmail(gomock.Any(), gomock.Any()).Return("", nil)
				m.utMock.Patch("GenerateToken", func(t time.Duration, jti string, email string) (string, error) {
					if email != "" && jti == "" {
						return "password-token", nil
					}
					if jti != "" && email == "" {
						return "refresh-token", nil
					}
					return "access-token", nil
				})
				m.authRepoMock.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil)
				m.authRepoMock.EXPECT().AccessTokenTTL().Return(time.Minute)
				m.authRepoMock.EXPECT().RefreshTokenTTL().Return(time.Hour)
				// m.mailerMock.EXPECT().SendEmailNotifyLogin(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			wantResp: &model.AuthLoginResponse{SID: "sid", AccessToken: "access-token", RefreshToken: "refresh-token"},
			wantErr:  false,
		},
		{
			name: "success login (multiple logged in)",
			svc:  &authService{},
			args: args{ctx: context.Background(), req: model.AuthLoginRequest{Email: "test@example.com", Password: "12345pass"}},
			prepareMocks: func(m *mocks) {
				m.cfgMock.Web.RefreshTokenTTL = time.Hour
				m.cfgMock.Web.AccessTokenTTL = time.Minute
				m.utMock.Patch("ValidateRequest", func(interface{}) error {
					return nil
				})
				m.ownerRepoMock.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(&model.Owner{Email: "test@example.com", Password: "12345pass"}, nil, nil)
				m.utMock.Patch("ValidatePassword", func(string, string) error {
					return nil
				})
				m.authRepoMock.EXPECT().GetSessionIDByEmail(gomock.Any(), gomock.Any()).Return("sid", nil)
			},
			wantResp: &model.AuthLoginResponse{SID: "sid"},
			wantErr:  false,
		},
		{
			name: "fail login (validate request error)",
			svc:  &authService{},
			args: args{
				ctx: context.Background(),
				req: model.AuthLoginRequest{
					Email: "invalid@email",
				},
			},
			prepareMocks: func(m *mocks) {
				m.cfgMock.Web.RefreshTokenTTL = time.Hour
				m.cfgMock.Web.AccessTokenTTL = time.Minute
				m.utMock.Patch("ValidateRequest", func(interface{}) error { return errors.New("oops! invalid request") })
			},
			wantErr: true,
		},
		{
			name: "fail login (validate password error)",
			svc:  &authService{},
			args: args{
				ctx: context.Background(),
				req: model.AuthLoginRequest{
					Email:    "test@example.com",
					Password: "wrong password",
				},
			},
			prepareMocks: func(m *mocks) {
				m.cfgMock.Web.RefreshTokenTTL = time.Hour
				m.cfgMock.Web.AccessTokenTTL = time.Minute
				m.utMock.Patch("ValidateRequest", func(interface{}) error { return nil })
				m.ownerRepoMock.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(&model.Owner{Email: "test@example.com", Password: "12345pass"}, nil, nil)
				m.utMock.Patch("ValidatePassword", func(string, string) error {
					return errors.New("oops! wrong password")
				})
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			utMock := utils.InitMock()
			config.InitMock()
			ownerRepoMock := repository.NewMockOwnerRepository(ctrl)
			authRepoMock := repository.NewMockAuthRepository(ctrl)
			mailerMock := NewMockMailer(ctrl)

			mocks := mocks{
				utMock:        &utMock,
				ownerRepoMock: ownerRepoMock,
				mailerMock:    mailerMock,
				authRepoMock:  authRepoMock,
				cfgMock:       config.Cfg(),
			}
			tt.svc.authRepo = authRepoMock
			tt.svc.ownerRepo = ownerRepoMock
			tt.svc.mailer = mailerMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks)
			}

			gotResp, err := tt.svc.Login(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil, err)
			assert.Condition(t, func() bool {
				if (gotResp == nil || tt.wantResp == nil) && !(gotResp == nil && tt.wantResp == nil) {
					return false
				}
				if gotResp == nil && tt.wantResp == nil {
					return true
				}

				return strings.Compare(gotResp.RefreshToken, tt.wantResp.RefreshToken)+strings.Compare(gotResp.AccessToken, tt.wantResp.AccessToken) == 0

			}, fmt.Sprintf("expected: %s\nactual: %s\n", tt.wantResp, gotResp))

			utMock.UnpatchAll()
			config.DestroyMock()
		})
	}
}

func Test_authService_Logout(t *testing.T) {
	type args struct {
		ctx context.Context
		req model.AuthLogoutRequest
	}
	type mocks struct {
		ownerRepoMock *repository.MockOwnerRepository
		authRepoMock  *repository.MockAuthRepository
		utMock        *utils.Mock
	}
	tests := []struct {
		name         string
		svc          *authService
		args         args
		prepareMocks func(*mocks)
		wantErr      bool
	}{
		{
			name: "success logout",
			svc:  &authService{},
			args: args{ctx: context.Background(), req: model.AuthLogoutRequest{SID: "sid", Password: "12345pass"}},
			prepareMocks: func(m *mocks) {
				m.utMock.Patch("ValueContext", func(ctx context.Context, s string) interface{} {
					if s == "Sid" {
						return "sid"
					}
					if s == "Authorization" {
						return "access-token"
					}
					return ""
				})
				m.utMock.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) { return &utils.JwtClaims{}, nil })
				m.utMock.Patch("ValidateRequest", func(interface{}) error { return nil })
				m.authRepoMock.EXPECT().Session(gomock.Any(), "sid").Return(&model.Auth{OwnerID: int64(1), Valid: true}, nil, nil)
				m.ownerRepoMock.EXPECT().Get(gomock.Any(), int64(1)).Return(&model.Owner{}, nil, nil)
				m.utMock.Patch("ValidatePassword", func(string, string) error { return nil })
				m.authRepoMock.EXPECT().DeleteSession(gomock.Any(), "sid").Return(nil)
			},
		},
		{
			name: "fail logout (error validate)",
			svc:  &authService{},
			args: args{ctx: context.Background(), req: model.AuthLogoutRequest{SID: "sid", Password: "12345pass"}},
			prepareMocks: func(m *mocks) {
				m.utMock.Patch("ValueContext", func(context.Context, string) interface{} { return "" })
				m.utMock.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) { return nil, errors.New("oops! error validate token") })
			},
			wantErr: true,
		},
		{
			name: "fail logout (error validate request)",
			svc:  &authService{},
			args: args{ctx: context.Background(), req: model.AuthLogoutRequest{SID: "sid", Password: "12345pass"}},
			prepareMocks: func(m *mocks) {
				m.utMock.Patch("ValueContext", func(context.Context, string) interface{} { return "access-token" })
				m.utMock.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) { return &utils.JwtClaims{}, nil })
				m.utMock.Patch("ValidateRequest", func(interface{}) error { return errors.New("oops! error validate request") })
			},
			wantErr: true,
		},
		{
			name: "fail logout (error get session)",
			svc:  &authService{},
			args: args{ctx: context.Background(), req: model.AuthLogoutRequest{SID: "sid", Password: "12345pass"}},
			prepareMocks: func(m *mocks) {
				m.utMock.Patch("ValueContext", func(ctx context.Context, s string) interface{} {
					if s == "Sid" {
						return "sid"
					}
					if s == "Authorization" {
						return "access-token"
					}
					return ""
				})
				m.utMock.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) { return &utils.JwtClaims{}, nil })
				m.utMock.Patch("ValidateRequest", func(interface{}) error { return nil })
				m.authRepoMock.EXPECT().Session(gomock.Any(), "sid").Return(nil, nil, errors.New("error get session"))
			},
			wantErr: true,
		},
		{
			name: "fail logout session is invalid",
			svc:  &authService{},
			args: args{ctx: context.Background(), req: model.AuthLogoutRequest{SID: "sid", Password: "12345pass"}},
			prepareMocks: func(m *mocks) {
				m.utMock.Patch("ValueContext", func(ctx context.Context, s string) interface{} {
					if s == "Sid" {
						return "sid"
					}
					if s == "Authorization" {
						return "access-token"
					}
					return ""
				})
				m.utMock.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) { return &utils.JwtClaims{}, nil })
				m.utMock.Patch("ValidateRequest", func(interface{}) error { return nil })
				m.authRepoMock.EXPECT().Session(gomock.Any(), "sid").Return(&model.Auth{OwnerID: int64(1), Valid: false}, nil, nil)
			},
			wantErr: true,
		},
		{
			name: "fail logout (error get owner)",
			svc:  &authService{},
			args: args{ctx: context.Background(), req: model.AuthLogoutRequest{SID: "sid", Password: "12345pass"}},
			prepareMocks: func(m *mocks) {
				m.utMock.Patch("ValueContext", func(ctx context.Context, s string) interface{} {
					if s == "Sid" {
						return "sid"
					}
					if s == "Authorization" {
						return "access-token"
					}
					return ""
				})
				m.utMock.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) { return &utils.JwtClaims{}, nil })
				m.utMock.Patch("ValidateRequest", func(interface{}) error { return nil })
				m.authRepoMock.EXPECT().Session(gomock.Any(), "sid").Return(&model.Auth{OwnerID: int64(1), Valid: true}, nil, nil)
				m.ownerRepoMock.EXPECT().Get(gomock.Any(), int64(1)).Return(nil, nil, errors.New("error get owner"))
			},
			wantErr: true,
		},
		{
			name: "fail logout (error validate password)",
			svc:  &authService{},
			args: args{ctx: context.Background(), req: model.AuthLogoutRequest{SID: "sid", Password: "12345pass"}},
			prepareMocks: func(m *mocks) {
				m.utMock.Patch("ValueContext", func(ctx context.Context, s string) interface{} {
					if s == "Sid" {
						return "sid"
					}
					if s == "Authorization" {
						return "access-token"
					}
					return ""
				})
				m.utMock.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) { return &utils.JwtClaims{}, nil })
				m.utMock.Patch("ValidateRequest", func(interface{}) error { return nil })
				m.authRepoMock.EXPECT().Session(gomock.Any(), "sid").Return(&model.Auth{OwnerID: int64(1), Valid: true}, nil, nil)
				m.ownerRepoMock.EXPECT().Get(gomock.Any(), int64(1)).Return(&model.Owner{}, nil, nil)
				m.utMock.Patch("ValidatePassword", func(string, string) error { return errors.New("oops! error validate password") })
			},
			wantErr: true,
		},
		{
			name: "fail logout (error delete owner's session)",
			svc:  &authService{},
			args: args{ctx: context.Background(), req: model.AuthLogoutRequest{SID: "sid", Password: "12345pass"}},
			prepareMocks: func(m *mocks) {
				m.utMock.Patch("ValueContext", func(ctx context.Context, s string) interface{} {
					if s == "Sid" {
						return "sid"
					}
					if s == "Authorization" {
						return "access-token"
					}
					return ""
				})
				m.utMock.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) { return &utils.JwtClaims{}, nil })
				m.utMock.Patch("ValidateRequest", func(interface{}) error { return nil })
				m.authRepoMock.EXPECT().Session(gomock.Any(), "sid").Return(&model.Auth{OwnerID: int64(1), Valid: true}, nil, nil)
				m.ownerRepoMock.EXPECT().Get(gomock.Any(), int64(1)).Return(&model.Owner{}, nil, nil)
				m.utMock.Patch("ValidatePassword", func(string, string) error { return nil })
				m.authRepoMock.EXPECT().DeleteSession(gomock.Any(), "sid").Return(errors.New("oops! error delete session"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			utMock := utils.InitMock()
			ownerRepoMock := repository.NewMockOwnerRepository(ctrl)
			authRepoMock := repository.NewMockAuthRepository(ctrl)

			mocks := mocks{
				utMock:        &utMock,
				ownerRepoMock: ownerRepoMock,
				authRepoMock:  authRepoMock,
			}
			tt.svc.authRepo = authRepoMock
			tt.svc.ownerRepo = ownerRepoMock
			tt.svc.mailer = nil

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks)
			}

			err := tt.svc.Logout(tt.args.ctx, tt.args.req)
			assert.Equal(t, err != nil, tt.wantErr, err)

			utMock.UnpatchAll()
		})
	}
}

func Test_authService_ForgotPassword(t *testing.T) {
	type args struct {
		ctx context.Context
		req model.AuthForgotPasswordRequest
	}
	type mocks struct {
		ownerRepoMock *repository.MockOwnerRepository
		authRepoMock  *repository.MockAuthRepository
		mailerMock    *MockMailer
		utMock        *utils.Mock
		cfgMock       *config.MockConfig
	}
	tests := []struct {
		name         string
		svc          *authService
		args         args
		prepareMocks func(*mocks)
		want         string
		wantErr      bool
	}{
		{
			name: "success forgot password",
			svc:  &authService{},
			args: args{
				ctx: context.Background(),
				req: model.AuthForgotPasswordRequest{
					Email: "test@example.com",
				},
			},
			prepareMocks: func(m *mocks) {
				m.cfgMock.Web.AccessTokenTTL = time.Minute
				m.utMock.Patch("ValidateRequest", func(interface{}) error { return nil })
				m.ownerRepoMock.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(&model.Owner{Email: "test@example.com", Name: "test"}, nil, nil)
				m.authRepoMock.EXPECT().AccessTokenTTL().Return(time.Minute)
				m.utMock.Patch("GenerateToken", func(t time.Duration, jti string, email string) (string, error) { return "password-token", nil }) // bot email and jti must not be empty
				m.mailerMock.EXPECT().SendEMailForgotPassword([]string{"test@example.com"}, gomock.Any(), gomock.Any(), "test", gomock.Any()).Return(nil)
			},
			want: "password-token",
		},
		{
			name: "fail forgot password (error validate request)",
			svc:  &authService{},
			args: args{
				ctx: context.Background(),
				req: model.AuthForgotPasswordRequest{
					Email: "test@example.com",
				},
			},
			prepareMocks: func(m *mocks) {
				m.cfgMock.Web.AccessTokenTTL = time.Minute
				m.utMock.Patch("ValidateRequest", func(interface{}) error { return errors.New("oops! error validate request") })
			},
			wantErr: true,
		},
		{
			name: "fail forgot password (error email not found)",
			svc:  &authService{},
			args: args{
				ctx: context.Background(),
				req: model.AuthForgotPasswordRequest{
					Email: "test@example.com",
				},
			},
			prepareMocks: func(m *mocks) {
				m.cfgMock.Web.AccessTokenTTL = time.Minute
				m.utMock.Patch("ValidateRequest", func(interface{}) error { return nil })
				m.ownerRepoMock.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(nil, errors.New("oops! error email not found"), nil)
			},
			wantErr: true,
		},
		{
			name: "fail forgot password (error get email)",
			svc:  &authService{},
			args: args{
				ctx: context.Background(),
				req: model.AuthForgotPasswordRequest{
					Email: "test@example.com",
				},
			},
			prepareMocks: func(m *mocks) {
				m.cfgMock.Web.AccessTokenTTL = time.Minute
				m.utMock.Patch("ValidateRequest", func(interface{}) error { return nil })
				m.ownerRepoMock.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(nil, nil, errors.New("oops! error server get email"))
			},
			wantErr: true,
		},
		{
			name: "fail forgot password (error generate token)",
			svc:  &authService{},
			args: args{
				ctx: context.Background(),
				req: model.AuthForgotPasswordRequest{
					Email: "test@example.com",
				},
			},
			prepareMocks: func(m *mocks) {
				m.cfgMock.Web.AccessTokenTTL = time.Minute
				m.utMock.Patch("ValidateRequest", func(interface{}) error { return nil })
				m.ownerRepoMock.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(&model.Owner{Email: "test@example.com", Name: "test"}, nil, nil)
				m.authRepoMock.EXPECT().AccessTokenTTL().Return(time.Minute)
				m.utMock.Patch("GenerateToken", func(t time.Duration, jti string, email string) (string, error) {
					return "", errors.New("oops! error generate token")
				})
			},
			wantErr: true,
		},
		{
			name: "fail forgot password (email not sent)",
			svc:  &authService{},
			args: args{
				ctx: context.Background(),
				req: model.AuthForgotPasswordRequest{
					Email: "test@example.com",
				},
			},
			prepareMocks: func(m *mocks) {
				m.cfgMock.Web.AccessTokenTTL = time.Minute
				m.utMock.Patch("ValidateRequest", func(interface{}) error { return nil })
				m.ownerRepoMock.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(&model.Owner{Email: "test@example.com", Name: "test"}, nil, nil)
				m.authRepoMock.EXPECT().AccessTokenTTL().Return(time.Minute)
				m.utMock.Patch("GenerateToken", func(t time.Duration, jti string, email string) (string, error) { return "password-token", nil }) // bot email and jti must not be empty
				m.mailerMock.EXPECT().SendEMailForgotPassword([]string{"test@example.com"}, gomock.Any(), gomock.Any(), "test", gomock.Any()).Return(errors.New("oops! some smtp error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			utMock := utils.InitMock()
			config.InitMock()
			ownerRepoMock := repository.NewMockOwnerRepository(ctrl)
			mailerMock := NewMockMailer(ctrl)
			authRepoMock := repository.NewMockAuthRepository(ctrl)

			mocks := mocks{
				utMock:        &utMock,
				ownerRepoMock: ownerRepoMock,
				mailerMock:    mailerMock,
				cfgMock:       config.Cfg(),
				authRepoMock:  authRepoMock,
			}

			tt.svc.ownerRepo = ownerRepoMock
			tt.svc.mailer = mailerMock
			tt.svc.authRepo = authRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks)
			}

			got, err := tt.svc.ForgotPassword(tt.args.ctx, tt.args.req)

			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, got, tt.want)

			utMock.UnpatchAll()
			config.DestroyMock()
		})
	}
}

func Test_authService_RenewAccessToken(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type mocks struct {
		utMock       *utils.Mock
		authRepoMock *repository.MockAuthRepository
		cfgMock      *config.MockConfig
	}
	tests := []struct {
		name         string
		svc          *authService
		args         args
		prepareMocks func(*mocks)
		want         *model.AuthRenewAccessTokenResponse
		wantErr      bool
	}{
		{
			name: "success renew access token",
			svc:  &authService{},
			args: args{ctx: context.Background()},
			prepareMocks: func(m *mocks) {
				m.cfgMock.Web.AccessTokenTTL = time.Minute
				m.utMock.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == consts.CtxKeySession {
						return &model.AuthSessionResponse{SID: "sid", Valid: true, Jti: "jti", OwnerID: 1}
					}
					if key == consts.CtxKeyAuthorization {
						return "refresh-token"
					}
					return nil
				})
				m.utMock.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return utils.NewJWTClaimTesting("jti"), nil
				})
				m.utMock.Patch("GenerateToken", func(time.Duration, string, string) (string, error) { return "access-token", nil })
				m.authRepoMock.EXPECT().AccessTokenTTL().Return(time.Minute).Times(2)

			},
			want: &model.AuthRenewAccessTokenResponse{
				AccessToken: "access-token",
			},
		},
		{
			name: "fail renew access token (error invalid token)",
			svc:  &authService{},
			args: args{ctx: context.Background()},
			prepareMocks: func(m *mocks) {
				m.cfgMock.Web.AccessTokenTTL = time.Minute
				m.utMock.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == "Session" {
						return &model.AuthSessionResponse{SID: "sid", Valid: true, Jti: "jti", OwnerID: 1}
					}
					if key == "Authorization" {
						return "invalid-refresh-token"
					}
					return nil
				})
				m.utMock.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return nil, errors.New("oops! error validate token")
				})
			},
			wantErr: true,
		},
		{
			name: "fail renew access token (error generate token)",
			svc:  &authService{},
			args: args{ctx: context.Background()},
			prepareMocks: func(m *mocks) {
				m.cfgMock.Web.AccessTokenTTL = time.Minute
				m.utMock.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == "Session" {
						return &model.AuthSessionResponse{SID: "sid", Valid: true, Jti: "jti", OwnerID: 1}
					}
					if key == "Authorization" {
						return "refresh-token"
					}
					return nil
				})
				m.utMock.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return utils.NewJWTClaimTesting("jti"), nil
				})
				m.utMock.Patch("GenerateToken", func(time.Duration, string, string) (string, error) {
					return "", errors.New("oops! error generate token")
				})
				m.authRepoMock.EXPECT().AccessTokenTTL().Return(time.Minute).Times(1)

			},
			wantErr: true,
		},
		{
			name: "fail renew access token (error session invalid)",
			svc:  &authService{},
			args: args{ctx: context.Background()},
			prepareMocks: func(m *mocks) {
				m.cfgMock.Web.AccessTokenTTL = time.Minute
				m.utMock.Patch("ValueContext", func(ctx context.Context, key string) interface{} {
					if key == "Session" {
						return &model.AuthSessionResponse{SID: "sid", Valid: false, Jti: "jti", OwnerID: 1}
					}
					if key == "Authorization" {
						return "refresh-token"
					}
					return nil
				})
				m.utMock.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return utils.NewJWTClaimTesting("jti"), nil
				})
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			utMock := utils.InitMock()
			config.InitMock()

			authRepoMock := repository.NewMockAuthRepository(ctrl)

			mocks := mocks{
				utMock:       &utMock,
				cfgMock:      config.Cfg(),
				authRepoMock: authRepoMock,
			}

			tt.svc.authRepo = authRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks)
			}

			got, err := tt.svc.RenewAccessToken(tt.args.ctx)
			var expf, gotf string
			if got != nil {
				gotf = got.AccessToken
			}
			if tt.want != nil {
				expf = tt.want.AccessToken
			}

			assert.Equal(t, tt.wantErr, err != nil)
			assert.Condition(t, func() bool {
				if got == nil && tt.want == nil {
					return true
				}

				if got == nil || tt.want == nil {
					return false
				}

				return strings.Compare(tt.want.AccessToken, got.AccessToken) == 0
			}, fmt.Sprintf("expected: %s\nactual: %s\n", expf, gotf))
		})
	}
}

func Test_authService_Session(t *testing.T) {
	type args struct {
		ctx context.Context
		sid string
	}
	type mocks struct {
		authRepoMock *repository.MockAuthRepository
	}
	tests := []struct {
		name         string
		svc          *authService
		args         args
		prepareMocks func(*mocks)
		want         *model.AuthSessionResponse
		wantErr      bool
	}{
		{
			name: "success get session",
			svc:  &authService{},
			args: args{
				ctx: context.Background(),
				sid: "sid",
			},
			prepareMocks: func(m *mocks) {
				m.authRepoMock.EXPECT().Session(gomock.Any(), "sid").Return(&model.Auth{SID: "sid", OwnerID: 1, Valid: true, Jti: "jti"}, nil, nil)
			},
			want: &model.AuthSessionResponse{SID: "sid", OwnerID: 1, Valid: true, Jti: "jti"},
		},
		{
			name: "fail get session (error server)",
			svc:  &authService{},
			args: args{
				ctx: context.Background(),
				sid: "sid",
			},
			prepareMocks: func(m *mocks) {
				m.authRepoMock.EXPECT().Session(gomock.Any(), "sid").Return(nil, nil, errors.New("error get session / server"))
			},
			wantErr: true,
		},
		{
			name: "fail get session (error no row)",
			svc:  &authService{},
			args: args{
				ctx: context.Background(),
				sid: "sid",
			},
			prepareMocks: func(m *mocks) {
				m.authRepoMock.EXPECT().Session(gomock.Any(), "sid").Return(nil, errors.New("error no session"), nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			authRepoMock := repository.NewMockAuthRepository(ctrl)

			mock := mocks{authRepoMock: authRepoMock}

			tt.svc.authRepo = authRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mock)
			}
			got, err := tt.svc.Session(tt.args.ctx, tt.args.sid)

			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, got, tt.want)
		})
	}
}
