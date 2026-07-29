package parser_test

import (
	"testing"

	"github.com/arturoeanton/goscim/scim/parser"
)

// Cada operador de comparación de RFC 7644 §3.4.2.2 debe traducirse al
// operador N1QL equivalente. gt/ge y lt/le estaban cruzados entre sí, así que
// toda búsqueda por rango devolvía el conjunto equivocado en silencio.
func TestOperadoresDeComparacion(t *testing.T) {
	cases := []struct {
		filtro string
		where  string
	}{
		{`userName eq "bjensen"`, "`userName` = \"bjensen\""},
		{`userName ne "bjensen"`, "`userName` <> \"bjensen\""},
		{`age gt 10`, "`age` > 10"},
		{`age ge 10`, "`age` >= 10"},
		{`age lt 10`, "`age` < 10"},
		{`age le 10`, "`age` <= 10"},
		{`title pr`, "`title` IS NOT NULL"},
	}

	for _, tc := range cases {
		t.Run(tc.filtro, func(t *testing.T) {
			page, count := parser.FilterToN1QL("User", tc.filtro)
			wantPage := "SELECT * FROM `User` WHERE " + tc.where
			if page != wantPage {
				t.Errorf("consulta de página:\n  obtenida: %s\n  esperada: %s", page, wantPage)
			}
			wantCount := "SELECT count(*) as count FROM `User` WHERE " + tc.where
			if count != wantCount {
				t.Errorf("consulta de conteo:\n  obtenida: %s\n  esperada: %s", count, wantCount)
			}
		})
	}
}

// gt/ge y lt/le tienen que ser distintos entre sí: si vuelven a cruzarse o a
// colapsar en el mismo operador, este test lo detecta aunque el de arriba se
// actualizara sin pensar.
func TestOperadoresDeRangoSonDistintos(t *testing.T) {
	gt, _ := parser.FilterToN1QL("User", "age gt 10")
	ge, _ := parser.FilterToN1QL("User", "age ge 10")
	lt, _ := parser.FilterToN1QL("User", "age lt 10")
	le, _ := parser.FilterToN1QL("User", "age le 10")

	if gt == ge {
		t.Error("gt y ge producen la misma consulta")
	}
	if lt == le {
		t.Error("lt y le producen la misma consulta")
	}
	// El operador estricto no debe llevar el "=" del inclusivo.
	if len(gt) >= len(ge) {
		t.Errorf("gt (%s) debería ser más estricto que ge (%s)", gt, ge)
	}
	if len(lt) >= len(le) {
		t.Errorf("lt (%s) debería ser más estricto que le (%s)", lt, le)
	}
}

// Los operadores de subcadena envuelven el valor con comodines según el caso.
func TestOperadoresDeSubcadena(t *testing.T) {
	cases := []struct {
		filtro string
		where  string
	}{
		{`emails co "example.com"`, "`emails` LIKE \"%example.com%\""},
		{`emails sw "example"`, "`emails` LIKE \"example%\""},
		{`emails ew ".com"`, "`emails` LIKE \"%.com\""},
	}
	for _, tc := range cases {
		t.Run(tc.filtro, func(t *testing.T) {
			page, _ := parser.FilterToN1QL("User", tc.filtro)
			want := "SELECT * FROM `User` WHERE " + tc.where
			if page != want {
				t.Errorf("\n  obtenida: %s\n  esperada: %s", page, want)
			}
		})
	}
}

// Un filtro vacío no debe generar cláusula WHERE.
func TestFiltroVacio(t *testing.T) {
	page, count := parser.FilterToN1QL("User", "")
	if page != "SELECT * FROM `User`" {
		t.Errorf("página = %s", page)
	}
	if count != "SELECT count(*)  as count FROM `User`" {
		t.Errorf("conteo = %s", count)
	}
}
