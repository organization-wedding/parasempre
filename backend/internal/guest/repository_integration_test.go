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

	fetched, err := repo.GetByIDAny(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByIDAny failed: %v", err)
	}
	if fetched.FirstName != created.FirstName {
		t.Fatalf("GetByIDAny returned different first_name: %q vs %q", fetched.FirstName, created.FirstName)
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

// TestIntegrationCoupleSharesGuestList reproduz o bug em que noivo e noiva
// (usuários distintos com URACFs diferentes, co-administradores do MESMO
// casamento) só enxergavam os convidados que eles próprios cadastraram.
// A lista de convidados é um dataset compartilhado do casamento: quem cria
// (created_by) é só auditoria, não dono. List e GetByIDAny não devem mais
// recortar por criador — o convidado cadastrado pelo NOIVO precisa aparecer
// independentemente de quem está consultando.
func TestIntegrationCoupleSharesGuestList(t *testing.T) {
	const groomRACF = "ZGRM1"

	pool := database.NewTestPool(t)
	tx := database.BeginTestTx(t, pool)
	repo := NewPostgresRepository(pool).WithTx(tx)
	ctx := context.Background()

	fgGroom := int64(90001)
	groomGuest, err := repo.Create(ctx, CreateGuestInput{
		FirstName:    "Convidado",
		LastName:     "DoNoivo",
		Relationship: "P",
		FamilyGroup:  &fgGroom,
	}, groomRACF)
	if err != nil {
		t.Fatalf("Create (groom) failed: %v", err)
	}

	// GetByIDAny não é recortado por criador: o convidado do NOIVO é visível.
	fetched, err := repo.GetByIDAny(ctx, groomGuest.ID)
	if err != nil {
		t.Fatalf("GetByIDAny should see groom's guest, got error: %v", err)
	}
	if fetched.ID != groomGuest.ID {
		t.Fatalf("GetByIDAny returned wrong guest: got %d, want %d", fetched.ID, groomGuest.ID)
	}

	// List é a lista compartilhada do casamento e deve incluir o convidado
	// criado pelo NOIVO, sem filtrar por created_by.
	guests, _, err := repo.List(ctx, 1000, 0, ListFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if !containsGuestID(guests, groomGuest.ID) {
		t.Fatalf("List did not include groom's guest id %d (created_by filter still scoping by creator)", groomGuest.ID)
	}
}

func containsGuestID(guests []Guest, id int64) bool {
	for _, g := range guests {
		if g.ID == id {
			return true
		}
	}
	return false
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

	// List is the wedding's shared roster (not scoped by created_by), so total
	// reflects every guest visible in the transaction — at least the 5 created
	// here. The assertion stays isolation-safe against pre-existing rows.
	guests, total, err := repo.List(ctx, 2, 0, ListFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total < 5 {
		t.Fatalf("expected total >= 5, got %d", total)
	}
	if len(guests) != 2 {
		t.Fatalf("expected 2 guests in page (limit=2), got %d", len(guests))
	}
}

func TestIntegrationListSearchSpansAllPages(t *testing.T) {
	pool := database.NewTestPool(t)
	tx := database.BeginTestTx(t, pool)
	repo := NewPostgresRepository(pool).WithTx(tx)
	ctx := context.Background()

	for i := range 25 {
		fg := int64(70000 + i)
		if _, err := repo.Create(ctx, CreateGuestInput{
			FirstName:    "Filler",
			LastName:     string(rune('A'+i%26)) + string(rune('a'+i)),
			Relationship: "P",
			FamilyGroup:  &fg,
		}, "TST01"); err != nil {
			t.Fatalf("Create filler %d failed: %v", i, err)
		}
	}
	fgTarget := int64(79999)
	if _, err := repo.Create(ctx, CreateGuestInput{
		FirstName:    "Aurelio",
		LastName:     "Buscavel",
		Relationship: "R",
		FamilyGroup:  &fgTarget,
	}, "TST01"); err != nil {
		t.Fatalf("Create target failed: %v", err)
	}

	guests, total, err := repo.List(ctx, 20, 0, ListFilters{Search: "aurelio buscavel"})
	if err != nil {
		t.Fatalf("List with search failed: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected filtered total 1, got %d", total)
	}
	if len(guests) != 1 || guests[0].FirstName != "Aurelio" {
		t.Fatalf("expected to find Aurelio across pages, got %+v", guests)
	}
}

func TestIntegrationStats(t *testing.T) {
	pool := database.NewTestPool(t)
	tx := database.BeginTestTx(t, pool)
	repo := NewPostgresRepository(pool).WithTx(tx)
	ctx := context.Background()

	before, err := repo.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats (before) failed: %v", err)
	}

	mk := func(fg int64, attending *bool) {
		g, err := repo.Create(ctx, CreateGuestInput{
			FirstName:    "Stat",
			LastName:     string(rune('A' + fg%26)),
			Relationship: "P",
			FamilyGroup:  &fg,
		}, "TST01")
		if err != nil {
			t.Fatalf("Create stat guest failed: %v", err)
		}
		if attending != nil {
			if _, err := repo.SetAttending(ctx, g.ID, *attending, "TST01"); err != nil {
				t.Fatalf("SetAttending failed: %v", err)
			}
		}
	}
	yes, no := true, false
	mk(60001, &yes)
	mk(60002, &yes)
	mk(60003, &no)
	mk(60004, nil)

	after, err := repo.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats (after) failed: %v", err)
	}
	if after.Total-before.Total != 4 {
		t.Fatalf("expected total +4, got +%d", after.Total-before.Total)
	}
	if after.Confirmed-before.Confirmed != 2 {
		t.Fatalf("expected confirmed +2, got +%d", after.Confirmed-before.Confirmed)
	}
	if after.Declined-before.Declined != 1 {
		t.Fatalf("expected declined +1, got +%d", after.Declined-before.Declined)
	}
	if after.Pending-before.Pending != 1 {
		t.Fatalf("expected pending +1, got +%d", after.Pending-before.Pending)
	}
}
