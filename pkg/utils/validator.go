package utils

import (
	"family-catering/pkg/apperrors"

	"github.com/go-playground/validator/v10"
)

// type StructValidator interface {
// 	Struct(s interface{}) error
// }

// type mockStructValidator struct {
// 	f func(s interface{}) error
// }

// func (m mockStructValidator) Struct(s interface{}) error {
// 	return m.f(s)
// }

// func InitMockValidator() {
// 	vdr = mockStructValidator{}
// }

// func DestroyMockValidator() { vdr = validator.New() }

// func resetValidator() { vdr = validator.New() }

var (
	vdr             *validator.Validate = validator.New()
	ValidateRequest func(s interface{}) error
)

func validateRequest(s interface{}) error {
	err := vdr.Struct(s)

	errs, ok := err.(validator.ValidationErrors)
	if ok {
		for _, e := range errs {
			if e != nil {
				if e.Tag() == "required" {
					return apperrors.ErrRequiredParam
				}
				return e
			}
		}
	}
	return nil
}
