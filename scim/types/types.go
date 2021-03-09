package types

import "time"

// Meta is of SCIM
type Meta struct {
	ResourceType string    `json:"resourceType"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"lastModified"`
	Version      string    `json:"version"`
	Location     string    `json:"location"`
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
	Schemas      []string   `json:"schemas"`
	TotalResults int        `json:"totalResults"`
	Resources    []Resource `json:"Resources"`
}
