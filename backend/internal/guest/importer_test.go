package guest

import (
	"strings"
	"testing"
)

func TestParseCSV(t *testing.T) {
	tests := []struct {
		name    string
		csv     string
		want    int
		wantErr bool
	}{
		{
			name: "valid CSV with header",
			csv:  "nome,sobrenome,telefone,relacionamento\nJoão,Silva,11999999999,noivo\nMaria,Santos,11888888888,noiva\n",
			want: 2,
		},
		{
			name: "valid CSV with extra whitespace",
			csv:  "nome,sobrenome,telefone,relacionamento\n João , Silva , 11999999999 , noivo \n",
			want: 1,
		},
		{
			name:    "empty CSV (header only)",
			csv:     "nome,sobrenome,telefone,relacionamento\n",
			want:    0,
			wantErr: false,
		},
		{
			name:    "missing columns",
			csv:     "nome,sobrenome\nJoão,Silva\n",
			wantErr: true,
		},
		{
			name:    "empty input",
			csv:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guests, err := ParseCSV(strings.NewReader(tt.csv))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(guests) != tt.want {
				t.Fatalf("expected %d guests, got %d", tt.want, len(guests))
			}
		})
	}
}

func TestParseCSVFieldMapping(t *testing.T) {
	csv := "nome,sobrenome,telefone,relacionamento\nJoão,Silva,11999999999,noivo\n"
	guests, err := ParseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(guests) != 1 {
		t.Fatalf("expected 1 guest, got %d", len(guests))
	}
	g := guests[0]
	if g.Nome != "João" {
		t.Errorf("expected nome 'João', got %q", g.Nome)
	}
	if g.Sobrenome != "Silva" {
		t.Errorf("expected sobrenome 'Silva', got %q", g.Sobrenome)
	}
	if g.Telefone != "11999999999" {
		t.Errorf("expected telefone '11999999999', got %q", g.Telefone)
	}
	if g.Relacionamento != "noivo" {
		t.Errorf("expected relacionamento 'noivo', got %q", g.Relacionamento)
	}
}

func TestParseXLSX(t *testing.T) {
	// XLSX parsing is tested via integration with excelize.
	// We test error handling for invalid input here.
	_, err := ParseXLSX(strings.NewReader("not a valid xlsx"))
	if err == nil {
		t.Fatal("expected error for invalid XLSX, got nil")
	}
}
