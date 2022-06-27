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
		_ = &parser.ItemKey{
			Brand:   "Gibco",
			Article: "PSC2014",
		}
		// mock item
		item := &parser.ItemData{
			Brand:   "Gibco",
			Article: "PSC2014",
			Info: parser.Info{
				Label:        "Gibco GM-CSF Recombinant Swine Protein",
				Description:  "None",
				Manufacturer: "Gibco PSC2014",
				Price:        100,
				Available:    true,
				CreatedAt:    time.Now().UTC().String(),
			},
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
		}}

	newParser := parser.NewParser(brandTags[1])
	newParser.FisherSciencific(storeClient)
	fmt.Println(time.Since(timeStart))
}
