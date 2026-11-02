package validator

import "testing"

func TestNIT(t *testing.T) {
	cases := []struct {
		in, norm string
		ok       bool
	}{
		{"5487963-2", "5487963-2", false},
		{"576937-K", "576937-K", true},
		{"576937k", "576937-K", true},
		{"cf", "CF", true},
		{" 1234567-0 ", "1234567-0", false},
		{"12345678", "1234567-8", false},
		{"123456-1", "123456-1", false},
		{"abc", "ABC", false},
	}
	for _, c := range cases {
		got := NormalizeNIT(c.in)
		if got != c.norm {
			t.Fatalf("NormalizeNIT(%q) = %q, esperaba %q", c.in, got, c.norm)
		}
	}
	// Cálculo manual: 1234567 -> pesos 2..8 desde la derecha.
	if !ValidNIT("1234567-"+checkDigit("1234567")) || !ValidNIT("576937-K") || !ValidNIT("CF") {
		t.Fatal("NIT válidos rechazados")
	}
	if ValidNIT("576937-1") || ValidNIT("ABC") || ValidNIT("") {
		t.Fatal("NIT inválidos aceptados")
	}
}

func checkDigit(cuerpo string) string {
	suma := 0
	for i := 0; i < len(cuerpo); i++ {
		suma += int(cuerpo[len(cuerpo)-1-i]-'0') * (i + 2)
	}
	d := (11 - suma%11) % 11
	if d == 10 {
		return "K"
	}
	return string(rune('0' + d))
}

func TestValidateClienteNIT(t *testing.T) {
	if errs := validateNIT(NormalizeNIT("576937k")); errs.HasErrors() {
		t.Fatalf("NIT valido rechazado: %v", errs)
	}
	if errs := validateNIT(NormalizeNIT("5487963-2")); !errs.HasErrors() || errs[0].Field != "nit" {
		t.Fatalf("NIT con verificador incorrecto aceptado: %v", errs)
	}
	if errs := validateNIT(""); errs.HasErrors() {
		t.Fatal("el NIT es opcional")
	}
}
