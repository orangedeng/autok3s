package clientbase

import (
	"github.com/rancher/apiserver/pkg/types"
)

// type SortOrder string

// type Sort struct {
// 	Name    string            `json:"name,omitempty"`
// 	Order   SortOrder         `json:"order,omitempty"`
// 	Reverse string            `json:"reverse,omitempty"`
// 	Links   map[string]string `json:"links,omitempty"`
// }

type ListOpts struct {
	Filters map[string]interface{}
}

type SchemaCollection struct {
	Data []APISchemaWithLinks
}

type APISchemaWithLinks struct {
	ID              string `json:"id,omitempty"`
	types.APISchema `json:",inline"`
	Links           map[string]string
}
