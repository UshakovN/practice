package main

import (
	"fmt"
	"time"

	"github.com/UshakovN/practice/internal/app/parser"
	"github.com/UshakovN/practice/internal/app/store"
)

func main() {
	/*
		// mock key
		_ = &store.ItemKey{
			Brand:   "Gibco",
			Article: "PSC2014",
		}
		// mock item
		_ = &store.ItemData{
			Brand:   "Gibco",
			Article: "PSC2014",

			Label:        "Gibco GM-CSF Recombinant Swine Protein",
			Description:  "None",
			Manufacturer: "Gibco PSC2014",
			Price:        100,
			Available:    true,
			CreatedAt:    time.Now().UTC().String(),
		}
	*/
	storeCfg := store.NewConfig()
	storeClient := store.NewClient(storeCfg)

	timeStart := time.Now()
	brandTags := []parser.Brand{
		{
			Name: "invitrogen",
			Code: "IIAM0WMR",
		},
		{
			Name: "gibco",
			Code: "IIAM3NW4",
		},
		{
			Name: "abbott-laboratories",
			Code: "I9C8LPZ9",
		},
		{
			Name: "buchi",
			Code: "I9C8LSEC",
		},
		{
			Name: "dwk-life-sciences",
			Code: "K8HKQ3DV",
		},
	}
	newParser := parser.NewParser(brandTags[1])
	/*
		if _, err := newParser.GetHtmlDocument("https://example.com/"); err != nil {
			log.Fatal(err)
		}
	*/
	newParser.FisherSciencific(storeClient)
	fmt.Println(time.Since(timeStart))
}
