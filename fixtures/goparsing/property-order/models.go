// Package propertyorder tests that generated specs preserve Go struct field order.
//
//	Schemes: http
//	Host: localhost
//	Version: 0.0.1
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
// swagger:meta
package propertyorder

// OrderedItem has fields in non-alphabetical order.
//
// swagger:model OrderedItem
type OrderedItem struct {
	// Zeta field comes first
	Zeta string `json:"zeta"`
	// Alpha field comes second
	Alpha string `json:"alpha"`
	// Mu field comes third
	Mu int64 `json:"mu"`
	// Beta field comes fourth
	Beta string `json:"beta"`
	// Epsilon field comes last
	Epsilon bool `json:"epsilon"`
}

// AnotherModel also has non-alphabetical field order.
//
// swagger:model AnotherModel
type AnotherModel struct {
	Charlie string `json:"charlie"`
	Able    string `json:"able"`
	Baker   int    `json:"baker"`
}

// swagger:route GET /items listItems
//
// # List ordered items
//
// Responses:
//
//	200: itemsResponse

// swagger:response itemsResponse
type itemsResponse struct {
	// in: body
	Body []OrderedItem
}
