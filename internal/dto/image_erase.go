package dto

import "encoding/json"

type ImageEraseRequest struct {
	ImageURL  string          `json:"imageUrl"`
	Boxes     json.RawMessage `json:"boxes"`
	Polygons  json.RawMessage `json:"polygons"`
	Shapes    json.RawMessage `json:"shapes"`
	Scope     string          `json:"scope"`
	Resource  string          `json:"resource"`
	ProductID uint64          `json:"productId"`
	SkuID     uint64          `json:"skuId"`
}

type ImageEraseFindRequest struct {
	ImageURL string `json:"imageUrl"`
	Query    string `json:"query"`
	Kinds    string `json:"kinds"`
	Shape    string `json:"shape"`
}
