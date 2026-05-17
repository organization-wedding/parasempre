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

type mockRepository struct {
	listFn                       func(ctx context.Context, limit, offset int, userRACF string) ([]Guest, int, error)
	listByFamilyGroupFn          func(ctx context.Context, familyGroup int64) ([]Guest, error)
	getByID                      func(ctx context.Context, id int64, userRACF string) (*Guest, error)
	getByIDAnyFn                 func(ctx context.Context, id int64) (*Guest, error)
	getByIDsFn                   func(ctx context.Context, ids []int64) ([]Guest, error)
	getByNameFn                  func(ctx context.Context, firstName, lastName string) (*Guest, error)
	familyGroupExistsFn          func(ctx context.Context, familyGroup int64) (bool, error)
	getNextFamilyGroupFn         func(ctx context.Context) (int64, error)
	createFn                     func(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error)
	updateFn                     func(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error)
	deleteFn                     func(ctx context.Context, id int64) error
	setConfirmedFn               func(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error)
	setConfirmedByFamilyGroupFn  func(ctx context.Context, familyGroup int64, confirmed bool, userRACF string) ([]Guest, error)
	setConfirmedByIDsFn          func(ctx context.Context, ids []int64, confirmed bool, userRACF string) ([]Guest, error)
	getFamilyGroupByPhoneFn      func(ctx context.Context, phone string) (*int64, error)
}

func (m *mockRepository) List(ctx context.Context, limit, offset int, userRACF string) ([]Guest, int, error) {
	return m.listFn(ctx, limit, offset, userRACF)
}

func (m *mockRepository) GetByID(ctx context.Context, id int64, userRACF string) (*Guest, error) {
	return m.getByID(ctx, id, userRACF)
}

func (m *mockRepository) ListByFamilyGroup(ctx context.Context, familyGroup int64) ([]Guest, error) {
	if m.listByFamilyGroupFn != nil {
		return m.listByFamilyGroupFn(ctx, familyGroup)
	}
	return []Guest{}, nil
}

func (m *mockRepository) GetByIDAny(ctx context.Context, id int64) (*Guest, error) {
	if m.getByIDAnyFn != nil {
		return m.getByIDAnyFn(ctx, id)
	}
	if m.getByID != nil {
		return m.getByID(ctx, id, "")
	}
	return nil, nil
}

func (m *mockRepository) GetByIDs(ctx context.Context, ids []int64) ([]Guest, error) {
	if m.getByIDsFn != nil {
		return m.getByIDsFn(ctx, ids)
	}
	return []Guest{}, nil
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

func (m *mockRepository) SetConfirmedByFamilyGroup(ctx context.Context, familyGroup int64, confirmed bool, userRACF string) ([]Guest, error) {
	if m.setConfirmedByFamilyGroupFn != nil {
		return m.setConfirmedByFamilyGroupFn(ctx, familyGroup, confirmed, userRACF)
	}
	return []Guest{}, nil
}

func (m *mockRepository) SetConfirmedByIDs(ctx context.Context, ids []int64, confirmed bool, userRACF string) ([]Guest, error) {
	if m.setConfirmedByIDsFn != nil {
		return m.setConfirmedByIDsFn(ctx, ids, confirmed, userRACF)
	}
	return []Guest{}, nil
}

func (m *mockRepository) GetFamilyGroupByPhone(ctx context.Context, phone string) (*int64, error) {
	if m.getFamilyGroupByPhoneFn != nil {
		return m.getFamilyGroupByPhoneFn(ctx, phone)
	}
	return nil, nil
}

func (m *mockRepository) WithTx(_ pgx.Tx) Repository {
	return m
}

type mockUserBridge struct {
	existsFn             func(ctx context.Context, uracf string) (bool, error)
	createFn             func(ctx context.Context, tx pgx.Tx, guestID int64, phone *string) error
	deleteFn             func(ctx context.Context, tx pgx.Tx, guestID int64) error
	getGuestIDByPhoneFn  func(ctx context.Context, phone string) (*int64, error)
	getGuestIDByUserIDFn func(ctx context.Context, userID int64) (*int64, error)
	getURACFByUserIDFn   func(ctx context.Context, userID int64) (string, error)
}

func (m *mockUserBridge) UserExistsByURACF(ctx context.Context, uracf string) (bool, error) {
	if m.existsFn != nil {
		return m.existsFn(ctx, uracf)
	}
	return true, nil
}

func (m *mockUserBridge) CreateGuestUserTx(ctx context.Context, tx pgx.Tx, guestID int64, phone *string) error {
	if m.createFn != nil {
		return m.createFn(ctx, tx, guestID, phone)
	}
	return nil
}

func (m *mockUserBridge) DeleteGuestUserTx(ctx context.Context, tx pgx.Tx, guestID int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, tx, guestID)
	}
	return nil
}

func (m *mockUserBridge) GetGuestIDByPhone(ctx context.Context, phone string) (*int64, error) {
	if m.getGuestIDByPhoneFn != nil {
		return m.getGuestIDByPhoneFn(ctx, phone)
	}
	return nil, nil
}

func (m *mockUserBridge) GetGuestIDByUserID(ctx context.Context, userID int64) (*int64, error) {
	if m.getGuestIDByUserIDFn != nil {
		return m.getGuestIDByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockUserBridge) GetURACFByUserID(ctx context.Context, userID int64) (string, error) {
	if m.getURACFByUserIDFn != nil {
		return m.getURACFByUserIDFn(ctx, userID)
	}
	return "TST01", nil
}

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

func defaultUserBridge() *mockUserBridge {
	return &mockUserBridge{
		getGuestIDByUserIDFn: func(ctx context.Context, userID int64) (*int64, error) {
			guestID := int64(1)
			return &guestID, nil
		},
	}
}

func newTestService(repo *mockRepository, users *mockUserBridge) *Service {
	return &Service{
		repo:     repo,
		users:    users,
		txRunner: &mockTxRunner{},
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
		name      string
		mockFn    func(ctx context.Context, limit, offset int, userRACF string) ([]Guest, int, error)
		page      int
		limit     int
		wantLen   int
		wantTotal int
		wantErr   bool
	}{
		{
			name: "returns guests with pagination",
			mockFn: func(ctx context.Context, limit, offset int, userRACF string) ([]Guest, int, error) {
				return []Guest{sampleGuest()}, 5, nil
			},
			page:      1,
			limit:     20,
			wantLen:   1,
			wantTotal: 5,
		},
		{
			name: "returns empty list",
			mockFn: func(ctx context.Context, limit, offset int, userRACF string) ([]Guest, int, error) {
				return []Guest{}, 0, nil
			},
			page:    1,
			limit:   20,
			wantLen: 0,
		},
		{
			name: "defaults for invalid page/limit",
			mockFn: func(ctx context.Context, limit, offset int, userRACF string) ([]Guest, int, error) {
				if limit != 20 || offset != 0 {
					return nil, 0, errors.New("expected default limit=20, offset=0")
				}
				return []Guest{}, 0, nil
			},
			page:  0,
			limit: 0,
		},
		{
			name: "caps limit at 100",
			mockFn: func(ctx context.Context, limit, offset int, userRACF string) ([]Guest, int, error) {
				if limit != 100 {
					return nil, 0, errors.New("expected limit capped at 100")
				}
				return []Guest{}, 0, nil
			},
			page:  1,
			limit: 500,
		},
		{
			name: "propagates error",
			mockFn: func(ctx context.Context, limit, offset int, userRACF string) ([]Guest, int, error) {
				return nil, 0, errors.New("db error")
			},
			page:    1,
			limit:   20,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(&mockRepository{listFn: tt.mockFn}, defaultUserBridge())
			result, err := svc.List(context.Background(), tt.page, tt.limit, "TST01")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Data) != tt.wantLen {
				t.Fatalf("expected %d guests, got %d", tt.wantLen, len(result.Data))
			}
			if result.Total != tt.wantTotal {
				t.Fatalf("expected total %d, got %d", tt.wantTotal, result.Total)
			}
		})
	}
}

func TestServiceGetByID(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		mockFn  func(ctx context.Context, id int64, userRACF string) (*Guest, error)
		wantErr bool
	}{
		{
			name: "returns guest",
			id:   1,
			mockFn: func(ctx context.Context, id int64, userRACF string) (*Guest, error) {
				g := sampleGuest()
				return &g, nil
			},
		},
		{
			name: "not found",
			id:   999,
			mockFn: func(ctx context.Context, id int64, userRACF string) (*Guest, error) {
				return nil, apperror.NotFound("guest not found")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(&mockRepository{getByID: tt.mockFn}, defaultUserBridge())
			guest, err := svc.GetByID(context.Background(), tt.id, "TST01")
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
			svc := newTestService(repo, defaultUserBridge())
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
		createFn: func(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
			return nil, apperror.Conflict("a guest named 'João Silva' already exists")
		},
	}
	svc := newTestService(repo, defaultUserBridge())

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
	svc := newTestService(repo, defaultUserBridge())

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
	svc := newTestService(repo, defaultUserBridge())

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
	users := &mockUserBridge{existsFn: func(ctx context.Context, uracf string) (bool, error) {
		return false, nil
	}}
	svc := newTestService(repo, users)

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
	users := &mockUserBridge{
		createFn: func(ctx context.Context, tx pgx.Tx, guestID int64, phone *string) error {
			userCreated = true
			capturedGuestID = guestID
			capturedPhone = phone
			return nil
		},
	}

	phone := "11999999999"
	svc := newTestService(repo, users)
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
	users := &mockUserBridge{
		createFn: func(ctx context.Context, tx pgx.Tx, guestID int64, phone *string) error {
			return apperror.Internal("user creation failed", errors.New("db error"))
		},
	}

	svc := newTestService(repo, users)
	_, err := svc.Create(context.Background(), CreateGuestInput{
		FirstName:    "Maria",
		LastName:     "Santos",
		Relationship: "R",
		FamilyGroup:  int64Ptr(1),
	}, "TST01")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
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
				getByID: func(ctx context.Context, id int64, userRACF string) (*Guest, error) {
					g := sampleGuest()
					return &g, nil
				},
			}
			svc := newTestService(repo, defaultUserBridge())
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
	users := &mockUserBridge{existsFn: func(ctx context.Context, uracf string) (bool, error) {
		return false, nil
	}}
	svc := newTestService(repo, users)

	relP := "P"
	_, err := svc.Update(context.Background(), 1, UpdateGuestInput{Relationship: &relP}, "TST01")

	assertAppError(t, err, http.StatusBadRequest, "user-racf does not match any registered user")
}

func TestServiceConfirm(t *testing.T) {
	tests := []struct {
		name           string
		getByIDFn      func(ctx context.Context, id int64, userRACF string) (*Guest, error)
		setConfirmedFn func(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error)
		userExists     bool
		wantErr        bool
	}{
		{
			name: "success",
			getByIDFn: func(ctx context.Context, id int64, userRACF string) (*Guest, error) {
				g := sampleGuest()
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
			getByIDFn: func(ctx context.Context, id int64, userRACF string) (*Guest, error) {
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
			users := &mockUserBridge{getGuestIDByUserIDFn: func(ctx context.Context, userID int64) (*int64, error) {
				if tt.userExists {
					guestID := int64(1)
					return &guestID, nil
				}
				return nil, nil
			}}
			svc := newTestService(repo, users)
			guest, err := svc.Confirm(context.Background(), 1, 1)
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
		getByID: func(ctx context.Context, id int64, userRACF string) (*Guest, error) {
			g := sampleGuest()
			g.Confirmed = true
			return &g, nil
		},
		setConfirmedFn: func(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error) {
			setConfirmedCalled = true
			return nil, nil
		},
	}
	users := &mockUserBridge{getGuestIDByUserIDFn: func(ctx context.Context, userID int64) (*int64, error) {
		guestID := int64(1)
		return &guestID, nil
	}}
	svc := newTestService(repo, users)
	guest, err := svc.Confirm(context.Background(), 1, 1)
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
		getByID: func(ctx context.Context, id int64, userRACF string) (*Guest, error) {
			g := sampleGuest()
			g.Confirmed = true
			return &g, nil
		},
		setConfirmedFn: func(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error) {
			g := sampleGuest()
			g.Confirmed = false
			return &g, nil
		},
	}
	users := &mockUserBridge{getGuestIDByUserIDFn: func(ctx context.Context, userID int64) (*int64, error) {
		guestID := int64(1)
		return &guestID, nil
	}}
	svc := newTestService(repo, users)
	guest, err := svc.Cancel(context.Background(), 1, 1)
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
		getByID: func(ctx context.Context, id int64, userRACF string) (*Guest, error) {
			g := sampleGuest()
			return &g, nil
		},
		setConfirmedFn: func(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error) {
			setConfirmedCalled = true
			return nil, nil
		},
	}
	users := &mockUserBridge{getGuestIDByUserIDFn: func(ctx context.Context, userID int64) (*int64, error) {
		guestID := int64(1)
		return &guestID, nil
	}}
	svc := newTestService(repo, users)
	guest, err := svc.Cancel(context.Background(), 1, 1)
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
		name         string
		phoneLookup  func(ctx context.Context, phone string) (*int64, error)
		userIDLookup func(ctx context.Context, userID int64) (*int64, error)
		wantErr      bool
		wantErrCode  int
		wantErrMsg   string
	}{
		{
			name: "success",
			phoneLookup: func(ctx context.Context, phone string) (*int64, error) {
				id := int64(1)
				return &id, nil
			},
			userIDLookup: func(ctx context.Context, userID int64) (*int64, error) {
				id := int64(1)
				return &id, nil
			},
		},
		{
			name: "phone not found",
			phoneLookup: func(ctx context.Context, phone string) (*int64, error) {
				return nil, nil
			},
			userIDLookup: func(ctx context.Context, userID int64) (*int64, error) {
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
				getByID: func(ctx context.Context, id int64, userRACF string) (*Guest, error) {
					g := sampleGuest()
					return &g, nil
				},
				setConfirmedFn: func(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error) {
					g := sampleGuest()
					g.Confirmed = true
					return &g, nil
				},
			}
			users := &mockUserBridge{
				getGuestIDByPhoneFn:  tt.phoneLookup,
				getGuestIDByUserIDFn: tt.userIDLookup,
			}
			svc := newTestService(repo, users)

			guest, err := svc.ConfirmByPhone(context.Background(), "43999999999", 1)
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
		getByID: func(ctx context.Context, id int64, userRACF string) (*Guest, error) {
			g := sampleGuest()
			g.Confirmed = true
			return &g, nil
		},
		setConfirmedFn: func(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error) {
			g := sampleGuest()
			g.Confirmed = false
			return &g, nil
		},
	}
	users := &mockUserBridge{
		getGuestIDByPhoneFn: func(ctx context.Context, phone string) (*int64, error) {
			id := int64(1)
			return &id, nil
		},
		getGuestIDByUserIDFn: func(ctx context.Context, userID int64) (*int64, error) {
			id := int64(1)
			return &id, nil
		},
	}
	svc := newTestService(repo, users)

	guest, err := svc.CancelByPhone(context.Background(), "43999999999", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if guest.Confirmed {
		t.Fatal("expected confirmed to be false")
	}
}

func TestServiceConfirmRejectsCrossFamily(t *testing.T) {
	repo := &mockRepository{
		getByID: func(ctx context.Context, id int64, userRACF string) (*Guest, error) {
			g := sampleGuest()
			if id == 1 {
				g.ID = 1
				g.FamilyGroup = 1
				return &g, nil
			}
			g.ID = 99
			g.FamilyGroup = 2
			return &g, nil
		},
	}
	users := &mockUserBridge{
		getGuestIDByUserIDFn: func(ctx context.Context, userID int64) (*int64, error) {
			id := int64(1)
			return &id, nil
		},
	}
	svc := newTestService(repo, users)

	_, err := svc.Confirm(context.Background(), 99, 1)
	assertAppError(t, err, http.StatusForbidden, "you can only confirm guests in your own family")
}

func TestServiceListMyFamily(t *testing.T) {
	t.Run("admin returns empty", func(t *testing.T) {
		users := &mockUserBridge{
			getGuestIDByUserIDFn: func(ctx context.Context, userID int64) (*int64, error) {
				return nil, nil
			},
		}
		svc := newTestService(&mockRepository{}, users)
		guests, err := svc.ListMyFamily(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(guests) != 0 {
			t.Fatalf("expected empty slice, got %d guests", len(guests))
		}
	})

	t.Run("returns family roster", func(t *testing.T) {
		repo := &mockRepository{
			getByID: func(ctx context.Context, id int64, userRACF string) (*Guest, error) {
				g := sampleGuest()
				return &g, nil
			},
			listByFamilyGroupFn: func(ctx context.Context, familyGroup int64) ([]Guest, error) {
				g1 := sampleGuest()
				g1.ID = 1
				g2 := sampleGuest()
				g2.ID = 2
				g2.FirstName = "Maria"
				return []Guest{g1, g2}, nil
			},
		}
		users := &mockUserBridge{
			getGuestIDByUserIDFn: func(ctx context.Context, userID int64) (*int64, error) {
				id := int64(1)
				return &id, nil
			},
		}
		svc := newTestService(repo, users)
		guests, err := svc.ListMyFamily(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(guests) != 2 {
			t.Fatalf("expected 2 guests, got %d", len(guests))
		}
	})
}

func TestServiceSetConfirmedBatch(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var updatedIDs []int64
		var updatedConfirmed bool
		repo := &mockRepository{
			getByID: func(ctx context.Context, id int64, userRACF string) (*Guest, error) {
				g := sampleGuest()
				return &g, nil
			},
			getByIDsFn: func(ctx context.Context, ids []int64) ([]Guest, error) {
				out := []Guest{}
				for _, id := range ids {
					g := sampleGuest()
					g.ID = id
					out = append(out, g)
				}
				return out, nil
			},
			setConfirmedByIDsFn: func(ctx context.Context, ids []int64, confirmed bool, userRACF string) ([]Guest, error) {
				updatedIDs = ids
				updatedConfirmed = confirmed
				out := []Guest{}
				for _, id := range ids {
					g := sampleGuest()
					g.ID = id
					g.Confirmed = confirmed
					out = append(out, g)
				}
				return out, nil
			},
		}
		users := &mockUserBridge{
			getGuestIDByUserIDFn: func(ctx context.Context, userID int64) (*int64, error) {
				id := int64(1)
				return &id, nil
			},
		}
		svc := newTestService(repo, users)
		guests, err := svc.SetConfirmedBatch(context.Background(), BatchConfirmInput{GuestIDs: []int64{1, 2}, Confirmed: true}, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(guests) != 2 {
			t.Fatalf("expected 2 updated guests, got %d", len(guests))
		}
		if !updatedConfirmed {
			t.Fatal("expected updatedConfirmed=true")
		}
		if len(updatedIDs) != 2 {
			t.Fatalf("expected 2 updated IDs, got %d", len(updatedIDs))
		}
	})

	t.Run("rejects cross-family", func(t *testing.T) {
		repo := &mockRepository{
			getByID: func(ctx context.Context, id int64, userRACF string) (*Guest, error) {
				g := sampleGuest()
				return &g, nil
			},
			getByIDsFn: func(ctx context.Context, ids []int64) ([]Guest, error) {
				g1 := sampleGuest()
				g1.ID = 1
				g1.FamilyGroup = 1
				g2 := sampleGuest()
				g2.ID = 2
				g2.FamilyGroup = 999
				return []Guest{g1, g2}, nil
			},
		}
		users := &mockUserBridge{
			getGuestIDByUserIDFn: func(ctx context.Context, userID int64) (*int64, error) {
				id := int64(1)
				return &id, nil
			},
		}
		svc := newTestService(repo, users)
		_, err := svc.SetConfirmedBatch(context.Background(), BatchConfirmInput{GuestIDs: []int64{1, 2}, Confirmed: true}, 1)
		assertAppError(t, err, http.StatusForbidden, "you can only confirm guests in your own family")
	})

	t.Run("rejects missing guests", func(t *testing.T) {
		repo := &mockRepository{
			getByID: func(ctx context.Context, id int64, userRACF string) (*Guest, error) {
				g := sampleGuest()
				return &g, nil
			},
			getByIDsFn: func(ctx context.Context, ids []int64) ([]Guest, error) {
				g := sampleGuest()
				return []Guest{g}, nil
			},
		}
		users := &mockUserBridge{
			getGuestIDByUserIDFn: func(ctx context.Context, userID int64) (*int64, error) {
				id := int64(1)
				return &id, nil
			},
		}
		svc := newTestService(repo, users)
		_, err := svc.SetConfirmedBatch(context.Background(), BatchConfirmInput{GuestIDs: []int64{1, 2}, Confirmed: true}, 1)
		assertAppError(t, err, http.StatusNotFound, "one or more guests not found")
	})

	t.Run("rejects empty ids", func(t *testing.T) {
		svc := newTestService(&mockRepository{}, defaultUserBridge())
		_, err := svc.SetConfirmedBatch(context.Background(), BatchConfirmInput{GuestIDs: nil, Confirmed: true}, 1)
		assertAppError(t, err, http.StatusBadRequest, "GuestIDs failed on required validation")
	})
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
			svc := newTestService(&mockRepository{deleteFn: tt.mockFn}, defaultUserBridge())
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
