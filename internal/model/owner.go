package model

import "database/sql"

type Owner struct {
	Id          int64          `db:"id"`
	Name        string         `db:"name"`
	Email       string         `db:"email"`
	PhoneNumber string         `db:"phone_number"`
	DateOfBirth sql.NullString `db:"date_of_birth"`
	Password    string         `db:"password"`
}

// // db model
// type CreateOwner struct {
// 	Name        string `db:"name"`
// 	Email       string `db:"email"`
// 	PhoneNumber string `db:"Phone_number"`
// 	Password    string `db:"password"`
// 	DateOfBirth string `db:"date_of_birth"`
// }
// type GetOwner struct {
// 	Id          int64          `db:"id"`
// 	Name        string         `db:"name"`
// 	Email       string         `db:"email"`
// 	PhoneNumber string         `db:"phone_number"`
// 	DateOfBirth sql.NullString `db:"date_of_birth"`
// 	Password    string         `db:"password"`
// }
// type UpdateOwner = GetOwner

// Requests model
type CreateOwnerRequest struct {
	Name        string `json:"name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,alphanum,min=8,omitempty"`
	PhoneNumber string `json:"phone_number" validate:"required,e164"`
} //	@name	create-owner_request
type UpdateOwnerRequest struct {
	Name        string `json:"name" validate:"required"`
	PhoneNumber string `json:"phone_number,omitempty" validate:"omitempty,e164"`
	DateOfBirth string `json:"date_of_birth,omitempty" validate:"omitempty,datetime"`
} //	@name	update-owner_request

type UpdateEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Password        string `json:"password" validate:"required,alphanum,uppercase,lowercase,min=8,omitempty"`
	PasswordConfirm string `json:"password_confirm" validate:"required,alphanum,uppercase,lowercase,min=8,omitempty"`
}

// Response model
type OwnerResponse struct {
	Owner interface{} `json:"owner"`
} //	@name	owner_response

type CreateOwnerResponse struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phone_number"`
} //	@name	owner

type GetOwnerResponse struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
} //	@name	get-update-owner_response

type UpdateOwnerResponse = GetOwnerResponse
