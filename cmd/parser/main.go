package main

import (
	"fmt"
	"time"

	"github.com/UshakovN/practice/internal/app/parser"
)

func main() {
	/*
		storeCfg := store.NewConfig()
		storeClient := store.NewClient(storeCfg)
		storeClient.PrintDataTables(storeCfg)
	*/
	timeStart := time.Now()
	brandTags := []string{"IIAM0WMR/invitrogen", "IIAM3NW4/gibco"}
	newParser := parser.NewParser(brandTags[1])
	newParser.FisherSciencific()
	fmt.Println(time.Since(timeStart))
}
