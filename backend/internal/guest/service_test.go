package guest

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockRepository implements Repository with function fields for easy stubbing.
type mockRepository struct {
	listFn   func(ctx context.Context) ([]Guest, error)
	getByID  func(ctx context.Context, id string) (*Guest, error)
	createFn func(ctx context.Context, input CreateGuestInput) (*Guest, error)
	updateFn func(ctx context.Context, id string, input UpdateGuestInput) (*Guest, error)
	deleteFn func(ctx context.Context, id string) error
}

func (m *mockRepository) List(ctx context.Context) ([]Guest, error) {
	return m.listFn(ctx)
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*Guest, error) {
	return m.getByID(ctx, id)
}

func (m *mockRepository) Create(ctx context.Context, input CreateGuestInput) (*Guest, error) {
	return m.createFn(ctx, input)
}

func (m *mockRepository) Update(ctx context.Context, id string, input UpdateGuestInput) (*Guest, error) {
	return m.updateFn(ctx, id, input)
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

func sampleGuest() Guest {
	return Guest{
		ID:             "abc-123",
		Nome:           "João",
		Sobrenome:      "Silva",
		Telefone:       "11999999999",
		Relacionamento: "noivo",
		Confirmacao:    false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
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
			svc := NewService(&mockRepository{listFn: tt.mockFn})
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
		id      string
		mockFn  func(ctx context.Context, id string) (*Guest, error)
		wantErr bool
	}{
		{
			name: "returns guest",
			id:   "abc-123",
			mockFn: func(ctx context.Context, id string) (*Guest, error) {
				g := sampleGuest()
				return &g, nil
			},
		},
		{
			name: "not found",
			id:   "not-exist",
			mockFn: func(ctx context.Context, id string) (*Guest, error) {
				return nil, ErrNotFound
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&mockRepository{getByID: tt.mockFn})
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
			name: "valid input",
			input: CreateGuestInput{
				Nome:           "Maria",
				Sobrenome:      "Santos",
				Telefone:       "11888888888",
				Relacionamento: "noiva",
			},
		},
		{
			name: "missing nome",
			input: CreateGuestInput{
				Sobrenome:      "Santos",
				Telefone:       "11888888888",
				Relacionamento: "noiva",
			},
			wantErr: "nome is required",
		},
		{
			name: "missing sobrenome",
			input: CreateGuestInput{
				Nome:           "Maria",
				Telefone:       "11888888888",
				Relacionamento: "noiva",
			},
			wantErr: "sobrenome is required",
		},
		{
			name: "missing telefone",
			input: CreateGuestInput{
				Nome:           "Maria",
				Sobrenome:      "Santos",
				Relacionamento: "noiva",
			},
			wantErr: "telefone is required",
		},
		{
			name: "invalid relacionamento",
			input: CreateGuestInput{
				Nome:           "Maria",
				Sobrenome:      "Santos",
				Telefone:       "11888888888",
				Relacionamento: "amigo",
			},
			wantErr: "relacionamento must be 'noivo' or 'noiva'",
		},
		{
			name: "missing relacionamento",
			input: CreateGuestInput{
				Nome:      "Maria",
				Sobrenome: "Santos",
				Telefone:  "11888888888",
			},
			wantErr: "relacionamento must be 'noivo' or 'noiva'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				createFn: func(ctx context.Context, input CreateGuestInput) (*Guest, error) {
					g := sampleGuest()
					return &g, nil
				},
			}
			svc := NewService(repo)
			_, err := svc.Create(context.Background(), tt.input)
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

func TestServiceUpdate(t *testing.T) {
	noivo := "noivo"
	invalid := "amigo"

	tests := []struct {
		name    string
		input   UpdateGuestInput
		wantErr string
	}{
		{
			name:  "valid partial update",
			input: UpdateGuestInput{Relacionamento: &noivo},
		},
		{
			name:  "empty update is valid",
			input: UpdateGuestInput{},
		},
		{
			name:    "invalid relacionamento",
			input:   UpdateGuestInput{Relacionamento: &invalid},
			wantErr: "relacionamento must be 'noivo' or 'noiva'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				updateFn: func(ctx context.Context, id string, input UpdateGuestInput) (*Guest, error) {
					g := sampleGuest()
					return &g, nil
				},
			}
			svc := NewService(repo)
			_, err := svc.Update(context.Background(), "abc-123", tt.input)
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

func TestServiceDelete(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(ctx context.Context, id string) error
		wantErr bool
	}{
		{
			name: "success",
			mockFn: func(ctx context.Context, id string) error {
				return nil
			},
		},
		{
			name: "not found",
			mockFn: func(ctx context.Context, id string) error {
				return ErrNotFound
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&mockRepository{deleteFn: tt.mockFn})
			err := svc.Delete(context.Background(), "abc-123")
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
