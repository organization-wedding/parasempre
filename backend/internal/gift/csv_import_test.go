package gift

import (
	"strings"
	"testing"
)

func TestParsePriceBRL(t *testing.T) {
	cases := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"150", 15000, false},
		{"150.00", 15000, false},
		{"150,00", 15000, false},
		{"150.5", 15050, false},
		{"150,5", 15050, false},
		{"1500.50", 150050, false},
		{"1500,50", 150050, false},
		{"R$ 150,00", 15000, false},
		{"  150  ", 15000, false},
		{"0", 0, false},
		{"", 0, true},
		{"abc", 0, true},
		{"150,000", 0, true},
		{"-10", 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := parsePriceBRL(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got %d", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("input %q: want %d cents, got %d", tc.input, tc.want, got)
			}
		})
	}
}

func TestParseCSVRows_HappyPath(t *testing.T) {
	csv := `name,description,price_brl,image_url,store_url
Panela Inox,Conjunto 5 peças,150.00,https://example.com/pan.jpg,https://example.com/pan
Jogo de Toalhas,,79,90,,
`
	rows, err := ParseCSVRows(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}

	if rows[0].Input.Name != "Panela Inox" {
		t.Fatalf("row 0 name: got %q", rows[0].Input.Name)
	}
	if rows[0].Input.PriceCents != 15000 {
		t.Fatalf("row 0 price_cents: want 15000, got %d", rows[0].Input.PriceCents)
	}
	if rows[0].DedupeKey != "panela inox" {
		t.Fatalf("row 0 dedupe_key: got %q", rows[0].DedupeKey)
	}
	if len(rows[0].Errors) != 0 {
		t.Fatalf("row 0 unexpected errors: %v", rows[0].Errors)
	}
	if rows[0].Input.Description == nil || *rows[0].Input.Description != "Conjunto 5 peças" {
		t.Fatalf("row 0 description mismatch")
	}
}

func TestParseCSVRows_SemicolonDelimiter(t *testing.T) {
	csv := `name;description;price_brl;image_url;store_url
Panela Inox;Conjunto 5 peças;150,00;https://example.com/pan.jpg;https://example.com/pan
`
	rows, err := ParseCSVRows(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	if rows[0].Input.Name != "Panela Inox" {
		t.Fatalf("name: got %q", rows[0].Input.Name)
	}
	if rows[0].Input.PriceCents != 15000 {
		t.Fatalf("price_cents: want 15000, got %d", rows[0].Input.PriceCents)
	}
	if len(rows[0].Errors) != 0 {
		t.Fatalf("unexpected errors: %v", rows[0].Errors)
	}
}

func TestParseCSVRows_UTF8BOM(t *testing.T) {
	csv := "\xef\xbb\xbfname;price_brl\nPanela Inox;150,00\n"
	rows, err := ParseCSVRows(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	if rows[0].Input.Name != "Panela Inox" {
		t.Fatalf("name with BOM stripped expected, got %q", rows[0].Input.Name)
	}
}

func TestParseCSVRows_MissingColumn(t *testing.T) {
	csv := `name,description
Só Nome,Apenas descrição
`
	_, err := ParseCSVRows(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for missing price_brl column")
	}
	if !strings.Contains(err.Error(), "price_brl") {
		t.Fatalf("error should mention price_brl, got %v", err)
	}
}

func TestParseCSVRows_InvalidRow(t *testing.T) {
	csv := `name,price_brl,image_url
,abc,not-a-url
Válido,50.00,https://example.com/x
`
	rows, err := ParseCSVRows(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	if len(rows[0].Errors) == 0 {
		t.Fatal("row 0 should have errors (empty name, bad price, bad url)")
	}
	if len(rows[1].Errors) != 0 {
		t.Fatalf("row 1 should be clean, got errors: %v", rows[1].Errors)
	}
}

func TestParseCSVRows_Empty(t *testing.T) {
	_, err := ParseCSVRows(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for empty file")
	}

	_, err = ParseCSVRows(strings.NewReader("name,price_brl\n"))
	if err == nil {
		t.Fatal("expected error for file with only header")
	}
}
