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
			csv:  "first_name,last_name,relationship,family_group\nJoão,Silva,P,1\nMaria,Santos,R,2\n",
			want: 2,
		},
		{
			name: "valid CSV with extra whitespace",
			csv:  "first_name,last_name,relationship,family_group\n João , Silva , P , 1 \n",
			want: 1,
		},
		{
			name:    "empty CSV (header only)",
			csv:     "first_name,last_name,relationship,family_group\n",
			want:    0,
			wantErr: false,
		},
		{
			name:    "missing columns",
			csv:     "first_name,last_name\nJoão,Silva\n",
			wantErr: true,
		},
		{
			name:    "missing family_group column",
			csv:     "first_name,last_name,relationship\nJoão,Silva,P\n",
			wantErr: true,
		},
		{
			name:    "empty input",
			csv:     "",
			wantErr: true,
		},
		{
			name:    "invalid family_group value",
			csv:     "first_name,last_name,relationship,family_group\nJoão,Silva,P,abc\n",
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
	csv := "first_name,last_name,relationship,family_group\nJoão,Silva,P,5\n"
	guests, err := ParseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(guests) != 1 {
		t.Fatalf("expected 1 guest, got %d", len(guests))
	}
	g := guests[0]
	if g.FirstName != "João" {
		t.Errorf("expected first_name 'João', got %q", g.FirstName)
	}
	if g.LastName != "Silva" {
		t.Errorf("expected last_name 'Silva', got %q", g.LastName)
	}
	if g.Relationship != "P" {
		t.Errorf("expected relationship 'P', got %q", g.Relationship)
	}
	if g.FamilyGroup == nil || *g.FamilyGroup != 5 {
		t.Errorf("expected family_group 5, got %v", g.FamilyGroup)
	}
}

func TestParseXLSX(t *testing.T) {
	_, err := ParseXLSX(strings.NewReader("not a valid xlsx"))
	if err == nil {
		t.Fatal("expected error for invalid XLSX, got nil")
	}
}
