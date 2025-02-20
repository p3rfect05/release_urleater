package validator

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/mail"
	"regexp"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	fmt.Println()
	if err := cv.validator.Struct(i); err != nil {
		return fmt.Errorf("error while validation data | %w", err)
	}

	return nil
}

func ValidateEmail(fl validator.FieldLevel) bool {
	_, err := mail.ParseAddress(fl.Field().String())
	return err == nil
}

func ValidateShortLink(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if len(val) == 0 || len(val) > 20 {
		return false
	}
	isAlphanumeric := regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(val)

	return isAlphanumeric
}

func NewValidator() (*CustomValidator, error) {
	validate := validator.New()
	if err := validate.RegisterValidation("email", ValidateEmail); err != nil {
		return nil, fmt.Errorf("error while register email | %w", err)
	}

	if err := validate.RegisterValidation("short_url", ValidateShortLink); err != nil {
		return nil, fmt.Errorf("error while register short_url | %w", err)
	}

	return &CustomValidator{validator: validate}, nil
}
