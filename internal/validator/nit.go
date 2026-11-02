package validator

import (
	"regexp"
	"strconv"
	"strings"
)

var nitRegex = regexp.MustCompile(`^[0-9]{1,12}[0-9K]$`)

// NormalizeNIT deja el NIT guatemalteco en formato "1234567-8" (o "CF" para consumidor final).
// Quita espacios y guiones y pasa a mayúsculas; si no puede normalizarse devuelve el texto limpio.
func NormalizeNIT(valor string) string {
	limpio := strings.ToUpper(strings.NewReplacer(" ", "", "-", "", ".", "").Replace(strings.TrimSpace(valor)))
	if limpio == "" || limpio == "CF" || !nitRegex.MatchString(limpio) {
		return limpio
	}
	return limpio[:len(limpio)-1] + "-" + limpio[len(limpio)-1:]
}

// ValidNIT verifica el dígito verificador del NIT de Guatemala (módulo 11, "K" equivale a 10).
func ValidNIT(normalizado string) bool {
	if normalizado == "CF" {
		return true
	}
	cuerpo, verificador, ok := strings.Cut(normalizado, "-")
	if !ok || cuerpo == "" || len(verificador) != 1 {
		return false
	}
	suma := 0
	for i := 0; i < len(cuerpo); i++ {
		digito, err := strconv.Atoi(string(cuerpo[len(cuerpo)-1-i]))
		if err != nil {
			return false
		}
		suma += digito * (i + 2)
	}
	esperado := (11 - suma%11) % 11
	if verificador == "K" {
		return esperado == 10
	}
	valor, err := strconv.Atoi(verificador)
	return err == nil && valor == esperado
}

func validateNIT(valor string) ValidationErrors {
	if valor == "" {
		return nil
	}
	if len(valor) > 20 || !ValidNIT(valor) {
		return ValidationErrors{{Field: "nit", Message: "NIT inválido: revisa los dígitos y el dígito verificador, o usa CF"}}
	}
	return nil
}
