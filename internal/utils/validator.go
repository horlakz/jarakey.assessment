package utils

import "github.com/go-playground/validator/v10"

var requestValidator = validator.New()

func Validate(v interface{}) error {
	return requestValidator.Struct(v)
}
