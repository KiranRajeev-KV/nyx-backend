package models

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type RegisterUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r RegisterUserRequest) Validate() (errorMsg string, err error) {
	err = v.ValidateStruct(&r,
		v.Field(&r.Name, v.Required, v.Length(3, 100)),
		v.Field(&r.Email, v.Required, is.Email),
		v.Field(&r.Password, v.Required, v.Length(8, 128)),
	)
	return "Invalid request format for registering a user", err
}

type VerifyOTPRequest struct {
	OTP string `json:"otp"`
}

func (r VerifyOTPRequest) Validate() (errorMsg string, err error) {
	err = v.ValidateStruct(&r,
		v.Field(&r.OTP, v.Required, v.Length(6, 6)),
	)
	return "Invalid request format for verifying OTP", err
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginUserRequest) Validate() (errorMsg string, err error) {
	err = v.ValidateStruct(&r,
		v.Field(&r.Email, v.Required, is.Email),
		v.Field(&r.Password, v.Required, v.Length(8, 128)),
	)
	return "Invalid request format for logging in a user", err
}