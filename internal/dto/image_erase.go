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
