package scim

// Meta is of SCIM
type Meta struct {
	ResourceType string `json:"resourceType"`
	Created      string `json:"created"`
	LastModified string `json:"lastModified"`
	Version      string `json:"version"`
	Location     string `json:"location"`
}

// SchemaExtension is of SCIM
type SchemaExtension struct {
	Schema   string `json:"schema"`
	Required bool   `json:"required"`
}

// ResoruceType is of SCIM
type ResoruceType struct {
	Schemas          []string          `json:"schemas"`
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Endpoint         string            `json:"endpoint"`
	Description      string            `json:"description"`
	Schema           string            `json:"schema"`
	SchemaExtensions []SchemaExtension `json:"schemaExtensions"`
	Meta             Meta              `json:"meta"`
}

// Attribute is of SCIM
type Attribute struct {
	Name          string      `json:"name"`
	Type          string      `json:"type"`
	MultiValued   bool        `json:"multiValued"`
	Description   string      `json:"description"`
	Required      bool        `json:"required"`
	CaseExact     bool        `json:"caseExact"`
	Mutability    string      `json:"mutability"`
	Returned      string      `json:"returned"`
	Uniqueness    string      `json:"uniqueness"`
	Writer        *[]string   `json:"$writer"`
	Read          *[]string   `json:"$reader"`
	SubAttributes []Attribute `json:"subAttributes"`
}

// Schema is of SCIM
type Schema struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Attributes  []Attribute `json:"attributes"`
	Meta        Meta        `json:"meta"`
}

// Resource is of SCIM
type Resource map[string]interface{}

// ListResponse is of SCIM
type ListResponse struct {
	Schemas      []string      `json:"schemas"`
	TotalResults int           `json:"totalResults,omitempty"`
	ItemsPerPage int           `json:"itemsPerPage,omitempty"`
	StartIndex   int           `json:"startIndex,omitempty"`
	Resources    []interface{} `json:"Resources,omitempty"`
}

// Error is ..
type Error struct {
	Schemas []string `json:"schemas"`
	Detail  string   `json:"detail"`
	Status  string   `json:"status"`
}

// Operation ...
type Operation struct {
	Op    string      `json:"op"` // "add", "remove", or "replace"
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// Patch is
// 3.5.2.  Modifying with PATCH
// HTTP PATCH is an OPTIONAL server function that enables clients to
// update one or more attributes of a SCIM resource using a sequence of
// operations to "add", "remove", or "replace" values.  Clients may
// discover service provider support for PATCH by querying the service
// provider configuration (see Section 4).
type Patch struct {
	Schemas    []string    `json:"schemas"`
	Operations []Operation `json:"Operations"`
}
