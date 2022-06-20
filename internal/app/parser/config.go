package parser

type Data struct {
	Price        float64
	Label        string
	Article      string
	Description  string
	Manufacturer string
	Available    bool
}

type Price struct {
	PriceAndAvailability struct {
		PartNumber []struct {
			Price string `json:"price"`
		} `json:"A1098801"` // variable
	} `json:"priceAndAvailability"`
}
