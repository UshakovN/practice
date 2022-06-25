package parser

import "time"

type ItemData struct {
	Brand   string `json:"brand"`
	Article string `json:"id"`
	Info    Info   `json:"info"`
}

type Info struct {
	Label        string    `json:"label"`
	Description  string    `json:"descpition"`
	Manufacturer string    `json:"manufacturer"`
	Price        float64   `json:"price"`
	Available    bool      `json:"available"`
	CreatedAt    time.Time `json:"created"`
}

type Price struct {
	PriceAndAvailability struct {
		PartNumber []struct {
			Price string `json:"price"`
		} `json:"A1098801"` // variable
	} `json:"priceAndAvailability"`
}
