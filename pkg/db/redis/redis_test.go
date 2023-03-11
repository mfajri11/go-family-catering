package redis

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisClient(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "success NewRedisClient"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := miniredis.RunT(t) // none running redis server is required
			got, err := New(s.Addr(), tt.args.password)
			if err != nil {
				panic(err)
			}

			assert.NotNil(t, got)
			assert.NoError(t, got.Close())
		})
	}
}

// func Test_redisClient_SetSession(t *testing.T) {
// 	type args struct {
// 		ctx     context.Context
// 		key     string
// 		ownerID int64
// 		jti     string
// 	}
// 	tests := []struct {
// 		name    string
// 		r       *redisClient
// 		args    args
// 		wantErr bool
// 		err     error
// 	}{
// 		{
// 			name: "success SetSession",
// 			args: args{
// 				ctx:     context.Background(),
// 				key:     "sid:value",
// 				ownerID: 1,
// 				jti:     "random-strings",
// 			},
// 		},
// 		{
// 			name: "failed SetSession",
// 			args: args{
// 				ctx:     context.Background(),
// 				key:     "sid:value",
// 				ownerID: 1,
// 				jti:     "random-strings",
// 			},
// 			err:     errors.New("oops! fail write to redis"),
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := miniredis.RunT(t)
// 			c := redis.NewClient(&redis.Options{
// 				Addr: s.Addr(),
// 			})

// 			tt.r = &redisClient{db: c}
// 			if tt.wantErr {
// 				s.SetError(tt.err.Error())
// 			}

// 			err := tt.r.SetSession(tt.args.ctx, tt.args.key, tt.args.ownerID, tt.args.jti)

// 			assert.Equal(t, tt.wantErr, err != nil, err)
// 		})
// 	}
// }

// func Test_redisClient_DeleteSession(t *testing.T) {
// 	type args struct {
// 		ctx context.Context
// 		key string
// 	}
// 	tests := []struct {
// 		name    string
// 		r       RedisClient
// 		args    args
// 		err     error
// 		wantErr bool
// 	}{
// 		{
// 			name: "success DeleteSession",
// 			args: args{
// 				ctx: context.Background(),
// 				key: "sid:value",
// 			},
// 		},
// 		{
// 			name: "fail DeleteSession",
// 			args: args{
// 				ctx: context.Background(),
// 				key: "sid:value",
// 			},
// 			err:     errors.New("oops! fail to delete key"),
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := miniredis.RunT(t)
// 			c := redis.NewClient(&redis.Options{
// 				Addr: s.Addr(),
// 			})
// 			// tt.r = &redisClient{db: c}
// 			r, _ := NewRedisClient(s.Addr(), "", 1)
// 			redismock.NewNiceMock(r)
// 			tt.r = r

// 			if tt.wantErr {
// 				s.SetError(tt.err.Error())
// 			}

// 			tt.r.DeleteSession(tt.args.ctx, tt.args.key)

// 		})
// 	}
// }
