//go:build integration
// +build integration

package guest

import (
	"context"
	"testing"

	"github.com/ferjunior7/parasempre/backend/internal/database"
)

func TestIntegrationCreateAndGet(t *testing.T) {
	pool := database.NewTestPool(t)
	tx := database.BeginTestTx(t, pool)
	repo := NewPostgresRepository(pool).WithTx(tx)
	ctx := context.Background()

	fg := int64(1)
	input := CreateGuestInput{
		FirstName:    "João",
		LastName:     "Integration",
		Relationship: "P",
		FamilyGroup:  &fg,
	}

	created, err := repo.Create(ctx, input, "TST01")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if created.FirstName != "João" {
		t.Fatalf("expected first_name João, got %q", created.FirstName)
	}

	fetched, err := repo.GetByID(ctx, created.ID, "TST01")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if fetched.FirstName != created.FirstName {
		t.Fatalf("GetByID returned different first_name: %q vs %q", fetched.FirstName, created.FirstName)
	}
}

func TestIntegrationUniqueConstraint(t *testing.T) {
	pool := database.NewTestPool(t)
	tx := database.BeginTestTx(t, pool)
	repo := NewPostgresRepository(pool).WithTx(tx)
	ctx := context.Background()

	fg := int64(1)
	input := CreateGuestInput{
		FirstName:    "Maria",
		LastName:     "Duplicada",
		Relationship: "R",
		FamilyGroup:  &fg,
	}

	_, err := repo.Create(ctx, input, "TST01")
	if err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	_, err = repo.Create(ctx, input, "TST01")
	if err == nil {
		t.Fatal("expected error on duplicate create, got nil")
	}
}

func TestIntegrationListPagination(t *testing.T) {
	pool := database.NewTestPool(t)
	tx := database.BeginTestTx(t, pool)
	repo := NewPostgresRepository(pool).WithTx(tx)
	ctx := context.Background()

	for i := range 5 {
		fg := int64(i + 1)
		input := CreateGuestInput{
			FirstName:    "Guest",
			LastName:     string(rune('A' + i)),
			Relationship: "P",
			FamilyGroup:  &fg,
		}
		if _, err := repo.Create(ctx, input, "TST01"); err != nil {
			t.Fatalf("Create guest %d failed: %v", i, err)
		}
	}

	guests, total, err := repo.List(ctx, 2, 0, "TST01")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 5 {
		t.Fatalf("expected total 5, got %d", total)
	}
	if len(guests) != 2 {
		t.Fatalf("expected 2 guests in page, got %d", len(guests))
	}
}
