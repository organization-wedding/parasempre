package guest

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

// mockRepository implements TxAwareRepository with function fields for easy stubbing.
type mockRepository struct {
	listFn               func(ctx context.Context) ([]Guest, error)
	getByID              func(ctx context.Context, id int64) (*Guest, error)
	getByNameFn          func(ctx context.Context, firstName, lastName string) (*Guest, error)
	familyGroupExistsFn  func(ctx context.Context, familyGroup int64) (bool, error)
	getNextFamilyGroupFn func(ctx context.Context) (int64, error)
	createFn             func(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error)
	updateFn             func(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error)
	deleteFn             func(ctx context.Context, id int64) error
	setConfirmedFn       func(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error)
}

func (m *mockRepository) List(ctx context.Context) ([]Guest, error) {
	return m.listFn(ctx)
}

func (m *mockRepository) GetByID(ctx context.Context, id int64) (*Guest, error) {
	return m.getByID(ctx, id)
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

func (m *mockRepository) SetConfirmed(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error) {
	return m.setConfirmedFn(ctx, id, confirmed, userRACF)
}

// WithTx returns itself (mock ignores transactions).
func (m *mockRepository) WithTx(_ pgx.Tx) Repository {
	return m
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

// mockUserCreator implements UserCreator for tests.
type mockUserCreator struct {
	createFn func(ctx context.Context, tx pgx.Tx, guestID int64, phone *string) error
}

func (m *mockUserCreator) CreateGuestUserTx(ctx context.Context, tx pgx.Tx, guestID int64, phone *string) error {
	if m.createFn != nil {
		return m.createFn(ctx, tx, guestID, phone)
	}
	return nil
}

// mockUserPhoneLookup implements UserPhoneLookup for tests.
type mockUserPhoneLookup struct {
	getGuestIDByPhoneFn func(ctx context.Context, phone string) (*int64, error)
}

func (m *mockUserPhoneLookup) GetGuestIDByPhone(ctx context.Context, phone string) (*int64, error) {
	if m.getGuestIDByPhoneFn != nil {
		return m.getGuestIDByPhoneFn(ctx, phone)
	}
	return nil, nil
}

func noopPhoneLookup() *mockUserPhoneLookup {
	return &mockUserPhoneLookup{}
}

// mockTxRunner implements database.TxRunner by running fn with a nil tx (pass-through).
type mockTxRunner struct{}

func (m *mockTxRunner) RunInTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return fn(nil)
}

func strPtr(s string) *string { return &s }

func int64Ptr(v int64) *int64 { return &v }

func sampleGuest() Guest {
	return Guest{
		ID:           1,
		FirstName:    "João",
		LastName:     "Silva",
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

func noopUserCreator() *mockUserCreator {
	return &mockUserCreator{}
}

// newTestService creates a Service wired with mocks. The mockTxRunner calls fn(nil)
// so WithTx returns the mock itself, and all repo calls go through the mockRepository.
func newTestService(repo *mockRepository, checker *mockUserChecker, creator *mockUserCreator) *Service {
	return newTestServiceWithPhone(repo, checker, creator, noopPhoneLookup())
}

func newTestServiceWithPhone(repo *mockRepository, checker *mockUserChecker, creator *mockUserCreator, phoneLookup *mockUserPhoneLookup) *Service {
	return &Service{
		repo:            repo,
		userChecker:     checker,
		userCreator:     creator,
		userPhoneLookup: phoneLookup,
		txRunner:        &mockTxRunner{},
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
		t.Fatalf("expected code %d, got %d", wantCode, ae.Code)
	}
	if !strings.Contains(ae.Message, wantMsg) {
		t.Fatalf("expected message containing %q, got %q", wantMsg, ae.Message)
	}
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
			svc := newTestService(&mockRepository{listFn: tt.mockFn}, alwaysExistsChecker(), noopUserCreator())
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
				return nil, apperror.NotFound("guest not found")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(&mockRepository{getByID: tt.mockFn}, alwaysExistsChecker(), noopUserCreator())
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
		name        string
		input       CreateGuestInput
		wantErr     bool
		wantErrMsg  string
		wantErrCode int
	}{
		{
			name: "valid input",
			input: CreateGuestInput{
				FirstName:    "Maria",
				LastName:     "Santos",
				Relationship: "R",
				FamilyGroup:  int64Ptr(1),
			},
		},
		{
			name: "missing first_name",
			input: CreateGuestInput{
				LastName:     "Santos",
				Relationship: "R",
				FamilyGroup:  int64Ptr(1),
			},
			wantErr:     true,
			wantErrMsg:  "first_name is required",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name: "missing last_name",
			input: CreateGuestInput{
				FirstName:    "Maria",
				Relationship: "R",
				FamilyGroup:  int64Ptr(1),
			},
			wantErr:     true,
			wantErrMsg:  "last_name is required",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name: "invalid relationship",
			input: CreateGuestInput{
				FirstName:    "Maria",
				LastName:     "Santos",
				Relationship: "X",
				FamilyGroup:  int64Ptr(1),
			},
			wantErr:     true,
			wantErrMsg:  "relationship must be 'P' or 'R'",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name: "missing relationship",
			input: CreateGuestInput{
				FirstName:   "Maria",
				LastName:    "Santos",
				FamilyGroup: int64Ptr(1),
			},
			wantErr:     true,
			wantErrMsg:  "relationship must be 'P' or 'R'",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name: "missing family_group auto-assigns",
			input: CreateGuestInput{
				FirstName:    "Maria",
				LastName:     "Santos",
				Relationship: "R",
			},
		},
		{
			name: "family_group zero",
			input: CreateGuestInput{
				FirstName:    "Maria",
				LastName:     "Santos",
				Relationship: "R",
				FamilyGroup:  int64Ptr(0),
			},
			wantErr:     true,
			wantErrMsg:  "family_group must be greater than 0",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name: "family_group negative",
			input: CreateGuestInput{
				FirstName:    "Maria",
				LastName:     "Santos",
				Relationship: "R",
				FamilyGroup:  int64Ptr(-1),
			},
			wantErr:     true,
			wantErrMsg:  "family_group must be greater than 0",
			wantErrCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				createFn: func(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
					g := sampleGuest()
					return &g, nil
				},
			}
			svc := newTestService(repo, alwaysExistsChecker(), noopUserCreator())
			_, err := svc.Create(context.Background(), tt.input, "TST01")
			if tt.wantErr {
				assertAppError(t, err, tt.wantErrCode, tt.wantErrMsg)
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
	}
	svc := newTestService(repo, alwaysExistsChecker(), noopUserCreator())

	_, err := svc.Create(context.Background(), CreateGuestInput{
		FirstName:    "João",
		LastName:     "Silva",
		Relationship: "P",
		FamilyGroup:  int64Ptr(1),
	}, "TST01")

	assertAppError(t, err, http.StatusConflict, "a guest named 'João Silva' already exists")
}

func TestServiceCreateFamilyGroupNotFound(t *testing.T) {
	repo := &mockRepository{
		getByNameFn: func(ctx context.Context, firstName, lastName string) (*Guest, error) {
			return nil, nil
		},
		familyGroupExistsFn: func(ctx context.Context, familyGroup int64) (bool, error) {
			return false, nil
		},
	}
	svc := newTestService(repo, alwaysExistsChecker(), noopUserCreator())

	_, err := svc.Create(context.Background(), CreateGuestInput{
		FirstName:    "Maria",
		LastName:     "Santos",
		Relationship: "R",
		FamilyGroup:  int64Ptr(10),
	}, "TST01")

	assertAppError(t, err, http.StatusBadRequest, "family_group not found")
}

func TestServiceCreateAutoAssignFamilyGroup(t *testing.T) {
	repo := &mockRepository{
		getByNameFn: func(ctx context.Context, firstName, lastName string) (*Guest, error) {
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
	svc := newTestService(repo, alwaysExistsChecker(), noopUserCreator())

	_, err := svc.Create(context.Background(), CreateGuestInput{
		FirstName:    "Maria",
		LastName:     "Santos",
		Relationship: "R",
	}, "TST01")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceCreateUserRACFNotFound(t *testing.T) {
	repo := &mockRepository{}
	checker := &mockUserChecker{existsFn: func(ctx context.Context, uracf string) (bool, error) {
		return false, nil
	}}
	svc := newTestService(repo, checker, noopUserCreator())

	_, err := svc.Create(context.Background(), CreateGuestInput{
		FirstName:    "Maria",
		LastName:     "Santos",
		Relationship: "R",
		FamilyGroup:  int64Ptr(1),
	}, "TST01")

	assertAppError(t, err, http.StatusBadRequest, "user-racf does not match any registered user")
}

func TestServiceCreateWithUser(t *testing.T) {
	var userCreated bool
	var capturedGuestID int64
	var capturedPhone *string

	repo := &mockRepository{
		createFn: func(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
			g := sampleGuest()
			g.ID = 42
			return &g, nil
		},
	}
	creator := &mockUserCreator{
		createFn: func(ctx context.Context, tx pgx.Tx, guestID int64, phone *string) error {
			userCreated = true
			capturedGuestID = guestID
			capturedPhone = phone
			return nil
		},
	}

	phone := "11999999999"
	svc := newTestService(repo, alwaysExistsChecker(), creator)
	g, err := svc.Create(context.Background(), CreateGuestInput{
		FirstName:    "Maria",
		LastName:     "Santos",
		Relationship: "R",
		FamilyGroup:  int64Ptr(1),
		Phone:        &phone,
	}, "TST01")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.ID != 42 {
		t.Fatalf("expected guest ID 42, got %d", g.ID)
	}
	if !userCreated {
		t.Fatal("expected user to be created")
	}
	if capturedGuestID != 42 {
		t.Fatalf("expected guest ID 42 in user creation, got %d", capturedGuestID)
	}
	if capturedPhone == nil || *capturedPhone != phone {
		t.Fatalf("expected phone %q in user creation, got %v", phone, capturedPhone)
	}
}

func TestServiceCreateUserFails(t *testing.T) {
	repo := &mockRepository{
		createFn: func(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
			g := sampleGuest()
			return &g, nil
		},
	}
	creator := &mockUserCreator{
		createFn: func(ctx context.Context, tx pgx.Tx, guestID int64, phone *string) error {
			return apperror.Internal("user creation failed", errors.New("db error"))
		},
	}

	svc := newTestService(repo, alwaysExistsChecker(), creator)
	_, err := svc.Create(context.Background(), CreateGuestInput{
		FirstName:    "Maria",
		LastName:     "Santos",
		Relationship: "R",
		FamilyGroup:  int64Ptr(1),
	}, "TST01")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// The error should propagate — in a real scenario the tx would have been rolled back
	assertAppError(t, err, http.StatusInternalServerError, "user creation failed")
}

func TestServiceUpdate(t *testing.T) {
	relP := "P"
	invalid := "X"

	tests := []struct {
		name        string
		input       UpdateGuestInput
		wantErr     bool
		wantErrMsg  string
		wantErrCode int
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
			name:        "invalid relationship",
			input:       UpdateGuestInput{Relationship: &invalid},
			wantErr:     true,
			wantErrMsg:  "relationship must be 'P' or 'R'",
			wantErrCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				updateFn: func(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error) {
					g := sampleGuest()
					return &g, nil
				},
				getByID: func(ctx context.Context, id int64) (*Guest, error) {
					g := sampleGuest()
					return &g, nil
				},
			}
			svc := newTestService(repo, alwaysExistsChecker(), noopUserCreator())
			_, err := svc.Update(context.Background(), 1, tt.input, "TST01")
			if tt.wantErr {
				assertAppError(t, err, tt.wantErrCode, tt.wantErrMsg)
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
	svc := newTestService(repo, checker, noopUserCreator())

	relP := "P"
	_, err := svc.Update(context.Background(), 1, UpdateGuestInput{Relationship: &relP}, "TST01")

	assertAppError(t, err, http.StatusBadRequest, "user-racf does not match any registered user")
}

func TestServiceConfirm(t *testing.T) {
	tests := []struct {
		name           string
		getByIDFn      func(ctx context.Context, id int64) (*Guest, error)
		setConfirmedFn func(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error)
		userExists     bool
		wantErr        bool
	}{
		{
			name: "success",
			getByIDFn: func(ctx context.Context, id int64) (*Guest, error) {
				g := sampleGuest() // Confirmed: false
				return &g, nil
			},
			setConfirmedFn: func(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error) {
				g := sampleGuest()
				g.Confirmed = true
				return &g, nil
			},
			userExists: true,
		},
		{
			name: "guest not found",
			getByIDFn: func(ctx context.Context, id int64) (*Guest, error) {
				return nil, apperror.NotFound("guest not found")
			},
			userExists: true,
			wantErr:    true,
		},
		{
			name:       "user racf not found",
			userExists: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{getByID: tt.getByIDFn, setConfirmedFn: tt.setConfirmedFn}
			checker := &mockUserChecker{existsFn: func(ctx context.Context, uracf string) (bool, error) {
				return tt.userExists, nil
			}}
			svc := newTestService(repo, checker, noopUserCreator())
			guest, err := svc.Confirm(context.Background(), 1, "TST01")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !guest.Confirmed {
				t.Fatal("expected confirmed to be true")
			}
		})
	}
}

func TestServiceConfirmAlreadyConfirmed(t *testing.T) {
	setConfirmedCalled := false
	repo := &mockRepository{
		getByID: func(ctx context.Context, id int64) (*Guest, error) {
			g := sampleGuest()
			g.Confirmed = true // já confirmado
			return &g, nil
		},
		setConfirmedFn: func(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error) {
			setConfirmedCalled = true
			return nil, nil
		},
	}
	svc := newTestService(repo, alwaysExistsChecker(), noopUserCreator())
	guest, err := svc.Confirm(context.Background(), 1, "TST01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !guest.Confirmed {
		t.Fatal("expected confirmed to be true")
	}
	if setConfirmedCalled {
		t.Fatal("expected SetConfirmed to NOT be called (idempotent skip)")
	}
}

func TestServiceCancel(t *testing.T) {
	repo := &mockRepository{
		getByID: func(ctx context.Context, id int64) (*Guest, error) {
			g := sampleGuest()
			g.Confirmed = true // está confirmado, cancel deve executar
			return &g, nil
		},
		setConfirmedFn: func(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error) {
			g := sampleGuest()
			g.Confirmed = false
			return &g, nil
		},
	}
	svc := newTestService(repo, alwaysExistsChecker(), noopUserCreator())
	guest, err := svc.Cancel(context.Background(), 1, "TST01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if guest.Confirmed {
		t.Fatal("expected confirmed to be false")
	}
}

func TestServiceCancelAlreadyCancelled(t *testing.T) {
	setConfirmedCalled := false
	repo := &mockRepository{
		getByID: func(ctx context.Context, id int64) (*Guest, error) {
			g := sampleGuest() // Confirmed: false — já cancelado
			return &g, nil
		},
		setConfirmedFn: func(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error) {
			setConfirmedCalled = true
			return nil, nil
		},
	}
	svc := newTestService(repo, alwaysExistsChecker(), noopUserCreator())
	guest, err := svc.Cancel(context.Background(), 1, "TST01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if guest.Confirmed {
		t.Fatal("expected confirmed to be false")
	}
	if setConfirmedCalled {
		t.Fatal("expected SetConfirmed to NOT be called (idempotent skip)")
	}
}

func TestServiceConfirmByPhone(t *testing.T) {
	tests := []struct {
		name        string
		phoneLookup func(ctx context.Context, phone string) (*int64, error)
		wantErr     bool
		wantErrCode int
		wantErrMsg  string
	}{
		{
			name: "success",
			phoneLookup: func(ctx context.Context, phone string) (*int64, error) {
				id := int64(1)
				return &id, nil
			},
		},
		{
			name: "phone not found",
			phoneLookup: func(ctx context.Context, phone string) (*int64, error) {
				return nil, nil
			},
			wantErr:     true,
			wantErrCode: http.StatusNotFound,
			wantErrMsg:  "no guest found for this phone number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				getByID: func(ctx context.Context, id int64) (*Guest, error) {
					g := sampleGuest() // Confirmed: false → confirm vai executar
					return &g, nil
				},
				setConfirmedFn: func(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error) {
					g := sampleGuest()
					g.Confirmed = true
					return &g, nil
				},
			}
			phoneLookup := &mockUserPhoneLookup{getGuestIDByPhoneFn: tt.phoneLookup}
			svc := newTestServiceWithPhone(repo, alwaysExistsChecker(), noopUserCreator(), phoneLookup)

			guest, err := svc.ConfirmByPhone(context.Background(), "43999999999", "TST01")
			if tt.wantErr {
				assertAppError(t, err, tt.wantErrCode, tt.wantErrMsg)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !guest.Confirmed {
				t.Fatal("expected confirmed to be true")
			}
		})
	}
}

func TestServiceCancelByPhone(t *testing.T) {
	repo := &mockRepository{
		getByID: func(ctx context.Context, id int64) (*Guest, error) {
			g := sampleGuest()
			g.Confirmed = true // está confirmado, cancel deve executar
			return &g, nil
		},
		setConfirmedFn: func(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error) {
			g := sampleGuest()
			g.Confirmed = false
			return &g, nil
		},
	}
	phoneLookup := &mockUserPhoneLookup{
		getGuestIDByPhoneFn: func(ctx context.Context, phone string) (*int64, error) {
			id := int64(1)
			return &id, nil
		},
	}
	svc := newTestServiceWithPhone(repo, alwaysExistsChecker(), noopUserCreator(), phoneLookup)

	guest, err := svc.CancelByPhone(context.Background(), "43999999999", "TST01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if guest.Confirmed {
		t.Fatal("expected confirmed to be false")
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
				return apperror.NotFound("guest not found")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(&mockRepository{deleteFn: tt.mockFn}, alwaysExistsChecker(), noopUserCreator())
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
