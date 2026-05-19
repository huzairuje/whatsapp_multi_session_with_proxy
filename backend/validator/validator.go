package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

func (v *Validator) Struct(s interface{}) error {
	return v.validate.Struct(s)
}

func ValidateStructResponseSliceString(data interface{}) error {
	if data == nil {
		return fmt.Errorf("data is nil")
	}

	switch v := data.(type) {
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("slice is empty")
		}
		for i, item := range v {
			if item == "" {
				return fmt.Errorf("item at index %d is empty", i)
			}
		}
		return nil
	default:
		return fmt.Errorf("expected []string, got %T", data)
	}
}

