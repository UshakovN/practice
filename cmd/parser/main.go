package main

import (
	"fmt"
	"time"

	"github.com/UshakovN/practice/internal/app/parser"
)

func main() {
	start := time.Now()
	tags := []string{"IIAM0WMR/invitrogen", "IIAM3NW4/gibco"}
	parser.FisherSciencific(tags[1])
	fmt.Println(time.Since(start))
}
