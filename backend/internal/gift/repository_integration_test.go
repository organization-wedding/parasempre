//go:build integration
// +build integration

package gift

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/database"
)

func setupRepo(t *testing.T) (*PostgresRepository, context.Context) {
	t.Helper()
	pool := database.NewTestPool(t)
	database.CleanTable(t, pool, "gift_transactions")
	database.CleanTable(t, pool, "gifts")
	return NewPostgresRepository(pool), context.Background()
}

func TestIntegrationCreateAndGet(t *testing.T) {
	repo, ctx := setupRepo(t)

	input := CreateGiftInput{
		Name:       "Panela Inox",
		PriceCents: 19990,
	}
	created, err := repo.Create(ctx, input, "panela inox", "TST01")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if created.Name != "Panela Inox" {
		t.Fatalf("expected name Panela Inox, got %q", created.Name)
	}
	if created.Status != "active" {
		t.Fatalf("expected default status 'active', got %q", created.Status)
	}

	fetched, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if fetched.Name != created.Name {
		t.Fatalf("GetByID returned different name: %q vs %q", fetched.Name, created.Name)
	}
}

func TestIntegrationCreateDuplicateDedupeKeyReturnsConflict(t *testing.T) {
	repo, ctx := setupRepo(t)

	input := CreateGiftInput{Name: "Panela", PriceCents: 10000}
	if _, err := repo.Create(ctx, input, "panela", "TST01"); err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	_, err := repo.Create(ctx, input, "panela", "TST01")
	if err == nil {
		t.Fatal("expected conflict, got nil")
	}
	ae, ok := apperror.IsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != 409 {
		t.Fatalf("expected 409, got %d", ae.Code)
	}
}

func TestIntegrationCreateZeroPriceReturnsValidation(t *testing.T) {
	repo, ctx := setupRepo(t)

	_, err := repo.Create(ctx, CreateGiftInput{
		Name:       "Broken",
		PriceCents: 0,
	}, "broken", "TST01")
	if err == nil {
		t.Fatal("expected validation error for zero price, got nil")
	}
	ae, ok := apperror.IsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != 400 {
		t.Fatalf("expected 400, got %d", ae.Code)
	}
}

func TestIntegrationCreateInvalidStatusReturnsValidation(t *testing.T) {
	repo, ctx := setupRepo(t)

	bogus := "bogus"
	_, err := repo.Create(ctx, CreateGiftInput{
		Name:       "Broken Status",
		PriceCents: 1000,
		Status:     &bogus,
	}, "broken status", "TST01")
	if err == nil {
		t.Fatal("expected validation error for invalid status, got nil")
	}
	ae, ok := apperror.IsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != 400 {
		t.Fatalf("expected 400, got %d", ae.Code)
	}
}

func TestIntegrationCreateNonHttpsImageURLFails(t *testing.T) {
	repo, ctx := setupRepo(t)

	insecure := "http://img.example.com/p.jpg"
	_, err := repo.Create(ctx, CreateGiftInput{
		Name:       "No SSL",
		PriceCents: 1000,
		ImageURL:   &insecure,
	}, "no ssl", "TST01")
	if err == nil {
		t.Fatal("expected validation error for http URL, got nil")
	}
}

func TestIntegrationUpdateRecalculatesFields(t *testing.T) {
	repo, ctx := setupRepo(t)

	created, err := repo.Create(ctx, CreateGiftInput{
		Name:       "Original",
		PriceCents: 10000,
	}, "original", "TST01")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	newName := "Updated Name"
	newPrice := int64(25000)
	newKey := "updated name"
	updated, err := repo.Update(ctx, created.ID, UpdateGiftInput{
		Name:       &newName,
		PriceCents: &newPrice,
	}, &newKey, "TST02")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Name != newName {
		t.Fatalf("expected name %q, got %q", newName, updated.Name)
	}
	if updated.PriceCents != newPrice {
		t.Fatalf("expected price %d, got %d", newPrice, updated.PriceCents)
	}
	if updated.DedupeKey != newKey {
		t.Fatalf("expected dedupe_key %q, got %q", newKey, updated.DedupeKey)
	}
	if updated.UpdatedBy != "TST02" {
		t.Fatalf("expected updated_by TST02, got %q", updated.UpdatedBy)
	}
	if !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Fatalf("expected updated_at to move forward: before=%v after=%v", created.UpdatedAt, updated.UpdatedAt)
	}
}

func TestIntegrationUpdateNotFound(t *testing.T) {
	repo, ctx := setupRepo(t)

	_, err := repo.Update(ctx, 99999, UpdateGiftInput{}, nil, "TST01")
	if err == nil {
		t.Fatal("expected NotFound, got nil")
	}
	ae, ok := apperror.IsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != 404 {
		t.Fatalf("expected 404, got %d", ae.Code)
	}
}

func TestIntegrationDeleteIsSoftAndRowPersists(t *testing.T) {
	repo, ctx := setupRepo(t)
	pool := database.NewTestPool(t)

	created, err := repo.Create(ctx, CreateGiftInput{
		Name:       "To Delete",
		PriceCents: 1000,
	}, "to delete", "TST01")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := repo.Delete(ctx, created.ID, "TST02"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, err := repo.GetByID(ctx, created.ID); err == nil {
		t.Fatal("expected NotFound after soft-delete, got nil")
	}

	var deletedAt *time.Time
	var updatedBy string
	err = pool.QueryRow(ctx,
		`SELECT deleted_at, updated_by FROM gifts WHERE id = $1`, created.ID).
		Scan(&deletedAt, &updatedBy)
	if err != nil {
		t.Fatalf("row should still exist after soft-delete: %v", err)
	}
	if deletedAt == nil {
		t.Fatal("expected deleted_at to be set after soft-delete")
	}
	if updatedBy != "TST02" {
		t.Fatalf("expected updated_by 'TST02' after soft-delete, got %q", updatedBy)
	}
}

func TestIntegrationCreateWithSameNameAfterDeleteIsAllowed(t *testing.T) {
	repo, ctx := setupRepo(t)

	original, err := repo.Create(ctx, CreateGiftInput{
		Name:       "Panela",
		PriceCents: 10000,
	}, "panela", "TST01")
	if err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	if err := repo.Delete(ctx, original.ID, "TST01"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	recreated, err := repo.Create(ctx, CreateGiftInput{
		Name:       "Panela",
		PriceCents: 12000,
	}, "panela", "TST01")
	if err != nil {
		t.Fatalf("expected to recreate after soft-delete, got: %v", err)
	}
	if recreated.ID == original.ID {
		t.Fatal("expected new row after soft-delete, got same ID")
	}
}

func TestIntegrationListExcludesDeleted(t *testing.T) {
	repo, ctx := setupRepo(t)

	g, err := repo.Create(ctx, CreateGiftInput{Name: "Visible", PriceCents: 1000}, "visible", "TST01")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if _, err := repo.Create(ctx, CreateGiftInput{Name: "Hidden", PriceCents: 1000}, "hidden", "TST01"); err != nil {
		t.Fatalf("Create hidden failed: %v", err)
	}
	if err := repo.Delete(ctx, g.ID, "TST01"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	gifts, total, err := repo.List(ctx, 10, 0, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1 after soft-delete, got %d", total)
	}
	if len(gifts) != 1 {
		t.Fatalf("expected 1 gift in list, got %d", len(gifts))
	}
	if gifts[0].Name != "Hidden" {
		t.Fatalf("expected 'Hidden' to remain, got %q", gifts[0].Name)
	}
}

func TestIntegrationFindByDedupeKeysIgnoresDeleted(t *testing.T) {
	repo, ctx := setupRepo(t)

	g, err := repo.Create(ctx, CreateGiftInput{Name: "Ghost", PriceCents: 1000}, "ghost", "TST01")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := repo.Delete(ctx, g.ID, "TST01"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	found, err := repo.FindByDedupeKeys(ctx, []string{"ghost"})
	if err != nil {
		t.Fatalf("FindByDedupeKeys failed: %v", err)
	}
	if found["ghost"] {
		t.Fatal("expected soft-deleted key to be ignored by FindByDedupeKeys")
	}
}

func TestIntegrationDeleteNotFound(t *testing.T) {
	repo, ctx := setupRepo(t)

	err := repo.Delete(ctx, 99999, "TST01")
	if err == nil {
		t.Fatal("expected NotFound, got nil")
	}
}

func TestIntegrationFindByDedupeKeys(t *testing.T) {
	repo, ctx := setupRepo(t)

	if _, err := repo.Create(ctx, CreateGiftInput{Name: "A", PriceCents: 1000}, "a", "TST01"); err != nil {
		t.Fatalf("Create A failed: %v", err)
	}
	if _, err := repo.Create(ctx, CreateGiftInput{Name: "B", PriceCents: 1000}, "b", "TST01"); err != nil {
		t.Fatalf("Create B failed: %v", err)
	}

	found, err := repo.FindByDedupeKeys(ctx, []string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("FindByDedupeKeys failed: %v", err)
	}
	if !found["a"] {
		t.Error("expected 'a' to be found")
	}
	if !found["b"] {
		t.Error("expected 'b' to be found")
	}
	if found["c"] {
		t.Error("expected 'c' to NOT be found")
	}
}

func TestIntegrationFindByDedupeKeysEmpty(t *testing.T) {
	repo, ctx := setupRepo(t)

	found, err := repo.FindByDedupeKeys(ctx, []string{})
	if err != nil {
		t.Fatalf("FindByDedupeKeys empty failed: %v", err)
	}
	if len(found) != 0 {
		t.Fatalf("expected empty map, got %+v", found)
	}
}

func TestIntegrationListPaginationAndStatusFilter(t *testing.T) {
	repo, ctx := setupRepo(t)

	inactive := "inactive"
	for i := 0; i < 5; i++ {
		status := "active"
		if i >= 3 {
			status = "inactive"
		}
		input := CreateGiftInput{
			Name:       fmt.Sprintf("Gift %d", i),
			PriceCents: 10000,
			Status:     &status,
		}
		if _, err := repo.Create(ctx, input, fmt.Sprintf("gift %d", i), "TST01"); err != nil {
			t.Fatalf("Create %d failed: %v", i, err)
		}
	}

	gifts, total, err := repo.List(ctx, 2, 0, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 5 {
		t.Fatalf("expected total 5, got %d", total)
	}
	if len(gifts) != 2 {
		t.Fatalf("expected 2 gifts in page, got %d", len(gifts))
	}

	inactiveGifts, inactiveTotal, err := repo.List(ctx, 10, 0, &inactive)
	if err != nil {
		t.Fatalf("List with status filter failed: %v", err)
	}
	if inactiveTotal != 2 {
		t.Fatalf("expected 2 inactive gifts, got %d", inactiveTotal)
	}
	if len(inactiveGifts) != 2 {
		t.Fatalf("expected 2 inactive gifts in list, got %d", len(inactiveGifts))
	}
	for _, g := range inactiveGifts {
		if g.Status != "inactive" {
			t.Errorf("expected status inactive, got %q", g.Status)
		}
	}
}
