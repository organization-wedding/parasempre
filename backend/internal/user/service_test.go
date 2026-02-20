package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/guest"
)

// mockUserRepo implements user.Repository with function fields.
type mockUserRepo struct {
	getByURACF   func(ctx context.Context, uracf string) (*User, error)
	getByGuestID func(ctx context.Context, guestID int64) (*User, error)
	createFn     func(ctx context.Context, u *User) (*User, error)
}

func (m *mockUserRepo) GetByURACF(ctx context.Context, uracf string) (*User, error) {
	return m.getByURACF(ctx, uracf)
}

func (m *mockUserRepo) GetByGuestID(ctx context.Context, guestID int64) (*User, error) {
	return m.getByGuestID(ctx, guestID)
}

func (m *mockUserRepo) Create(ctx context.Context, u *User) (*User, error) {
	return m.createFn(ctx, u)
}

// mockGuestRepo implements guest.Repository with function fields.
type mockGuestRepo struct {
	listFn               func(ctx context.Context) ([]guest.Guest, error)
	getByID              func(ctx context.Context, id int64) (*guest.Guest, error)
	getByPhone           func(ctx context.Context, phone string) (*guest.Guest, error)
	getByName            func(ctx context.Context, firstName, lastName string) (*guest.Guest, error)
	familyGroupExistsFn  func(ctx context.Context, familyGroup int64) (bool, error)
	getNextFamilyGroupFn func(ctx context.Context) (int64, error)
	createFn             func(ctx context.Context, input guest.CreateGuestInput, userRACF string) (*guest.Guest, error)
	updateFn             func(ctx context.Context, id int64, input guest.UpdateGuestInput, userRACF string) (*guest.Guest, error)
	deleteFn             func(ctx context.Context, id int64) error
}

func (m *mockGuestRepo) List(ctx context.Context) ([]guest.Guest, error) {
	return m.listFn(ctx)
}

func (m *mockGuestRepo) GetByID(ctx context.Context, id int64) (*guest.Guest, error) {
	return m.getByID(ctx, id)
}

func (m *mockGuestRepo) GetByPhone(ctx context.Context, phone string) (*guest.Guest, error) {
	return m.getByPhone(ctx, phone)
}

func (m *mockGuestRepo) GetByName(ctx context.Context, firstName, lastName string) (*guest.Guest, error) {
	if m.getByName != nil {
		return m.getByName(ctx, firstName, lastName)
	}
	return nil, nil
}

func (m *mockGuestRepo) Create(ctx context.Context, input guest.CreateGuestInput, userRACF string) (*guest.Guest, error) {
	return m.createFn(ctx, input, userRACF)
}

func (m *mockGuestRepo) FamilyGroupExists(ctx context.Context, familyGroup int64) (bool, error) {
	if m.familyGroupExistsFn != nil {
		return m.familyGroupExistsFn(ctx, familyGroup)
	}
	return true, nil
}

func (m *mockGuestRepo) GetNextFamilyGroup(ctx context.Context) (int64, error) {
	if m.getNextFamilyGroupFn != nil {
		return m.getNextFamilyGroupFn(ctx)
	}
	return 1, nil
}

func (m *mockGuestRepo) Update(ctx context.Context, id int64, input guest.UpdateGuestInput, userRACF string) (*guest.Guest, error) {
	return m.updateFn(ctx, id, input, userRACF)
}

func (m *mockGuestRepo) Delete(ctx context.Context, id int64) error {
	return m.deleteFn(ctx, id)
}

func sampleGuest() *guest.Guest {
	phone := "11999999999"
	return &guest.Guest{
		ID:           1,
		FirstName:    "João",
		LastName:     "Silva",
		Phone:        &phone,
		Relationship: "P",
		Confirmed:    false,
		FamilyGroup:  0,
		CreatedBy:    "TST01",
		UpdatedBy:    "TST01",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func sampleUser() *User {
	guestID := int64(1)
	return &User{
		ID:        1,
		GuestID:   &guestID,
		Role:      "guest",
		URACF:     "USR01",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestServiceRegister(t *testing.T) {
	tests := []struct {
		name       string
		input      RegisterInput
		guestFound bool
		userExists bool
		uracfTaken bool
		wantErr    string
	}{
		{
			name:       "valid registration",
			input:      RegisterInput{Phone: "11999999999", URACF: "USR01"},
			guestFound: true,
		},
		{
			name:    "missing phone",
			input:   RegisterInput{Phone: "", URACF: "USR01"},
			wantErr: "phone is required",
		},
		{
			name:    "invalid phone format",
			input:   RegisterInput{Phone: "1188888888", URACF: "USR01"},
			wantErr: "phone must be a valid BR mobile number (11 digits: DDD + 9 + 8 digits)",
		},
		{
			name:    "missing uracf",
			input:   RegisterInput{Phone: "11999999999", URACF: ""},
			wantErr: "uracf is required",
		},
		{
			name:    "invalid uracf format",
			input:   RegisterInput{Phone: "11999999999", URACF: "toolong1"},
			wantErr: "uracf must be exactly 5 uppercase alphanumeric characters",
		},
		{
			name:       "guest not found",
			input:      RegisterInput{Phone: "11999999999", URACF: "USR01"},
			guestFound: false,
			wantErr:    ErrGuestNotFound.Error(),
		},
		{
			name:       "user already registered for guest",
			input:      RegisterInput{Phone: "11999999999", URACF: "USR01"},
			guestFound: true,
			userExists: true,
			wantErr:    ErrAlreadyRegistered.Error(),
		},
		{
			name:       "uracf already taken",
			input:      RegisterInput{Phone: "11999999999", URACF: "USR01"},
			guestFound: true,
			uracfTaken: true,
			wantErr:    ErrURACFTaken.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{
				getByGuestID: func(ctx context.Context, guestID int64) (*User, error) {
					if tt.userExists {
						return sampleUser(), nil
					}
					return nil, nil
				},
				getByURACF: func(ctx context.Context, uracf string) (*User, error) {
					if tt.uracfTaken {
						return sampleUser(), nil
					}
					return nil, nil
				},
				createFn: func(ctx context.Context, u *User) (*User, error) {
					return &User{
						ID:        1,
						GuestID:   u.GuestID,
						Role:      u.Role,
						URACF:     u.URACF,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil
				},
			}

			guestRepo := &mockGuestRepo{
				getByPhone: func(ctx context.Context, phone string) (*guest.Guest, error) {
					if tt.guestFound {
						return sampleGuest(), nil
					}
					return nil, nil
				},
			}

			svc := NewService(userRepo, guestRepo)
			u, err := svc.Register(context.Background(), tt.input)
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
			if u == nil {
				t.Fatal("expected user, got nil")
			}
			if u.URACF != tt.input.URACF {
				t.Fatalf("expected uracf %q, got %q", tt.input.URACF, u.URACF)
			}
		})
	}
}

func TestServiceCheckByPhone(t *testing.T) {
	tests := []struct {
		name       string
		phone      string
		guestFound bool
		userExists bool
		wantExists bool
		wantRole   string
		wantErr    string
	}{
		{
			name:       "user exists",
			phone:      "11999999999",
			guestFound: true,
			userExists: true,
			wantExists: true,
			wantRole:   "guest",
		},
		{
			name:       "guest found but no user",
			phone:      "11999999999",
			guestFound: true,
			userExists: false,
			wantExists: false,
		},
		{
			name:       "guest not found",
			phone:      "11999999999",
			guestFound: false,
			wantExists: false,
		},
		{
			name:    "missing phone",
			phone:   "",
			wantErr: "phone is required",
		},
		{
			name:    "invalid phone format",
			phone:   "abc",
			wantErr: "phone must be a valid BR mobile number (11 digits: DDD + 9 + 8 digits)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{
				getByGuestID: func(ctx context.Context, guestID int64) (*User, error) {
					if tt.userExists {
						return sampleUser(), nil
					}
					return nil, nil
				},
			}

			guestRepo := &mockGuestRepo{
				getByPhone: func(ctx context.Context, phone string) (*guest.Guest, error) {
					if tt.guestFound {
						return sampleGuest(), nil
					}
					return nil, nil
				},
			}

			svc := NewService(userRepo, guestRepo)
			resp, err := svc.CheckByPhone(context.Background(), tt.phone)
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
			if resp.Exists != tt.wantExists {
				t.Fatalf("expected exists=%v, got %v", tt.wantExists, resp.Exists)
			}
			if resp.Role != tt.wantRole {
				t.Fatalf("expected role=%q, got %q", tt.wantRole, resp.Role)
			}
		})
	}
}

func TestServiceSeedCouple(t *testing.T) {
	t.Run("creates both groom and bride", func(t *testing.T) {
		var created []string
		userRepo := &mockUserRepo{
			getByURACF: func(ctx context.Context, uracf string) (*User, error) {
				return nil, nil
			},
			createFn: func(ctx context.Context, u *User) (*User, error) {
				created = append(created, u.Role)
				return &User{ID: 1, Role: u.Role, URACF: u.URACF, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
			},
		}

		svc := NewService(userRepo, &mockGuestRepo{})
		svc.SeedCouple(context.Background(),
			CoupleData{URACF: "GRM01"},
			CoupleData{URACF: "BRD01"},
		)

		if len(created) != 2 {
			t.Fatalf("expected 2 users created, got %d", len(created))
		}
		if created[0] != "groom" {
			t.Fatalf("expected first created to be groom, got %q", created[0])
		}
		if created[1] != "bride" {
			t.Fatalf("expected second created to be bride, got %q", created[1])
		}
	})

	t.Run("skips existing users", func(t *testing.T) {
		var createCalled int
		userRepo := &mockUserRepo{
			getByURACF: func(ctx context.Context, uracf string) (*User, error) {
				return sampleUser(), nil
			},
			createFn: func(ctx context.Context, u *User) (*User, error) {
				createCalled++
				return sampleUser(), nil
			},
		}

		svc := NewService(userRepo, &mockGuestRepo{})
		svc.SeedCouple(context.Background(),
			CoupleData{URACF: "GRM01"},
			CoupleData{URACF: "BRD01"},
		)

		if createCalled != 0 {
			t.Fatalf("expected 0 creates, got %d", createCalled)
		}
	})

	t.Run("skips when uracf is empty", func(t *testing.T) {
		var createCalled int
		userRepo := &mockUserRepo{
			getByURACF: func(ctx context.Context, uracf string) (*User, error) {
				return nil, nil
			},
			createFn: func(ctx context.Context, u *User) (*User, error) {
				createCalled++
				return sampleUser(), nil
			},
		}

		svc := NewService(userRepo, &mockGuestRepo{})
		svc.SeedCouple(context.Background(),
			CoupleData{URACF: ""},
			CoupleData{URACF: ""},
		)

		if createCalled != 0 {
			t.Fatalf("expected 0 creates, got %d", createCalled)
		}
	})

	t.Run("handles repo error gracefully", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getByURACF: func(ctx context.Context, uracf string) (*User, error) {
				return nil, errors.New("db error")
			},
		}

		svc := NewService(userRepo, &mockGuestRepo{})
		// Should not panic
		svc.SeedCouple(context.Background(),
			CoupleData{URACF: "GRM01"},
			CoupleData{URACF: "BRD01"},
		)
	})
}
