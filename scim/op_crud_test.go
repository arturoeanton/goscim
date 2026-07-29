package scim

import (
	"encoding/json"
	"net/http"
	"testing"
)

// Estos tests recorren el router SCIM real de punta a punta contra el store en
// memoria. Fijan el comportamiento vigente para que el resto del plan de
// release pueda cambiarlo de forma deliberada y visible: donde el
// comportamiento actual incumple el RFC se marca con TODO(Bn).

func TestCreateElement(t *testing.T) {
	r, store := newTestServer(t)

	body := createElement(t, r, "Element1", 1)

	id, ok := body["id"].(string)
	if !ok || id == "" {
		t.Fatalf("la respuesta no trae id: %v", body)
	}
	if body["name"] != "Element1" {
		t.Errorf("name = %v", body["name"])
	}
	// Create no aplica el filtrado por rol, a diferencia de Read/Search.
	if body["description"] != "descripción de Element1" {
		t.Errorf("description = %v", body["description"])
	}

	meta, ok := body["meta"].(map[string]interface{})
	if !ok {
		t.Fatalf("meta ausente o con forma inesperada: %v", body["meta"])
	}
	if meta["resourceType"] != "Element" {
		t.Errorf("meta.resourceType = %v", meta["resourceType"])
	}
	if meta["created"] == "" || meta["created"] != meta["lastModified"] {
		t.Errorf("created/lastModified inconsistentes: %v", meta)
	}
	if meta["version"] == "" {
		t.Error("meta.version vacío")
	}
	// TODO(B9): location debe ser absoluta e incluir el prefijo /scim/v2.
	if meta["location"] != "/Elements/"+id {
		t.Errorf("meta.location = %v", meta["location"])
	}

	// El documento quedó persistido en el bucket con el nombre del resource
	// type ("Element"), no con el del endpoint ("/Elements").
	stored, err := store.Get("Element", id)
	if err != nil {
		t.Fatalf("no se persistió en el bucket Element: %v", err)
	}
	if stored["name"] != "Element1" {
		t.Errorf("documento almacenado = %v", stored)
	}
}

func TestCreateElementRechazaPayloadsInvalidos(t *testing.T) {
	cases := []struct {
		nombre string
		body   string
	}{
		{
			"sin schemas",
			`{"name":"x","$ref":"/x"}`,
		},
		{
			"falta un atributo requerido del core",
			`{"schemas":["` + schemaCore + `","` + schemaExt + `"],"name":"x","` + schemaExt + `":{"required":1}}`,
		},
		{
			"atributo no declarado en el schema",
			`{"schemas":["` + schemaCore + `","` + schemaExt + `"],"name":"x","$ref":"/x","noExiste":1,"` + schemaExt + `":{"required":1}}`,
		},
		{
			"tipo equivocado en el core",
			`{"schemas":["` + schemaCore + `","` + schemaExt + `"],"name":123,"$ref":"/x","` + schemaExt + `":{"required":1}}`,
		},
		{
			"tipo equivocado en la extensión",
			`{"schemas":["` + schemaCore + `","` + schemaExt + `"],"name":"x","$ref":"/x","` + schemaExt + `":{"required":"no-es-entero"}}`,
		},
		{
			"schema no declarado en el resource type",
			`{"schemas":["` + schemaCore + `","urn:ietf:params:scim:schemas:extension:inventado:2.0:X"],"name":"x","$ref":"/x"}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.nombre, func(t *testing.T) {
			r, _ := newTestServer(t)
			w := do(t, r, http.MethodPost, elementsPath, tc.body)
			requireStatus(t, w, http.StatusBadRequest)
			out := decode(t, w)
			schemas, _ := out["schemas"].([]interface{})
			if len(schemas) != 1 || schemas[0] != "urn:ietf:params:scim:api:messages:2.0:Error" {
				t.Errorf("no devolvió un error SCIM: %v", out)
			}
		})
	}
}

func TestReadElement(t *testing.T) {
	r, _ := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)

	w := do(t, r, http.MethodGet, elementsPath+"/"+id, "")
	requireStatus(t, w, http.StatusOK)
	body := decode(t, w)

	if body["id"] != id {
		t.Errorf("id = %v", body["id"])
	}
	// name declara $reader ["role2","role1"] y los roles hardcodeados incluyen
	// role1, así que se devuelve.
	if body["name"] != "Element1" {
		t.Errorf("name = %v", body["name"])
	}
	// description declara $reader ["role2","role3"]: ningún rol coincide, así
	// que el valor se censura.
	// TODO(B8): debería omitirse la clave, no devolverla vacía.
	if body["description"] != "" {
		t.Errorf("description debía venir censurada, vino %v", body["description"])
	}
	// $ref no declara $reader, así que no se filtra.
	if body["$ref"] != "/Element1" {
		t.Errorf("$ref = %v", body["$ref"])
	}
}

func TestReadElementInexistente(t *testing.T) {
	r, _ := newTestServer(t)
	w := do(t, r, http.MethodGet, elementsPath+"/no-existe", "")
	requireStatus(t, w, http.StatusNotFound)
}

func TestReplaceElement(t *testing.T) {
	r, _ := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)
	metaOriginal := created["meta"].(map[string]interface{})

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(validElement("Element1-modificado", 7)), &payload); err != nil {
		t.Fatal(err)
	}
	payload["meta"] = metaOriginal
	raw, _ := json.Marshal(payload)

	w := do(t, r, http.MethodPut, elementsPath+"/"+id, string(raw))
	requireStatus(t, w, http.StatusOK)
	body := decode(t, w)

	if body["id"] != id {
		t.Errorf("el PUT cambió el id: %v", body["id"])
	}
	if body["name"] != "Element1-modificado" {
		t.Errorf("name = %v", body["name"])
	}
	meta := body["meta"].(map[string]interface{})
	if meta["created"] != metaOriginal["created"] {
		t.Errorf("created no se preservó: %v vs %v", meta["created"], metaOriginal["created"])
	}
	if meta["version"] == metaOriginal["version"] {
		t.Error("version no cambió tras el PUT")
	}
}

func TestReplaceElementInexistente(t *testing.T) {
	r, _ := newTestServer(t)
	created := createElement(t, r, "Element1", 1)

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(validElement("Element1", 1)), &payload); err != nil {
		t.Fatal(err)
	}
	payload["meta"] = created["meta"]
	raw, _ := json.Marshal(payload)

	w := do(t, r, http.MethodPut, elementsPath+"/no-existe", string(raw))
	requireStatus(t, w, http.StatusNotFound)
}

func TestPatchElement(t *testing.T) {
	r, store := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)

	patch := `{"schemas":["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
	           "Operations":[{"op":"replace","path":"description","value":"nueva descripción"}]}`
	w := do(t, r, http.MethodPatch, elementsPath+"/"+id, patch)
	requireStatus(t, w, http.StatusOK)

	stored, err := store.Get("Element", id)
	if err != nil {
		t.Fatal(err)
	}
	if stored["description"] != "nueva descripción" {
		t.Errorf("description almacenada = %v", stored["description"])
	}
	if stored["name"] != "Element1" {
		t.Errorf("el PATCH tocó un atributo que no debía: %v", stored["name"])
	}
}

func TestPatchElementSobreExtension(t *testing.T) {
	r, store := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)

	patch := `{"schemas":["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
	           "Operations":[{"op":"replace","path":"` + schemaExt + `.required","value":42}]}`
	w := do(t, r, http.MethodPatch, elementsPath+"/"+id, patch)
	requireStatus(t, w, http.StatusOK)

	stored, err := store.Get("Element", id)
	if err != nil {
		t.Fatal(err)
	}
	ext := stored[schemaExt].(map[string]interface{})
	if ext["required"] != float64(42) {
		t.Errorf("extensión tras el patch = %v", ext)
	}
}

func TestDeleteElement(t *testing.T) {
	r, _ := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)

	w := do(t, r, http.MethodDelete, elementsPath+"/"+id, "")
	requireStatus(t, w, http.StatusNoContent)
	if w.Body.Len() != 0 {
		t.Errorf("204 con cuerpo: %s", w.Body.String())
	}

	w = do(t, r, http.MethodGet, elementsPath+"/"+id, "")
	requireStatus(t, w, http.StatusNotFound)
}

func TestDeleteElementInexistente(t *testing.T) {
	r, _ := newTestServer(t)
	w := do(t, r, http.MethodDelete, elementsPath+"/no-existe", "")
	requireStatus(t, w, http.StatusNotFound)
}

func TestSearchElements(t *testing.T) {
	r, _ := newTestServer(t)
	createElement(t, r, "AAA", 1)
	createElement(t, r, "BBB", 2)
	createElement(t, r, "CCC", 3)

	w := do(t, r, http.MethodGet, elementsPath+"?sortBy=name", "")
	requireStatus(t, w, http.StatusOK)
	body := decode(t, w)

	schemas := body["schemas"].([]interface{})
	if len(schemas) != 1 || schemas[0] != "urn:ietf:params:scim:api:messages:2.0:ListResponse" {
		t.Errorf("schemas = %v", schemas)
	}
	if body["totalResults"] != float64(3) {
		t.Errorf("totalResults = %v", body["totalResults"])
	}
	if body["startIndex"] != float64(1) {
		t.Errorf("startIndex = %v", body["startIndex"])
	}

	resources := body["Resources"].([]interface{})
	if len(resources) != 3 {
		t.Fatalf("se devolvieron %d recursos", len(resources))
	}
	nombres := make([]string, 0, 3)
	for _, res := range resources {
		nombres = append(nombres, res.(map[string]interface{})["name"].(string))
	}
	if nombres[0] != "AAA" || nombres[1] != "BBB" || nombres[2] != "CCC" {
		t.Errorf("orden ascendente incorrecto: %v", nombres)
	}
}

func TestSearchElementsOrdenDescendente(t *testing.T) {
	r, _ := newTestServer(t)
	createElement(t, r, "AAA", 1)
	createElement(t, r, "BBB", 2)

	w := do(t, r, http.MethodGet, elementsPath+"?sortBy=name&sortOrder=descending", "")
	requireStatus(t, w, http.StatusOK)
	resources := decode(t, w)["Resources"].([]interface{})
	if resources[0].(map[string]interface{})["name"] != "BBB" {
		t.Errorf("orden descendente incorrecto: %v", resources)
	}
}

func TestSearchElementsPaginacion(t *testing.T) {
	r, _ := newTestServer(t)
	createElement(t, r, "AAA", 1)
	createElement(t, r, "BBB", 2)
	createElement(t, r, "CCC", 3)

	w := do(t, r, http.MethodGet, elementsPath+"?sortBy=name&startIndex=2&count=2", "")
	requireStatus(t, w, http.StatusOK)
	body := decode(t, w)

	if body["totalResults"] != float64(3) {
		t.Errorf("totalResults debe contar todo el conjunto, no la página: %v", body["totalResults"])
	}
	if body["startIndex"] != float64(2) {
		t.Errorf("startIndex = %v", body["startIndex"])
	}
	resources := body["Resources"].([]interface{})
	if len(resources) != 2 {
		t.Fatalf("se esperaban 2 recursos, vinieron %d", len(resources))
	}
	if resources[0].(map[string]interface{})["name"] != "BBB" {
		t.Errorf("la página empieza en %v", resources[0])
	}
}

func TestSearchElementsCensuraPorRol(t *testing.T) {
	r, _ := newTestServer(t)
	createElement(t, r, "AAA", 1)

	w := do(t, r, http.MethodGet, elementsPath, "")
	requireStatus(t, w, http.StatusOK)
	resources := decode(t, w)["Resources"].([]interface{})
	primero := resources[0].(map[string]interface{})
	if primero["description"] != "" {
		t.Errorf("description debía venir censurada en el search: %v", primero["description"])
	}
}

func TestSearchElementsParametrosInvalidos(t *testing.T) {
	cases := []string{"?startIndex=abc", "?count=abc"}
	for _, query := range cases {
		t.Run(query, func(t *testing.T) {
			r, _ := newTestServer(t)
			w := do(t, r, http.MethodGet, elementsPath+query, "")
			requireStatus(t, w, http.StatusBadRequest)
		})
	}
}

func TestDiscoveryNoImplementada(t *testing.T) {
	r, _ := newTestServer(t)
	// TODO(B9/mejora 8): los tres endpoints deben servir la configuración real
	// y colgar de /scim/v2.
	for _, path := range []string{"/ServiceProviderConfig", "/ResourceTypes", "/Schemas"} {
		w := do(t, r, http.MethodGet, path, "")
		requireStatus(t, w, http.StatusNotImplemented)
	}
}

func TestRutasRegistradasParaCadaResourceType(t *testing.T) {
	r, _ := newTestServer(t)
	rutas := make(map[string]bool)
	for _, info := range r.Routes() {
		rutas[info.Method+" "+info.Path] = true
	}
	for _, endpoint := range []string{"/Users", "/Groups", "/Elements"} {
		for _, esperada := range []string{
			"POST " + PREFIX + endpoint,
			"GET " + PREFIX + endpoint,
			"GET " + PREFIX + endpoint + "/:id",
			"PUT " + PREFIX + endpoint + "/:id",
			"PATCH " + PREFIX + endpoint + "/:id",
			"DELETE " + PREFIX + endpoint + "/:id",
		} {
			if !rutas[esperada] {
				t.Errorf("ruta no registrada: %s", esperada)
			}
		}
	}
}
