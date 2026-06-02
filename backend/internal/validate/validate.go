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

		instance.RegisterValidation("cpf", func(fl validator.FieldLevel) bool {
			return IsValidCPF(fl.Field().String())
		})
	})
	return instance
}

var nonDigitRegex = regexp.MustCompile(`\D`)

func StripCPF(raw string) string {
	return nonDigitRegex.ReplaceAllString(raw, "")
}

func IsValidCPF(raw string) bool {
	digits := StripCPF(raw)
	if len(digits) != 11 {
		return false
	}
	allEqual := true
	for i := 1; i < 11; i++ {
		if digits[i] != digits[0] {
			allEqual = false
			break
		}
	}
	if allEqual {
		return false
	}
	for k := 9; k <= 10; k++ {
		sum := 0
		for i := 0; i < k; i++ {
			sum += int(digits[i]-'0') * (k + 1 - i)
		}
		check := (sum * 10) % 11
		if check == 10 {
			check = 0
		}
		if check != int(digits[k]-'0') {
			return false
		}
	}
	return true
}

var fieldMessages = map[string]map[string]string{
	"FirstName":       {"required": "first_name is required"},
	"LastName":        {"required": "last_name is required"},
	"Phone":           {"required": "phone is required", "brphone": "phone must be a valid BR mobile number (11 digits: DDD + 9 + 8 digits)"},
	"Relationship":    {"required": "relationship must be 'P' or 'R'", "relationship": "relationship must be 'P' or 'R'"},
	"FamilyGroup":     {"gt": "family_group must be greater than 0"},
	"URACF":           {"required": "uracf is required", "uracf": "uracf must be exactly 5 uppercase alphanumeric characters"},
	"Code":            {"required": "code is required", "len": "code must be exactly 6 digits"},
	"Name":            {"required": "name is required", "min": "name must not be empty", "max": "name must be at most 200 characters"},
	"Description":     {"max": "description must be at most 2000 characters"},
	"PriceCents":      {"required": "price_cents is required", "gt": "price_cents must be greater than 0"},
	"ImageURL":        {"url": "image_url must be a valid URL", "startswith": "image_url must start with https://"},
	"StoreURL":        {"url": "store_url must be a valid URL", "startswith": "store_url must start with https://"},
	"Status":          {"giftstatus": "status must be 'active' or 'inactive'"},
	"URL":             {"required": "url é obrigatória", "url": "url deve ser válida", "startswith": "url deve começar com https://"},
	"PaymentMethodID": {"required": "payment_method_id é obrigatório"},
	"Token":           {"required_if": "token é obrigatório para pagamento com cartão"},
	"Installments":    {"min": "installments deve ser pelo menos 1", "max": "installments deve ser no máximo 12"},
	"Email":           {"required": "payer.email é obrigatório", "email": "payer.email deve ser um e-mail válido"},
	"Type":            {"required": "payer.identification.type é obrigatório", "eq": "payer.identification.type deve ser CPF"},
	"Number":          {"required": "payer.identification.number é obrigatório", "cpf": "payer.identification.number deve ser um CPF válido"},
	"IdempotencyKey":  {"required": "idempotency_key é obrigatório", "min": "idempotency_key muito curto", "max": "idempotency_key muito longo"},
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
