package common

import (
	"encoding/json"
	"fmt"

	"github.com/UshakovN/practice/internal/app/store"
)

func PrettyPrint(data interface{}) {
	pb, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("%s\n\n", pb)
}

func BatchContains(itemsBatch []*store.ItemData, addItem *store.ItemData) bool {
	for _, batchItem := range itemsBatch {
		if addItem.Article == batchItem.Article {
			return true
		}
	}
	return false
}
