package main

import (
	"fmt"
	"go-plate/internal/api"
)

func main() {
	fmt.Println("Hello World")

	server := api.NewAPIServer()

	server.Run()
}
