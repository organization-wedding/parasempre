package validate

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

var (
	brPhoneRegex = regexp.MustCompile(`^\d{2}9\d{8}$`)
	uracfRegex   = regexp.MustCompile(`^[A-Z0-9]{5}$`)

	instance *validator.Validate
	once     sync.Once
)

func getValidator() *validator.Validate {
	once.Do(func() {
		instance = validator.New()

		instance.RegisterValidation("brphone", func(fl validator.FieldLevel) bool {
			return brPhoneRegex.MatchString(fl.Field().String())
		})

		instance.RegisterValidation("uracf", func(fl validator.FieldLevel) bool {
			return uracfRegex.MatchString(fl.Field().String())
		})

		instance.RegisterValidation("relationship", func(fl validator.FieldLevel) bool {
			val := fl.Field().String()
			return val == "P" || val == "R"
		})

		instance.RegisterValidation("giftstatus", func(fl validator.FieldLevel) bool {
			val := fl.Field().String()
			return val == "active" || val == "inactive"
		})
	})
	return instance
}

var fieldMessages = map[string]map[string]string{
	"FirstName":    {"required": "first_name is required"},
	"LastName":     {"required": "last_name is required"},
	"Phone":        {"required": "phone is required", "brphone": "phone must be a valid BR mobile number (11 digits: DDD + 9 + 8 digits)"},
	"Relationship": {"required": "relationship must be 'P' or 'R'", "relationship": "relationship must be 'P' or 'R'"},
	"FamilyGroup":  {"gt": "family_group must be greater than 0"},
	"URACF":        {"required": "uracf is required", "uracf": "uracf must be exactly 5 uppercase alphanumeric characters"},
	"Code":         {"required": "code is required", "len": "code must be exactly 6 digits"},
	"Name":         {"required": "name is required", "min": "name must not be empty", "max": "name must be at most 200 characters"},
	"Description":  {"max": "description must be at most 2000 characters"},
	"PriceCents":   {"required": "price_cents is required", "gt": "price_cents must be greater than 0"},
	"ImageURL":     {"url": "image_url must be a valid URL", "startswith": "image_url must start with https://"},
	"StoreURL":     {"url": "store_url must be a valid URL", "startswith": "store_url must start with https://"},
	"Status":       {"giftstatus": "status must be 'active' or 'inactive'"},
}

func Struct(s any) error {
	err := getValidator().Struct(s)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return apperror.Validation(err.Error())
	}

	var msgs []string
	for _, fe := range validationErrors {
		if fieldMsgs, exists := fieldMessages[fe.Field()]; exists {
			if msg, exists := fieldMsgs[fe.Tag()]; exists {
				msgs = append(msgs, msg)
				continue
			}
		}
		msgs = append(msgs, fmt.Sprintf("%s failed on %s validation", fe.Field(), fe.Tag()))
	}

	return apperror.Validation(strings.Join(msgs, "; "))
}
