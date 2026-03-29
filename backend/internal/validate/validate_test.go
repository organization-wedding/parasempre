package validate

import (
	"strings"
	"testing"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

type testCreateGuest struct {
	FirstName    string `validate:"required"`
	LastName     string `validate:"required"`
	Phone        string `validate:"omitempty,brphone"`
	Relationship string `validate:"required,relationship"`
	FamilyGroup  *int64 `validate:"omitempty,gt=0"`
}

type testRegister struct {
	Phone string `validate:"required,brphone"`
	URACF string `validate:"required,uracf"`
}

func int64Ptr(v int64) *int64 { return &v }

func TestStructValidCreateGuest(t *testing.T) {
	input := testCreateGuest{
		FirstName:    "Maria",
		LastName:     "Santos",
		Phone:        "11988888888",
		Relationship: "R",
		FamilyGroup:  int64Ptr(1),
	}
	if err := Struct(input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStructValidCreateGuestWithoutPhone(t *testing.T) {
	input := testCreateGuest{
		FirstName:    "Maria",
		LastName:     "Santos",
		Relationship: "P",
	}
	if err := Struct(input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStructMissingFirstName(t *testing.T) {
	input := testCreateGuest{
		LastName:     "Santos",
		Relationship: "R",
	}
	err := Struct(input)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "first_name is required") {
		t.Fatalf("expected first_name message, got %q", err.Error())
	}
}

func TestStructMissingLastName(t *testing.T) {
	input := testCreateGuest{
		FirstName:    "Maria",
		Relationship: "R",
	}
	err := Struct(input)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "last_name is required") {
		t.Fatalf("expected last_name message, got %q", err.Error())
	}
}

func TestStructInvalidPhone(t *testing.T) {
	input := testCreateGuest{
		FirstName:    "Maria",
		LastName:     "Santos",
		Phone:        "1188888888",
		Relationship: "R",
	}
	err := Struct(input)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "phone must be a valid BR mobile number") {
		t.Fatalf("expected phone message, got %q", err.Error())
	}
}

func TestStructInvalidRelationship(t *testing.T) {
	input := testCreateGuest{
		FirstName:    "Maria",
		LastName:     "Santos",
		Relationship: "X",
	}
	err := Struct(input)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "relationship must be 'P' or 'R'") {
		t.Fatalf("expected relationship message, got %q", err.Error())
	}
}

func TestStructMissingRelationship(t *testing.T) {
	input := testCreateGuest{
		FirstName: "Maria",
		LastName:  "Santos",
	}
	err := Struct(input)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "relationship must be 'P' or 'R'") {
		t.Fatalf("expected relationship message, got %q", err.Error())
	}
}

func TestStructFamilyGroupZero(t *testing.T) {
	fg := int64(0)
	input := testCreateGuest{
		FirstName:    "Maria",
		LastName:     "Santos",
		Relationship: "R",
		FamilyGroup:  &fg,
	}
	err := Struct(input)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "family_group must be greater than 0") {
		t.Fatalf("expected family_group message, got %q", err.Error())
	}
}

func TestStructFamilyGroupNegative(t *testing.T) {
	fg := int64(-1)
	input := testCreateGuest{
		FirstName:    "Maria",
		LastName:     "Santos",
		Relationship: "R",
		FamilyGroup:  &fg,
	}
	err := Struct(input)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "family_group must be greater than 0") {
		t.Fatalf("expected family_group message, got %q", err.Error())
	}
}

func TestStructRegisterValid(t *testing.T) {
	input := testRegister{
		Phone: "11999999999",
		URACF: "USR01",
	}
	if err := Struct(input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStructRegisterInvalidURACF(t *testing.T) {
	input := testRegister{
		Phone: "11999999999",
		URACF: "toolong1",
	}
	err := Struct(input)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "uracf must be exactly 5 uppercase alphanumeric characters") {
		t.Fatalf("expected uracf message, got %q", err.Error())
	}
}

func TestStructReturnsAppError(t *testing.T) {
	input := testCreateGuest{}
	err := Struct(input)
	if err == nil {
		t.Fatal("expected error")
	}
	ae, ok := apperror.IsAppError(err)
	if !ok {
		t.Fatal("expected AppError")
	}
	if ae.Code != 400 {
		t.Fatalf("expected 400, got %d", ae.Code)
	}
}

func TestBrPhoneRegexEdgeCases(t *testing.T) {
	tests := []struct {
		phone string
		valid bool
	}{
		{"11999999999", true},
		{"21912345678", true},
		{"1199999999", false},
		{"119999999990", false},
		{"11888888888", false},
		{"abc", false},
		{"", false},
	}

	type phoneOnly struct {
		Phone string `validate:"required,brphone"`
	}

	for _, tt := range tests {
		err := Struct(phoneOnly{Phone: tt.phone})
		if tt.valid && err != nil {
			t.Errorf("phone %q: expected valid, got error: %v", tt.phone, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("phone %q: expected invalid, got nil", tt.phone)
		}
	}
}
