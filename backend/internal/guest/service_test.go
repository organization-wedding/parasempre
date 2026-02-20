package guest

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockRepository implements Repository with function fields for easy stubbing.
type mockRepository struct {
	listFn               func(ctx context.Context) ([]Guest, error)
	getByID              func(ctx context.Context, id int64) (*Guest, error)
	getByPhone           func(ctx context.Context, phone string) (*Guest, error)
	getByNameFn          func(ctx context.Context, firstName, lastName string) (*Guest, error)
	familyGroupExistsFn  func(ctx context.Context, familyGroup int64) (bool, error)
	getNextFamilyGroupFn func(ctx context.Context) (int64, error)
	createFn             func(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error)
	updateFn             func(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error)
	deleteFn             func(ctx context.Context, id int64) error
}

func (m *mockRepository) List(ctx context.Context) ([]Guest, error) {
	return m.listFn(ctx)
}

func (m *mockRepository) GetByID(ctx context.Context, id int64) (*Guest, error) {
	return m.getByID(ctx, id)
}

func (m *mockRepository) GetByPhone(ctx context.Context, phone string) (*Guest, error) {
	return m.getByPhone(ctx, phone)
}

func (m *mockRepository) GetByName(ctx context.Context, firstName, lastName string) (*Guest, error) {
	if m.getByNameFn != nil {
		return m.getByNameFn(ctx, firstName, lastName)
	}
	return nil, nil
}

func (m *mockRepository) FamilyGroupExists(ctx context.Context, familyGroup int64) (bool, error) {
	if m.familyGroupExistsFn != nil {
		return m.familyGroupExistsFn(ctx, familyGroup)
	}
	return true, nil
}

func (m *mockRepository) GetNextFamilyGroup(ctx context.Context) (int64, error) {
	if m.getNextFamilyGroupFn != nil {
		return m.getNextFamilyGroupFn(ctx)
	}
	return 1, nil
}

func (m *mockRepository) Create(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
	return m.createFn(ctx, input, userRACF)
}

func (m *mockRepository) Update(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error) {
	return m.updateFn(ctx, id, input, userRACF)
}

func (m *mockRepository) Delete(ctx context.Context, id int64) error {
	return m.deleteFn(ctx, id)
}

// mockUserChecker implements UserChecker for tests.
type mockUserChecker struct {
	existsFn func(ctx context.Context, uracf string) (bool, error)
}

func (m *mockUserChecker) UserExistsByURACF(ctx context.Context, uracf string) (bool, error) {
	if m.existsFn != nil {
		return m.existsFn(ctx, uracf)
	}
	return true, nil
}

func strPtr(s string) *string { return &s }

func int64Ptr(v int64) *int64 { return &v }

func sampleGuest() Guest {
	phone := "11999999999"
	return Guest{
		ID:           1,
		FirstName:    "João",
		LastName:     "Silva",
		Phone:        &phone,
		Relationship: "P",
		Confirmed:    false,
		FamilyGroup:  1,
		CreatedBy:    "TST01",
		UpdatedBy:    "TST01",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// alwaysExistsChecker returns a UserChecker that always reports the user exists.
func alwaysExistsChecker() *mockUserChecker {
	return &mockUserChecker{existsFn: func(ctx context.Context, uracf string) (bool, error) {
		return true, nil
	}}
}

func TestServiceList(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(ctx context.Context) ([]Guest, error)
		wantLen int
		wantErr bool
	}{
		{
			name: "returns guests",
			mockFn: func(ctx context.Context) ([]Guest, error) {
				return []Guest{sampleGuest()}, nil
			},
			wantLen: 1,
		},
		{
			name: "returns empty list",
			mockFn: func(ctx context.Context) ([]Guest, error) {
				return []Guest{}, nil
			},
			wantLen: 0,
		},
		{
			name: "propagates error",
			mockFn: func(ctx context.Context) ([]Guest, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&mockRepository{listFn: tt.mockFn}, alwaysExistsChecker())
			guests, err := svc.List(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(guests) != tt.wantLen {
				t.Fatalf("expected %d guests, got %d", tt.wantLen, len(guests))
			}
		})
	}
}

func TestServiceGetByID(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		mockFn  func(ctx context.Context, id int64) (*Guest, error)
		wantErr bool
	}{
		{
			name: "returns guest",
			id:   1,
			mockFn: func(ctx context.Context, id int64) (*Guest, error) {
				g := sampleGuest()
				return &g, nil
			},
		},
		{
			name: "not found",
			id:   999,
			mockFn: func(ctx context.Context, id int64) (*Guest, error) {
				return nil, ErrNotFound
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&mockRepository{getByID: tt.mockFn}, alwaysExistsChecker())
			guest, err := svc.GetByID(context.Background(), tt.id)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if guest == nil {
				t.Fatal("expected guest, got nil")
			}
		})
	}
}

func TestServiceCreate(t *testing.T) {
	tests := []struct {
		name    string
		input   CreateGuestInput
		wantErr string
	}{
		{
			name: "valid input with phone",
			input: CreateGuestInput{
				FirstName:    "Maria",
				LastName:     "Santos",
				Phone:        "11988888888",
				Relationship: "R",
				FamilyGroup:  int64Ptr(1),
			},
		},
		{
			name: "valid input without phone",
			input: CreateGuestInput{
				FirstName:    "Maria",
				LastName:     "Santos",
				Relationship: "P",
				FamilyGroup:  int64Ptr(1),
			},
		},
		{
			name: "missing first_name",
			input: CreateGuestInput{
				LastName:     "Santos",
				Phone:        "11988888888",
				Relationship: "R",
				FamilyGroup:  int64Ptr(1),
			},
			wantErr: "first_name is required",
		},
		{
			name: "missing last_name",
			input: CreateGuestInput{
				FirstName:    "Maria",
				Phone:        "11988888888",
				Relationship: "R",
				FamilyGroup:  int64Ptr(1),
			},
			wantErr: "last_name is required",
		},
		{
			name: "invalid phone format",
			input: CreateGuestInput{
				FirstName:    "Maria",
				LastName:     "Santos",
				Phone:        "1188888888",
				Relationship: "R",
				FamilyGroup:  int64Ptr(1),
			},
			wantErr: "phone must be a valid BR mobile number (11 digits: DDD + 9 + 8 digits)",
		},
		{
			name: "phone without leading 9",
			input: CreateGuestInput{
				FirstName:    "Maria",
				LastName:     "Santos",
				Phone:        "11888888888",
				Relationship: "R",
				FamilyGroup:  int64Ptr(1),
			},
			wantErr: "phone must be a valid BR mobile number (11 digits: DDD + 9 + 8 digits)",
		},
		{
			name: "invalid relationship",
			input: CreateGuestInput{
				FirstName:    "Maria",
				LastName:     "Santos",
				Phone:        "11988888888",
				Relationship: "X",
				FamilyGroup:  int64Ptr(1),
			},
			wantErr: "relationship must be 'P' or 'R'",
		},
		{
			name: "missing relationship",
			input: CreateGuestInput{
				FirstName:   "Maria",
				LastName:    "Santos",
				Phone:       "11988888888",
				FamilyGroup: int64Ptr(1),
			},
			wantErr: "relationship must be 'P' or 'R'",
		},
		{
			name: "missing family_group",
			input: CreateGuestInput{
				FirstName:    "Maria",
				LastName:     "Santos",
				Phone:        "11988888888",
				Relationship: "R",
			},
		},
		{
			name: "family_group zero",
			input: CreateGuestInput{
				FirstName:    "Maria",
				LastName:     "Santos",
				Phone:        "11988888888",
				Relationship: "R",
				FamilyGroup:  int64Ptr(0),
			},
			wantErr: "family_group must be greater than 0",
		},
		{
			name: "family_group negative",
			input: CreateGuestInput{
				FirstName:    "Maria",
				LastName:     "Santos",
				Phone:        "11988888888",
				Relationship: "R",
				FamilyGroup:  int64Ptr(-1),
			},
			wantErr: "family_group must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				createFn: func(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
					g := sampleGuest()
					return &g, nil
				},
				getByPhone: func(ctx context.Context, phone string) (*Guest, error) {
					return nil, nil
				},
			}
			svc := NewService(repo, alwaysExistsChecker())
			_, err := svc.Create(context.Background(), tt.input, "TST01")
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestServiceCreateDuplicateName(t *testing.T) {
	repo := &mockRepository{
		getByNameFn: func(ctx context.Context, firstName, lastName string) (*Guest, error) {
			g := sampleGuest()
			return &g, nil
		},
		getByPhone: func(ctx context.Context, phone string) (*Guest, error) {
			return nil, nil
		},
	}
	svc := NewService(repo, alwaysExistsChecker())

	_, err := svc.Create(context.Background(), CreateGuestInput{
		FirstName:    "João",
		LastName:     "Silva",
		Phone:        "11988888888",
		Relationship: "P",
		FamilyGroup:  int64Ptr(1),
	}, "TST01")

	if err == nil {
		t.Fatal("expected duplicate name error, got nil")
	}
	if err.Error() != "a guest named 'João Silva' already exists" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceCreateFamilyGroupNotFound(t *testing.T) {
	repo := &mockRepository{
		getByNameFn: func(ctx context.Context, firstName, lastName string) (*Guest, error) {
			return nil, nil
		},
		getByPhone: func(ctx context.Context, phone string) (*Guest, error) {
			return nil, nil
		},
		familyGroupExistsFn: func(ctx context.Context, familyGroup int64) (bool, error) {
			return false, nil
		},
	}
	svc := NewService(repo, alwaysExistsChecker())

	_, err := svc.Create(context.Background(), CreateGuestInput{
		FirstName:    "Maria",
		LastName:     "Santos",
		Phone:        "11988888888",
		Relationship: "R",
		FamilyGroup:  int64Ptr(10),
	}, "TST01")

	if err == nil {
		t.Fatal("expected family_group validation error, got nil")
	}
	if err.Error() != "family_group not found" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceCreateAutoAssignFamilyGroup(t *testing.T) {
	repo := &mockRepository{
		getByNameFn: func(ctx context.Context, firstName, lastName string) (*Guest, error) {
			return nil, nil
		},
		getByPhone: func(ctx context.Context, phone string) (*Guest, error) {
			return nil, nil
		},
		getNextFamilyGroupFn: func(ctx context.Context) (int64, error) {
			return 42, nil
		},
		createFn: func(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
			if input.FamilyGroup == nil || *input.FamilyGroup != 42 {
				t.Fatalf("expected family_group to be auto-assigned as 42, got %+v", input.FamilyGroup)
			}
			g := sampleGuest()
			g.FamilyGroup = 42
			return &g, nil
		},
	}
	svc := NewService(repo, alwaysExistsChecker())

	_, err := svc.Create(context.Background(), CreateGuestInput{
		FirstName:    "Maria",
		LastName:     "Santos",
		Phone:        "11988888888",
		Relationship: "R",
	}, "TST01")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceCreateDuplicatePhone(t *testing.T) {
	repo := &mockRepository{
		getByNameFn: func(ctx context.Context, firstName, lastName string) (*Guest, error) {
			return nil, nil
		},
		getByPhone: func(ctx context.Context, phone string) (*Guest, error) {
			g := sampleGuest()
			return &g, nil
		},
	}
	svc := NewService(repo, alwaysExistsChecker())

	_, err := svc.Create(context.Background(), CreateGuestInput{
		FirstName:    "Maria",
		LastName:     "Santos",
		Phone:        "11999999999",
		Relationship: "R",
		FamilyGroup:  int64Ptr(1),
	}, "TST01")

	if err == nil {
		t.Fatal("expected duplicate phone error, got nil")
	}
	if err.Error() != "a guest with phone '11999999999' already exists" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceCreateUserRACFNotFound(t *testing.T) {
	repo := &mockRepository{}
	checker := &mockUserChecker{existsFn: func(ctx context.Context, uracf string) (bool, error) {
		return false, nil
	}}
	svc := NewService(repo, checker)

	_, err := svc.Create(context.Background(), CreateGuestInput{
		FirstName:    "Maria",
		LastName:     "Santos",
		Relationship: "R",
		FamilyGroup:  int64Ptr(1),
	}, "TST01")

	if err == nil {
		t.Fatal("expected user not found error, got nil")
	}
	if err.Error() != "user-racf does not match any registered user" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceUpdate(t *testing.T) {
	relP := "P"
	invalid := "X"
	badPhone := "1188888888"

	tests := []struct {
		name    string
		input   UpdateGuestInput
		wantErr string
	}{
		{
			name:  "valid partial update",
			input: UpdateGuestInput{Relationship: &relP},
		},
		{
			name:  "empty update is valid",
			input: UpdateGuestInput{},
		},
		{
			name:    "invalid relationship",
			input:   UpdateGuestInput{Relationship: &invalid},
			wantErr: "relationship must be 'P' or 'R'",
		},
		{
			name:    "invalid phone format",
			input:   UpdateGuestInput{Phone: &badPhone},
			wantErr: "phone must be a valid BR mobile number (11 digits: DDD + 9 + 8 digits)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				updateFn: func(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error) {
					g := sampleGuest()
					return &g, nil
				},
				getByPhone: func(ctx context.Context, phone string) (*Guest, error) {
					return nil, nil
				},
				getByID: func(ctx context.Context, id int64) (*Guest, error) {
					g := sampleGuest()
					return &g, nil
				},
			}
			svc := NewService(repo, alwaysExistsChecker())
			_, err := svc.Update(context.Background(), 1, tt.input, "TST01")
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestServiceUpdateUserRACFNotFound(t *testing.T) {
	repo := &mockRepository{}
	checker := &mockUserChecker{existsFn: func(ctx context.Context, uracf string) (bool, error) {
		return false, nil
	}}
	svc := NewService(repo, checker)

	relP := "P"
	_, err := svc.Update(context.Background(), 1, UpdateGuestInput{Relationship: &relP}, "TST01")

	if err == nil {
		t.Fatal("expected user not found error, got nil")
	}
	if err.Error() != "user-racf does not match any registered user" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceDelete(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(ctx context.Context, id int64) error
		wantErr bool
	}{
		{
			name: "success",
			mockFn: func(ctx context.Context, id int64) error {
				return nil
			},
		},
		{
			name: "not found",
			mockFn: func(ctx context.Context, id int64) error {
				return ErrNotFound
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&mockRepository{deleteFn: tt.mockFn}, alwaysExistsChecker())
			err := svc.Delete(context.Background(), 1)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
