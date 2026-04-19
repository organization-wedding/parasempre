package gift

import "testing"

func TestNormalizeDedupeKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "lowercase simple", input: "Panela", want: "panela"},
		{name: "trim and collapse spaces", input: "  Panela   de  Inox  ", want: "panela de inox"},
		{name: "removes accents", input: "Máquina de Café", want: "maquina de cafe"},
		{name: "mixed case with accents", input: "JOGO de Xícaras", want: "jogo de xicaras"},
		{name: "cedilla", input: "Almofaça Elétrica", want: "almofaca eletrica"},
		{name: "empty string", input: "", want: ""},
		{name: "only spaces", input: "   ", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeDedupeKey(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeDedupeKey(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
