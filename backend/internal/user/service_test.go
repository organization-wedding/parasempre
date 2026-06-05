package user

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/guest"
)

type mockUserRepo struct {
	getByURACF      func(ctx context.Context, uracf string) (*User, error)
	getByGuestID    func(ctx context.Context, guestID int64) (*User, error)
	getByPhone      func(ctx context.Context, phone string) (*User, error)
	getByID         func(ctx context.Context, id int64) (*User, error)
	getByRole       func(ctx context.Context, role string) (*User, error)
	getMeByURACF    func(ctx context.Context, uracf string) (*MeResponse, error)
	createFn        func(ctx context.Context, u *User) (*User, error)
	updateFn        func(ctx context.Context, id int64, input UpdateInput) (*User, error)
	deleteFn        func(ctx context.Context, id int64) error
	listFn          func(ctx context.Context) ([]UserListItem, error)
	updateLastLogin func(ctx context.Context, userID int64) error
	logAction       func(ctx context.Context, userID int64, action string, details map[string]any) error
}

func (m *mockUserRepo) GetByURACF(ctx context.Context, uracf string) (*User, error) {
	return m.getByURACF(ctx, uracf)
}

func (m *mockUserRepo) GetByGuestID(ctx context.Context, guestID int64) (*User, error) {
	if m.getByGuestID != nil {
		return m.getByGuestID(ctx, guestID)
	}
	return nil, nil
}

func (m *mockUserRepo) GetByPhone(ctx context.Context, phone string) (*User, error) {
	if m.getByPhone != nil {
		return m.getByPhone(ctx, phone)
	}
	return nil, nil
}

func (m *mockUserRepo) Create(ctx context.Context, u *User) (*User, error) {
	return m.createFn(ctx, u)
}

func (m *mockUserRepo) List(ctx context.Context) ([]UserListItem, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return []UserListItem{}, nil
}

func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, userID int64) error {
	if m.updateLastLogin != nil {
		return m.updateLastLogin(ctx, userID)
	}
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*User, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepo) GetByRole(ctx context.Context, role string) (*User, error) {
	if m.getByRole != nil {
		return m.getByRole(ctx, role)
	}
	return nil, nil
}

func (m *mockUserRepo) GetMeByURACF(ctx context.Context, uracf string) (*MeResponse, error) {
	if m.getMeByURACF != nil {
		return m.getMeByURACF(ctx, uracf)
	}
	return nil, nil
}

func (m *mockUserRepo) Update(ctx context.Context, id int64, input UpdateInput) (*User, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockUserRepo) DeleteByGuestID(ctx context.Context, guestID int64) error {
	return nil
}

func (m *mockUserRepo) LogAction(ctx context.Context, userID int64, action string, details map[string]any) error {
	if m.logAction != nil {
		return m.logAction(ctx, userID, action, details)
	}
	return nil
}

func (m *mockUserRepo) WithTx(_ pgx.Tx) Repository {
	return m
}

type mockGuestRepo struct {
	listFn                      func(ctx context.Context, limit, offset int) ([]guest.Guest, int, error)
	listByFamilyGroupFn         func(ctx context.Context, familyGroup int64) ([]guest.Guest, error)
	getByIDAnyFn                func(ctx context.Context, id int64) (*guest.Guest, error)
	getByIDsFn                  func(ctx context.Context, ids []int64) ([]guest.Guest, error)
	getByNameFn                 func(ctx context.Context, firstName, lastName string) (*guest.Guest, error)
	familyGroupExistsFn         func(ctx context.Context, familyGroup int64) (bool, error)
	getNextFamilyGroupFn        func(ctx context.Context) (int64, error)
	createFn                    func(ctx context.Context, input guest.CreateGuestInput, userRACF string) (*guest.Guest, error)
	updateFn                    func(ctx context.Context, id int64, input guest.UpdateGuestInput, userRACF string) (*guest.Guest, error)
	deleteFn                    func(ctx context.Context, id int64) error
	setConfirmedFn              func(ctx context.Context, id int64, confirmed bool, userRACF string) (*guest.Guest, error)
	setConfirmedByFamilyGroupFn func(ctx context.Context, familyGroup int64, confirmed bool, userRACF string) ([]guest.Guest, error)
	setConfirmedByIDsFn         func(ctx context.Context, ids []int64, confirmed bool, userRACF string) ([]guest.Guest, error)
	getFamilyGroupByPhoneFn     func(ctx context.Context, phone string) (*int64, error)
}

func (m *mockGuestRepo) List(ctx context.Context, limit, offset int) ([]guest.Guest, int, error) {
	if m.listFn != nil {
		return m.listFn(ctx, limit, offset)
	}
	return []guest.Guest{}, 0, nil
}

func (m *mockGuestRepo) ListByFamilyGroup(ctx context.Context, familyGroup int64) ([]guest.Guest, error) {
	if m.listByFamilyGroupFn != nil {
		return m.listByFamilyGroupFn(ctx, familyGroup)
	}
	return []guest.Guest{}, nil
}

func (m *mockGuestRepo) GetByIDAny(ctx context.Context, id int64) (*guest.Guest, error) {
	if m.getByIDAnyFn != nil {
		return m.getByIDAnyFn(ctx, id)
	}
	return nil, nil
}

func (m *mockGuestRepo) GetByIDs(ctx context.Context, ids []int64) ([]guest.Guest, error) {
	if m.getByIDsFn != nil {
		return m.getByIDsFn(ctx, ids)
	}
	return []guest.Guest{}, nil
}

func (m *mockGuestRepo) GetByName(ctx context.Context, firstName, lastName string) (*guest.Guest, error) {
	if m.getByNameFn != nil {
		return m.getByNameFn(ctx, firstName, lastName)
	}
	return nil, nil
}

func (m *mockGuestRepo) Create(ctx context.Context, input guest.CreateGuestInput, userRACF string) (*guest.Guest, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input, userRACF)
	}
	return nil, nil
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
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input, userRACF)
	}
	return nil, nil
}

func (m *mockGuestRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockGuestRepo) SetConfirmed(ctx context.Context, id int64, confirmed bool, userRACF string) (*guest.Guest, error) {
	return nil, nil
}

func (m *mockGuestRepo) SetConfirmedByFamilyGroup(ctx context.Context, familyGroup int64, confirmed bool, userRACF string) ([]guest.Guest, error) {
	if m.setConfirmedByFamilyGroupFn != nil {
		return m.setConfirmedByFamilyGroupFn(ctx, familyGroup, confirmed, userRACF)
	}
	return []guest.Guest{}, nil
}

func (m *mockGuestRepo) SetConfirmedByIDs(ctx context.Context, ids []int64, confirmed bool, userRACF string) ([]guest.Guest, error) {
	if m.setConfirmedByIDsFn != nil {
		return m.setConfirmedByIDsFn(ctx, ids, confirmed, userRACF)
	}
	return []guest.Guest{}, nil
}

func (m *mockGuestRepo) GetFamilyGroupByPhone(ctx context.Context, phone string) (*int64, error) {
	if m.getFamilyGroupByPhoneFn != nil {
		return m.getFamilyGroupByPhoneFn(ctx, phone)
	}
	return nil, nil
}

func (m *mockGuestRepo) WithTx(_ pgx.Tx) guest.Repository {
	return m
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

func assertAppError(t *testing.T, err error, wantCode int, wantMsg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", wantMsg)
	}
	ae, ok := apperror.IsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != wantCode {
		t.Fatalf("expected code %d, got %d (message: %s)", wantCode, ae.Code, ae.Message)
	}
	if !strings.Contains(ae.Message, wantMsg) {
		t.Fatalf("expected message containing %q, got %q", wantMsg, ae.Message)
	}
}

func TestServiceRegister(t *testing.T) {
	tests := []struct {
		name        string
		input       RegisterInput
		phoneExists bool
		uracfTaken  bool
		wantErr     bool
		wantErrMsg  string
		wantErrCode int
	}{
		{
			name:  "valid registration",
			input: RegisterInput{Phone: "11999999999", URACF: "USR01"},
		},
		{
			name:        "missing phone",
			input:       RegisterInput{Phone: "", URACF: "USR01"},
			wantErr:     true,
			wantErrMsg:  "phone is required",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name:        "invalid phone format",
			input:       RegisterInput{Phone: "1188888888", URACF: "USR01"},
			wantErr:     true,
			wantErrMsg:  "phone must be a valid BR mobile number",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name:        "missing uracf",
			input:       RegisterInput{Phone: "11999999999", URACF: ""},
			wantErr:     true,
			wantErrMsg:  "uracf is required",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name:        "invalid uracf format",
			input:       RegisterInput{Phone: "11999999999", URACF: "toolong1"},
			wantErr:     true,
			wantErrMsg:  "uracf must be exactly 5 uppercase alphanumeric characters",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name:        "phone already registered",
			input:       RegisterInput{Phone: "11999999999", URACF: "USR01"},
			phoneExists: true,
			wantErr:     true,
			wantErrMsg:  "user already registered with this phone",
			wantErrCode: http.StatusConflict,
		},
		{
			name:        "uracf already taken",
			input:       RegisterInput{Phone: "11999999999", URACF: "USR01"},
			uracfTaken:  true,
			wantErr:     true,
			wantErrMsg:  "uracf already in use",
			wantErrCode: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{
				getByPhone: func(ctx context.Context, phone string) (*User, error) {
					if tt.phoneExists {
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
						Role:      u.Role,
						URACF:     u.URACF,
						Phone:     u.Phone,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil
				},
			}

			svc := NewService(userRepo, &mockGuestRepo{})
			u, err := svc.Register(context.Background(), tt.input)
			if tt.wantErr {
				assertAppError(t, err, tt.wantErrCode, tt.wantErrMsg)
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
		name        string
		phone       string
		userFound   bool
		wantExists  bool
		wantErr     bool
		wantErrMsg  string
		wantErrCode int
	}{
		{
			name:       "user exists",
			phone:      "11999999999",
			userFound:  true,
			wantExists: true,
		},
		{
			name:       "user not found",
			phone:      "11999999999",
			userFound:  false,
			wantExists: false,
		},
		{
			name:        "missing phone",
			phone:       "",
			wantErr:     true,
			wantErrMsg:  "phone is required",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name:        "invalid phone format",
			phone:       "abc",
			wantErr:     true,
			wantErrMsg:  "phone must be a valid BR mobile number",
			wantErrCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{
				getByPhone: func(ctx context.Context, phone string) (*User, error) {
					if tt.userFound {
						return sampleUser(), nil
					}
					return nil, nil
				},
			}

			svc := NewService(userRepo, &mockGuestRepo{})
			resp, err := svc.CheckByPhone(context.Background(), tt.phone)
			if tt.wantErr {
				assertAppError(t, err, tt.wantErrCode, tt.wantErrMsg)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Exists != tt.wantExists {
				t.Fatalf("expected exists=%v, got %v", tt.wantExists, resp.Exists)
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

	t.Run("creates with phone", func(t *testing.T) {
		var createdPhones []*string
		userRepo := &mockUserRepo{
			getByURACF: func(ctx context.Context, uracf string) (*User, error) {
				return nil, nil
			},
			createFn: func(ctx context.Context, u *User) (*User, error) {
				createdPhones = append(createdPhones, u.Phone)
				return &User{ID: 1, Role: u.Role, URACF: u.URACF, Phone: u.Phone, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
			},
		}

		svc := NewService(userRepo, &mockGuestRepo{})
		svc.SeedCouple(context.Background(),
			CoupleData{URACF: "GRM01", Phone: "11999999999"},
			CoupleData{URACF: "BRD01", Phone: "11988888888"},
		)

		if len(createdPhones) != 2 {
			t.Fatalf("expected 2 users created, got %d", len(createdPhones))
		}
		if createdPhones[0] == nil || *createdPhones[0] != "11999999999" {
			t.Fatalf("expected groom phone 11999999999, got %v", createdPhones[0])
		}
		if createdPhones[1] == nil || *createdPhones[1] != "11988888888" {
			t.Fatalf("expected bride phone 11988888888, got %v", createdPhones[1])
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
		svc.SeedCouple(context.Background(),
			CoupleData{URACF: "GRM01"},
			CoupleData{URACF: "BRD01"},
		)
	})
}

func TestServicePhoneExists(t *testing.T) {
	t.Run("found in users table", func(t *testing.T) {
		phone := "11999999999"
		userRepo := &mockUserRepo{
			getByPhone: func(ctx context.Context, p string) (*User, error) {
				return &User{ID: 1, Role: "groom", URACF: "GRM01", Phone: &phone}, nil
			},
		}

		svc := NewService(userRepo, &mockGuestRepo{})
		exists, err := svc.PhoneExists(context.Background(), "11999999999")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Fatal("expected exists=true")
		}
	})

	t.Run("not found", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getByPhone: func(ctx context.Context, phone string) (*User, error) {
				return nil, nil
			},
		}

		svc := NewService(userRepo, &mockGuestRepo{})
		exists, err := svc.PhoneExists(context.Background(), "11999999999")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exists {
			t.Fatal("expected exists=false")
		}
	})
}

func TestServiceRecordLogin(t *testing.T) {
	t.Run("calls UpdateLastLogin and LogAction", func(t *testing.T) {
		var lastLoginCalled, logActionCalled bool
		userRepo := &mockUserRepo{
			updateLastLogin: func(ctx context.Context, userID int64) error {
				lastLoginCalled = true
				if userID != 42 {
					t.Fatalf("expected userID 42, got %d", userID)
				}
				return nil
			},
			logAction: func(ctx context.Context, userID int64, action string, details map[string]any) error {
				logActionCalled = true
				if action != "login" {
					t.Fatalf("expected action login, got %q", action)
				}
				return nil
			},
		}

		svc := NewService(userRepo, &mockGuestRepo{})
		svc.RecordLogin(context.Background(), 42)

		if !lastLoginCalled {
			t.Fatal("expected UpdateLastLogin to be called")
		}
		if !logActionCalled {
			t.Fatal("expected LogAction to be called")
		}
	})

	t.Run("handles errors gracefully", func(t *testing.T) {
		userRepo := &mockUserRepo{
			updateLastLogin: func(ctx context.Context, userID int64) error {
				return errors.New("db error")
			},
			logAction: func(ctx context.Context, userID int64, action string, details map[string]any) error {
				return errors.New("db error")
			},
		}

		svc := NewService(userRepo, &mockGuestRepo{})
		svc.RecordLogin(context.Background(), 42)
	})
}

func TestServiceCreateGuestUserTx(t *testing.T) {
	t.Run("creates user with guest_id and phone", func(t *testing.T) {
		var createdUser *User
		userRepo := &mockUserRepo{
			createFn: func(ctx context.Context, u *User) (*User, error) {
				createdUser = u
				return &User{
					ID:        10,
					GuestID:   u.GuestID,
					Role:      u.Role,
					URACF:     u.URACF,
					Phone:     u.Phone,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil
			},
		}

		svc := NewServiceWithTx(userRepo, &mockGuestRepo{})
		phone := "11999999999"
		err := svc.CreateGuestUserTx(context.Background(), nil, 42, &phone)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if createdUser == nil {
			t.Fatal("expected user to be created")
		}
		if createdUser.GuestID == nil || *createdUser.GuestID != 42 {
			t.Fatalf("expected guest_id 42, got %v", createdUser.GuestID)
		}
		if createdUser.Role != "guest" {
			t.Fatalf("expected role guest, got %q", createdUser.Role)
		}
		if createdUser.Phone == nil || *createdUser.Phone != phone {
			t.Fatalf("expected phone %q, got %v", phone, createdUser.Phone)
		}
		if len(createdUser.URACF) != 5 {
			t.Fatalf("expected URACF of length 5, got %q (%d)", createdUser.URACF, len(createdUser.URACF))
		}
		for _, c := range createdUser.URACF {
			if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
				t.Fatalf("URACF contains invalid char %q in %q", string(c), createdUser.URACF)
			}
		}
	})

	t.Run("creates user without phone", func(t *testing.T) {
		var createdUser *User
		userRepo := &mockUserRepo{
			createFn: func(ctx context.Context, u *User) (*User, error) {
				createdUser = u
				return &User{
					ID:        10,
					GuestID:   u.GuestID,
					Role:      u.Role,
					URACF:     u.URACF,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil
			},
		}

		svc := NewServiceWithTx(userRepo, &mockGuestRepo{})
		err := svc.CreateGuestUserTx(context.Background(), nil, 42, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if createdUser == nil {
			t.Fatal("expected user to be created")
		}
		if createdUser.Phone != nil {
			t.Fatalf("expected nil phone, got %v", createdUser.Phone)
		}
	})

	t.Run("returns error on repo failure", func(t *testing.T) {
		userRepo := &mockUserRepo{
			createFn: func(ctx context.Context, u *User) (*User, error) {
				return nil, errors.New("db error")
			},
		}

		svc := NewServiceWithTx(userRepo, &mockGuestRepo{})
		err := svc.CreateGuestUserTx(context.Background(), nil, 42, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGenerateURACF(t *testing.T) {
	uracf, err := GenerateURACF()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(uracf) != 5 {
		t.Fatalf("expected length 5, got %d: %q", len(uracf), uracf)
	}
	for _, c := range uracf {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			t.Fatalf("invalid char %q in URACF %q", string(c), uracf)
		}
	}

	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		u, err := GenerateURACF()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		seen[u] = true
	}
	if len(seen) < 50 {
		t.Fatalf("expected at least 50 unique URACFs in 100 generations, got %d", len(seen))
	}
}
