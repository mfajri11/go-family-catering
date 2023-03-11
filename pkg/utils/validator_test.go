package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRequest(t *testing.T) {
	type args struct {
		s interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success ValidateRequest",
			args: args{s: struct {
				CharMorethan3 string `validate:"required,min=3"`
				Email         string `validate:"email"`
			}{
				CharMorethan3: "morethan3chars",
				Email:         "test@example.com",
			},
			},
			wantErr: false,
		},
		{
			name: "invalid ValidateRequest (required)",
			args: args{
				s: struct {
					Invalid string `validate:"required"`
				}{},
			},
			wantErr: true,
		},
		{
			name: "invalid ValidateRequest (another error)",
			args: args{
				s: struct {
					Email string `validate:"required,email"`
				}{
					Email: "test@example",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequest(tt.args.s)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
