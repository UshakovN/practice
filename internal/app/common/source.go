package common

import (
	"encoding/json"
	"fmt"
)

func PrettyPrint(data interface{}) {
	pb, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("%s\n\n", pb)
}
