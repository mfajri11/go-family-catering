package postgres

import (
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type args struct {
		source string
		opts   []Option
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Success NewPostgreClient without option",
			args: args{
				source: DataSourcef("postgres", "root", "localhost", 5432, "test"),
			},
		},

		{
			name: "Success NewPostgreClient with options",
			args: args{
				source: DataSourcef("postgres", "root", "localhost", 5432, "test"),
				opts:   []Option{WithMaxIdleConns(2), WithMaxLifeTime(1 * time.Minute), WithMaxOpenConnection(2)},
			},
		},

		{
			name:    "Error NewPostgreClient",
			args:    args{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.source, tt.args.opts...)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, got == nil, tt.wantErr)
			if got != nil {
				assert.NoError(t, got.Close())
			}
		})
	}
}
