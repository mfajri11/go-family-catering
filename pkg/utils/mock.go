package utils

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Mock struct{}

func InitMock() Mock {
	return Mock{}
}

func (m Mock) Patch(targetFunc string, f interface{}) {
	switch strings.ToLower(targetFunc) {

	case "validaterequest":
		newF, ok := f.(func(interface{}) error)
		if !ok {
			err := fmt.Errorf("utils.Mock.Patch: ValidateRequest type miss match want func(s interface{}) error, got %T", f)
			panic(err)
		}
		ValidateRequest = newF

	case "hashpassword":
		newF, ok := f.(func(string) (string, error))
		if !ok {
			err := fmt.Errorf("utils.Mock.Patch: HashPassword type miss match want func(string) (string, error) got %T", f)
			panic(err)
		}
		HashPassword = newF
	case "validatepassword":
		newF, ok := f.(func(string, string) error)
		if !ok {
			err := fmt.Errorf("utils.Mock.Patch: ValidatePassword type miss match want func(string, string) error got %T", f)
			panic(err)
		}
		ValidatePassword = newF
	// case "generaterandomint64":
	// 	newF, ok := f.(func() (int64, error))
	// 	if !ok {
	// 		err := fmt.Errorf("utils.Mock.Patch: GenerateRandomInt64 type miss match want func() (int64, error) got %T", f)
	// 		panic(err)
	// 	}
	// 	GenerateRandomInt64 = newF

	// case "generateaccesstoken":
	// 	newF, ok := f.(func(time.Duration, string) (string, error))
	// 	if !ok {
	// 		err := fmt.Errorf("utils.Mock.Patch: GenerateAccessToken type miss match, want func(time.Duration) (string, error), got %T", f)
	// 		panic(err)
	// 	}
	// 	GenerateAccessToken = newF

	case "generatetoken":
		newF, ok := f.(func(time.Duration, string, string) (string, error))
		if !ok {
			err := fmt.Errorf("utils.Mock.Patch: GenerateToken type miss match, want func(time.Duration, string, string) (string, error), got %T", f)
			panic(err)
		}
		GenerateToken = newF
	case "validatetoken":
		newF, ok := f.(func(string) (*JwtClaims, error))
		if !ok {
			err := fmt.Errorf("utils.Mock.Patch: ValidateToken type miss match, want func(string) (interface{}, error), got %T", f)
			panic(err)
		}
		ValidateToken = newF
	case "contextwithvalue":
		newF, ok := f.(func(context.Context, string, interface{}) context.Context)
		if !ok {
			panic("err")
		}
		ContextWithValue = newF
	case "valuecontext":
		newF, ok := f.(func(ctx context.Context, key string) interface{})
		if !ok {
			panic("err")
		}
		ValueContext = newF
	}
}

func (m Mock) Unpatch(targetFunc string) {
	switch strings.ToLower(targetFunc) {
	case "validaterequest":
		ValidateRequest = validateRequest
	case "hashpassword":
		HashPassword = hashPassword
	case "validatepassword":
		ValidatePassword = validatePassword
	case "generatetoken":
		GenerateToken = generateToken
	// case "generaterandomint64":
	// 	GenerateRandomInt64 = generateRandomInt64
	// case "generateaccesstoken":
	// 	GenerateAccessToken = generateAccessToken
	// case "generaterefreshtoken":
	// 	GenerateRefreshToken = generateRefreshToken
	case "contextwithvalue":
		ContextWithValue = contextWithValue
	case "valuecontext":
		ValueContext = valueContext
	}
}

func (m Mock) UnpatchAll() {
	HashPassword = hashPassword
	ValidatePassword = validatePassword
	GenerateToken = generateToken
	// GenerateRandomInt64 = generateRandomInt64
	// GenerateAccessToken = generateAccessToken
	// GenerateRefreshToken = generateRefreshToken
	ValidateRequest = validateRequest
	ValueContext = valueContext
	ContextWithValue = contextWithValue
}
