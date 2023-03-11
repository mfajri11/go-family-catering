package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	type args struct {
		key      string
		fallback string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test GetEnv return string",
			args: args{
				key:      "TESTENV",
				fallback: "fallback-key",
			},
			want: "TESTENV",
		},
		{
			name: "Test GetEnv return fallback",
			args: args{
				key:      "",
				fallback: "fallback-key",
			},
			want: "fallback-key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.args.key, tt.want)
			env := GetEnv(tt.args.key, tt.args.fallback)
			assert.Equal(t, env, tt.want)
		})
	}
}
