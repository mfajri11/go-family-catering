package service

import (
	"context"
	"errors"
	"family-catering/internal/model"
	"family-catering/internal/repository"
	"family-catering/pkg/consts"
	"family-catering/pkg/utils"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewMenuService(t *testing.T) {
	type args struct {
		menuRepo repository.MenuRepository
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success NewMenuService",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, NewMenuService(tt.args.menuRepo))
		})
	}
}

func Test_menuService_GetByID(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	type mocks struct {
		utMocks      utils.Mock
		menuRepoMock *repository.MockMenuRepository
	}
	tests := []struct {
		name         string
		svc          *menuService
		args         args
		prepareMocks func(*mocks)
		want         *model.GetMenuResponse
		wantErr      bool
	}{
		{
			name: "success GetByID",
			svc:  &menuService{},
			args: args{
				ctx: utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				id:  1,
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.menuRepoMock.EXPECT().GetByID(gomock.Any(), int64(1)).Return(&model.Menu{ID: 1, Name: "sate", Price: 25_000, Categories: "Indonesian food"}, nil, nil)
			},
			want: &model.GetMenuResponse{
				ID:         1,
				Name:       "sate",
				Price:      25_000,
				Categories: "Indonesian food",
			},
		},
		{
			name: "fail GetByID (invalid token)",
			svc:  &menuService{},
			args: args{
				ctx: utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "invalid-token"),
				id:  1,
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
			name: "fail GetByID (err no row)",
			svc:  &menuService{},
			args: args{
				ctx: utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				id:  1,
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.menuRepoMock.EXPECT().GetByID(gomock.Any(), int64(1)).Return(nil, errors.New("oops! db error no row"), nil)
			},
			wantErr: true,
		},
		{
			name: "fail GetByID (err db)",
			svc:  &menuService{},
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.menuRepoMock.EXPECT().GetByID(gomock.Any(), int64(1)).Return(nil, nil, errors.New("oops! db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			utMocks := utils.InitMock()
			menuRepoMock := repository.NewMockMenuRepository(ctrl)

			tt.svc.menuRepo = menuRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{utMocks: utMocks, menuRepoMock: menuRepoMock})
			}

			got, err := tt.svc.GetByID(tt.args.ctx, tt.args.id)

			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
			utMocks.UnpatchAll()
		})
	}
}

func Test_menuService_GetByName(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}
	type mocks struct {
		utMocks      utils.Mock
		menuRepoMock *repository.MockMenuRepository
	}
	tests := []struct {
		name         string
		svc          *menuService
		args         args
		prepareMocks func(*mocks)
		want         *model.GetMenuResponse
		wantErr      bool
	}{
		{
			name: "success GetByName",
			svc:  &menuService{},
			args: args{
				ctx:  utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				name: "soto betawi",
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.menuRepoMock.EXPECT().GetByName(gomock.Any(), "soto betawi").Return(&model.Menu{ID: 6, Name: "soto betawi", Price: 30_000, Categories: "Indonesian food"}, nil, nil)
			},
			want: &model.GetMenuResponse{
				ID:         6,
				Name:       "soto betawi",
				Price:      30_000,
				Categories: "Indonesian food",
			},
		},
		{
			name: "fail GetByName (invalid token)",
			svc:  &menuService{},
			args: args{
				ctx:  utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "invalid-token"),
				name: "soto betawi",
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "invalid token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return nil, errors.New("oops! invalid token")
				})

			},
			wantErr: true,
		},
		{
			name: "fail GetByName (no rows)",
			svc:  &menuService{},
			args: args{
				ctx:  utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				name: "not-exists-food-name",
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.menuRepoMock.EXPECT().GetByName(gomock.Any(), "not-exists-food-name").Return(nil, errors.New("oops! error no rows"), nil)
			},
			wantErr: true,
		},
		{
			name: "fail GetByName (db error)",
			svc:  &menuService{},
			args: args{
				ctx:  utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				name: "soto betawi",
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.menuRepoMock.EXPECT().GetByName(gomock.Any(), "soto betawi").Return(nil, nil, errors.New("oops! db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			utMocks := utils.InitMock()
			menuRepoMock := repository.NewMockMenuRepository(ctrl)

			tt.svc.menuRepo = menuRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{utMocks: utMocks, menuRepoMock: menuRepoMock})
			}

			got, err := tt.svc.GetByName(tt.args.ctx, tt.args.name)

			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
			utMocks.UnpatchAll()
		})
	}
}

func Test_menuService_List(t *testing.T) {
	type args struct {
		ctx    context.Context
		limit  int
		offset int
	}
	type mocks struct {
		utMocks      utils.Mock
		menuRepoMock *repository.MockMenuRepository
	}
	tests := []struct {
		name         string
		svc          *menuService
		args         args
		prepareMocks func(*mocks)
		want         []*model.GetMenuResponse
		wantErr      bool
	}{
		{
			name: "success GetListMenu",
			svc:  &menuService{},
			args: args{
				ctx:    utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				limit:  2,
				offset: 1,
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.menuRepoMock.EXPECT().List(gomock.Any(), 2, 1).Return([]*model.Menu{
					{ID: 1, Name: "sate", Price: 25_000, Categories: "Indonesian food"},
					{ID: 1, Name: "kerang saus tiram", Price: 44_000, Categories: "Indonesian food"}}, nil, nil)
			},
			want: []*model.GetMenuResponse{
				{ID: 1, Name: "sate", Price: 25_000, Categories: "Indonesian food"},
				{ID: 1, Name: "kerang saus tiram", Price: 44_000, Categories: "Indonesian food"}},
		},
		{
			name: "fail GetListMenu (invalid token)",
			svc:  &menuService{},
			args: args{
				ctx:    utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "invalid-token"),
				limit:  2,
				offset: 1,
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
			name: "success GetListMenu (but no rows)",
			svc:  &menuService{},
			args: args{
				ctx:    utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				limit:  2,
				offset: 1,
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.menuRepoMock.EXPECT().List(gomock.Any(), 2, 1).Return(nil, errors.New("oops! no rows"), nil)
			},
			want:    []*model.GetMenuResponse{},
			wantErr: false, // for list err no rows is not consider as an error (will be used for http 200 with no data)
		},
		{
			name: "success GetListMenu (but no rows)",
			svc:  &menuService{},
			args: args{
				ctx:    utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				limit:  2,
				offset: 1,
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.menuRepoMock.EXPECT().List(gomock.Any(), 2, 1).Return(nil, nil, errors.New("oops! db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			utMocks := utils.InitMock()
			menuRepoMock := repository.NewMockMenuRepository(ctrl)

			tt.svc.menuRepo = menuRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{utMocks: utMocks, menuRepoMock: menuRepoMock})
			}
			got, err := tt.svc.List(tt.args.ctx, tt.args.limit, tt.args.offset)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
			utMocks.UnpatchAll()
		})
	}
}

func Test_menuService_Create(t *testing.T) {
	type args struct {
		ctx context.Context
		req model.CreateMenuRequest
	}
	type mocks struct {
		utMocks      utils.Mock
		menuRepoMock *repository.MockMenuRepository
	}
	tests := []struct {
		name         string
		svc          *menuService
		args         args
		prepareMocks func(*mocks)
		want         *model.CreateMenuResponse
		wantErr      bool
	}{
		{
			name: "success Create",
			svc:  &menuService{},
			args: args{
				ctx: utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				req: model.CreateMenuRequest{
					Name:       "Udon Rice",
					Price:      40_000,
					Categories: "Japanese food",
				},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.utMocks.Patch("ValidateRequest", func(interface{}) error {
					return nil
				})
				m.menuRepoMock.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(10), nil)
			},
			want: &model.CreateMenuResponse{
				ID:         10,
				Name:       "Udon Rice",
				Price:      40_000,
				Categories: "Japanese food",
			},
		},
		{
			name: "fail Create menu (invalid token)",
			svc:  &menuService{},
			args: args{
				ctx: utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "invalid-token"),
				req: model.CreateMenuRequest{
					Name:       "Udon Rice",
					Price:      40_000,
					Categories: "Japanese food",
				},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "invalid token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return nil, errors.New("oops! invalid token")
				})
			},
			wantErr: true,
		},
		{
			name: "fail Create menu (invalid request)",
			svc:  &menuService{},
			args: args{
				ctx: utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				req: model.CreateMenuRequest{
					Name:       "Udon Rice",
					Price:      1e-10, // must be greater than 0.05
					Categories: "Japanese food",
				},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.utMocks.Patch("ValidateRequest", func(interface{}) error {
					return errors.New("invalid request")
				})
			},
			wantErr: true,
		},
		{
			name: "fail Create menu (db error)",
			svc:  &menuService{},
			args: args{
				ctx: utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				req: model.CreateMenuRequest{
					Name:       "Udon Rice",
					Price:      40_000,
					Categories: "Japanese food",
				},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.utMocks.Patch("ValidateRequest", func(interface{}) error {
					return nil
				})
				m.menuRepoMock.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("oops! db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			utMocks := utils.InitMock()
			menuRepoMock := repository.NewMockMenuRepository(ctrl)

			tt.svc.menuRepo = menuRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{utMocks: utMocks, menuRepoMock: menuRepoMock})
			}

			got, err := tt.svc.Create(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
			utMocks.UnpatchAll()
		})
	}
}

func Test_menuService_Update(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
		req model.UpdateMenuRequest
	}
	type mocks struct {
		utMocks      utils.Mock
		menuRepoMock *repository.MockMenuRepository
	}
	tests := []struct {
		name         string
		svc          *menuService
		args         args
		prepareMocks func(*mocks)
		want         *model.UpdateMenuResponse
		wantErr      bool
	}{
		{
			name: "success Update",
			svc:  &menuService{},
			args: args{
				ctx: utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				id:  11,
				req: model.UpdateMenuRequest{
					Name:       "Kerak Telor",
					Price:      30_000,
					Categories: "Indonesian food",
				},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.utMocks.Patch("ValidateRequest", func(interface{}) error {
					return nil
				})
				m.menuRepoMock.EXPECT().Update(gomock.Any(), gomock.Any()).Return(int64(1), nil, nil)
			},
			want: &model.UpdateMenuResponse{
				ID:         11,
				Name:       "Kerak Telor",
				Price:      30_000,
				Categories: "Indonesian food",
			},
		},
		{
			name: "fail Update (no rows)",
			svc:  &menuService{},
			args: args{
				ctx: utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				id:  11,
				req: model.UpdateMenuRequest{
					Name:       "Kerak Telor",
					Price:      30_000,
					Categories: "Indonesian food",
				},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.utMocks.Patch("ValidateRequest", func(interface{}) error {
					return nil
				})
				m.menuRepoMock.EXPECT().Update(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("oops! err no rows"), nil)
			},
			wantErr: true,
		},
		{
			name: "fail Update (invalid token)",
			svc:  &menuService{},
			args: args{
				ctx: utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "invalid-token"),
				id:  11,
				req: model.UpdateMenuRequest{
					Name:       "Kerak Telor",
					Price:      30_000,
					Categories: "Indonesian food",
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
			name: "fail Update (invalid request price must be greater than 0.05)",
			svc:  &menuService{},
			args: args{
				ctx: utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				id:  11,
				req: model.UpdateMenuRequest{
					Name:       "Kerak Telor",
					Price:      3e-5,
					Categories: "Indonesian food",
				},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.utMocks.Patch("ValidateRequest", func(interface{}) error {
					return errors.New("oops! invalid request")
				})
			},
			wantErr: true,
		},
		{
			name: "fail Update (db error)",
			svc:  &menuService{},
			args: args{
				ctx: utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				id:  11,
				req: model.UpdateMenuRequest{
					Name:       "Kerak Telor",
					Price:      35_000,
					Categories: "Indonesian food",
				},
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.utMocks.Patch("ValidateRequest", func(interface{}) error {
					return nil
				})
				m.menuRepoMock.EXPECT().Update(gomock.Any(), gomock.Any()).Return(int64(0), nil, errors.New("oops! db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			utMocks := utils.InitMock()
			menuRepoMock := repository.NewMockMenuRepository(ctrl)

			tt.svc.menuRepo = menuRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{utMocks: utMocks, menuRepoMock: menuRepoMock})
			}

			got, err := tt.svc.Update(tt.args.ctx, tt.args.id, tt.args.req)

			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
			utMocks.UnpatchAll()
		})
	}
}

func Test_menuService_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	type mocks struct {
		utMocks      utils.Mock
		menuRepoMock *repository.MockMenuRepository
	}
	tests := []struct {
		name          string
		svc           *menuService
		args          args
		prepareMocks  func(*mocks)
		wantNAffected int64
		wantErr       bool
	}{
		{
			name: "success Delete",
			svc:  &menuService{},
			args: args{
				ctx: utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				id:  10,
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.menuRepoMock.EXPECT().Delete(gomock.Any(), int64(10)).Return(int64(1), nil, nil)
			},
			wantNAffected: 1,
		},
		{
			name: "fail Delete (invalid token)",
			svc:  &menuService{},
			args: args{
				ctx: utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "invalid-token"),
				id:  10,
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
			name: "fail Delete (err no rows)",
			svc:  &menuService{},
			args: args{
				ctx: utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				id:  1_000_000_000_000_000, // assume no menu with this id
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.menuRepoMock.EXPECT().Delete(gomock.Any(), int64(1_000_000_000_000_000)).Return(int64(0), errors.New("oops! no row"), nil)
			},
			wantErr: true,
		}, {
			name: "fail Delete (db error)",
			svc:  &menuService{},
			args: args{
				ctx: utils.ContextWithValue(context.Background(), consts.CtxKeyAuthorization, "access-token"),
				id:  10,
			},
			prepareMocks: func(m *mocks) {
				m.utMocks.Patch("ValueContext", func(context.Context, string) interface{} {
					return "access-token"
				})
				m.utMocks.Patch("ValidateToken", func(string) (*utils.JwtClaims, error) {
					return &utils.JwtClaims{}, nil
				})
				m.menuRepoMock.EXPECT().Delete(gomock.Any(), int64(10)).Return(int64(0), nil, errors.New("oops! db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			utMocks := utils.InitMock()
			menuRepoMock := repository.NewMockMenuRepository(ctrl)

			tt.svc.menuRepo = menuRepoMock

			if tt.prepareMocks != nil {
				tt.prepareMocks(&mocks{utMocks: utMocks, menuRepoMock: menuRepoMock})
			}

			gotNAffected, err := tt.svc.Delete(tt.args.ctx, tt.args.id)

			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantNAffected, gotNAffected)
			utMocks.UnpatchAll()
		})
	}
}
