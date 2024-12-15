package receipt

// Receipt - model for JSON receipt body
type Receipt struct {
	Id           string        `json:"id"`
	Retailer     string        `json:"retailer"`
	PurchaseDate string        `json:"purchaseDate"`
	PurchaseTime string        `json:"purchaseTime"`
	Items        []ReceiptItem `json:"items"`
	Points       int           `json:"points,string"`
	Total        float64       `json:"total,string"`
}

// ReceiptItem - model for Json receipt items
type ReceiptItem struct {
	ShortDescription string  `json:"shortDescription"`
	Price            float64 `json:"price,string"`
}
