package main

import "github.com/UshakovN/practice/internal/app/apiserver"

func main() {
	config := apiserver.NewConfig()
	server := apiserver.NewServer(config)
	server.Start()
}
