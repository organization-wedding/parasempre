package validate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

var (
	brPhoneRegex = regexp.MustCompile(`^\d{2}9\d{8}$`)
	uracfRegex   = regexp.MustCompile(`^[A-Z0-9]{5}$`)

	v *validator.Validate
)

func init() {
	v = validator.New()

	v.RegisterValidation("brphone", func(fl validator.FieldLevel) bool {
		return brPhoneRegex.MatchString(fl.Field().String())
	})

	v.RegisterValidation("uracf", func(fl validator.FieldLevel) bool {
		return uracfRegex.MatchString(fl.Field().String())
	})

	v.RegisterValidation("relationship", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return val == "P" || val == "R"
	})
}

var fieldMessages = map[string]map[string]string{
	"FirstName":    {"required": "first_name is required"},
	"LastName":     {"required": "last_name is required"},
	"Phone":        {"required": "phone is required", "brphone": "phone must be a valid BR mobile number (11 digits: DDD + 9 + 8 digits)"},
	"Relationship": {"required": "relationship must be 'P' or 'R'", "relationship": "relationship must be 'P' or 'R'"},
	"FamilyGroup":  {"gt": "family_group must be greater than 0"},
	"URACF":        {"required": "uracf is required", "uracf": "uracf must be exactly 5 uppercase alphanumeric characters"},
	"Code":         {"required": "code is required", "len": "code must be exactly 6 digits"},
}

func Struct(s any) error {
	err := v.Struct(s)
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
