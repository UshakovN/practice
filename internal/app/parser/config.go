package parser

type Price struct {
	PriceAndAvailability struct {
		PartNumber []struct {
			Price string `json:"price"`
		} `json:""` // variable
	} `json:"priceAndAvailability"`
}
