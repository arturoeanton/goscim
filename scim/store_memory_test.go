package scim

import (
	"errors"
	"testing"
)

// El MemoryStore es infraestructura de test: si miente, mienten todos los tests
// que se apoyan en él. Estos casos fijan su contrato contra el de Store.

func TestMemoryStoreCRUD(t *testing.T) {
	s := NewMemoryStore()
	if err := s.EnsureBucket("Element"); err != nil {
		t.Fatal(err)
	}

	if _, err := s.Get("Element", "1"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get sobre bucket vacío = %v, se esperaba ErrNotFound", err)
	}
	if err := s.Replace("Element", "1", map[string]interface{}{}); !errors.Is(err, ErrNotFound) {
		t.Errorf("Replace sobre inexistente = %v, se esperaba ErrNotFound", err)
	}
	if err := s.Remove("Element", "1"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Remove sobre inexistente = %v, se esperaba ErrNotFound", err)
	}

	if err := s.Upsert("Element", "1", map[string]interface{}{"name": "uno"}); err != nil {
		t.Fatal(err)
	}
	doc, err := s.Get("Element", "1")
	if err != nil || doc["name"] != "uno" {
		t.Fatalf("Get tras Upsert = %v, %v", doc, err)
	}

	if err := s.Replace("Element", "1", map[string]interface{}{"name": "dos"}); err != nil {
		t.Fatal(err)
	}
	doc, _ = s.Get("Element", "1")
	if doc["name"] != "dos" {
		t.Errorf("Replace no aplicó: %v", doc)
	}

	if err := s.Remove("Element", "1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get("Element", "1"); !errors.Is(err, ErrNotFound) {
		t.Errorf("el documento sobrevivió al Remove")
	}
}

// Upsert sobre un bucket que nadie declaró debe funcionar igual que en
// Couchbase, donde el bucket ya existe desde el arranque.
func TestMemoryStoreUpsertCreaBucket(t *testing.T) {
	s := NewMemoryStore()
	if err := s.Upsert("Nuevo", "1", map[string]interface{}{"a": 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get("Nuevo", "1"); err != nil {
		t.Fatal(err)
	}
}

// El store debe aislar al llamante de su estado interno, como haría un driver
// real que serializa.
func TestMemoryStoreAislaElEstado(t *testing.T) {
	s := NewMemoryStore()
	original := map[string]interface{}{"anidado": map[string]interface{}{"v": "inicial"}}
	if err := s.Upsert("Element", "1", original); err != nil {
		t.Fatal(err)
	}

	original["anidado"].(map[string]interface{})["v"] = "mutado-por-fuera"
	doc, _ := s.Get("Element", "1")
	if doc["anidado"].(map[string]interface{})["v"] != "inicial" {
		t.Error("mutar el mapa original alteró lo almacenado")
	}

	doc["anidado"].(map[string]interface{})["v"] = "mutado-en-la-copia"
	doc2, _ := s.Get("Element", "1")
	if doc2["anidado"].(map[string]interface{})["v"] != "inicial" {
		t.Error("mutar el mapa devuelto alteró lo almacenado")
	}
}

// Un struct Go escrito al store debe volver como map con números float64,
// igual que al pasar por Couchbase. Sin esto el fake escondería los type
// asserts que sí rompen en producción.
func TestMemoryStoreNormalizaComoJSON(t *testing.T) {
	s := NewMemoryStore()
	doc := map[string]interface{}{
		"meta":   Meta{ResourceType: "Element", Created: "2020-01-01T00:00:00Z"},
		"entero": int64(42),
	}
	if err := s.Upsert("Element", "1", doc); err != nil {
		t.Fatal(err)
	}
	out, err := s.Get("Element", "1")
	if err != nil {
		t.Fatal(err)
	}
	meta, ok := out["meta"].(map[string]interface{})
	if !ok {
		t.Fatalf("meta volvió como %T, se esperaba map[string]interface{}", out["meta"])
	}
	if meta["resourceType"] != "Element" {
		t.Errorf("meta = %v", meta)
	}
	if _, ok := out["entero"].(float64); !ok {
		t.Errorf("entero volvió como %T, se esperaba float64", out["entero"])
	}
}

func TestMemoryStoreSearchOrdenYPaginacion(t *testing.T) {
	s := NewMemoryStore()
	for _, name := range []string{"c", "a", "b"} {
		if err := s.Upsert("Element", name, map[string]interface{}{"id": name, "name": name}); err != nil {
			t.Fatal(err)
		}
	}

	total, page, err := s.Search(SearchQuery{Bucket: "Element", SortBy: "name", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 {
		t.Errorf("total = %d", total)
	}
	if page[0]["name"] != "a" || page[2]["name"] != "c" {
		t.Errorf("orden ascendente = %v", page)
	}

	_, page, _ = s.Search(SearchQuery{Bucket: "Element", SortBy: "name", SortDescending: true, Limit: 10})
	if page[0]["name"] != "c" {
		t.Errorf("orden descendente = %v", page)
	}

	total, page, _ = s.Search(SearchQuery{Bucket: "Element", SortBy: "name", Offset: 1, Limit: 1})
	if total != 3 {
		t.Errorf("total debe ignorar la paginación, fue %d", total)
	}
	if len(page) != 1 || page[0]["name"] != "b" {
		t.Errorf("página = %v", page)
	}

	// Un offset más allá del final devuelve una página vacía, no un error.
	total, page, err = s.Search(SearchQuery{Bucket: "Element", Offset: 99, Limit: 10})
	if err != nil || total != 3 || len(page) != 0 {
		t.Errorf("offset fuera de rango = %d, %v, %v", total, page, err)
	}

	// Un bucket desconocido es un conjunto vacío, no un error.
	total, page, err = s.Search(SearchQuery{Bucket: "NoExiste", Limit: 10})
	if err != nil || total != 0 || len(page) != 0 {
		t.Errorf("bucket desconocido = %d, %v, %v", total, page, err)
	}
}

// El fake no evalúa filtros: debe decirlo en voz alta en vez de devolver
// resultados silenciosamente incorrectos.
func TestMemoryStoreRechazaFiltros(t *testing.T) {
	s := NewMemoryStore()
	if _, _, err := s.Search(SearchQuery{Bucket: "Element", Filter: `name eq "x"`}); !errors.Is(err, ErrFilterUnsupported) {
		t.Errorf("Search con filtro = %v, se esperaba ErrFilterUnsupported", err)
	}
}

// Comprobación estática de que ambas implementaciones satisfacen Store.
var (
	_ Store = (*MemoryStore)(nil)
	_ Store = (*CouchbaseStore)(nil)
)
